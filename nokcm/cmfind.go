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
	cmap "github.com/orcaman/concurrent-map"
	"github.com/unidoc/unioffice/spreadsheet"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type CmFinder struct {
	log *zap.Logger
	cmpath []string
	paras string
	maxgo int
	mocDb cmap.ConcurrentMap // key=MOC_CAT.MOC_NAME, val=MOC full name
	paraDb map[string][]string // key=MOC_CAT.MOC_NAME, val=list of parameters
	db cmap.ConcurrentMap // [key1=MOC_CAT.MOC_NAME, val1=[key2=dn, val2=[key3=paraName, val3=paraVal]]]
	debug bool
}

func (p *CmFinder) Init(log *zap.Logger, cmpath, paras string, debug bool) {
	p.log = log
	p.cmpath = strings.Split(cmpath, ",")
	p.paras = paras
	p.mocDb = cmap.New()
	p.paraDb = make(map[string][]string)
	p.db = cmap.New()
	p.debug = debug

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing CM Finder..."))
}

func (p *CmFinder) Search() {
	p.LoadParas()
	// p.writeLog(zapcore.DebugLevel, fmt.Sprintf("%v", p.mocDb))
	// p.writeLog(zapcore.DebugLevel, fmt.Sprintf("%v", p.paraDb))

	for _, sname := range p.mocDb.Keys() {
		p.db.Set(sname, cmap.New())
	}

	for _, cmp := range p.cmpath {
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Searching CM files...[path=%v]", cmp))
		fileInfo, err := ioutil.ReadDir(cmp)
		if err != nil {
			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Fail to read directory: %s.", cmp))
			return
		}

		wg := &sync.WaitGroup{}
		for _, file := range fileInfo {
			if !file.IsDir() {
				wg.Add(1)
				go func(fn, ts string) {
					defer wg.Done()
					p.parseDat(fn, ts)
				}(filepath.Join(cmp, file.Name()), filepath.Base(cmp))
			}
		}
		wg.Wait()
	}

	/*
	for _, sname := range p.db.Keys() {
		m1, _ := p.db.Get(sname)
		for _, dn := range m1.(cmap.ConcurrentMap).Keys() {
			m2, _ := m1.(cmap.ConcurrentMap).Get(dn)
			for _, pn := range m2.(cmap.ConcurrentMap).Keys() {
				pv, _ := m2.(cmap.ConcurrentMap).Get(pn)
				p.writeLog(zapcore.DebugLevel, fmt.Sprintf("sname=%v,dn=%v,pn=%v,pv=%v", sname, dn, pn, pv))
			}
		}
	}
	 */

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Exporting results to excel..."))
	workbook := spreadsheet.New()
	wrapped := workbook.StyleSheet.AddCellStyle()
	wrapped.SetWrapped(true)
	for _, sname := range p.db.Keys() {
		sheet := workbook.AddSheet()
		sheet.SetName(sname)
		// write header
		row := sheet.AddRow()
		mocName, _ := p.mocDb.Get(sname)
		header := append([]string{fmt.Sprintf("DN(%v)", mocName.(string)), "TS"}, p.paraDb[sname]...)
		for _, h := range header {
			cell := row.AddCell()
			cell.SetString(h)
			cell.SetStyle(wrapped)
		}

		m1, _ := p.db.Get(sname)
		for _, dn := range m1.(cmap.ConcurrentMap).Keys() {
			row := sheet.AddRow()

			tokens := strings.Split(dn, ",")
			tmp := strings.Split(tokens[1], "/")
			idList := make([]string, 0)
			for _, t := range tmp {
				pair := strings.Split(t, "-")
				idList = append(idList, pair[1])
			}
			rowData := []string{strings.Join(idList, "_"), tokens[0]}

			m2, _ := m1.(cmap.ConcurrentMap).Get(dn)
			for _, pn := range p.paraDb[sname] {
				pv, ok := m2.(cmap.ConcurrentMap).Get(pn)
				if ok {
					rowData = append(rowData, pv.(string))
				} else {
					rowData = append(rowData, "-")
				}
			}

			for _, d := range rowData{
				row.AddCell().SetString(d)
			}
		}

		sheet.SetFrozen(true, true)
		sheet.SetAutoFilter(fmt.Sprintf("A1:%s%d", p.int2Col(sheet.MaxColumnIdx()+1), len(sheet.Rows())))
	}

	workbook.SaveToFile(filepath.Join(filepath.Dir(p.paras), fmt.Sprintf("cm_find_result_%s.xlsx", time.Now().Format("20060102_150405"))))
	workbook.Close()
}

func (p *CmFinder) int2Col(i uint32) string {
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

func (p *CmFinder) LoadParas() {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Loading parameter list...[%s]", filepath.Base(p.paras)))

	fin, err := os.Open(p.paras)
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

		tokens := strings.Split(line, ":")
		// nrbts:NRCELL-msg1FrequencyStart:PRACH Frequency start
		if len(tokens) == 3 {
			names := strings.Split(tokens[1], "-")
			mocDn := fmt.Sprintf("%s.%s", tokens[0], names[0])
			if _, e := p.mocDb.Get(mocDn); !e {
				p.mocDb.Set(mocDn, "")
				p.paraDb[mocDn] = make([]string, 0)
			}
			p.paraDb[mocDn] = append(p.paraDb[mocDn], names[1])
		}
	}

	fin.Close()
}

func (p *CmFinder) parseDat(dat, ts string) {
	// p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing CM file...[%s]", filepath.Base(dat)))

	fin, err := os.Open(dat)
	if err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
		return
	}

	reader := bufio.NewReader(fin)
	var dn, moc, sname string
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
			dn = strings.Split(line, "===")[1]
			tokens := strings.Split(dn, "/")
			mocList := make([]string, 0)
			for _, t := range tokens {
				pair := strings.Split(t, "-")
				mocList = append(mocList, pair[0])
			}

			// update dn to include timestamp
			dn = fmt.Sprintf("%v,%v", ts, dn)

			moc = strings.Join(mocList, ".")
			valid = false
			for _, k := range p.mocDb.Keys() {
				names := strings.Split(k, ".")
				if names[0] == "eqm" && strings.Contains(moc, "EQM_R") {
					continue
				} else {
					if strings.HasPrefix(moc, mocCatMap[names[0]].prefix) && mocList[len(mocList)-1] == names[1] {
						valid = true
						sname = k
						// fmt.Printf("dn=%v, sname=%v, valid=%v\n", dn, sname, valid)
						break
					}
				}
			}

			if valid {
				p.mocDb.Set(sname, strings.Join(mocList, "_"))
				m, _ := p.db.Get(sname)
				m.(cmap.ConcurrentMap).Set(dn, cmap.New())
				p.db.Set(sname, m)
			}
		} else {
			if valid {
				tokens := strings.Split(line, "===")
				if utils.ContainsStr(p.paraDb[sname], tokens[0]) {
					m1, _ := p.db.Get(sname)
					m2, _ := m1.(cmap.ConcurrentMap).Get(dn)
					m2.(cmap.ConcurrentMap).Set(tokens[0], tokens[1])
					m1.(cmap.ConcurrentMap).Set(dn, m2)
					p.db.Set(sname, m1)
				}
			}
		}
	}

	fin.Close()
}

func (p *CmFinder) writeLog(level zapcore.Level, s string) {
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