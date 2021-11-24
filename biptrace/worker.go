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
	"bufio"
	"bytes"
	"fmt"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BIN_TSHARK        string = "tshark.exe"
	BIN_TSHARK_LINUX  string = "tshark"
	BIN_EDITCAP string = "editcap.exe"
	BIN_EDITCAP_LINUX string = "editcap"
	LUA_LUASHARK      string = "luashark.lua"
	BIP_PUCCH_REQ     int    = 0
	BIP_PUCCH_RESP_PS int    = 1
	BIP_PUSCH_REQ     int    = 2
	BIP_PUSCH_RESP_PS int    = 3
	VG_IMG_WIDTH      int    = 6
	VG_IMG_HEIGHT     int    = 3
)

type BipTraceParser struct {
	log          *zap.Logger
	wsharkPath   string
	luasharkPath string
	bipTracePath string
	pattern      string
	scs          string
	chbw         string
	maxgo        int
	debug        bool

	headerWritten cmap.ConcurrentMap
}

func (p *BipTraceParser) Init(log *zap.Logger, lua, wshark, trace, pattern, scs, chbw string, maxgo int, debug bool) {
	p.log = log
	p.luasharkPath = lua
	p.wsharkPath = wshark
	p.bipTracePath = trace
	p.pattern = pattern
	p.scs = strings.ToLower(scs)
	p.chbw = strings.ToLower(chbw)
	p.maxgo = utils.MaxInt([]int{2, maxgo})
	p.debug = debug
	p.headerWritten = cmap.New()

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing BIP trace parser...(working dir: %v)", trace))
}

func (p *BipTraceParser) Exec() {
	// recreate dir for parsed bip trace
	outPath := filepath.Join(p.bipTracePath, "parsed_biptrace")
	if err := os.RemoveAll(outPath); err != nil {
		panic(fmt.Sprintf("Fail to remove directory: %v", err))
	}

	if err := os.MkdirAll(outPath, 0775); err != nil {
		panic(fmt.Sprintf("Fail to create directory: %v", err))
	}

	if p.pattern == ".pcap" {
		fileInfo, err := ioutil.ReadDir(p.bipTracePath)
		if err != nil {
			p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", p.bipTracePath))
			return
		}

		/*
		wg := &sync.WaitGroup{}
		for _, file := range fileInfo {
			if !file.IsDir() && strings.HasPrefix(filepath.Ext(file.Name()), ".pcap") {
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
				}(file.Name())
			}
		}
		wg.Wait()
		 */

		for _, file := range fileInfo {
			if !file.IsDir() && strings.HasPrefix(filepath.Ext(file.Name()), ".pcap") {
				p.parse(file.Name())
			}
		}

		// write header
		for _, event := range p.headerWritten.Keys() {
			outFn := filepath.Join(outPath, fmt.Sprintf("%s.csv", event))
			tmpFn := filepath.Join(outPath, fmt.Sprintf("%s.csv.tmp", event))
			fout, err := os.OpenFile(tmpFn, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0664)
			if err != nil {
				p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", tmpFn))
				break
			}

			data, _ := ioutil.ReadFile(outFn)
			m, _ := p.headerWritten.Get(event)
			fout.WriteString(m.(string) + "\n")
			fout.Write(data)
			fout.Close()
			os.Rename(tmpFn, outFn)
		}
	}

	// noisePower from BIP can be used to calculate offset of DDR4 RSSI
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Collecting PUCCH/PUSCH noisePower..."))
	bipMsgs := []string{"PucchReceiveReq.csv", "PucchReceiveRespPs.csv", "PuschReceiveReq.csv", "PuschReceiveRespPs.csv"}
	msgFields := [][]string{
		{"timestamp", "sfn (Uint16)", "slot (Uint8)", "subcellId (Uint8)", "rnti (Uint16)", "pucchFormat (Enum)", "startPrb (Uint16)", "numOfPrb (Uint8)", "frequencyHopping (Enum)", "secondHopPrb (Uint16)"},
		{"timestamp", "sfn (Uint16)", "slot (Uint8)", "subcellId (Uint8)", "rnti (Uint16)", "noisePower (Float32)"},
		{"timestamp", "sfn (Uint16)", "slot (Uint8)", "subcellId (Uint8)", "rnti (Uint16)", "startPrb (Uint16)", "numOfPrb (Uint16)"},
		{"timestamp", "sfn (Uint16)", "slot (Uint8)", "subcellId (Uint8)", "noisePower (Float32)", "rnti (Uint16)"},
	}

	posMap := make(map[string]map[string][]int)
	dataMap := make(map[string]map[string][]string)
	for i, msg := range bipMsgs {
		fin, err := os.Open(filepath.Join(outPath, msg))
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, err.Error())
			// return
			continue
		}

		reader := bufio.NewReader(fin)
		firstLine := true
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

			tokens := strings.Split(line, ",")
			if firstLine {
				posMap[msg] = make(map[string][]int)
				dataMap[msg] = make(map[string][]string)

				for k, tok := range tokens {
					if _, e := posMap[msg][tok]; !e {
						posMap[msg][tok] = []int{k}
					} else {
						posMap[msg][tok] = append(posMap[msg][tok], k)
					}
				}

				// special handing of pucchReceiveReq.startPrb for ABIO
				lenStartPrb, lenNumOfPrb := len(posMap[bipMsgs[BIP_PUCCH_REQ]]["startPrb (Uint16)"]), len(posMap[bipMsgs[BIP_PUCCH_REQ]]["numOfPrb (Uint8)"])
				if lenStartPrb > lenNumOfPrb {
					posMap[bipMsgs[BIP_PUCCH_REQ]]["startPrb (Uint16)"] = posMap[bipMsgs[BIP_PUCCH_REQ]]["startPrb (Uint16)"][lenStartPrb-lenNumOfPrb:]
				}

				firstLine = false

				msgFieldsInfo := fmt.Sprintf("posMap[%v]:", msg)
				for _, field := range msgFields[i] {
					msgFieldsInfo += fmt.Sprintf(" %v=%v", field, posMap[msg][field])
				}
				p.writeLog(zapcore.DebugLevel, msgFieldsInfo)
			} else {
				// key prefix: timestamp_sfn_slot_subcellId
				keyPrefix := make([]string, 4)
				for k := 0; k < 4; k++ {
					pos := posMap[msg][msgFields[i][k]][0]
					keyPrefix[k] = tokens[pos]
					//p.writeLog(zapcore.DebugLevel, tokens[pos])
					if k == 0 {
						keyPrefix[k] = keyPrefix[k][:len("2006-01-02_15:04:05")]
					} else {
						keyPrefix[k] = strings.Split(strings.TrimSuffix(keyPrefix[k], ")"), "(")[1]
					}
				}
				for _, field := range msgFields[i][4:] {
					key := strings.Join(append(keyPrefix, strings.Split(field, " ")[0]), "_")
					if len(posMap[msg][field]) == 1 {
						if len(tokens) < posMap[msg][field][0] + 1 {
							continue
						}

						val := strings.Split(tokens[posMap[msg][field][0]], " ")[0]
						if strings.HasSuffix(val, ")") {
							val = strings.Split(strings.TrimSuffix(val, ")"), "(")[1]
						}
						dataMap[msg][key] = []string{val}
					} else {
						dataMap[msg][key] = []string{}
						for _, pos := range posMap[msg][field] {
							if pos < len(tokens) {
								val := strings.Split(tokens[pos], " ")[0]
								if strings.HasSuffix(val, ")") {
									val = strings.Split(strings.TrimSuffix(val, ")"), "(")[1]
								}
								dataMap[msg][key] = append(dataMap[msg][key], val)
							}
						}
					}
				}
			}
		}
	}

	/*
		for msg := range dataMap {
			p.writeLog(zapcore.DebugLevel, fmt.Sprintf("dataMap[%v] info:", msg))
			for key := range dataMap[msg] {
				p.writeLog(zapcore.DebugLevel, fmt.Sprintf("key=%v;%v, val=%v", key, msg, dataMap[msg][key]))
			}
		}
	 */

	// key = scs_bw, val = PRB number
	nbrPrbMap := map[string]int{
		"15k_5m":    25,
		"15k_10m":   52,
		"15k_15m":   79,
		"15k_20m":   106,
		"15k_30m":   160,
		"15k_40m":   216,
		"30k_20m":   51,
		"30k_40m":   106,
		"30k_50m":   133,
		"30k_60m":   162,
		"30k_80m":   217,
		"30k_90m":   245,
		"30k_100m":  273,
		"120k_100m": 66,
	}

	nbrPrb := nbrPrbMap[p.scs+"_"+p.chbw]
	pucchRssiMap := make(map[string]map[int][]float64)
	puschRssiMap := make(map[string]map[int][]float64)
	/*
	for i := 0; i < nbrPrb; i++ {
		pucchRssiMap[i] = make([]float64, 0)
		puschRssiMap[i] = make([]float64, 0)
	}
	 */

	// collect PUSCH noisePower
	for key := range dataMap[bipMsgs[BIP_PUCCH_RESP_PS]] {
		subcell := strings.Split(key, "_")[4] // timestamp has format of 2006-01-02_15:04:05
		if _, e := pucchRssiMap[subcell]; !e {
			pucchRssiMap[subcell] = make(map[int][]float64)
			for i := 0; i < nbrPrb; i++ {
				pucchRssiMap[subcell][i] = make([]float64, 0)
			}
		}
		if strings.HasSuffix(key, "rnti") {
			rntiPucchReq := dataMap[bipMsgs[BIP_PUCCH_REQ]][key]
			startPrbPucchReq := dataMap[bipMsgs[BIP_PUCCH_REQ]][strings.Replace(key, "rnti", "startPrb", -1)]
			numOfPrbPucchReq := dataMap[bipMsgs[BIP_PUCCH_REQ]][strings.Replace(key, "rnti", "numOfPrb", -1)]
			freqHopPucchReq := dataMap[bipMsgs[BIP_PUCCH_REQ]][strings.Replace(key, "rnti", "frequencyHopping", -1)]
			secondHopPrbPucchReq := dataMap[bipMsgs[BIP_PUCCH_REQ]][strings.Replace(key, "rnti", "secondHopPrb", -1)]

			rntiPucchRespPs := dataMap[bipMsgs[BIP_PUCCH_RESP_PS]][key]
			noisePucchRespPs := dataMap[bipMsgs[BIP_PUCCH_RESP_PS]][strings.Replace(key, "rnti", "noisePower", -1)]

			for i, rnti := range rntiPucchRespPs {
				j := utils.IndexStr(rntiPucchReq, rnti)

				if j == -1 {
					p.writeLog(zapcore.DebugLevel, fmt.Sprintf("RNTI mismatch(known timestamp issue), key=%v, rnti=%v, skipping", key, rnti))
					continue
				}

				noisePower, _ := strconv.ParseFloat(noisePucchRespPs[i], 64)
				startPrb, _ := strconv.ParseInt(startPrbPucchReq[j], 10, 32)
				numOfPrb, _ := strconv.ParseInt(numOfPrbPucchReq[j], 10, 32)
				freqHop, _ := strconv.ParseBool(freqHopPucchReq[j])
				secondHopPrb, _ := strconv.ParseInt(secondHopPrbPucchReq[j], 10, 32)

				for k := 0; k < int(numOfPrb); k++ {
					pucchRssiMap[subcell][int(startPrb)+k] = append(pucchRssiMap[subcell][int(startPrb)+k], noisePower)
				}

				if freqHop {
					for k := 0; k < int(numOfPrb); k++ {
						pucchRssiMap[subcell][int(secondHopPrb)+k] = append(pucchRssiMap[subcell][int(secondHopPrb)+k], noisePower)
					}
				}
			}
		}
	}

	pucchInfo := make(map[string][]int)
	for subcell := range pucchRssiMap {
		pucchInfo[subcell] = make([]int, nbrPrb)
		for i := 0; i < nbrPrb; i++ {
			pucchInfo[subcell][i] = len(pucchRssiMap[subcell][i])
		}
	}
	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("pucchInfo = %v", pucchInfo))

	// collect PUSCH noisePower
	for key := range dataMap[bipMsgs[BIP_PUSCH_RESP_PS]] {
		subcell := strings.Split(key, "_")[4] // timestamp has format of 2006-01-02_15:04:05
		if _, e := puschRssiMap[subcell]; !e {
			puschRssiMap[subcell] = make(map[int][]float64)
			for i := 0; i < nbrPrb; i++ {
				puschRssiMap[subcell][i] = make([]float64, 0)
			}
		}
		if strings.HasSuffix(key, "rnti") {
			rntiPuschReq := dataMap[bipMsgs[BIP_PUSCH_REQ]][key]
			startPrbPuschReq := dataMap[bipMsgs[BIP_PUSCH_REQ]][strings.Replace(key, "rnti", "startPrb", -1)]
			numOfPrbPuschReq := dataMap[bipMsgs[BIP_PUSCH_REQ]][strings.Replace(key, "rnti", "numOfPrb", -1)]

			rntiPuschRespPs := dataMap[bipMsgs[BIP_PUSCH_RESP_PS]][key]
			noisePuschRespPs := dataMap[bipMsgs[BIP_PUSCH_RESP_PS]][strings.Replace(key, "rnti", "noisePower", -1)]

			noisePower, _ := strconv.ParseFloat(noisePuschRespPs[0], 64)
			for _, rnti := range rntiPuschRespPs {
				j := utils.IndexStr(rntiPuschReq, rnti)

				if j == -1 {
					p.writeLog(zapcore.DebugLevel, fmt.Sprintf("RNTI mismatch(known timestamp issue), key=%v, rnti=%v, skipping", key, rnti))
					continue
				}

				startPrb, _ := strconv.ParseInt(startPrbPuschReq[j], 10, 32)
				numOfPrb, _ := strconv.ParseInt(numOfPrbPuschReq[j], 10, 32)

				for k := 0; k < int(numOfPrb); k++ {
					puschRssiMap[subcell][int(startPrb)+k] = append(puschRssiMap[subcell][int(startPrb)+k], noisePower)
				}
			}
		}
	}

	puschInfo := make(map[string][]int)
	for subcell := range puschRssiMap {
		puschInfo[subcell] = make([]int, nbrPrb)
		for i := 0; i < nbrPrb; i++ {
			puschInfo[subcell][i] = len(puschRssiMap[subcell][i])
		}
	}
	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("puschInfo = %v", puschInfo))

	rssi := make(map[string][]float64)
	for subcell := range puschRssiMap {
		rssi[subcell] = make([]float64, nbrPrb)
	}
	for i := 0; i < nbrPrb; i++ {
		for subcell := range puschRssiMap {
			for j := range pucchRssiMap[subcell][i] {
				rssi[subcell][i] += math.Pow(10, pucchRssiMap[subcell][i][j]/10)
			}

			for j := range puschRssiMap[subcell][i] {
				rssi[subcell][i] += math.Pow(10, puschRssiMap[subcell][i][j]/10)
			}

			if len(pucchRssiMap[subcell][i])+len(puschRssiMap[subcell][i]) > 0 {
				rssi[subcell][i] = 10 * math.Log10(rssi[subcell][i]/float64(len(pucchRssiMap[subcell][i])+len(puschRssiMap[subcell][i])))
			} else {
				scs, _ := strconv.ParseFloat(strings.TrimSuffix(p.scs, "k"), 64)
				rssi[subcell][i] = -174 + 10*math.Log10(scs*12*1000)
			}
		}
	}

	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("RSSI = %v", rssi))

	// save per PRB RSSI as .png using gonum/plot
	for subcell := range rssi {
		pts := make(plotter.XYs, nbrPrb)
		pts2 := make(plotter.XYs, nbrPrb)
		pts3 := make(plotter.XYs, nbrPrb)
		for i := range pts {
			pts[i].X = float64(i)
			if _, e := pucchInfo[subcell]; e {
				pts[i].Y = float64(pucchInfo[subcell][i])
			} else {
				pts[i].Y = 0
			}

			pts2[i].X = float64(i)
			if _, e := puschInfo[subcell]; e {
				pts2[i].Y = float64(puschInfo[subcell][i])
			} else {
				pts2[i].Y = 0
			}

			pts3[i].X = float64(i)
			pts3[i].Y = rssi[subcell][i]
		}

		const rows, cols = 2, 1
		plots := make([][]*plot.Plot, rows)
		for j := 0; j < rows; j++ {
			plots[j] = make([]*plot.Plot, cols)
			for i := 0; i < cols; i++ {
				pl := plot.New()
				pl.Add(plotter.NewGrid())

				pl.Legend.Top = true

				if i == 0 && j == 0 {
					pl.Title.Text = fmt.Sprintf("BIP PUCCH/PUSCH Usage(subcell%v)", subcell)
					pl.X.Label.Text = "PRB"
					pl.Y.Label.Text = "Count(#)"
					pl.X.Min = 0
					pl.X.Max = float64(nbrPrb - 1)
					plotutil.AddLines(pl, "PUCCH_count", pts, "PUSCH_count", pts2)
				}

				if i == 0 && j == 1 {
					pl.Title.Text = fmt.Sprintf("BIP PUCCH/PUSCH noisePower(subcell%v)", subcell)
					pl.X.Label.Text = "PRB"
					pl.Y.Label.Text = "noisePower(dBm)"
					pl.X.Min = 0
					pl.X.Max = float64(nbrPrb - 1)
					pl.Y.Min = -140
					pl.Y.Max = -60
					plotutil.AddLines(pl, "noisePower_per_PRB", pts3)
				}

				plots[j][i] = pl
			}
		}

		width, _ := vg.ParseLength(fmt.Sprintf("%vin", cols*VG_IMG_WIDTH))
		height, _ := vg.ParseLength(fmt.Sprintf("%vin", rows*VG_IMG_HEIGHT))
		img := vgimg.New(width, height)
		dc := draw.New(img)
		t := draw.Tiles{
			Rows:      rows,
			Cols:      cols,
			PadX:      vg.Millimeter,
			PadY:      vg.Millimeter,
			PadTop:    vg.Points(2),
			PadBottom: vg.Points(2),
			PadLeft:   vg.Points(2),
			PadRight:  vg.Points(2),
		}
		canvases := plot.Align(plots, t, dc)
		for j := 0; j < rows; j++ {
			for i := 0; i < cols; i++ {
				if plots[j][i] != nil {
					plots[j][i].Draw(canvases[j][i])
				}
			}
		}

		w, err := os.Create(filepath.Join(outPath, fmt.Sprintf("bip_noisePower_subcell%v.png", subcell)))
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, err.Error())
		}
		defer w.Close()

		png := vgimg.PngCanvas{Canvas: img}
		if _, err := png.WriteTo(w); err != nil {
			p.writeLog(zapcore.ErrorLevel, err.Error())
		}
	}
}

func (p *BipTraceParser) parse(fn string) {
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Splitting BIP trace using editcap...[%s]", fn))

	ecPath := filepath.Join(p.bipTracePath, strings.Replace(fn, ".", "_", -1))
	if err := os.RemoveAll(ecPath); err != nil {
		panic(fmt.Sprintf("Fail to remove directory: %v", err))
	}
	if err := os.MkdirAll(ecPath, 0775); err != nil {
		panic(fmt.Sprintf("Fail to create directory: %v", err))
	}

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command(filepath.Join(p.wsharkPath, BIN_EDITCAP_LINUX), "-c", "50000",  filepath.Join(p.bipTracePath, fn), filepath.Join(ecPath, "ec.pcap"))
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command(filepath.Join(p.wsharkPath, BIN_EDITCAP), "-c", "50000",  filepath.Join(p.bipTracePath, fn), filepath.Join(ecPath, "ec.pcap"))
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

	outPath := filepath.Join(p.bipTracePath, "parsed_biptrace")
	//mapEventRecord := cmap.New()

	ecl, _ := filepath.Glob(filepath.Join(ecPath, "ec*.pcap"))
	wg := &sync.WaitGroup{}
	for _, ec := range ecl {
		for {
			if runtime.NumGoroutine() >= p.maxgo {
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		wg.Add(1)
		go func(ec string) {
			defer wg.Done()

			mapEventHeader := cmap.New()
			mapEventRecord := cmap.New()

			var ecStdOut bytes.Buffer
			var ecStdErr bytes.Buffer
			var ecCmd *exec.Cmd
			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Parsing BIP trace using tshark... [%s/%s]", fn, filepath.Base(ec)))
			if runtime.GOOS == "linux" {
				ecCmd = exec.Command(filepath.Join(p.wsharkPath, BIN_TSHARK_LINUX), "-r", ec, "-X", fmt.Sprintf("lua_script:%s", filepath.Join(p.luasharkPath, LUA_LUASHARK)), "-V")
			} else if runtime.GOOS == "windows" {
				ecCmd = exec.Command(filepath.Join(p.wsharkPath, BIN_TSHARK), "-r", ec, "-X", fmt.Sprintf("lua_script:%s", filepath.Join(p.luasharkPath, LUA_LUASHARK)), "-V")
			} else {
				p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Unsupported OS: runtime.GOOS=%s", runtime.GOOS))
				return
			}
			ecCmd.Stdout = &ecStdOut
			ecCmd.Stderr = &ecStdErr
			p.writeLog(zapcore.DebugLevel, ecCmd.String())
			if err := ecCmd.Run(); err != nil {
				p.writeLog(zapcore.ErrorLevel, err.Error())
			}
			if ecStdOut.Len() > 0 {
				// TODO use bytes.Buffer.readString("\n") to postprocessing text files
				p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Splitting BIP trace into csv... [%s/%s]", fn, filepath.Base(ec)))
				icomRec := false
				var ts string
				var event string
				var fields string
				for {
					line, err := ecStdOut.ReadString('\n')
					if err != nil {
						break
					}

					// remove leading and tailing spaces
					line = strings.TrimSpace(line)
					if len(line) > 0 {
						// skip [...] lines such as: [8 TB Payload fragments (65580 bytes): #33(8960), #34(8960), #35(8960), #36(8960), #37(8960), #39(8960), #40(8960), #41(2860)]
						if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
							continue
						}

						if strings.HasPrefix(line, "Frame") && strings.Contains(line, "on wire") && strings.Contains(line, "captured") {
							icomRec = false

							// SKipping event DlData_EmptySendReq
							if len(fields) > 0 && event != "EmptySendReq" {
								m, _ := mapEventRecord.Get(event)
								m.(cmap.ConcurrentMap).Set(ts, fields)
								mapEventRecord.Set(event, m)

								m2, _ := mapEventHeader.Get(event)
								header := strings.Join(m2.([]string), ",")
								m, e := p.headerWritten.Get(event)
								if !e {
									p.headerWritten.Set(event, header)
									// p.writeLog(zapcore.DebugLevel, fmt.Sprintf("event=%v, headerWrittern=%v", event, header))
								} else {
									if len(header) == len(m.(string)) && header != m.(string) {
										p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("event=%v, header mismatch while having same length", event))
									}
									if len(header) > len(m.(string)) && strings.HasPrefix(header, m.(string)) {
										p.headerWritten.Set(event, header)
										// p.writeLog(zapcore.DebugLevel, fmt.Sprintf("event=%v, headerWrittern=%v", event, header))
									}
								}
							}
						}

						//if bipEvent && strings.Split(line, ":")[0] == "Epoch Time" {
						if strings.Split(line, ":")[0] == "Epoch Time" {
							// Epoch Time: 1621323617.322338000 seconds
							tokens := strings.Split(strings.Split(line, " ")[2], ".")
							sec, _ := strconv.ParseInt(tokens[0], 10, 64)
							nsec, _ := strconv.ParseInt(tokens[1], 10, 64)
							ts = time.Unix(sec, nsec).Format("2006-01-02_15:04:05.999999999")
						}

						if line == "ICOM 5G Protocol" {
							icomRec = true

							nextLine, err := ecStdOut.ReadString('\n')
							if err != nil {
								break
							}

							event = strings.Split(strings.TrimSpace(nextLine), " ")[0]
							if !mapEventRecord.Has(event) {
								mapEventRecord.Set(event, cmap.New())
							}

							mapEventHeader.Set(event, []string{"eventType", "timestamp"})
							fields = fmt.Sprintf("%s,%s", event, ts)
						}

						if icomRec {
							if strings.Contains(line, "padding") {
								continue
							}

							if strings.Contains(line, "Structure") {
								m, _ := mapEventHeader.Get(event)
								mapEventHeader.Set(event, append(m.([]string), line))
								fields = fields + ",|"
							}

							tokens := strings.Split(line, ":")
							if len(tokens) == 2 {
								m, _ := mapEventHeader.Get(event)
								mapEventHeader.Set(event, append(m.([]string), tokens[0]))
								fields = fields + "," + strings.Replace(strings.TrimSpace(tokens[1]), ",", "_", -1)
							}
						}
					}
				}
			}
			if ecStdErr.Len() > 0 {
				p.writeLog(zapcore.DebugLevel, ecStdErr.String())
			}

			for _, k1 := range mapEventRecord.Keys() {
				// SKipping event DlData_EmptySendReq
				if k1 == "EmptySendReq" {
					continue
				}

				outFn := filepath.Join(outPath, fmt.Sprintf("%s.csv", k1))
				fout, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
				if err != nil {
					p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", outFn))
					break
				}

				m, _ := mapEventRecord.Get(k1)
				mks := m.(cmap.ConcurrentMap).Keys()
				sort.Strings(mks)
				for _, k2 := range mks {
					v2, _ := m.(cmap.ConcurrentMap).Get(k2)
					fout.WriteString(v2.(string) + "\n")
				}

				fout.Close()
			}
		} (ec)
	}
	wg.Wait()

	if err := os.RemoveAll(ecPath); err != nil {
		panic(fmt.Sprintf("Fail to remove directory: %v", err))
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
