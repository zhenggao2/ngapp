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
package nokcm

import (
	"bufio"
	"fmt"
	"github.com/unidoc/unioffice/spreadsheet"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type MocCategory struct {
	prefix string
	suffix []string
}

var mocCatMap = map[string]MocCategory {
	"sbts": {prefix: "MRBTS", suffix: []string{"MRBTS", "MRBTSDESC"}},
	"nrbts": {prefix: "MRBTS.NRBTS", suffix: nil},
	"mnl": {prefix: "MRBTS.MNL", suffix: nil},
	"tnl": {prefix: "MRBTS.TNL", suffix: nil},
	"eqm": {prefix: "MRBTS.EQM", suffix: nil},
	"eqmr": {prefix: "MRBTS.EQM_R", suffix: nil},
}

type CmDiffer struct {
	log *zap.Logger
	cmpath []string
	ins []string
	moc []string // list of MOC catagories to be analyzed
	ignore map[string][]string // key=MOC catagory, val=list of ignored MOCs
	db map[string]map[string]*utils.OrderedMap // [k1=moc, v1=[k2=paraName, v2=[k3=instanceId, v3=paraVal]]]
	db2 *utils.OrderedMap
	debug bool
}

func (p *CmDiffer) Init(log *zap.Logger, cmpath, ins, moc, ignore string, debug bool) {
	p.log = log
	p.cmpath = strings.Split(cmpath, ",")
	p.ins = strings.Split(ins, ",")
	p.moc = strings.Split(moc, ",")
	if utils.ContainsStr(p.moc, "all") {
		p.moc = []string{"all"}
	}

	p.ignore = make(map[string][]string)
	tokens := strings.Split(ignore, ",")
	for _, t := range tokens {
		fields := strings.Split(t, ":")
		if len(fields) == 2 {
			catName := fields[0]
			mocName := fields[1]
			if utils.ContainsStr(p.moc, "all") || utils.ContainsStr(p.moc, catName) {
				if _, e := p.ignore[catName]; !e {
					p.ignore[catName] = []string{mocName}
				} else {
					p.ignore[catName] = append(p.ignore[catName], mocName)
				}
			}
		}
	}

	p.db = make(map[string]map[string]*utils.OrderedMap)
	p.db2 = utils.NewOrderedMap()
	p.debug = debug
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing CM differ..."))
}

func (p *CmDiffer) Compare() {
	for _, fn := range p.ins {
		for _, cmp := range p.cmpath {
			dat := path.Join(cmp, fn)
			p.parseDat(dat)
		}
	}

	for _, k := range p.db2.Sort() {
		if p.db2.Val(k).(bool) {
			p.writeLog(zapcore.DebugLevel, fmt.Sprintf("moc=%v, valid=%v", k, p.db2.Val(k)))
		}
	}
	//p.writeLog(zapcore.DebugLevel, fmt.Sprintf("db=%v\n", p.db))

	headerMap := make(map[string][]string)
	for k1 := range p.db {
		headerMap[k1] = make([]string, 0)
		count := 0
		for k2 := range p.db[k1] {
			ids := make([]string, 0)
			for _, k3 := range p.db[k1][k2].Sort() {
				ids = append(ids, k3)
			}
			if len(ids) > count {
				count = len(ids)
				headerMap[k1] = append([]string{k1, "Diff"}, ids...)
			}
		}
	}

	hasDiff := false
	workbook := spreadsheet.New()
	wrapped := workbook.StyleSheet.AddCellStyle()
	wrapped.SetWrapped(true)
	for k1 := range p.db {
		sheet := workbook.AddSheet()
		hasDiff = false

		// write header
		row := sheet.AddRow()
		for _, h := range headerMap[k1] {
			cell := row.AddCell()
			cell.SetString(h)
			cell.SetStyle(wrapped)
		}

		for k2 := range p.db[k1] {
			diff := "NO"
			ids := make([]string, 0)
			vals := make([]string, 0)
			vset := make(map[string]bool)
			for _, k3 := range p.db[k1][k2].Sort() {
				v3 := p.db[k1][k2].Val(k3).(string)
				ids = append(ids, k3)
				vals = append(vals, v3)
				vset[v3] = true
			}

			if len(vset) > 1 {
				diff = "YES"
				hasDiff = true
			}

			// write row
			row := sheet.AddRow()
			rowData := append([]string{k2, diff}, vals...)
			for _, d := range rowData{
				row.AddCell().SetString(d)
			}
		}

		// set sheet name
		sheetName := k1
		if hasDiff {
			sheetName += "#"
		}
		if len(sheetName) > 31 {
			sheet.SetName(sheetName[len(sheetName)-31:])
		} else {
			sheet.SetName(sheetName)
		}

		sheet.SetFrozen(true, true)
		sheet.SetAutoFilter(fmt.Sprintf("A1:%s%d", p.int2Col(sheet.MaxColumnIdx()+1), len(sheet.Rows())))
	}

	workbook.SaveToFile(filepath.Join(filepath.Dir(p.cmpath[0]), fmt.Sprintf("cm_diff_report_%s.xlsx", time.Now().Format("20060102_150406"))))
	workbook.Close()
}

func (p *CmDiffer) parseDat(dat string) {
	//p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing CM file...[%s]", path.Base(dat)))
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing CM file...[%s]", dat))

	fin, err := os.Open(dat)
	if err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
		return
	}

	reader := bufio.NewReader(fin)
	var moc, id string
	var valid bool
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

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			line = line[1:len(line)-1]
			dn := strings.Split(line, "===")[1]
			tokens := strings.Split(dn, "/")
			mocList := make([]string, 0)
			idList := make([]string, 0)
			for _, t := range tokens {
				pair := strings.Split(t, "-")
				mocList = append(mocList, pair[0])
				idList = append(idList, pair[1])
			}

			moc = strings.Join(mocList, ".")
			id = strings.Join(idList, ".")

			// check against p.moc
			valid = false
			for _, m := range p.moc {
				if m == "all" {
					valid = true
					break
				}

				if strings.HasPrefix(moc, mocCatMap[m].prefix) && (mocCatMap[m].suffix == nil || (mocCatMap[m].suffix != nil && utils.ContainsStr(mocCatMap[m].suffix, mocList[len(mocList)-1]))) {
					if m == "eqm" && strings.Contains(moc, "EQM_R") {
						valid = false
					} else {
						valid = true
					}
				}
			}

			// check against p.ignore
			for k := range p.ignore {
				if strings.HasPrefix(moc, mocCatMap[k].prefix) {
					for _, m := range p.ignore[k] {
						if utils.ContainsStr(mocList, m) {
							valid = false
						}
					}
				}
			}

			p.db2.Add(moc, valid)
			if valid {
				if _, e := p.db[moc]; !e {
					p.db[moc] = make(map[string]*utils.OrderedMap)
				}
			}
		} else {
			if valid {
				tokens := strings.Split(line, "===")
				if _, e := p.db[moc][tokens[0]]; !e {
					p.db[moc][tokens[0]] = utils.NewOrderedMap()
				}
				p.db[moc][tokens[0]].Add(fmt.Sprintf("%v-%v", id, dat), tokens[1])
			}
		}
	}

	fin.Close()
}

func (p *CmDiffer) int2Col(i uint32) string {
	var s string
	for {
		if i / 26 > 0 {
			s = fmt.Sprintf("%s%s", string('A' + i % 26 - 1), s)
			i = (i - i % 26) / 26
		} else {
			s = fmt.Sprintf("%s%s", string('A' + i % 26 - 1), s)
			break
		}
	}

	return s
}

func (p *CmDiffer) writeLog(level zapcore.Level, s string) {
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
