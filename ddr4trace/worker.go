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
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"io/ioutil"
	"math"
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
	PY3_SNAPSHOT_TOOL   string = "SnapshotAnalyzer.py"
	LOKI_IQ_NORM_FACTOR int    = 16384
)

type Ddr4TraceParser struct {
	log           *zap.Logger
	py3Path       string
	snapToolPath  string
	ddr4TracePath string
	pattern       string
	scs           string
	chbw          string
	maxgo         int
	gain          float64
	debug         bool

	traceFiles []string
	iqData     cmap.ConcurrentMap
	rssiData   cmap.ConcurrentMap
	rssiCount  cmap.ConcurrentMap
}

func (p *Ddr4TraceParser) Init(log *zap.Logger, py3, snaptool, trace, pattern, scs, chbw string, maxgo, gain int, debug bool) {
	p.log = log
	p.py3Path = py3
	p.snapToolPath = snaptool
	p.ddr4TracePath = trace
	p.pattern = pattern
	p.scs = strings.ToLower(scs)
	p.chbw = strings.ToLower(chbw)
	p.maxgo = utils.MaxInt([]int{2, maxgo})
	p.gain = float64(gain)
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

	p.iqData = cmap.New()
	p.rssiData = cmap.New()
	p.rssiCount = cmap.New()
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

	if p.pattern == ".bin" {
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
				}(file.Name())
			}
		}
		wg.Wait()
	}

	// calculate sampling constants
	// key = scs_bw, val = FFT size
	fftSizeMap := map[string]int{
		"15k_5m":    512,
		"15k_10m":   1024,
		"15k_15m":   1536,
		"15k_20m":   2048,
		"15k_30m":   4096,
		"15k_40m":   4096,
		"30k_20m":   1024,
		"30k_40m":   2048,
		"30k_50m":   2048,
		"30k_60m":   4096,
		"30k_80m":   4096,
		"30k_90m":   4096,
		"30k_100m":  4096,
		"120k_100m": 1024,
	}

	// key = scs_bw, val = samples rate(Msps)
	samplingRateMap := map[string]float64{
		"15k_5m":    7.68,
		"15k_10m":   15.35,
		"15k_15m":   23.04,
		"15k_20m":   30.72,
		"15k_30m":   61.44,
		"15k_40m":   61.44,
		"30k_20m":   30.72,
		"30k_40m":   61.44,
		"30k_50m":   61.44,
		"30k_60m":   122.88,
		"30k_80m":   122.88,
		"30k_90m":   122.88,
		"30k_100m":  122.88,
		"120k_100m": 122.88,
	}

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

	var slotsPerRf int
	var tc, k, u, deltaF, exp2Negu, symbsPerSubf float64
	var nu, ncpSymb07, ncpSymbOth float64
	// OFDM symbol length(nomal, l=0 or l=7*2^u) and OFDM symbol length(nomal, l<>0 and l<>7*2^u), unit is ms
	var lenSymb07, lenSymbOth float64

	tc = 1.0 / (480 * 1000 * 4096) * 1000
	k = 64
	switch p.scs {
	case "15k":
		u = 0
	case "30k":
		u = 1
	case "120k":
		u = 3
	default:
		u = -1
	}
	deltaF = 15 * math.Exp2(u)
	slotsPerRf = int(10 * math.Exp2(u))
	exp2Negu = 1 / math.Exp2(u)
	symbsPerSubf = 14 * math.Exp2(u)
	nu = 2048 * k * exp2Negu
	ncpSymb07 = 144*k*exp2Negu + 16*k
	ncpSymbOth = 144 * k * exp2Negu
	lenSymb07 = (nu + ncpSymb07) * tc
	lenSymbOth = (nu + ncpSymbOth) * tc

	// Sampling rate, unit is Msps
	samplingRate := samplingRateMap[p.scs+"_"+p.chbw]
	fftSize := fftSizeMap[p.scs+"_"+p.chbw]
	// Sampling time, unit is ms
	samplingTime := 1 / samplingRate / 1000

	// Transmission bandwidth in PRB
	nbrPrb := nbrPrbMap[p.scs+"_"+p.chbw]
	nbrRe := nbrPrb * 12
	firstRe := (fftSize - nbrRe) / 2
	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("nbrPrb=%v, nbrRe=%v [firstRe=%v, lastRe=%v]", nbrPrb, nbrRe, firstRe, firstRe+nbrRe-1))

	var samplesSymb07, samplesSymbOth, samplesSlot, samplesCpSymb07, samplesCpSymbOth int
	samplesSymb07 = int(lenSymb07 / samplingTime)
	samplesSymbOth = int(lenSymbOth / samplingTime)
	samplesSlot = 2*samplesSymb07 + 12*samplesSymbOth
	samplesCpSymb07 = int(ncpSymb07 * tc / samplingTime)
	samplesCpSymbOth = int(ncpSymbOth * tc / samplingTime)

	samplesExCpSymb07 := samplesSymb07 - samplesCpSymb07
	samplesExCpSymbOth := samplesSymbOth - samplesCpSymbOth

	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("tc=%v, k=%v, u=%v, deltaF=%v, slotsPerRf=%v, exp2Negu=%v", tc, k, u, deltaF, slotsPerRf, exp2Negu))
	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("nu=%v, ncpSymb07=%v, ncpSymbOth=%v, lenSymb07=%vms, lenSymbOth=%vms, lenSubframe=%vms", nu, ncpSymb07, ncpSymbOth, lenSymb07, lenSymbOth, 2*lenSymb07+(symbsPerSubf-2)*lenSymbOth))
	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("SamplingRate=%vMsps, fftSize=%v, samplingTime=%vms", samplingRate, fftSize, samplingTime))
	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("samplesSymb07=%v, samplesSymbOth=%v, samplesSlot=%v, samplesCpSymb07=%v, samplesCpSymbOth=%v, samplesExCpSymb07=%v, samplesExCpSymbOth=%v", samplesSymb07, samplesSymbOth, samplesSlot, samplesCpSymb07, samplesCpSymbOth, samplesExCpSymb07, samplesExCpSymbOth))

	wg := &sync.WaitGroup{}
	for _, key := range p.iqData.Keys() {
		for {
			if runtime.NumGoroutine() >= p.maxgo {
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			m, _ := p.iqData.Get(key)
			totSamples := len(m.([]complex128))
			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Processing iqData: key=%v, len=%v", key, totSamples))
			/*
				for i := 0; i < 100; i++ {
					p.writeLog(zapcore.DebugLevel, fmt.Sprintf("iq = %v", m.([]complex128)[i]))
				}
			*/

			tks := strings.Split(strings.Split(key, ".")[0], "_")
			ant := tks[len(tks)-1]
			p.rssiData.SetIfAbsent(ant, cmap.New())
			p.rssiCount.SetIfAbsent(ant, cmap.New())

			nrf := totSamples / (samplesSlot * slotsPerRf)
			for irf := range utils.PyRange(0, nrf, 1) {
				for isl := range utils.PyRange(0, slotsPerRf, 1) {
					k := irf*slotsPerRf + isl
					z := m.([]complex128)[k*samplesSlot : (k+1)*samplesSlot]

					c := 0
					for isymb := range utils.PyRange(0, 14, 1) {
						var uiq []complex128
						if isymb == 0 || isymb == int(7*math.Exp2(u)) {
							s := z[c : c+samplesSymb07]
							c += samplesSymb07
							uiq = s[samplesCpSymb07:]
						} else {
							s := z[c : c+samplesSymbOth]
							c += samplesSymbOth
							uiq = s[samplesCpSymbOth:]
						}

						fft := fourier.NewCmplxFFT(len(uiq))
						coeff := fft.Coefficients(nil, uiq)

						// iq := make([]complex128, 0)
						amp := make([]float64, 0)
						for i := range coeff {
							i := fft.ShiftIdx(i)
							// iq = append(iq, coeff[i])
							amp = append(amp, cmplx.Abs(coeff[i]))
						}

						// p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Frame%v,Slot%v,Symbol%v: len=%v", irf, isl, isymb, len(amp)))

						key2 := fmt.Sprintf("symbol%v", isymb)
						m, _ := p.rssiData.Get(ant)
						m.(cmap.ConcurrentMap).SetIfAbsent(key2, make([]float64, len(amp)))
						m2, _ := m.(cmap.ConcurrentMap).Get(key2)
						for i := range amp {
							// unit is mW assuming 50 Ohm impedance, P = V^2 / 2R and V = sqrt(I^2 + Q^2)
							m2.([]float64)[i] += 10 * math.Pow(amp[i], 2)
						}
						m.(cmap.ConcurrentMap).Set(key2, m2)
						p.rssiData.Set(ant, m)

						c, _ := p.rssiCount.Get(ant)
						c.(cmap.ConcurrentMap).SetIfAbsent(key2, 0)
						c2, _ := c.(cmap.ConcurrentMap).Get(key2)
						c2 = c2.(int) + 1
						c.(cmap.ConcurrentMap).Set(key2, c2)
						p.rssiCount.Set(ant, c)
					}
				}
			}
		}(key)
	}
	wg.Wait()

	// calculate average RSSI in mW
	for _, ant := range p.rssiData.Keys() {
		m, _ := p.rssiData.Get(ant)
		c, _ := p.rssiCount.Get(ant)
		for _, symb := range m.(cmap.ConcurrentMap).Keys() {
			m2, _ := m.(cmap.ConcurrentMap).Get(symb)
			c2, _ := c.(cmap.ConcurrentMap).Get(symb)
			// p.writeLog(zapcore.DebugLevel, fmt.Sprintf("ant=%v,symb=%v,count=%v", ant, symb, c2.(int)))
			for i := range m2.([]float64) {
				m2.([]float64)[i] = m2.([]float64)[i] / float64(c2.(int))
			}

			m.(cmap.ConcurrentMap).Set(symb, m2)
		}
		p.rssiData.Set(ant, m)
	}

	// save per ant/symbol RSSI as rssi_per_ant_symbol_x.csv
	fout, err := os.OpenFile(path.Join(outPath, fmt.Sprintf("rssi_per_ant_symbol_re.csv")), os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
		return
	}
	fout.WriteString("Antenna Port,Symbol,FFT Bin,RSSI(dBm)\n")

	fout2, err2 := os.OpenFile(path.Join(outPath, fmt.Sprintf("rssi_per_ant_symbol_prb.csv")), os.O_WRONLY|os.O_CREATE, 0664)
	if err2 != nil {
		p.writeLog(zapcore.ErrorLevel, err2.Error())
		return
	}
	fout2.WriteString("Antenna Port,Symbol,PRB,RSSI(dBm)\n")

	// key = symbol, val = RSSI per RE or RSSI per PRB
	rssiRe := make(map[string][]float64)
	rssiPrb := make(map[string][]float64)
	for _, ant := range p.rssiData.Keys() {
		m, _ := p.rssiData.Get(ant)
		for _, symb := range m.(cmap.ConcurrentMap).Keys() {
			m2, _ := m.(cmap.ConcurrentMap).Get(symb)

			if _, e := rssiRe[symb]; !e {
				rssiRe[symb] = make([]float64, len(m2.([]float64)))
				rssiPrb[symb] = make([]float64, nbrPrb)
			}

			pts := make(plotter.XYs, len(m2.([]float64)))
			pts2 := make(plotter.XYs, nbrPrb)
			for i := range pts {
				pts[i].X = float64(i)
				pts[i].Y = math.Max(10 * math.Log10(m2.([]float64)[i]) - p.gain, -174 + 10 * math.Log10(15000))
				rssiRe[symb][i] += m2.([]float64)[i]

				fout.WriteString(fmt.Sprintf("%v,%v,%v,%v\n", ant, symb, pts[i].X, pts[i].Y))

				if i >= firstRe && i < (firstRe+nbrRe) {
					iprb := math.Floor(float64(i-firstRe) / 12)
					j := int(iprb)
					pts2[j].Y += m2.([]float64)[i]
					rssiPrb[symb][j] += m2.([]float64)[i]
				}
			}

			for i := range pts2 {
				pts2[i].X = float64(i)
				// pts2[i].Y = 10 * math.Log10(pts2[i].Y / 12 / math.Exp2(p.gain))
				pts2[i].Y = 10 * math.Log10(pts2[i].Y) - p.gain

				fout2.WriteString(fmt.Sprintf("%v,%v,%v,%v\n", ant, symb, pts2[i].X, pts2[i].Y))
			}

			// save per ant/symb RSSI as .png using gonum/plot
			const rows, cols = 2, 1
			plots := make([][]*plot.Plot, rows)
			for j := 0; j < rows; j++ {
				plots[j] = make([]*plot.Plot, cols)
				for i := 0; i < cols; i++ {
					pl := plot.New()
					pl.Add(plotter.NewGrid())
					pl.Title.Text = fmt.Sprintf("RSSI (%v-%v)", ant, symb)
					pl.Y.Label.Text = "RSSI(dBm)"
					pl.Y.Min = -140
					pl.Y.Max = -60
					pl.Legend.Top = true

					if i == 0 && j == 0 {
						pl.X.Label.Text = "FFT Bin"
						pl.X.Min = 0
						pl.X.Max = float64(len(m2.([]float64)) - 1)
						plotutil.AddLines(pl, "RSSI_per_RE", pts)
					}

					if i == 0 && j == 1 {
						pl.X.Label.Text = "PRB"
						pl.X.Min = 0
						pl.X.Max = float64(nbrPrb - 1)
						plotutil.AddLines(pl, "RSSI_per_PRB", pts2)
					}

					plots[j][i] = pl
				}
			}

			img := vgimg.New(8*vg.Inch, 8*vg.Inch)
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

			w, err := os.Create(path.Join(outPath, fmt.Sprintf("rssi_%v_%v.png", ant, symb)))
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

	fout.Close()
	fout2.Close()

	// save per symb RSSI as .png using gonum/plot and as rssi_per_symbol_x.csv
	fout3, err3 := os.OpenFile(path.Join(outPath, fmt.Sprintf("rssi_per_symbol_re.csv")), os.O_WRONLY|os.O_CREATE, 0664)
	if err3 != nil {
		p.writeLog(zapcore.ErrorLevel, err3.Error())
		return
	}
	fout3.WriteString("Symbol,FFT Bin,RSSI(dBm)\n")

	fout4, err4 := os.OpenFile(path.Join(outPath, fmt.Sprintf("rssi_per_symbol_prb.csv")), os.O_WRONLY|os.O_CREATE, 0664)
	if err4 != nil {
		p.writeLog(zapcore.ErrorLevel, err4.Error())
		return
	}
	fout4.WriteString("Symbol,PRB,RSSI(dBm)\n")

	for symb := range rssiRe {
		ptsRe := make(plotter.XYs, len(rssiRe[symb]))
		ptsPrb := make(plotter.XYs, len(rssiPrb[symb]))
		for i := range ptsRe {
			ptsRe[i].X = float64(i)
			ptsRe[i].Y = math.Max(10 * math.Log10(rssiRe[symb][i]) - p.gain, -174 + 10 * math.Log10(15000))

			fout3.WriteString(fmt.Sprintf("%v,%v,%v\n", symb, ptsRe[i].X, ptsRe[i].Y))
		}

		for i := range ptsPrb {
			ptsPrb[i].X = float64(i)
			ptsPrb[i].Y = 10 * math.Log10(rssiPrb[symb][i]) - p.gain

			fout4.WriteString(fmt.Sprintf("%v,%v,%v\n", symb, ptsPrb[i].X, ptsPrb[i].Y))
		}

		const rows, cols = 2, 1
		plots := make([][]*plot.Plot, rows)
		for j := 0; j < rows; j++ {
			plots[j] = make([]*plot.Plot, cols)
			for i := 0; i < cols; i++ {
				pl := plot.New()
				pl.Add(plotter.NewGrid())
				pl.Title.Text = fmt.Sprintf("RSSI (%v)", symb)
				pl.Y.Label.Text = "RSSI(dBm)"
				pl.Y.Min = -140
				pl.Y.Max = -60
				pl.Legend.Top = true

				if i == 0 && j == 0 {
					pl.X.Label.Text = "FFT Bin"
					pl.X.Min = 0
					pl.X.Max = float64(len(rssiRe[symb]) - 1)
					plotutil.AddLines(pl, "RSSI_per_RE", ptsRe)
				}

				if i == 0 && j == 1 {
					pl.X.Label.Text = "PRB"
					pl.X.Min = 0
					pl.X.Max = float64(len(rssiPrb[symb]) - 1)
					plotutil.AddLines(pl, "RSSI_per_PRB", ptsPrb)
				}

				plots[j][i] = pl
			}
		}

		img := vgimg.New(8*vg.Inch, 8*vg.Inch)
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

		w, err := os.Create(path.Join(outPath, fmt.Sprintf("rssi_%v.png", symb)))
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, err.Error())
		}
		defer w.Close()

		png := vgimg.PngCanvas{Canvas: img}
		if _, err := png.WriteTo(w); err != nil {
			p.writeLog(zapcore.ErrorLevel, err.Error())
		}
	}
	fout3.Close()
	fout4.Close()
}

func (p *Ddr4TraceParser) parse(fn string) {
	// 1st step: parse DDR4(.bin) using Loki snapshot_tool
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing DDR4 trace using Loki snapshot_tool... [%s]", fn))
	outPath := path.Join(p.ddr4TracePath, "parsed_ddr4trace")
	decodedDir := path.Join(outPath, strings.Split(fn, ".")[0]+"_result")

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

	// 2nd step: parsing ant_x.txt where x=0..nbrAntennaPorts-1
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
			p.iqData.SetIfAbsent(key, make([]complex128, 0))

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

				m, _ := p.iqData.Get(key)
				m = append(m.([]complex128), complex(i/float64(LOKI_IQ_NORM_FACTOR), q/float64(LOKI_IQ_NORM_FACTOR)))
				p.iqData.Set(key, m)
			}
		}(file)
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
