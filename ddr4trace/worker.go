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
	VG_IMG_WIDTH int = 6
	VG_IMG_HEIGHT int = 3
)

type Ddr4TraceParser struct {
	log           *zap.Logger
	py3Path       string
	snapToolPath  string
	ddr4TracePath string
	pattern       string
	scs           string
	chbw          string
	filter        string
	maxgo         int
	gain          float64
	debug         bool

	iqData    cmap.ConcurrentMap

	// key1 = ant, val1 = [key2 = symbol, val2 = list of FFT bins]]
	rssiData  cmap.ConcurrentMap
	// key1 = ant, val1 = [key2 = symbol, val2 = count of a specific symbol]]
	rssiCount cmap.ConcurrentMap
}

func (p *Ddr4TraceParser) Init(log *zap.Logger, py3, snaptool, trace, pattern, scs, chbw, filter string, maxgo, gain int, debug bool) {
	p.log = log
	p.py3Path = py3
	p.snapToolPath = snaptool
	p.ddr4TracePath = trace
	p.pattern = pattern
	p.scs = strings.ToLower(scs)
	p.chbw = strings.ToLower(chbw)
	p.filter = filter
	if !(p.filter == "ul" || p.filter == "dl") {
		p.writeLog(zapcore.ErrorLevel, "The --filter option should be either ul or dl for DDR4 parser!")
		return
	}
	p.maxgo = utils.MaxInt([]int{2, maxgo})
	p.gain = float64(gain)
	p.debug = debug

	p.iqData = cmap.New()
	p.rssiData = cmap.New()
	p.rssiCount = cmap.New()

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing DDR4 trace parser...(working dir: %v)", trace))
}

func (p *Ddr4TraceParser) Exec() {
	// recreate dir for parsed ddr4 trace
	outPath := filepath.Join(p.ddr4TracePath, "parsed_ddr4trace")
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
			if !file.IsDir() && strings.HasPrefix(filepath.Ext(file.Name()),".bin") {
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

						// p.writeLog(zapcore.DebugLevel, fmt.Sprintf("%v,Frame%v,Slot%v,Symbol%v: len=%v", ant, irf, isl, isymb, len(amp)))

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
	fout1, err1 := os.OpenFile(filepath.Join(outPath, fmt.Sprintf("rssi_per_ant_symbol_re.csv")), os.O_WRONLY|os.O_CREATE, 0664)
	if err1 != nil {
		p.writeLog(zapcore.ErrorLevel, err1.Error())
		return
	}
	defer fout1.Close()
	fout1.WriteString("Antenna Port,Symbol,FFT Bin,RSSI(dBm)\n")

	fout2, err2 := os.OpenFile(filepath.Join(outPath, fmt.Sprintf("rssi_per_ant_symbol_prb.csv")), os.O_WRONLY|os.O_CREATE, 0664)
	if err2 != nil {
		p.writeLog(zapcore.ErrorLevel, err2.Error())
		return
	}
	defer fout2.Close()
	fout2.WriteString("Antenna Port,Symbol,PRB,RSSI(dBm)\n")

	// key = symbol, val = RSSI per RE or RSSI per PRB
	rssiSymbRe := make(map[string][]float64)
	rssiSymbPrb := make(map[string][]float64)
	ptsAntSymbRe := make(map[string]plotter.XYs)
	ptsAntSymbPrb := make(map[string]plotter.XYs)
	for _, ant := range p.rssiData.Keys() {
		m, _ := p.rssiData.Get(ant)
		for _, symb := range m.(cmap.ConcurrentMap).Keys() {
			m2, _ := m.(cmap.ConcurrentMap).Get(symb)

			if _, e := rssiSymbRe[symb]; !e {
				rssiSymbRe[symb] = make([]float64, len(m2.([]float64)))
				rssiSymbPrb[symb] = make([]float64, nbrPrb)
			}

			key := ant + "_" + symb
			ptsAntSymbRe[key] = make(plotter.XYs, len(m2.([]float64)))
			ptsAntSymbPrb[key] = make(plotter.XYs, nbrPrb)
			for i := range ptsAntSymbRe[key] {
				ptsAntSymbRe[key][i].X = float64(i)
				// scs, _ := strconv.ParseFloat(strings.TrimSuffix(p.scs, "k"), 64)
				// ptsAntSymbRe[i].Y = math.Max(10 * math.Log10(m2.([]float64)[i]) - p.gain, -174 + 10 * math.Log10(scs * 1000))
				ptsAntSymbRe[key][i].Y = 10*math.Log10(m2.([]float64)[i]) - p.gain
				rssiSymbRe[symb][i] += m2.([]float64)[i]

				fout1.WriteString(fmt.Sprintf("%v,%v,%v,%v\n", ant, symb, ptsAntSymbRe[key][i].X, ptsAntSymbRe[key][i].Y))

				if i >= firstRe && i < (firstRe+nbrRe) {
					iprb := math.Floor(float64(i-firstRe) / 12)
					j := int(iprb)
					ptsAntSymbPrb[key][j].Y += m2.([]float64)[i]
					rssiSymbPrb[symb][j] += m2.([]float64)[i]
				}
			}

			for i := range ptsAntSymbPrb[key] {
				ptsAntSymbPrb[key][i].X = float64(i)
				// scs, _ := strconv.ParseFloat(strings.TrimSuffix(p.scs, "k"), 64)
				// ptsAntSymbPrb[i].Y = math.Max(10 * math.Log10(ptsAntSymbPrb[i].Y) - p.gain, -174 + 10 * math.Log10(scs * 12 * 1000))
				ptsAntSymbPrb[key][i].Y = 10*math.Log10(ptsAntSymbPrb[key][i].Y) - p.gain

				fout2.WriteString(fmt.Sprintf("%v,%v,%v,%v\n", ant, symb, ptsAntSymbPrb[key][i].X, ptsAntSymbPrb[key][i].Y))
			}
		}
	}

	// save per ant/symb RSSI as .png using gonum/plot
	//rows, cols := len(ptsAntSymbRe) / len(p.rssiData.Keys()), 2 * len(p.rssiData.Keys())
	rows, cols := 2*len(p.rssiData.Keys()), len(ptsAntSymbRe)/len(p.rssiData.Keys())
	plots := make([][]*plot.Plot, rows)
	for i := 0; i < rows; i++ {
		ant := fmt.Sprintf("ant%v", i/2)

		plots[i] = make([]*plot.Plot, cols)
		for j := 0; j < cols; j++ {
			symb := fmt.Sprintf("symbol%v", j)
			key := ant + "_" + symb

			pl := plot.New()
			pl.Add(plotter.NewGrid())
			pl.Title.Text = fmt.Sprintf("RSSI(%v-%v)", ant, symb)
			pl.Y.Label.Text = "RSSI(dBm)"
			pl.Y.Min = -140
			pl.Y.Max = -60
			pl.Legend.Top = true

			if i%2 == 0 {
				pl.X.Label.Text = "FFT Bin"
				pl.X.Min = 0
				pl.X.Max = float64(len(ptsAntSymbRe[key]) - 1)
				plotutil.AddLines(pl, "RSSI_per_RE", ptsAntSymbRe[key])
			} else if i%2 == 1 {
				pl.X.Label.Text = "PRB"
				pl.X.Min = 0
				pl.X.Max = float64(nbrPrb - 1)
				plotutil.AddLines(pl, "RSSI_per_PRB", ptsAntSymbPrb[key])
			}

			plots[i][j] = pl
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

	w, err := os.Create(filepath.Join(outPath, "rssi_per_ant_per_symb.png"))
	if err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
	}
	defer w.Close()

	png := vgimg.PngCanvas{Canvas: img}
	if _, err := png.WriteTo(w); err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
	}

	// save per symb RSSI as .png using gonum/plot and as rssi_per_symbol_x.csv
	fout3, err3 := os.OpenFile(filepath.Join(outPath, fmt.Sprintf("rssi_per_symbol_re.csv")), os.O_WRONLY|os.O_CREATE, 0664)
	if err3 != nil {
		p.writeLog(zapcore.ErrorLevel, err3.Error())
		return
	}
	defer fout3.Close()
	fout3.WriteString("Symbol,FFT Bin,RSSI(dBm)\n")

	fout4, err4 := os.OpenFile(filepath.Join(outPath, fmt.Sprintf("rssi_per_symbol_prb.csv")), os.O_WRONLY|os.O_CREATE, 0664)
	if err4 != nil {
		p.writeLog(zapcore.ErrorLevel, err4.Error())
		return
	}
	defer fout4.Close()
	fout4.WriteString("Symbol,PRB,RSSI(dBm)\n")

	ptsSymbRe := make(map[string]plotter.XYs)
	ptsSymbPrb := make(map[string]plotter.XYs)
	for symb := range rssiSymbRe {
		ptsSymbRe[symb] = make(plotter.XYs, len(rssiSymbRe[symb]))
		ptsSymbPrb[symb] = make(plotter.XYs, len(rssiSymbPrb[symb]))
		for i := range ptsSymbRe[symb] {
			ptsSymbRe[symb][i].X = float64(i)
			// scs, _ := strconv.ParseFloat(strings.TrimSuffix(p.scs, "k"), 64)
			// ptsSymbRe[i].Y = math.Max(10 * math.Log10(rssiSymbRe[symb][i]) - p.gain, -174 + 10 * math.Log10(scs * 1000))
			ptsSymbRe[symb][i].Y = 10*math.Log10(rssiSymbRe[symb][i]) - p.gain

			fout3.WriteString(fmt.Sprintf("%v,%v,%v\n", symb, ptsSymbRe[symb][i].X, ptsSymbRe[symb][i].Y))
		}

		for i := range ptsSymbPrb[symb] {
			ptsSymbPrb[symb][i].X = float64(i)
			// scs, _ := strconv.ParseFloat(strings.TrimSuffix(p.scs, "k"), 64)
			// ptsSymbPrb[i].Y = math.Max(10 * math.Log10(rssiSymbPrb[symb][i]) - p.gain, -174 + 10 * math.Log10(scs * 12 * 1000))
			ptsSymbPrb[symb][i].Y = 10*math.Log10(rssiSymbPrb[symb][i]) - p.gain

			fout4.WriteString(fmt.Sprintf("%v,%v,%v\n", symb, ptsSymbPrb[symb][i].X, ptsSymbPrb[symb][i].Y))
		}
	}

	// rows, cols := len(ptsSymbRe), 2
	rows, cols = 2, len(ptsSymbRe)
	plots = make([][]*plot.Plot, rows)
	for i := 0; i < rows; i++ {
		plots[i] = make([]*plot.Plot, cols)
		for j := 0; j < cols; j++ {
			symb := fmt.Sprintf("symbol%v", j)

			pl := plot.New()
			pl.Add(plotter.NewGrid())
			pl.Title.Text = fmt.Sprintf("RSSI(symbol%v)", j)
			pl.Y.Label.Text = "RSSI(dBm)"
			pl.Y.Min = -140
			pl.Y.Max = -60
			pl.Legend.Top = true

			if i%2 == 0 {
				pl.X.Label.Text = "FFT Bin"
				pl.X.Min = 0
				pl.X.Max = float64(len(rssiSymbRe[symb]) - 1)
				plotutil.AddLines(pl, "RSSI_per_RE", ptsSymbRe[symb])
			} else if i%2 == 1 {
				pl.X.Label.Text = "PRB"
				pl.X.Min = 0
				pl.X.Max = float64(len(rssiSymbPrb[symb]) - 1)
				plotutil.AddLines(pl, "RSSI_per_PRB", ptsSymbPrb[symb])
			}

			plots[i][j] = pl
		}
	}

	width, _ = vg.ParseLength(fmt.Sprintf("%vin", cols*VG_IMG_WIDTH))
	height, _ = vg.ParseLength(fmt.Sprintf("%vin", rows*VG_IMG_HEIGHT))
	img = vgimg.New(width, height)
	dc = draw.New(img)
	t = draw.Tiles{
		Rows:      rows,
		Cols:      cols,
		PadX:      vg.Millimeter,
		PadY:      vg.Millimeter,
		PadTop:    vg.Points(2),
		PadBottom: vg.Points(2),
		PadLeft:   vg.Points(2),
		PadRight:  vg.Points(2),
	}
	canvases = plot.Align(plots, t, dc)
	for j := 0; j < rows; j++ {
		for i := 0; i < cols; i++ {
			if plots[j][i] != nil {
				plots[j][i].Draw(canvases[j][i])
			}
		}
	}

	w, err = os.Create(filepath.Join(outPath, "rssi_per_symbol.png"))
	if err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
	}
	defer w.Close()

	png = vgimg.PngCanvas{Canvas: img}
	if _, err := png.WriteTo(w); err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
	}

	// RSSI per PRB by averaging per-Symbol RSSI
	var fnSuffix string
	if p.filter == "dl" {
		fnSuffix = "_unreliable"
	}
	fout5, err5 := os.OpenFile(filepath.Join(outPath, fmt.Sprintf("rssi_per_prb%v.csv", fnSuffix)), os.O_WRONLY|os.O_CREATE, 0664)
	if err5 != nil {
		p.writeLog(zapcore.ErrorLevel, err5.Error())
		return
	}
	defer fout5.Close()
	fout5.WriteString("PRB,RSSI(dBm)\n")

	ptsPrb := make(plotter.XYs, nbrPrb)
	for i := range ptsPrb {
		ptsPrb[i].X = float64(i)

		for symb := range rssiSymbPrb {
			ptsPrb[i].Y += rssiSymbPrb[symb][i]
		}

		ptsPrb[i].Y = 10*math.Log10(ptsPrb[i].Y/float64(len(rssiSymbPrb))) - p.gain

		fout5.WriteString(fmt.Sprintf("%v,%v\n", ptsPrb[i].X, ptsPrb[i].Y))
	}

	pl := plot.New()
	pl.Add(plotter.NewGrid())
	pl.Title.Text = "RSSI"
	pl.X.Label.Text = "PRB"
	pl.X.Min = 0
	pl.X.Max = float64(nbrPrb - 1)
	pl.Y.Label.Text = "RSSI(dBm)"
	pl.Y.Min = -140
	pl.Y.Max = -60
	pl.Legend.Top = true
	plotutil.AddLines(pl, "RSSI_per_PRB", ptsPrb)
	width, _ = vg.ParseLength(fmt.Sprintf("%vin", VG_IMG_WIDTH))
	height, _ = vg.ParseLength(fmt.Sprintf("%vin", VG_IMG_HEIGHT))
	if err := pl.Save(width, height, filepath.Join(outPath, fmt.Sprintf("rssi_prb%v.png", fnSuffix))); err != nil {
		p.writeLog(zapcore.ErrorLevel, err.Error())
	}
}

func (p *Ddr4TraceParser) parse(fn string) {
	// 1st step: parse DDR4(.bin) using Loki snapshot_tool
	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing DDR4 trace using Loki snapshot_tool... [%s]", fn))
	outPath := filepath.Join(p.ddr4TracePath, "parsed_ddr4trace")
	decodedDir := filepath.Join(outPath, strings.Split(fn, ".")[0]+"_result")

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command(filepath.Join(p.py3Path, BIN_PYTHON3_LINUX), filepath.Join(p.snapToolPath, "SnapshotAnalyzer.py"), "--iqdata", filepath.Join(p.ddr4TracePath, fn), "--output", decodedDir)
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command(filepath.Join(p.py3Path, BIN_PYTHON3), filepath.Join(p.snapToolPath, "SnapshotAnalyzer.py"), "--iqdata", filepath.Join(p.ddr4TracePath, fn), "--output", decodedDir)
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
	var subcells []string
	if p.filter == "ul" {
		subcells, _ = filepath.Glob(filepath.Join(filepath.Join(decodedDir, "nr", "ul"), "subcellid*"))
	} else if p.filter == "dl" {
		subcells, _ = filepath.Glob(filepath.Join(filepath.Join(decodedDir, "nr", "dl"), "subcellid*", "*"))
	} else {
		p.writeLog(zapcore.ErrorLevel, "The --filter option should be either ul or dl for DDR4 parser!")
		return
	}

	for _, sc := range subcells {
		ants, _ := filepath.Glob(filepath.Join(sc, "ant*.txt"))
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
			key := strings.Replace(strings.Replace(strings.Replace(fn, filepath.Join(p.ddr4TracePath, "parsed_ddr4trace"), "ddr4", -1), "/", "_", -1), "\\", "_", -1)
			key = fmt.Sprintf("%v%v", strings.Replace(key[:len(key)-len(filepath.Ext(key))], ".", "_", -1), filepath.Ext(key))
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
