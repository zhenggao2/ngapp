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
package ddr4trace

import (
	"fmt"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

const (
	BIN_PYTHON3         string = "python.exe"
	BIN_PYTHON3_LINUX   string = "python3"
	PY3_SNAPSHOT_TOOL string = "SnapshotAnalyzer.py"
)

type Ddr4TraceParser struct {
	log          *zap.Logger
	ddr4TracePath string
	pattern      string
	maxgo int
	debug        bool

	traceFiles []string
}

func (p *Ddr4TraceParser) Init(log *zap.Logger, trace, pattern string, maxgo int, debug bool) {
	p.log = log
	p.ddr4TracePath = trace
	p.pattern = pattern
	p.maxgo = utils.MaxInt([]int{2, maxgo})
	p.debug = debug

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing DDR4 trace parser...(working dir: %v)", p.ddr4TracePath))

	fileInfo, err := ioutil.ReadDir(p.ddr4TracePath)
	if err != nil {
		p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", p.ddr4TracePath))
		return
	}

	for _, file := range fileInfo {
		if !file.IsDir() && path.Ext(file.Name()) == p.pattern {
			p.traceFiles = append(p.traceFiles, path.Join(p.ddr4TracePath, file.Name()))
		}
	}
}

func (p *Ddr4TraceParser) Exec() {
	// recreate dir for parsed ddr4 trace
	outPath := path.Join(p.ddr4TracePath, "parsed_ddr4trace")
	if err := os.RemoveAll(outPath); err != nil {
		panic(fmt.Sprintf("Fail to remove directory: %v", err))
	}

	if err := os.MkdirAll(outPath, 0775); err != nil {
		panic(fmt.Sprintf("Fail to create directory: %v", err))
	}

	if p.pattern == ".bin"  {
		fileInfo, err := ioutil.ReadDir(p.ddr4TracePath)
		if err != nil {
			p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", p.ddr4TracePath))
			return
		}

		wg := &sync.WaitGroup{}
		for _, file := range fileInfo {
			if !file.IsDir() && path.Ext(file.Name())[:len(".bin")] == ".bin" {
				for {
					if runtime.NumGoroutine() >= p.maxgo {
						time.Sleep(1 * time.Second)
					} else {
						break
					}
				}

				wg.Add(1)
				go func(fn string) {
					defer wg.Done()
					p.parse(fn)
				} (file.Name())
			}
		}
		wg.Wait()
	}
}

func (p *Ddr4TraceParser) parse(fn string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing DDR4 trace using Loki snapshot_tool... [%s]", fn))
	// outPath := path.Join(p.ddr4TracePath, "parsed_biptrace")
}

func (p *Ddr4TraceParser) writeLog(level zapcore.Level, s string) {
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
