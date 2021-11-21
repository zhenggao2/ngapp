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
	"github.com/xuri/excelize/v2"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"math"
	"os"
	"path/filepath"
	"runtime"
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

var PmAggMax = []string {
	// 5G NSA
	"M55114C00010",
	"M55114C00013",
	"M55114C00036",
	"M55308C02001",
	"M55308C02003",
	"M55308C20001",
	"M55308C20002",
	"M55308C21002",
	"M55308C21004",
	// 5G SA
	"M55138C00014",
	"M55138C00015",
	"M55138C00016",
	"M55138C00019",
	"M55138C00022",
	"M55138C01007",
	"M55138C01010",
	"M55351C00001", // new in 5G22R1
	"M55351C01001", // new in 5G22R1
}

var PmAggMin = []string {
	// 5G SA
	"M55139C00512",
}

func (p *KpiParser) Init(log *zap.Logger, op, db string, maxgo int, debug bool) {
	p.log = log
	p.op = op
	p.db = db
	p.maxgo = utils.MaxInt([]int{2, maxgo})
	p.debug = debug
	p.kpis = utils.NewOrderedMap()

	// For TWM XINOS M55145(NRANS), the aggregation is NRBTS_PLMN
	if p.op == "twm" {
		aggPerMeasType["NRANS"] = "NRBTS_PLMN"
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing KPI parser..."))
}

func (p *KpiParser) ParseKpiDef(kdf string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing KPI definitions...[%s]", filepath.Base(kdf)))

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
				p.kpis.Remove(name)
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

	if p.debug {
		for _, k := range p.kpis.Keys() {
			p.writeLog(zapcore.DebugLevel, fmt.Sprintf("kpiName=%s, kpiDef=%v", k, p.kpis.Val(k)))
		}
	}
}

func (p *KpiParser) LoadPmDb(db, btsid, stime, etime string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Loading PM DB..."))

	var btsList []string
	if btsid != "all" {
		btsList = strings.Split(btsid, ",")
	}

	p.pms = cmap.New()
	for _, k := range p.kpis.Keys() {
		for _, c := range p.kpis.Val(k).(*KpiDef).counters {
			p.pms.SetIfAbsent(c, cmap.New())
		}
	}

	wg := &sync.WaitGroup{}
	for _, c := range p.pms.Keys() {
		// avoid 'too many open files' error of os.Open
		// ulimit -n 1024/2048 or ulimit -a
		for {
			if runtime.NumGoroutine() >= 512 {
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		fn := filepath.Join(db, c + ".gz")
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Loading PM...[%s]", filepath.Base(fn)))

		wg.Add(1)
		go func(c, fn string) {
			defer wg.Done()

			fin, err := os.Open(fn)
			if err != nil {
				p.writeLog(zapcore.DebugLevel, err.Error())
				fmt.Printf("Fail to load PM: %v, please check NRPMRNL.\n", filepath.Base(fn))
				return
			}

			gr, err2 := gzip.NewReader(fin)
			if err2 != nil {
				p.writeLog(zapcore.ErrorLevel, err.Error())
				return
			}

			br := bufio.NewReader(gr)
			tsMap := cmap.New()
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
					// TWM: Timestamp should be "2006-01-02"
					// CMCC: Timestamp should be "startTime.interval"
					ts := strings.Replace(tokens2[len(tokens2)-1], "-", "", -1)

					// check duplication
					_, e := tsMap.Get(tokens[0])
					if !e {
						tsMap.SetIfAbsent(tokens[0], true)
					} else {
						continue
					}

					// check against BTS ID
					if len(btsList) > 0 && !utils.ContainsStr(btsList, bts) {
						continue
					}

					if p.op == "cmcc" {
						startTime := strings.Split(ts, ".")[0]
						ts = startTime[:len(stime)]

						// override tokens[0]
						tokens2[len(tokens2)-1] = ts
						tokens[0] = strings.Join(tokens2, "_")
					}

					if ts < stime || ts > etime {
						continue
					}

					if len(tokens[1]) == 0 {
						m, _ := p.pms.Get(c)
						_, e := m.(cmap.ConcurrentMap).Get(tokens[0])
						if !e {
							m.(cmap.ConcurrentMap).Set(tokens[0], float64(0))
						}
						p.pms.Set(c, m)
					} else {
						v, err := strconv.ParseFloat(tokens[1], 64)
						if err != nil {
							p.writeLog(zapcore.ErrorLevel, err.Error())
							continue
						} else {
							m, _ := p.pms.Get(c)
							v0, e := m.(cmap.ConcurrentMap).Get(tokens[0])
							if !e {
								m.(cmap.ConcurrentMap).Set(tokens[0], v)
							} else {
								// check PM object aggregation method
								if utils.ContainsStr(PmAggMax, c) {
									m.(cmap.ConcurrentMap).Set(tokens[0], math.Max(v0.(float64), v))
								} else if utils.ContainsStr(PmAggMin, c) {
									m.(cmap.ConcurrentMap).Set(tokens[0], math.Min(v0.(float64), v))
								} else {
									m.(cmap.ConcurrentMap).Set(tokens[0], v0.(float64) + v)
								}
							}
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

	if p.debug {
		for _, c := range p.pms.Keys() {
			m, _ := p.pms.Get(c)
			p.writeLog(zapcore.DebugLevel, fmt.Sprintf("counterName=%s, count=%d", c, m.(cmap.ConcurrentMap).Count()))
		}
	}
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

	// key1 = agg, key2 = aggKey, key3 = kpiName, val3 = kpiVal
	report := make(map[string]*utils.OrderedMap)
	// key = agg, val = list of kpiName
	reportHeader := make(map[string][]string)
	reportHeaderWiUnit := make(map[string][]string)
	timestamp := time.Now().Format("20060102_150405")

	for _, kpi := range p.kpis.Keys() {
		headerWritten := false
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

			agg := p.kpis.Val(kpi).(*KpiDef).agg
			precision := p.kpis.Val(kpi).(*KpiDef).precision
			unit := p.kpis.Val(kpi).(*KpiDef).unit
			keyPat := keyPerAgg[agg]
			if _, exist := report[agg]; !exist {
				report[agg] = utils.NewOrderedMap()
				reportHeader[agg] = []string{strings.Replace(keyPat, "_", ",", -1)}
				reportHeaderWiUnit[agg] = []string{strings.Replace(keyPat, "_", ",", -1)}
			}
			if !report[agg].Exist(key) {
				report[agg].Add(key, utils.NewOrderedMap())
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
				if p.debug {
					p.writeLog(zapcore.DebugLevel, fmt.Sprintf("kpiName=%v, kpiAgg=%v, aggKey=%v, paras=%v, ret=%.*f", kpi, agg, key, paras, precision, ret))
				}
				report[agg].Val(key).(*utils.OrderedMap).Add(kpi, strconv.FormatFloat(ret.(float64), 'f', precision, 64))
				if !headerWritten {
					reportHeader[agg] = append(reportHeader[agg], kpi.(string))
					reportHeaderWiUnit[agg] = append(reportHeaderWiUnit[agg], fmt.Sprintf("%s[%s]", kpi.(string), unit))
					headerWritten = true
				}
			}
		}
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Generating KPI report..."))
	/*
	for agg := range report {
		ofn := filepath.Join(rptPath, fmt.Sprintf("kpi_report_%s_%s.csv", agg, timestamp))
		fout, err := os.OpenFile(ofn, os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", ofn))
			return
		}

		fout.WriteString(strings.Join(reportHeaderWiUnit[agg], ",") + "\n")
		for _, aggKey := range report[agg].Keys() {
			v := report[agg].Val(aggKey).(*utils.OrderedMap)
			line := []string{strings.Replace(aggKey.(string), "_", ",", -1)}
			for i := 1; i < len(reportHeader[agg]); i += 1 {
				if v.Exist(reportHeader[agg][i]) {
					line = append(line, v.Val(reportHeader[agg][i]).(string))
				} else {
					line = append(line, "-")
				}
			}
			fout.WriteString(strings.Join(line, ",") + "\n")
		}

		fout.Close()
	}
	 */

	wb := excelize.NewFile()
	for agg := range report {
		if wb.GetSheetName(wb.GetActiveSheetIndex()) == "Sheet1" {
			wb.SetSheetName("Sheet1", agg)
		} else {
			wb.NewSheet(agg)
		}

		// write header
		row := 1
		reportHeaderWiUnit[agg] = append(strings.Split(reportHeaderWiUnit[agg][0], ","), reportHeaderWiUnit[agg][1:]...)
		for i, h := range reportHeaderWiUnit[agg] {
			wb.SetCellValue(agg, fmt.Sprintf("%v%v", p.int2Col(i+1), row), h)
		}

		for _, aggKey := range report[agg].Keys() {
			row++
			tokens := strings.Split(aggKey.(string), "_")
			for i, k := range tokens {
				wb.SetCellValue(agg, fmt.Sprintf("%v%v", p.int2Col(i+1), row), k)
			}

			v := report[agg].Val(aggKey).(*utils.OrderedMap)
			for i := 1; i < len(reportHeader[agg]); i += 1 {
				if v.Exist(reportHeader[agg][i]) {
					fv, err := strconv.ParseFloat(v.Val(reportHeader[agg][i]).(string), 64)
					if err != nil {
						p.writeLog(zapcore.WarnLevel, fmt.Sprintf("strconv.ParseFloat failed (v = %v, error=%v)", v, err.Error()))
						wb.SetCellValue(agg, fmt.Sprintf("%v%v", p.int2Col(len(tokens)+i), row), "")
					} else {
						if math.IsNaN(fv) {
							wb.SetCellValue(agg, fmt.Sprintf("%v%v", p.int2Col(len(tokens)+i), row), "NA")
						} else {
							wb.SetCellValue(agg, fmt.Sprintf("%v%v", p.int2Col(len(tokens)+i), row), fv)
						}
					}
				} else {
					wb.SetCellValue(agg, fmt.Sprintf("%v%v", p.int2Col(len(tokens)+i), row), "-")
				}
			}
		}

		//wb.SetPanes(agg, `{"freeze":true,"split":false,"x_split":1,"y_split":1}`)
		wb.AutoFilter(agg, "A1", fmt.Sprintf("%v%v", p.int2Col(len(reportHeaderWiUnit[agg])), row), "")
	}

	if err := wb.SaveAs(filepath.Join(rptPath, fmt.Sprintf("kpi_report_%s.xlsx", timestamp))); err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
		return
	}
}

func (p *KpiParser) int2Col(i int) string {
	var s string
	azm := make(map[int]byte)
	for i := 0; i < 26; i++ {
		azm[i] = byte('A' + i)
	}

	for {
		if i > 26 {
			rem := (i - 1) % 26
			s = string(azm[rem]) + s
			i = (i - rem) / 26
		} else {
			rem := (i - 1) % 26
			s = string(azm[rem]) + s
			break
		}
	}

	return s
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

