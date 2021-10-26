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
	"fmt"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"math"
	"strconv"
	"strings"
)

type CmPdcch struct {
	log *zap.Logger
	scs string
	bwpid int
	coreset []string
	css []string
	uss string
	debug bool
}

type Coreset struct {
	size int
	duration int
}

type SearchSpace struct {
	monitoringSymbs []int
	mapCandidates map[int]int
	period int
	coreset string
}

func (p *CmPdcch) Init(log *zap.Logger, scs string, bwpid int, coreset, css []string, uss string, debug bool) {
	p.log = log
	p.debug = debug
	p.scs = scs
	p.bwpid = bwpid
	p.coreset = coreset
	p.css = css
	p.uss = uss
}

func (p *CmPdcch) Exec() {
	mapCoreset := make(map[string]Coreset)
	mapSearchSpace := make(map[string]SearchSpace)

	for _, k := range p.coreset {
		toks := strings.Split(k, "_")
		if len(toks) != 3 {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Invalid CORESET settings: %v. Format should be: coresetId_size_duration.", k))
			return
		}
		mapCoreset[toks[0]] = Coreset {
			size: p.unsafeAtoi(toks[1]),
			duration: p.unsafeAtoi(toks[2]),
		}
	}

	ss := append(p.css, p.uss)
	for _, k := range ss {
		toks := strings.Split(k, "_")
		if len(toks) != 9 {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Invalid SearchSpace settings: %v. Format should be: searchSpaceType_monitoringSymbolWithinSlot_pdcchCandidatesAL1_pdcchCandidatesAL2_pdcchCandidatesAL4_pdcchCandidatesAL8_pdcchCandidatesAL16_periodicity_coresetId.", k))
			return
		}
		monitoringSymbolWithinSlot := toks[1]
		if len(monitoringSymbolWithinSlot) != 3 {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Invalid monitoringSymbolWithinSlot: %v. Format should be: 110", k))
			return
		}
		monitoringSymbs := make([]int, 0)
		for i := range monitoringSymbolWithinSlot {
			if monitoringSymbolWithinSlot[i] == '1' {
				monitoringSymbs = append(monitoringSymbs, i)
			}
		}

		mapSearchSpace[strings.ToLower(toks[0])] = SearchSpace {
			monitoringSymbs: monitoringSymbs,
			mapCandidates: map[int]int {
				1 : p.unsafeAtoi(toks[2][1:]),
				2 : p.unsafeAtoi(toks[3][1:]),
				4 : p.unsafeAtoi(toks[4][1:]),
				8 : p.unsafeAtoi(toks[5][1:]),
				16 : p.unsafeAtoi(toks[6][1:]),
			},
			period : p.unsafeAtoi(toks[7][2:]),
			coreset : toks[8],
		}
	}

	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("mapCoreset=%v\nmapSearchSpace=%v", mapCoreset, mapSearchSpace))

	// Validate against Table 10.1-2: Maximum number of monitored PDCCH candidates per slot for a DL BWP
	mapScs2MaxCandidatesPerSlot := map[string]int { "15k" : 44, "30k" : 36, "60k" : 22, "120k" : 20}

	cssMonitoringSymbs := utils.MaxInt([]int{len(mapSearchSpace["type0a"].monitoringSymbs), len(mapSearchSpace["type1"].monitoringSymbs), len(mapSearchSpace["type2"].monitoringSymbs), len(mapSearchSpace["type3"].monitoringSymbs)})
	cssCandiatesPerSymb := 0
	ussCandidatesPerSymb := 0
	for _, al := range []int{1,2,4,8,16} {
		cssCandiatesPerSymb += utils.MaxInt([]int{mapSearchSpace["type0a"].mapCandidates[al], mapSearchSpace["type1"].mapCandidates[al], mapSearchSpace["type2"].mapCandidates[al], mapSearchSpace["type3"].mapCandidates[al]})
		ussCandidatesPerSymb += mapSearchSpace["uss"].mapCandidates[al]
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Max number of monitored PDCCH candidates per slot is %v when scs = %v", mapScs2MaxCandidatesPerSlot[p.scs], p.scs))
	totCssCandidatesPerSlot := cssCandiatesPerSymb * cssMonitoringSymbs + ussCandidatesPerSymb * len(mapSearchSpace["uss"].monitoringSymbs) * 2
	if totCssCandidatesPerSlot > mapScs2MaxCandidatesPerSlot[p.scs] {
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Max number of monitored PDCCH candidates validation FAILED: cssCandiatesPerSymb = %v, cssMonitoringSymbs = %v, ussCandidatesPerSymb = %v, ussMonitoringSymbs = %v and totCandidatesPerSlot = %v", cssCandiatesPerSymb, cssMonitoringSymbs, ussCandidatesPerSymb, len(mapSearchSpace["uss"].monitoringSymbs), totCssCandidatesPerSlot))
	} else {
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Max number of monitored PDCCH candidates validation PASSED: cssCandiatesPerSymb = %v, cssMonitoringSymbs = %v, ussCandidatesPerSymb = %v, ussMonitoringSymbs = %v and totCandidatesPerSlot = %v", cssCandiatesPerSymb, cssMonitoringSymbs, ussCandidatesPerSymb, len(mapSearchSpace["uss"].monitoringSymbs), totCssCandidatesPerSlot))
	}

	// Validate against Table 10.1-3: Maximum number of non-overlapped CCEs per slot for a DL BWP
	mapScs2MaxNonOverlapCcesPerSlot := map[string]int { "15k" : 56, "30k" : 56, "60k" : 48, "120k" : 32}
	mapScs2SlotsPerRf:= map[string]int { "15k" : 10, "30k" : 20, "60k" : 40, "120k" : 80}
	// key = coresetId_monitoringSymbol, val = per CCE flag (1=used,0=not used)
	mapNonOverlapCces := make(map[string]map[int][]int)
	for coresetId, coreset := range mapCoreset {
		N_CCE := coreset.size / 6
		mapNonOverlapCces[coresetId] = make(map[int][]int)
		for i := 0; i < 3; i++ {
			mapNonOverlapCces[coresetId][i] = make([]int, N_CCE)
		}

		for sstype, ss := range mapSearchSpace {
			if ss.coreset != coresetId {
				continue
			}

			for _, al := range []int{1,2,4,8,16} {
				L := al
				M := ss.mapCandidates[al]
				if M > 0 {
					if sstype != "uss" {
						Y := 0
						for m := 0; m < M; m++ {
							startCce := p.CalcStartCceIndex(float64(N_CCE), float64(L), float64(M), float64(m), float64(Y), -1)
							for _, isymb := range ss.monitoringSymbs {
								for ial := 0; ial < L; ial++ {
									mapNonOverlapCces[coresetId][isymb][startCce+ial] = 1
								}
							}
						}
					} else {
						p_ := p.unsafeAtoi(coresetId[len("CORESETT"):])
						var Ap int
						switch p_ % 3 {
						case 0:
							Ap = 39827
						case 1:
							Ap = 39829
						case 2:
							Ap = 39839
						}

						Y := make([]int, mapScs2SlotsPerRf[p.scs])
						D := 65537
						for ns := 0; ns < mapScs2SlotsPerRf[p.scs]; ns++ {
							if ns == 0 {
								Y[ns] = (Ap * 100) % D // assume RNTI=100 by default
							} else {
								Y[ns] = (Ap * Y[ns-1]) % D
							}
							for m := 0; m < M; m++ {
								startCce := p.CalcStartCceIndex(float64(N_CCE), float64(L), float64(M), float64(m), float64(Y[ns]), ns)
								for _, isymb := range ss.monitoringSymbs {
									for ial := 0; ial < L; ial++ {
										mapNonOverlapCces[coresetId][isymb][startCce+ial] = 1
									}
								}
							}
						}
					}
				}
			}
		}
	}

	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("mapNonOverlapCces = %v", mapNonOverlapCces))

	mapNonOverlapCcesPerCoreset := make(map[string][]string)
	totNonOverlapCcesPerSlot := 0
	for coreset := range mapNonOverlapCces {
		mapNonOverlapCcesPerCoreset[coreset] = make([]string, 0)
		for symb := range mapNonOverlapCces[coreset] {
			totNonOverlapCcesPerSlot += utils.SumInt(mapNonOverlapCces[coreset][symb])
			mapNonOverlapCcesPerCoreset[coreset] = append(mapNonOverlapCcesPerCoreset[coreset], fmt.Sprintf("symbol%v_%vCCEs", symb, utils.SumInt(mapNonOverlapCces[coreset][symb])))
		}
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Max number of non-overlapped CCEs per slot is %v when scs = %v", mapScs2MaxNonOverlapCcesPerSlot[p.scs], p.scs))
	if totNonOverlapCcesPerSlot > mapScs2MaxNonOverlapCcesPerSlot[p.scs] {
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Max number of non-overlapped CCEs validation FAILED: mapNonOverlapCcesPerCoreset = %v and totNonOverlapCcesPerSlot = %v", mapNonOverlapCcesPerCoreset, totNonOverlapCcesPerSlot))
	} else {
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Max number of non-overlapped CCEs validation PASSED: mapNonOverlapCcesPerCoreset = %v and totNonOverlapCcesPerSlot = %v", mapNonOverlapCcesPerCoreset, totNonOverlapCcesPerSlot))
	}
}

func (p *CmPdcch) CalcStartCceIndex(N_CCE, L, M, m, Y float64, ns int) int {
	startCce := int(L * math.Mod(Y + math.Floor(m * N_CCE / (L * M)), math.Floor(N_CCE / L)))
	if ns < 0 {
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("ns=all, N_CCE=%v,L=%v,M=%v,m=%v,Y=%v -> startCce=%v", N_CCE, L, M, m, Y, startCce))
	} else {
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("ns=%v, N_CCE=%v,L=%v,M=%v,m=%v,Y=%v -> startCce=%v", ns, N_CCE, L, M, m, Y, startCce))
	}

	return startCce
}

func (p *CmPdcch) unsafeAtoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func (p *CmPdcch) writeLog(level zapcore.Level, s string) {
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
