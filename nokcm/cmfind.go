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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"strings"
)

type ParaDef struct {
	index int
	mocCat string
	mocName string
	paraName string
	comments string
}

type CmFinder struct {
	log *zap.Logger
	cmpath string
	paras string
	db map[string]map[string]map[string]string // [k1=moc, v1=[k2=instanceId, v2=[k3=paraName, v3=paraVal]]]
	debug bool
}

func (p *CmFinder) Init(log *zap.Logger, cmpath, paras string, debug bool) {
	p.log = log
	p.cmpath = cmpath
	p.paras = paras
	p.db = make(map[string]map[string]map[string]string)
	p.debug = debug
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing CM Finder..."))
}

func (p *CmFinder) parseDat(dat string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing CM file...[%s]", path.Base(dat)))

	fin, err := os.Open(dat)
	if err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
		return
	}

	reader := bufio.NewReader(fin)
	var moc, id string
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
			if _, e := p.db[moc]; !e {
				p.db[moc] = make(map[string]map[string]string)
			}

			if _, e := p.db[moc][id]; !e {
				p.db[moc][id] = make(map[string]string)
			}
		} else {
			tokens := strings.Split(line, "===")
			p.db[moc][id][tokens[0]] = tokens[1]
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