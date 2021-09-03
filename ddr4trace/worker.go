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
	"bufio"
	"bytes"
	"fmt"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gonum.org/v1/gonum/dsp/fourier"
	"io/ioutil"
	"math/cmplx"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BIN_PYTHON3         string = "python.exe"
	BIN_PYTHON3_LINUX   string = "python3"
	PY3_SNAPSHOT_TOOL string = "SnapshotAnalyzer.py"
	LOKI_IQ_NORM_FACTOR int = 16384
)

type Ddr4TraceParser struct {
	log          *zap.Logger
	py3Path string
	snapToolPath string
	ddr4TracePath string
	pattern      string
	maxgo int
	debug        bool

	traceFiles []string
	iqdata cmap.ConcurrentMap
}

func (p *Ddr4TraceParser) Init(log *zap.Logger, py3, snaptool, trace, pattern string, maxgo int, debug bool) {
	p.log = log
	p.py3Path = py3
	p.snapToolPath = snaptool
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

	p.iqdata = cmap.New()
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

	for _, key := range p.iqdata.Keys() {
		m, _ := p.iqdata.Get(key)
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("iqdata: key=%v, len=%v", key, len(m.([]complex128))))
		/*
		for i := 0; i < 100; i++ {
			p.writeLog(zapcore.DebugLevel, fmt.Sprintf("iq = %v", m.([]complex128)[i]))
		}
		 */

		rf0sl0symb0u := m.([]complex128)[320:4416]
		fft := fourier.NewCmplxFFT(len(rf0sl0symb0u))
		coeff := fft.Coefficients(nil, rf0sl0symb0u)

		iq := make([]complex128, 0)
		amp := make([]float64, 0)
		for i := range coeff {
			i := fft.ShiftIdx(i)
			iq = append(iq, coeff[i])
			amp = append(amp, cmplx.Abs(coeff[i]))
		}

		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Frame 0 Slot 0 Symbol 0: len=%v", len(amp)))
		fout, err := os.OpenFile(path.Join(outPath, strings.Split(key, ".")[0] + "_frame0_slot0_symbol0.csv"), os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, err.Error())
			return
		}

		for i := range iq {
			fout.WriteString(fmt.Sprintf("%v,%v,%v,%v\n", i+1, real(iq[i]), imag(iq[i]), amp[i]))
		}

		fout.Close()
	}
}

func (p *Ddr4TraceParser) parse(fn string) {
	// 1st step: parse DDR4(.bin) using Loki snapshot_tool
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing DDR4 trace using Loki snapshot_tool... [%s]", fn))
	outPath := path.Join(p.ddr4TracePath, "parsed_ddr4trace")
	decodedDir := path.Join(outPath, strings.Split(fn, ".")[0] + "_result")

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command(path.Join(p.py3Path, BIN_PYTHON3_LINUX), path.Join(p.snapToolPath, "SnapshotAnalyzer.py"), "--iqdata", path.Join(p.ddr4TracePath, fn), "--output", decodedDir)
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command(path.Join(p.py3Path, BIN_PYTHON3), path.Join(p.snapToolPath, "SnapshotAnalyzer.py"), "--iqdata", path.Join(p.ddr4TracePath, fn), "--output", decodedDir)
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

	// 2nd step: parsing ant_x.txt where x=0..nap-1
	antFiles := make([]string, 0)
	subcells, _ := filepath.Glob(path.Join(path.Join(decodedDir, "nr", "ul"), "subcellid*"))
	for _, sc := range subcells {
		ants, _ := filepath.Glob(path.Join(sc, "ant*.txt"))
		antFiles = append(antFiles, ants...)
	}

	if p.debug {
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("fn = %v, antFiles=%v\n", fn, antFiles))
	}

	wg := &sync.WaitGroup{}
	for _, file := range antFiles {
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

			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Loading per antenna I/Q samples... [%s]", fn))
			key := strings.Replace(strings.Replace(fn, path.Join(p.ddr4TracePath, "parsed_ddr4trace"), "ddr4", -1), "/", "_", -1)
			p.iqdata.SetIfAbsent(key, make([]complex128, 0))

			fin, err := os.Open(fn)
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
				if len(line) == 0 {
					continue
				}

				iq := strings.Split(line, " ")
				i, _ := strconv.ParseFloat(iq[0], 64)
				q, _ := strconv.ParseFloat(iq[1], 64)

				m, _ := p.iqdata.Get(key)
				m = append(m.([]complex128), complex(i/float64(LOKI_IQ_NORM_FACTOR), q/float64(LOKI_IQ_NORM_FACTOR)))
				p.iqdata.Set(key, m)
			}
		} (file)
	}
	wg.Wait()
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
