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
	uss []string
	rnti int
	debug bool
}

type Coreset struct {
	id int
	size int
	duration int
}

type SearchSpace struct {
	id              int
	searchSpaceType string
	monitoringSymbs []int
	mapCandidates   map[int]int
	period          int
	coreset         string
}

func (p *CmPdcch) Init(log *zap.Logger, scs string, bwpid int, coreset, css, uss []string, rnti int, debug bool) {
	p.log = log
	p.debug = debug
	p.scs = scs
	p.bwpid = bwpid
	p.coreset = coreset
	p.css = css
	p.uss = uss
	p.rnti = rnti
}

func (p *CmPdcch) Exec() {
	mapCoreset := make(map[string]Coreset) //key=coresetId, val=Coreset struct
	mapSearchSpace := make(map[int]SearchSpace) //key=id, val=SearchSpace struct

	for _, k := range p.coreset {
		toks := strings.Split(k, "_")
		if len(toks) != 3 {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Invalid CORESET settings: %v. Format should be: coresetId_size_duration.", k))
			return
		}
		mapCoreset[toks[0]] = Coreset {
			id: p.unsafeAtoi(toks[0][len("CORESET"):]),
			size: p.unsafeAtoi(toks[1]),
			duration: p.unsafeAtoi(toks[2]),
		}
	}

	ss := append(p.css, p.uss...)
	for _, k := range ss {
		toks := strings.Split(k, "_")
		if len(toks) != 10 {
			p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Invalid SearchSpace settings: %v. Format should be: searchSpaceType_searchSpaceId_monitoringSymbolWithinSlot_pdcchCandidatesAL1_pdcchCandidatesAL2_pdcchCandidatesAL4_pdcchCandidatesAL8_pdcchCandidatesAL16_periodicity_coresetId.", k))
			return
		}
		searchSpaceId := p.unsafeAtoi(toks[1])
		monitoringSymbolWithinSlot := toks[2]
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

		//mapSearchSpace[strings.ToLower(toks[0])] = SearchSpace {
		mapSearchSpace[searchSpaceId] = SearchSpace {
			id:              searchSpaceId,
			searchSpaceType: toks[0],
			monitoringSymbs: monitoringSymbs,
			mapCandidates: map[int]int {
				1 : p.unsafeAtoi(toks[3][1:]),
				2 : p.unsafeAtoi(toks[4][1:]),
				4 : p.unsafeAtoi(toks[5][1:]),
				8 : p.unsafeAtoi(toks[6][1:]),
				16 : p.unsafeAtoi(toks[7][1:]),
			},
			period : p.unsafeAtoi(toks[8][2:]),
			coreset : toks[9],
		}

		/*
		if _, e := mapSearchSpace[strings.ToLower(toks[0])]; !e {
			mapSearchSpace[strings.ToLower(toks[0])] = []SearchSpace{sss}
		} else {
			mapSearchSpace[strings.ToLower(toks[0])] = append(mapSearchSpace[strings.ToLower(toks[0])], sss)
		}
		 */
	}

	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("mapCoreset=%v\nmapSearchSpace=%v", mapCoreset, mapSearchSpace))

	// Validate against Table 10.1-2: Maximum number of monitored PDCCH candidates per slot for a DL BWP
	mapScs2MaxCandidatesPerSlot := map[string]int { "15k" : 44, "30k" : 36, "60k" : 22, "120k" : 20}
	cssCandidatesPerSymb := make(map[string]map[int]int) //key=corestId, key2=monitoringSymbol, val=count
	ussCandidatesPerSymb := make(map[string]map[int]int) //key=corestId, key2=monitoringSymbol, val=count
	for coresetId := range mapCoreset {
		cssCandidatesPerSymb[coresetId] = make(map[int]int)
		ussCandidatesPerSymb[coresetId] = make(map[int]int)
		for i := 0; i < 3; i++ {
			cssCandidatesPerSymb[coresetId][i] = 0
			ussCandidatesPerSymb[coresetId][i] = 0
		}

		for _, ss := range mapSearchSpace {
			if ss.coreset != coresetId {
				continue
			}

			if ss.searchSpaceType == "uss" {
				for _, i  := range ss.monitoringSymbs {
					for _, al := range []int{1,2,4,8,16} {
						ussCandidatesPerSymb[coresetId][i] += ss.mapCandidates[al]
					}
				}
			} else {
				for _, i := range ss.monitoringSymbs {
					totCandiates := 0
					for _, al := range []int{1,2,4,8,16} {
						totCandiates += ss.mapCandidates[al]
					}
					cssCandidatesPerSymb[coresetId][i] = utils.MaxInt([]int{cssCandidatesPerSymb[coresetId][i], totCandiates})
				}
			}
		}
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Max number of monitored PDCCH candidates per slot is %v when scs = %v", mapScs2MaxCandidatesPerSlot[p.scs], p.scs))
	totCandidatesPerCoreset := make(map[string]int) //key=coresetId, val=count
	totCandidates := 0
	for coresetId := range mapCoreset {
		totCandidatesPerCoreset[coresetId] = 0
		for i := 0; i < 3; i++ {
			totCandidatesPerCoreset[coresetId] += (cssCandidatesPerSymb[coresetId][i] + ussCandidatesPerSymb[coresetId][i] * 2)
		}
		totCandidates += totCandidatesPerCoreset[coresetId]
	}
	//totCssCandidatesPerSlot := cssCandidatesPerSymb * cssMonitoringSymbs + ussCandidatesPerSymb * len(mapSearchSpace["uss"].monitoringSymbs) * 2
	p.writeLog(zapcore.InfoLevel, "By YangYang's method:")
	if totCandidates > mapScs2MaxCandidatesPerSlot[p.scs] {
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("-Max number of monitored PDCCH candidates validation FAILED: cssCandidatesPerSymb = %v, ussCandidatesPerSymb = %v, totCandidatesPerCoreset = %v and totCandidatesPerSlot = %v", cssCandidatesPerSymb, ussCandidatesPerSymb, totCandidatesPerCoreset, totCandidates))
	} else {
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("-Max number of monitored PDCCH candidates validation PASSED: cssCandidatesPerSymb = %v, ussCandidatesPerSymb = %v, totCandidatesPerCoreset = %v and totCandidatesPerSlot = %v", cssCandidatesPerSymb, ussCandidatesPerSymb, totCandidatesPerCoreset, totCandidates))
	}

	// Validate against Table 10.1-3: Maximum number of non-overlapped CCEs per slot for a DL BWP
	mapScs2MaxNonOverlapCcesPerSlot := map[string]int { "15k" : 56, "30k" : 56, "60k" : 48, "120k" : 32}
	mapScs2SlotsPerRf:= map[string]int { "15k" : 10, "30k" : 20, "60k" : 40, "120k" : 80}
	// key = slot_coresetId_monitoringSymbol, val = per CCE flag (1=used,0=not used)
	mapNonOverlapCces := make(map[int]map[string]map[int][]int)
	// key = slot_coresetId_searchSpaceId_monitoringSymbol_al, val = startCce
	mapStartCce := make(map[int]map[string]map[int]map[int]map[int][]int)
	Y := make([]int, mapScs2SlotsPerRf[p.scs])
	for ns := 0; ns < mapScs2SlotsPerRf[p.scs]; ns++ {
		mapNonOverlapCces[ns] = make(map[string]map[int][]int)
		mapStartCce[ns] = make(map[string]map[int]map[int]map[int][]int)

		for coresetId, coreset := range mapCoreset {
			N_CCE := coreset.size / 6
			mapNonOverlapCces[ns][coresetId] = make(map[int][]int)
			mapStartCce[ns][coresetId] = make(map[int]map[int]map[int][]int)
			for i := 0; i < 3; i++ {
				mapNonOverlapCces[ns][coresetId][i] = make([]int, N_CCE)
			}

			for _, ss := range mapSearchSpace {
				if ss.coreset != coresetId {
					continue
				}

				mapStartCce[ns][coresetId][ss.id] = make(map[int]map[int][]int)
				for i := 0; i < 3; i++ {
					mapStartCce[ns][coresetId][ss.id][i] = make(map[int][]int)
					for _, al := range []int{1, 2, 4, 8, 16} {
						mapStartCce[ns][coresetId][ss.id][i][al] = make([]int, 0)
					}
				}

				for _, al := range []int{1, 2, 4, 8, 16} {
					L := al
					M := ss.mapCandidates[al]
					if M > 0 {
						if ss.searchSpaceType != "uss" {
							Y := 0
							for m := 0; m < M; m++ {
								startCce := p.CalcStartCceIndex(ss.searchSpaceType, float64(N_CCE), float64(L), float64(M), float64(m), float64(Y), ns)
								for _, isymb := range ss.monitoringSymbs {
									for ial := 0; ial < L; ial++ {
										mapNonOverlapCces[ns][coresetId][isymb][startCce+ial] = 1
									}
									if !utils.ContainsInt(mapStartCce[ns][coresetId][ss.id][isymb][al], startCce) {
										mapStartCce[ns][coresetId][ss.id][isymb][al] = append(mapStartCce[ns][coresetId][ss.id][isymb][al], startCce)
									}
								}
							}
						} else {
							var Ap int
							switch coreset.id % 3 {
							case 0:
								Ap = 39827
							case 1:
								Ap = 39829
							case 2:
								Ap = 39839
							}

							D := 65537
							if ns == 0 {
								Y[ns] = (Ap * p.rnti) % D
							} else {
								Y[ns] = (Ap * Y[ns-1]) % D
							}
							for m := 0; m < M; m++ {
								startCce := p.CalcStartCceIndex(ss.searchSpaceType, float64(N_CCE), float64(L), float64(M), float64(m), float64(Y[ns]), ns)
								for _, isymb := range ss.monitoringSymbs {
									for ial := 0; ial < L; ial++ {
										mapNonOverlapCces[ns][coresetId][isymb][startCce+ial] = 1
									}
									if !utils.ContainsInt(mapStartCce[ns][coresetId][ss.id][isymb][al], startCce) {
										mapStartCce[ns][coresetId][ss.id][isymb][al] = append(mapStartCce[ns][coresetId][ss.id][isymb][al], startCce)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Validate against Table 10.1-2: Maximum number of monitored PDCCH candidates per slot for a DL BWP
	// key = slot_coresetId_searchSpaceId_monitoringSymbol, val = count
	totCandidatesPerSymb := make(map[int]map[string]map[int]map[int]int)
	totCandidatesPerSlot := make(map[int]int)
	for ns := 0; ns < mapScs2SlotsPerRf[p.scs]; ns++ {
		totCandidatesPerSymb[ns] = make(map[string]map[int]map[int]int)
		totCandidatesPerSlot[ns] = 0
		for coresetId := range mapStartCce[ns] {
			totCandidatesPerSymb[ns][coresetId] = make(map[int]map[int]int)
			for searchSpaceId := range mapStartCce[ns][coresetId] {
				totCandidatesPerSymb[ns][coresetId][searchSpaceId] = make(map[int]int)
				for isymb := range mapStartCce[ns][coresetId][searchSpaceId] {
					totCandidatesPerSymb[ns][coresetId][searchSpaceId][isymb] = 0
					for al := range mapStartCce[ns][coresetId][searchSpaceId][isymb] {
						//p.writeLog(zapcore.DebugLevel, fmt.Sprintf("ns=%v,coresetId=%v,searchSpaceId=%v,isymb=%v,al=AL%v,startCce=%v", ns, coresetId, searchSpaceId, isymb, al, mapStartCce[ns][coresetId][searchSpaceId][isymb][al]))
						if mapSearchSpace[searchSpaceId].searchSpaceType == "uss" {
							totCandidatesPerSymb[ns][coresetId][searchSpaceId][isymb] += 2 * len(mapStartCce[ns][coresetId][searchSpaceId][isymb][al])
						} else {
							totCandidatesPerSymb[ns][coresetId][searchSpaceId][isymb] += len(mapStartCce[ns][coresetId][searchSpaceId][isymb][al])
						}
					}

					if totCandidatesPerSymb[ns][coresetId][searchSpaceId][isymb] > 0 {
						p.writeLog(zapcore.DebugLevel, fmt.Sprintf("ns=%v,coresetId=%v,searchSpaceId=%v,searchSpaceType=%v,monitoringSymbol=%v,totCandidatesPerSymb=%v", ns, coresetId, searchSpaceId, mapSearchSpace[searchSpaceId].searchSpaceType, isymb, totCandidatesPerSymb[ns][coresetId][searchSpaceId][isymb]))
					}
					totCandidatesPerSlot[ns] += totCandidatesPerSymb[ns][coresetId][searchSpaceId][isymb]
				}
			}
		}
	}

	p.writeLog(zapcore.InfoLevel, "By 3GPP 38.213's method:")
	for ns := 0; ns < mapScs2SlotsPerRf[p.scs]; ns++ {
		if totCandidatesPerSlot[ns] > mapScs2MaxCandidatesPerSlot[p.scs] {
			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("-Max number of monitored PDCCH candidates validation FAILED: ns=%v: totCandidatesPerSlot = %v", ns, totCandidatesPerSlot[ns]))
		} else {
			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("-Max number of monitored PDCCH candidates validation PASSED: ns=%v: totCandidatesPerSlot = %v", ns, totCandidatesPerSlot[ns]))
		}
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Max number of non-overlapped CCEs per slot is %v when scs = %v", mapScs2MaxNonOverlapCcesPerSlot[p.scs], p.scs))
	for ns := 0; ns < mapScs2SlotsPerRf[p.scs]; ns++ {
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("mapNonOverlapCces[ns=%v] = %v", ns, mapNonOverlapCces[ns]))

		numNonOverlapPerSlot := 0
		for coreset := range mapNonOverlapCces[ns] {
			for symb := range mapNonOverlapCces[ns][coreset] {
				numNonOverlapCces := utils.SumInt(mapNonOverlapCces[ns][coreset][symb])
				p.writeLog(zapcore.DebugLevel, fmt.Sprintf("ns=%v,coreset=%v,symbol=%v: numNonOverlapCces=%v", ns, coreset, symb, numNonOverlapCces))

				numNonOverlapPerSlot += numNonOverlapCces
			}
		}

		if numNonOverlapPerSlot > mapScs2MaxNonOverlapCcesPerSlot[p.scs] {
			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("-Max number of non-overlapped CCEs validation FAILED: ns=%v: totNonOverlapCcesPerSlot=%v", ns, numNonOverlapPerSlot))
		} else {
			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("-Max number of non-overlapped CCEs validation PASSED: ns=%v: totNonOverlapCcesPerSlot=%v", ns, numNonOverlapPerSlot))
		}
	}
}

func (p *CmPdcch) CalcStartCceIndex(sstype string, N_CCE, L, M, m, Y float64, ns int) int {
	startCce := int(L * math.Mod(Y + math.Floor(m * N_CCE / (L * M)), math.Floor(N_CCE / L)))
	p.writeLog(zapcore.DebugLevel, fmt.Sprintf("%v, ns=%v, N_CCE=%v, L=%v, M=%v, m=%v, Y=%v -> startCce=%v", sstype, ns, N_CCE, L, M, m, Y, startCce))

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
