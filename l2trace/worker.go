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
package l2trace

import (
	"bytes"
	"fmt"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BIN_PYTHON2         string = "python.exe"
	BIN_PYTHON2_LINUX   string = "python2"
	PY2_TLDA            string = "tlda.py"
	LUA_LUASHARK        string = "luashark.lua"
	BIN_TSHARK          string = "tshark.exe"
	BIN_TSHARK_LINUX    string = "tshark"
	BIN_TEXT2PCAP       string = "text2pcap.exe"
	BIN_TEXT2PCAP_LINUX string = "text2pcap"
)

type L2TraceParser struct {
	log          *zap.Logger
	py2Path      string
	tldaPath     string
	luasharkPath string
	wsharkPath   string
	l2TracePath  string
	pattern      string
	maxgo int
	debug        bool
}

func (p *L2TraceParser) Init(log *zap.Logger, py2, tlda, lua, wshark, trace, pattern string, maxgo int, debug bool) {
	p.log = log
	p.py2Path = py2
	p.tldaPath = tlda
	p.luasharkPath = lua
	p.wsharkPath = wshark
	p.l2TracePath = trace
	p.pattern = pattern
	p.maxgo = utils.MaxInt([]int{2, maxgo})
	p.debug = debug

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing L2 trace parser...(working dir: %v)", trace))
}

func (p *L2TraceParser) Exec() {
	// recreate dir for parsed l2 trace
	outPath := filepath.Join(p.l2TracePath, "parsed_l2trace")
	if err := os.RemoveAll(outPath); err != nil {
		panic(fmt.Sprintf("Fail to remove directory: %v", err))
	}

	if err := os.MkdirAll(outPath, 0775); err != nil {
		panic(fmt.Sprintf("Fail to create directory: %v", err))
	}

	if p.pattern == ".pcap"  {
		fileInfo, err := ioutil.ReadDir(p.l2TracePath)
		if err != nil {
			p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", p.l2TracePath))
			return
		}

		wg := &sync.WaitGroup{}
		for _, file := range fileInfo {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".pcap" {
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
	} else if p.pattern == ".dat" {
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing L2Trace using TLDA..."))
		var stdOut bytes.Buffer
		var stdErr bytes.Buffer
		var cmd *exec.Cmd
		if runtime.GOOS == "linux" {
			cmd = exec.Command(filepath.Join(p.py2Path, BIN_PYTHON2_LINUX), filepath.Join(p.tldaPath, PY2_TLDA), "--techlog_path", p.l2TracePath, "--only", "decode", "--components", "UPLANE", "-o", outPath)
		} else if runtime.GOOS == "windows" {
			cmd = exec.Command(filepath.Join(p.py2Path, BIN_PYTHON2), filepath.Join(p.tldaPath, PY2_TLDA), "--techlog_path", p.l2TracePath, "--only", "decode", "--components", "UPLANE", "-o", outPath)
		} else {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Unsupported OS: runtime.GOOS=%s", runtime.GOOS))
			return
		}
		cmd.Stdout = &stdOut
		cmd.Stderr = &stdErr
		p.writeLog(zapcore.DebugLevel, cmd.String())
		if err := cmd.Run(); err != nil {
			p.writeLog(zapcore.ErrorLevel, err.Error())
		}
		if stdOut.Len() > 0 {
			p.writeLog(zapcore.DebugLevel, stdOut.String())
		}
		if stdErr.Len() > 0 {
			p.writeLog(zapcore.DebugLevel, stdErr.String())
		}

		decodedDir := filepath.Join(outPath, "decoded", "UPLANE_snapshot_decoder", "decode")
		fileInfo, err := ioutil.ReadDir(decodedDir)
		if err != nil {
			p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", decodedDir))
			return
		}

		wg := &sync.WaitGroup{}
		for _, file := range fileInfo {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".pcap" {
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

		// remove decoded folder
		decodedDir = filepath.Join(outPath, "decoded")
		if err := os.RemoveAll(decodedDir); err != nil {
			panic(fmt.Sprintf("Fail to remove directory: %v", err))
		}
	}
}

func (p *L2TraceParser) parse(fn string) {
	if p.pattern == ".pcap" {
		p.parseWithTshark(fn)
	} else if p.pattern == ".dat" {
		p.convert2Pcap(fn)
		p.parseWithTshark(fn)
	}
}

func (p *L2TraceParser) convert2Pcap(fn string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("converting L2Trace using text2pcap... [%s]", fn))
	outPath := filepath.Join(p.l2TracePath, "parsed_l2trace")
	decodedDir := filepath.Join(outPath, "decoded", "UPLANE_snapshot_decoder", "decode")

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command(filepath.Join(p.wsharkPath, BIN_TEXT2PCAP_LINUX), "-u", "5678,0", filepath.Join(decodedDir, fn), filepath.Join(outPath, fn))
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command(filepath.Join(p.wsharkPath, BIN_TEXT2PCAP), "-u", "5678,0", filepath.Join(decodedDir, fn), filepath.Join(outPath, fn))
	} else {
		p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Unsupported OS: runtime.GOOS=%s", runtime.GOOS))
		return
	}
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	p.writeLog(zapcore.DebugLevel, cmd.String())
	if err := cmd.Run(); err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
	}
	if stdOut.Len() > 0 {
		// p.writeLog(zapcore.DebugLevel, stdOut.String())
	}
	if stdErr.Len() > 0 {
		// p.writeLog(zapcore.DebugLevel, stdErr.String())
	}
}

func (p *L2TraceParser) parseWithTshark(fn string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing L2Trace using tshark... [%s]", fn))
	outPath := filepath.Join(p.l2TracePath, "parsed_l2trace")

	mapEventHeader := make(map[string][]string)
	mapEventHeaderOk := make(map[string]bool)
	mapEventRecord := make(map[string]*utils.OrderedMap)

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command(filepath.Join(p.wsharkPath, BIN_TSHARK_LINUX), "-r", filepath.Join(outPath, fn), "-X", fmt.Sprintf("lua_script:%s", filepath.Join(p.luasharkPath, LUA_LUASHARK)), "-P", "-V")
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command(filepath.Join(p.wsharkPath, BIN_TSHARK), "-r", filepath.Join(outPath, fn), "-X", fmt.Sprintf("lua_script:%s", filepath.Join(p.luasharkPath, LUA_LUASHARK)), "-P", "-V")
	} else {
		p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Unsupported OS: runtime.GOOS=%s", runtime.GOOS))
		return
	}
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	p.writeLog(zapcore.DebugLevel, cmd.String())
	if err := cmd.Run(); err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
	}
	if stdOut.Len() > 0 {
		// TODO use bytes.Buffer.readString("\n") to postprocessing text files
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("splitting L2Trace into csv... [%s]", fn))
		icomRec := false
		pduDump := false
		payload := false
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
					pduDump = false
					payload = false

					if len(fields) > 0 {
						mapEventRecord[event].Add(ts, fields)
						mapEventHeaderOk[event] = true
					}
				}

				if strings.Split(line, ":")[0] == "Epoch Time" {
					// Epoch Time: 1621323617.322338000 seconds
					tokens := strings.Split(strings.Split(line, " ")[2], ".")
					sec, _ := strconv.ParseInt(tokens[0], 10, 64)
					nsec, _ := strconv.ParseInt(tokens[1], 10, 64)
					ts = time.Unix(sec, nsec).Format("2006-01-02_15:04:05.999999999")
				}

				if line == "ICOM 5G Protocol" {
					icomRec = true
				}

				if line == "pduDump-Payload" {
					pduDump = true
				}

				if line == "Event-Payload" {
					payload = true

					line, err := stdOut.ReadString('\n')
					if err != nil {
						break
					}
					line = strings.TrimSpace(line)
					event = strings.Split(line, " ")[0]
					if _, exist := mapEventHeader[event]; !exist {
						mapEventHeader[event] = make([]string, 0)
						mapEventHeader[event] = append(mapEventHeader[event], []string{"eventType", "timestamp"}...)
						mapEventHeaderOk[event] = false
					}
					if _, exist := mapEventRecord[event]; !exist {
						mapEventRecord[event] = utils.NewOrderedMap()
					}
					fields = fmt.Sprintf("%s,%s", event, ts)
				}

				if icomRec && !pduDump && payload {
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
			outFn := filepath.Join(outPath, fmt.Sprintf("%s_%s.csv", strings.Replace(filepath.Base(fn), "uplane_ttitrace_decoder_", "", -1), k1))
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

	// remove intermediate pcap
	os.Remove(filepath.Join(outPath, fn))
}

func (p *L2TraceParser) writeLog(level zapcore.Level, s string) {
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