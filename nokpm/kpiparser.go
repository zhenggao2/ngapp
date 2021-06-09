/*
Copyright © 2020 Zhengwei Gao<zhengwei.gao@yahoo.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package nokpm

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/Knetic/govaluate"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

type KpiParser struct {
	log   *zap.Logger
	op    string
	db    string
	maxgo int
	bts []string
	date []string
	debug bool

	//kpis map[string]*KpiDef // key = KPI name, val = KpiDef
	kpis *utils.OrderedMap
	//pms is equivalent to map[string]map[string]float64 with key1 = counter id, key2 = key as specified in keyPerAgg, key3 = counter value
	pms cmap.ConcurrentMap
}

var NUM_KPI_DEF_FIELDS = 8

type KpiDef struct {
	name string // the KPI_NAME field
	perPlmn bool // the PER_PLMN field
	perSlice bool // the PER_SLICE field
	perRelation bool // the PER_RELATION field
	formula string // the KPI_FORMULA field
	precision int // the KPI_PRECISION field
	unit string // the KPI_UNIT field
	aggMethod string // the KPI_AGG_METHOD field

	measTypes []string
	counters []string
	agg string // used aggregation level
}

func (p *KpiParser) Init(log *zap.Logger, op, db string, maxgo int, debug bool) {
	p.log = log
	p.op = op
	p.db = db
	p.maxgo = utils.MaxInt([]int{2, maxgo})
	p.debug = debug
	//p.kpis = make(map[string]*KpiDef)
	p.kpis = utils.NewOrderedMap()

	// For TWM XINOS M55145(NRANS), the aggregation is NRBTS_PLMN
	if p.op == "twm" {
		aggPerMeasType["NRANS"] = "NRBTS_PLMN"
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing KPI parser..."))
}

func (p *KpiParser) ParseKpiDef(kdf string) {
	if !strings.Contains(kdf, "5g21a") {
		return
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing KPI definitions...[%s]", path.Base(kdf)))

	fin, err := os.Open(kdf)
	if err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
		return
	}

	reader := bufio.NewReader(fin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		// remove leading and tailing spaces
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		tokens := strings.Split(line, ",")
		if len(tokens) != NUM_KPI_DEF_FIELDS {
			p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Invalid KPI definition: incorrect number of fields. (%s)", line))
			continue
		}

		name := tokens[0]
		//if _, exist := p.kpis[name]; !exist {
		if !p.kpis.Exist(name) {
			//p.kpis[name] = &KpiDef{
			p.kpis.Add(name, &KpiDef{
				name: tokens[0],
				perPlmn: p.unsafeParseBool(tokens[1]),
				perSlice: p.unsafeParseBool(tokens[2]),
				perRelation: p.unsafeParseBool(tokens[3]),
				formula: tokens[4],
				precision: p.unsafeAtoi(tokens[5]),
				unit: tokens[6],
				aggMethod: tokens[7],
				measTypes: make([]string, 0),
				counters: make([]string, 0),
			})

			// validate KPI definition
			//expression, err := govaluate.NewEvaluableExpression(p.kpis[name].formula)
			expression, err := govaluate.NewEvaluableExpression(p.kpis.Val(name).(*KpiDef).formula)
			if err != nil {
				p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Invalid KPI definition: fail to parse formula. (kpidef.name=%s, kpiDef.formula=%s)", name, p.kpis.Val(name).(*KpiDef).formula))
				continue
			}
			for _, v := range expression.Vars() {
				measId := strings.ToUpper(strings.Split(v, "C")[0])
				measType := measId2MeasType[measId]
				if !utils.ContainsStr(p.kpis.Val(name).(*KpiDef).measTypes, measType) {
					p.kpis.Val(name).(*KpiDef).measTypes = append(p.kpis.Val(name).(*KpiDef).measTypes, measType)
				}
				p.kpis.Val(name).(*KpiDef).counters = append(p.kpis.Val(name).(*KpiDef).counters, v)
			}

			valid := true
			for _, k := range p.kpis.Val(name).(*KpiDef).measTypes {
				if len(p.kpis.Val(name).(*KpiDef).agg) == 0 {
					p.kpis.Val(name).(*KpiDef).agg = aggPerMeasType[k]
				} else {
					if p.kpis.Val(name).(*KpiDef).agg != aggPerMeasType[k] {
						valid = false
						p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Invalid KPI definition: all counters must have the same aggregation level." +
							" (kpiDef.name=%s, kpidef.agg=%s while measType[%s].agg=%s)", name, p.kpis.Val(name).(*KpiDef).agg, k, aggPerMeasType[k]))
					}
				}
			}
			if !valid {
				p.kpis.Remove(name)
				continue
			}
		} else {
			p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Invalid KPI definition: duplicate KPI name. (kpiDef.name=%s)", tokens[0]))
		}
	}

	fin.Close()

	/*
	for k := range p.kpis {
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("kpiName=%s, kpiDef=%v", k, p.kpis[k]))
	}
	 */
}

func (p *KpiParser) LoadPmDb(db, btsid, stime, etime string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Loading PM DB..."))

	btsList := strings.Split(btsid, ",")

	p.pms = cmap.New()
	for _, k := range p.kpis.Keys() {
		for _, c := range p.kpis.Val(k).(*KpiDef).counters {
			p.pms.SetIfAbsent(c, cmap.New())
		}
	}

	wg := &sync.WaitGroup{}
	for _, c := range p.pms.Keys() {
		/*
		for {
			if runtime.NumGoroutine() >= p.maxgo {
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}
		 */

		fn := path.Join(db, c + ".gz")
		//p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Loading PM...[%s]", path.Base(fn)))

		wg.Add(1)
		go func(c, fn string) {
			defer wg.Done()

			fin, err := os.Open(fn)
			if err != nil {
				p.writeLog(zapcore.ErrorLevel, err.Error())
				return
			}

			gr, err2 := gzip.NewReader(fin)
			if err2 != nil {
				p.writeLog(zapcore.ErrorLevel, err.Error())
				return
			}

			br := bufio.NewReader(gr)
			for {
				line, err := br.ReadString('\n')
				if err != nil {
					break
				}

				// remove leading and tailing spaces
				line = strings.TrimSpace(line)
				if len(line) > 0 {
					tokens := strings.Split(line, ",")
					tokens2 := strings.Split(tokens[0], "_")
					bts := tokens2[0]
					// Timestamp should be "2006-01-02"
					ts := strings.Replace(tokens2[len(tokens2)-1], "-", "", -1)
					if !utils.ContainsStr(btsList, bts) {
						continue
					}
					if ts < stime || ts > etime {
						continue
					}

					if len(tokens[1]) == 0 {
						m, _ := p.pms.Get(c)
						m.(cmap.ConcurrentMap).Set(tokens[0], float64(0))
						p.pms.Set(c, m)
					} else {
						v, err := strconv.ParseFloat(tokens[1], 64)
						if err != nil {
							p.writeLog(zapcore.ErrorLevel, err.Error())
							continue
						} else {
							m, _ := p.pms.Get(c)
							m.(cmap.ConcurrentMap).Set(tokens[0], v)
							p.pms.Set(c, m)
						}
					}
				}
			}

			gr.Close()
			fin.Close()
		}(c, fn)
	}
	wg.Wait()

	/*
	for _, c := range p.pms.Keys() {
		m, _ := p.pms.Get(c)
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("counterName=%s, count=%d", c, m.(cmap.ConcurrentMap).Count()))
	}
	 */
}

func (p *KpiParser) unsafeAtoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func (p *KpiParser) unsafeParseBool(s string) bool {
	v, _ := strconv.ParseBool(s)
	return v
}

func (p *KpiParser) CalcKpi(rptPath string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Calculating KPI..."))

	//rptHeader := []string{"kpiName, aggKey, kpiVal"}
	ofn := path.Join(rptPath, fmt.Sprintf("kpi_report_%s.csv", time.Now().Format("20060102_150406")))
	fout, err := os.OpenFile(ofn, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", ofn))
		return
	}
	fout.WriteString("kpiName, kpiAgg, aggKey, kpiVal\n")

	for _, kpi := range p.kpis.Keys() {
		//p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Calculating KPI...[%s]", kpi))
		expr, _ := govaluate.NewEvaluableExpression(p.kpis.Val(kpi).(*KpiDef).formula)
		paras := make(map[string]interface{})
		m, _ := p.pms.Get(p.kpis.Val(kpi).(*KpiDef).counters[0])
		for _, key := range m.(cmap.ConcurrentMap).Keys() {
			// validate key
			valid := true
			for _, c := range p.kpis.Val(kpi).(*KpiDef).counters {
				m, _ := p.pms.Get(c)
				if !m.(cmap.ConcurrentMap).Has(key) {
					valid = false
					break
				}
			}

			if !valid {
				continue
			}

			for _, c := range p.kpis.Val(kpi).(*KpiDef).counters {
				m, _ := p.pms.Get(c)
				v, _ := m.(cmap.ConcurrentMap).Get(key)
				paras[c] = v.(float64)
			}

			ret, err := expr.Evaluate(paras)
			if err != nil {
				p.writeLog(zapcore.ErrorLevel, err.Error())
			} else {
				//p.writeLog(zapcore.DebugLevel, fmt.Sprintf("kpi_name=%v, key=%v, paras=%v, ret=%.*f", kpi, key, paras, p.kpis[kpi].precision, ret))
				p.writeLog(zapcore.DebugLevel, fmt.Sprintf("kpiName=%v, kpiAgg=%v, aggKey=%v, ret=%.*f", kpi, p.kpis.Val(kpi).(*KpiDef).agg, key, p.kpis.Val(kpi).(*KpiDef).precision, ret))
				fout.WriteString(fmt.Sprintf("%s[%s],%s,%s,%.*f\n", kpi, p.kpis.Val(kpi).(*KpiDef).unit, p.kpis.Val(kpi).(*KpiDef).agg, key, p.kpis.Val(kpi).(*KpiDef).precision, ret))
			}
		}
	}

	fout.Close()
}

func (p *KpiParser) writeLog(level zapcore.Level, s string) {
	switch level {
	case zapcore.DebugLevel:
		p.log.Debug(s)
	case zapcore.InfoLevel:
		p.log.Info(s)
	case zapcore.WarnLevel:
		p.log.Warn(s)
	case zapcore.ErrorLevel:
		p.log.Error(s)
	case zapcore.FatalLevel:
		p.log.Fatal(s)
	case zapcore.PanicLevel:
		p.log.Panic(s)
	default:
	}

	if level != zapcore.DebugLevel {
		fmt.Println(s)
	}
}

