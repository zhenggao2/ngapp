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
package biptrace

import (
	"bytes"
	"fmt"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BIN_TSHARK    string = "tshark.exe"
	LUA_LUASHARK  string = "luashark.lua"
)

type BipTraceParser struct {
	log          *zap.Logger
	wsharkPath   string
	luasharkPath string
	bipTracePath string
	pattern      string
	debug        bool

	traceFiles []string
}

func (p *BipTraceParser) Init(log *zap.Logger, lua, wshark, trace, pattern string, debug bool) {
	p.log = log
	p.luasharkPath = lua
	p.wsharkPath = wshark
	p.bipTracePath = trace
	p.pattern = pattern
	p.debug = debug

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing BipTrace parser...(working dir: %v)", p.bipTracePath))

	fileInfo, err := ioutil.ReadDir(p.bipTracePath)
	if err != nil {
		p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", p.bipTracePath))
		return
	}

	for _, file := range fileInfo {
		if !file.IsDir() && path.Ext(file.Name()) == p.pattern {
			p.traceFiles = append(p.traceFiles, path.Join(p.bipTracePath, file.Name()))
		}
	}
}

func (p *BipTraceParser) Exec() {
	// recreate dir for parsed ttis
	outPath := path.Join(p.bipTracePath, "parsed_biptrace")
	if err := os.RemoveAll(outPath); err != nil {
		panic(fmt.Sprintf("Fail to remove directory: %v", err))
	}

	if err := os.MkdirAll(outPath, 0775); err != nil {
		panic(fmt.Sprintf("Fail to create directory: %v", err))
	}

	if p.pattern == ".pcap"  {
		fileInfo, err := ioutil.ReadDir(p.bipTracePath)
		if err != nil {
			p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", p.bipTracePath))
			return
		}

		wg := &sync.WaitGroup{}
		for _, file := range fileInfo {
			if !file.IsDir() && path.Ext(file.Name()) == ".pcap" {
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

func (p *BipTraceParser) parse(fn string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing BIP trace using tshark... [%s]", fn))
	outPath := path.Join(p.bipTracePath, "parsed_biptrace")

	mapEventHeader := make(map[string][]string)
	mapEventHeaderOk := make(map[string]bool)
	mapEventRecord := make(map[string]*utils.OrderedMap)

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd := exec.Command(path.Join(p.wsharkPath, BIN_TSHARK), "-r", path.Join(p.bipTracePath, fn), "-X", fmt.Sprintf("lua_script:%s", path.Join(p.luasharkPath, LUA_LUASHARK)), "-P", "-V")
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	p.writeLog(zapcore.DebugLevel, cmd.String())
	if err := cmd.Run(); err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
	}
	if stdOut.Len() > 0 {
		// TODO use bytes.Buffer.readString("\n") to postprocessing text files
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("splitting BIP trace into csv... [%s]", fn))
		icomRec := false
		var ts string
		var event string
		var fields string
		for {
			line, err := stdOut.ReadString('\n')
			if err != nil {
				break
			}

			// remove leading and tailing spaces
			line = strings.TrimSpace(line)

			if len(line) > 0 {
				if strings.Contains(line, "ICOM_5G") {
					icomRec = false

					if len(fields) > 0 {
						mapEventRecord[event].Add(ts, fields)
						mapEventHeaderOk[event] = true
					}

					tokens := strings.Split(line, "/")
					event = strings.Replace(tokens[len(tokens)-1], ",", "_", -1)
					if _, exist := mapEventHeader[event]; !exist {
						mapEventHeader[event] = make([]string, 0)
						mapEventHeaderOk[event] = false
					}
					if _, exist := mapEventRecord[event]; !exist {
						mapEventRecord[event] = utils.NewOrderedMap()
					}
				}

				if strings.Split(line, ":")[0] == "Epoch Time" {
					// Epoch Time: 1621323617.322338000 seconds
					tokens := strings.Split(strings.Split(line, " ")[2], ".")
					sec, _ := strconv.ParseInt(tokens[0], 10, 64)
					nsec, _ := strconv.ParseInt(tokens[1], 10, 64)
					ts = time.Unix(sec, nsec).Format("2006-01-02_15:04:05.999999999")

					if !mapEventHeaderOk[event] {
						mapEventHeader[event] = append(mapEventHeader[event], []string{"eventType", "timestamp"}...)
					}
					fields = fmt.Sprintf("%s,%s", event, ts)
				}

				if line == "ICOM 5G Protocol" {
					icomRec = true
				}

				if icomRec {
					tokens := strings.Split(line, ":")
					if len(tokens) == 2 {
						if !mapEventHeaderOk[event] {
							mapEventHeader[event] = append(mapEventHeader[event], tokens[0])
						}
						fields = fields + "," + strings.Replace(strings.TrimSpace(tokens[1]), ",", "_", -1)
					}
				}
			}
		}

		for k1, v1 := range mapEventHeader {
			outFn := path.Join(outPath, fmt.Sprintf("%s_%s.csv", path.Base(fn), k1))
			fout, err := os.OpenFile(outFn, os.O_WRONLY|os.O_CREATE, 0664)
			if err != nil {
				p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", outFn))
				break
			}

			fout.WriteString(strings.Join(v1, ",") + "\n")
			for p := 0; p < mapEventRecord[k1].Len(); p += 1{
				k2 := mapEventRecord[k1].Keys()[p].(string)
				v2 := mapEventRecord[k1].Val(k2).(string)
				fout.WriteString(v2 + "\n")
			}

			fout.Close()
		}
	}
	if stdErr.Len() > 0 {
		p.writeLog(zapcore.DebugLevel, stdErr.String())
	}
}

func (p *BipTraceParser) writeLog(level zapcore.Level, s string) {
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
