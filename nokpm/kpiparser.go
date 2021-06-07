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
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"strings"
)

type KpiParser struct {
	log   *zap.Logger
	op    string
	db    string
	debug bool
}

func (p *KpiParser) Init(log *zap.Logger, op, db string, debug bool) {
	p.log = log
	p.op = op
	p.db = db
	p.debug = debug

	// For TWM XINOS M55145(NRANS), the aggregation is NRBTS_PLMN
	if p.op == "twm" {
		aggPerMeasType["NRANS"] = "NRBTS_PLMN"
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing KPI parser..."))
}

func (p *KpiParser) Parse(kpi string) {
	if !strings.Contains(kpi, "5g21a") {
		return
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing KPI definitions...[%s]", path.Base(kpi)))

	fin, err := os.Open(kpi)
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
		if len(tokens) != 8 {
			p.writeLog(zapcore.DebugLevel, line)
		}
	}

	fin.Close()
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

