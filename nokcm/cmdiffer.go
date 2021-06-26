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
	"fmt"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
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
	cmpath string
	ins []string
	moc []string // list of MOC catagories to be analyzed
	ignore map[string][]string // key=MOC catagory, val=list of ignored MOCs
	debug bool
}

func (p *CmDiffer) Init(log *zap.Logger, cmpath, ins, moc, ignore string, debug bool) {
	p.log = log
	p.cmpath = cmpath
	p.ins = strings.Split(ins, ",")
	p.moc = strings.Split(moc, ",")

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

	p.debug = debug
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing CM differ..."))
}

func (p *CmDiffer) parseDat() {

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
