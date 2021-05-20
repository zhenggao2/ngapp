package l2trace

import (
	"bytes"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"fmt"
	"os"
	"path"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	BIN_PYTHON    string = "python.exe"
	PY_TLDA       string = "tlda.py"
	LUA_LUASHARK  string = "luashark.lua"
	BIN_TSHARK    string = "tshark.exe"
	BIN_TEXT2PCAP string = "text2pcap.exe"
)

type L2TraceParser struct {
	log          *zap.Logger
	py2Path      string
	tldaPath     string
	luasharkPath string
	wsharkPath   string
	l2TracePath  string
	pattern      string
	debug        bool

	traceFiles []string
}

func (p *L2TraceParser) Init(log *zap.Logger, py2, tlda, lua, wshark, trace, pattern string, debug bool) {
	p.log = log
	p.py2Path = py2
	p.tldaPath = tlda
	p.luasharkPath = lua
	p.wsharkPath = wshark
	p.l2TracePath = trace
	p.pattern = pattern
	p.debug = debug

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing L2Trace parser...(working dir: %v)", p.l2TracePath))

	fileInfo, err := ioutil.ReadDir(p.l2TracePath)
	if err != nil {
		p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", p.l2TracePath))
		return
	}

	for _, file := range fileInfo {
		if !file.IsDir() && path.Ext(file.Name()) == p.pattern {
			p.traceFiles = append(p.traceFiles, path.Join(p.l2TracePath, file.Name()))
		}
	}
}

func (p *L2TraceParser) Exec() {
	// recreate dir for parsed ttis
	outPath := path.Join(p.l2TracePath, "parsed_l2trace")
	if err := os.RemoveAll(outPath); err != nil {
		panic(fmt.Sprintf("Fail to remove directory: %v", err))
	}

	if err := os.MkdirAll(outPath, 0775); err != nil {
		panic(fmt.Sprintf("Fail to create directory: %v", err))
	}

	if p.pattern == ".pcap"  {
		var cmd *exec.Cmd
		var err error
		var stdOut bytes.Buffer
		var stdErr bytes.Buffer

		fileInfo, err := ioutil.ReadDir(p.l2TracePath)
		if err != nil {
			p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", p.l2TracePath))
			return
		}

		for _, file := range fileInfo {
			if !file.IsDir() && path.Ext(file.Name()) == ".pcap" {
				p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing L2Trace using tshark... [%s]", file.Name()))

				mapEventHeader := make(map[string][]string)
				mapEventHeaderOk := make(map[string]bool)
				mapEventRecord := make(map[string]*utils.OrderedMap)

				cmd = exec.Command(path.Join(p.wsharkPath, BIN_TSHARK), "-r", path.Join(p.l2TracePath, file.Name()), "-X", fmt.Sprintf("lua_script:%s", path.Join(p.luasharkPath, LUA_LUASHARK)), "-P", "-V")
				cmd.Stdout = &stdOut
				cmd.Stderr = &stdErr
				p.writeLog(zapcore.DebugLevel, cmd.String())
				err = cmd.Run()
				if err != nil {
					p.writeLog(zapcore.ErrorLevel, err.Error())
				}

				if stdOut.Len() > 0 {
					// TODO use bytes.Buffer.readString("\n") to postprocessing text files
					p.writeLog(zapcore.InfoLevel, fmt.Sprintf("splitting L2Trace into csv..."))
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
						outFn := path.Join(outPath, fmt.Sprintf("%s_%s.csv", path.Base(file.Name()), k1))
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
				stdOut.Reset()
				stdErr.Reset()
			}
		}
	} else if p.pattern == ".dat" {
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing L2Trace using TLDA..."))

		var cmd *exec.Cmd
		var err error
		var stdOut bytes.Buffer
		var stdErr bytes.Buffer
		cmd = exec.Command(path.Join(p.py2Path, BIN_PYTHON), path.Join(p.tldaPath, PY_TLDA), "--techlog_path", p.l2TracePath, "--only", "decode", "--components", "UPLANE", "-o", outPath)
		cmd.Stdout = &stdOut
		cmd.Stderr = &stdErr
		p.writeLog(zapcore.DebugLevel, cmd.String())
		err = cmd.Run()
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, err.Error())
		}
		if stdOut.Len() > 0 {
			p.writeLog(zapcore.DebugLevel, stdOut.String())
		}
		if stdErr.Len() > 0 {
			p.writeLog(zapcore.DebugLevel, stdErr.String())
		}
		stdOut.Reset()
		stdErr.Reset()

		decodedDir := path.Join(outPath, "decoded", "UPLANE_snapshot_decoder", "decode")
		fileInfo, err := ioutil.ReadDir(decodedDir)
		if err != nil {
			p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", decodedDir))
			return
		}

		for _, file := range fileInfo {
			if !file.IsDir() && path.Ext(file.Name()) == ".pcap" {
				p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing L2Trace using text2pcap/tshark... [%s]", file.Name()))

				mapEventHeader := make(map[string][]string)
				mapEventHeaderOk := make(map[string]bool)
				mapEventRecord := make(map[string]*utils.OrderedMap)

				cmd = exec.Command(path.Join(p.wsharkPath, BIN_TEXT2PCAP), "-u", "5678,0", path.Join(decodedDir, file.Name()), path.Join(outPath, file.Name()))
				cmd.Stdout = &stdOut
				cmd.Stderr = &stdErr
				p.writeLog(zapcore.DebugLevel, cmd.String())
				err = cmd.Run()
				if err != nil {
					p.writeLog(zapcore.ErrorLevel, err.Error())
				}
				if stdOut.Len() > 0 {
					// p.writeLog(zapcore.DebugLevel, stdOut.String())
				}
				if stdErr.Len() > 0 {
					// p.writeLog(zapcore.DebugLevel, stdErr.String())
				}
				stdOut.Reset()
				stdErr.Reset()

				cmd = exec.Command(path.Join(p.wsharkPath, BIN_TSHARK), "-r", path.Join(outPath, file.Name()), "-X", fmt.Sprintf("lua_script:%s", path.Join(p.luasharkPath, LUA_LUASHARK)), "-P", "-V")
				cmd.Stdout = &stdOut
				cmd.Stderr = &stdErr
				p.writeLog(zapcore.DebugLevel, cmd.String())
				err = cmd.Run()
				if err != nil {
					p.writeLog(zapcore.ErrorLevel, err.Error())
				}
				if stdOut.Len() > 0 {
					// TODO use bytes.Buffer.readString("\n") to postprocessing text files
					p.writeLog(zapcore.InfoLevel, fmt.Sprintf("splitting L2Trace into csv..."))
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
						outFn := path.Join(outPath, fmt.Sprintf("%s_%s.csv", strings.Replace(path.Base(file.Name()), "uplane_ttitrace_decoder_", "", -1), k1))
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
				stdOut.Reset()
				stdErr.Reset()

				// remove intermediate pcap
				os.Remove(path.Join(outPath, file.Name()))
			}
		}

		// remove decoded folder
		decodedDir = path.Join(outPath, "decoded")
		if err := os.RemoveAll(decodedDir); err != nil {
			panic(fmt.Sprintf("Fail to remove directory: %v", err))
		}
	}
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