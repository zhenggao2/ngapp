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
package ttitrace

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

type L2TtiTraceParser struct {
	log *zap.Logger
	ttiDir string
	ttiPattern string
	ttiRat string
	ttiScs string
	ttiFilter string
	maxgo int
	debug bool
	nok21a bool

	slotsPerRf int
	ttiFiles []string
}

type SfnInfo struct {
	lastSfn int
	hsfn int
}

func (p *L2TtiTraceParser) Init(log *zap.Logger, dir, pattern, rat, scs, filter string, maxgo int, debug bool) {
	p.log = log
	p.ttiDir = dir
	p.ttiPattern = strings.ToLower(pattern)
	p.ttiRat = strings.ToLower(rat)
	p.ttiScs = strings.ToLower(scs)
	p.ttiFilter = strings.ToLower(filter)
	p.maxgo = maxgo
	p.debug = debug
	p.nok21a = false

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing L2 TTI trace parser...(working dir: %v)", p.ttiDir))

	fileInfo, err := ioutil.ReadDir(p.ttiDir)
	if err != nil {
		p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", p.ttiDir))
		return
	}

	for _, file := range fileInfo {
	    if !file.IsDir() && path.Ext(file.Name()) == p.ttiPattern {
			p.ttiFiles = append(p.ttiFiles, path.Join(p.ttiDir, file.Name()))
		}
	}
}

// func (p *L2TtiTraceParser) onOkBtnClicked(checked bool) {
func (p *L2TtiTraceParser) Exec() {
	scs2nslots := map[string]int{"15k": 10, "30k": 20, "120k": 80}
	p.slotsPerRf = scs2nslots[strings.ToLower(p.ttiScs)]

	// recreate dir for parsed l2 tti trace
	outPath := path.Join(p.ttiDir, "parsed_tti")
	os.RemoveAll(outPath)
	if err := os.MkdirAll(outPath, 0775); err != nil {
		panic(fmt.Sprintf("Fail to create directory: %v", err))
	}

	// key=EventName_PCI_RNTI or EventName for both mapFieldName and mapSfnInfo
	mapFieldName := make(map[string][]string)
	mapSfnInfo := make(map[string]SfnInfo)

	// Field positions per event
	// TODO - variable definition
	var posDlBeam TtiDlBeamDataPos
	var posDlPreSched TtiDlPreSchedDataPos
	var posDlTdSched TtiDlTdSchedSubcellDataPos
	var posDlFdSched TtiDlFdSchedDataPos
	var posDlHarq TtiDlHarqRxDataPos
	var posDlLaAvgCqi TtiDlLaAverageCqiPos
	var posCsiSrReport TtiCsiSrReportDataPos
	var posDlFlowControl TtiDlFlowControlDataPos
	var posDlLaDeltaCqi TtiDlLaDeltaCqiPos
	var posUlBsr TtiUlBsrRxDataPos
	var posUlPreSched TtiUlPreSchedDataPos
	var posUlTdSched TtiUlTdSchedSubcellDataPos
	var posUlFdSched TtiUlFdSchedDataPos
	var posUlHarq TtiUlHarqRxDataPos
	var posDrx TtiUlIntraDlToUlDrxSyncDlDataPos
	var posUlLaDeltaSinr TtiUlLaDeltaSinrPos
	var posUlLaAvgSinr TtiUlLaAverageSinrPos
	var posUlLaPhr TtiUlLaPhrPos
	var posUlPucch TtiUlPucchReceiveRespPsDataPos
	var posUlPusch TtiUlPuschReceiveRespPsDataPos
	var posUlPduDemux TtiUlPduDemuxDataPos
	var mapEventRecord = map[string]*utils.OrderedMap{
		"dlBeamData":           utils.NewOrderedMap(),
		"dlPreSchedData":       utils.NewOrderedMap(),
		"dlTdSchedSubcellData": utils.NewOrderedMap(),
		"dlFdSchedData":        utils.NewOrderedMap(),
		"dlHarqRxData":         utils.NewOrderedMap(),
		"dlLaAverageCqi":       utils.NewOrderedMap(),
		"csiSrReportData":       utils.NewOrderedMap(),
		"dlFlowControlData":       utils.NewOrderedMap(),
		"dlLaDeltaCqi":       utils.NewOrderedMap(),
		"ulBsrRxData":       utils.NewOrderedMap(),
		"ulPreSchedData":       utils.NewOrderedMap(),
		"ulTdSchedSubcellData": utils.NewOrderedMap(),
		"ulFdSchedData":       utils.NewOrderedMap(),
		"ulHarqRxData":       utils.NewOrderedMap(),
		"ulIntraDlToUlDrxSyncDlData":       utils.NewOrderedMap(),
		"ulLaDeltaSinr":       utils.NewOrderedMap(),
		"ulLaAverageSinr":       utils.NewOrderedMap(),
		"ulLaPhr":       utils.NewOrderedMap(),
		"ulPucchReceiveRespPsData":       utils.NewOrderedMap(),
		"ulPuschReceiveRespPsData":       utils.NewOrderedMap(),
		"ulPduDemuxData":       utils.NewOrderedMap(),
	}
	var dlSchedAggFields string
	var ulSchedAggFields string
	var dlPerBearerProcessed bool
	var mapPdschSliv map[string][]int = p.initPdschSliv()
	var mapPuschSliv map[string][]int = p.initPuschSliv()

	mapAntPort := map[int]string {
		32768: "0",
		16384: "1",
		8192: "2",
		4096: "3",
		49152: "0;1",
		12288: "2;3",
		57344: "0;1;2",
		61440: "0;1;2;3",
		40960: "0;2",
		2048: "4",
		1024: "5",
		512: "6",
		256: "7",
		3072: "4;5",
		768: "6;7",
		34816: "0;4",
		8704: "2;6",
		51200: "0;1;4",
		12800: "2;3;6",
		52224: "0;1;4;5",
		13056: "2;3;6;7",
		43520: "0;2;4;6",
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("parsing tti files...(%d in total)", len(p.ttiFiles)))
	for _, fn := range p.ttiFiles {
		p.writeLog(zapcore.DebugLevel, fmt.Sprintf("parsing: %s", fn))

		fin, err := os.Open(fn)
		// defer fin.Close()
		if err != nil {
			p.writeLog(zapcore.ErrorLevel, err.Error())
			continue
		}

		reader := bufio.NewReader(fin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			// remove leading and tailing spaces
			line = strings.TrimSpace(line)

			if len(line) > 0 {
				tokens := strings.Split(line, ":")
				if len(tokens) == 2 {
					eventName, eventRecord := tokens[0], tokens[1]
					tokens = strings.Split(eventRecord, ",")
					if len(tokens) > 0 {
						// special handing of tokens[0]
						tokens[0] = strings.TrimSpace(tokens[0])

						// differentiate field names and field values, also keep track of PCI and RNTI
						posSfn, posPci, posRnti, valStart := -1, -1, -1, -1
						for pos, item := range tokens {
							if strings.ToLower(item) == "sfn" && posSfn < 0 {
								posSfn = pos
							}

							if strings.ToLower(item) == "rnti" && posRnti < 0 {
								posRnti = pos
							}

							if strings.ToLower(item) == "physcellid" && posPci < 0 {
								posPci = pos
							}

							_, err := strconv.Atoi(item)
							if err == nil {
								valStart = pos
								break
							}
						}

						var key string
						if posPci >= 0 && posRnti >= 0 {
							key = fmt.Sprintf("%s_pci%s_rnti%s", eventName, tokens[valStart+posPci], tokens[valStart+posRnti])
						} else {
							key = eventName
						}

						outFn := path.Join(outPath, fmt.Sprintf("%s.csv", key))
						_, exist := mapFieldName[key]
						if !exist {
							mapFieldName[key] = make([]string, valStart)
							copy(mapFieldName[key], tokens[:valStart])

							sfn, _ := strconv.Atoi(tokens[valStart+posSfn])
							mapSfnInfo[key] = SfnInfo{sfn, 1} // assume HSFN start with 1

							// Step-1: write event header only once
							fout, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
							//defer fout.Close()
							if err != nil {
								p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", outFn))
								break
							}

							// Step-1.1: write HSFN header field
							fout.WriteString("hsfn,")

							row := strings.Join(mapFieldName[key], ",")
							fout.WriteString(fmt.Sprintf("%s\n", row))
							fout.Close()

							// update dlSchedAggFields
							if len(dlSchedAggFields) == 0 && eventName == "dlFdSchedData" {
								dlSchedAggFields = "hsfn," + row
							}

							// update ulSchedAggFields
							if len(ulSchedAggFields) == 0 && eventName == "ulFdSchedData" {
								ulSchedAggFields = "hsfn," + row
							}
						} else {
							curSfn, _ := strconv.Atoi(tokens[valStart+posSfn])
							if mapSfnInfo[key].lastSfn > curSfn {
								mapSfnInfo[key] = SfnInfo{curSfn, mapSfnInfo[key].hsfn + 1}
							} else {
								mapSfnInfo[key] = SfnInfo{curSfn, mapSfnInfo[key].hsfn}
							}
						}

						// Step-2: write event record
						fout2, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
						//defer fout2.Close()
						if err != nil {
							p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", outFn))
							break
						}

						// Step-2.1: write hsfn
						fout2.WriteString(fmt.Sprintf("%d,", mapSfnInfo[key].hsfn))

						row := strings.Join(tokens[valStart:], ",")
						fout2.WriteString(fmt.Sprintf("%s\n", row))
						fout2.Close()

						// Step-3: aggregate events
						if eventName == "dlBeamData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlBeamData
							if posDlBeam.Ready == false {
								posDlBeam = FindTtiDlBeamDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlBeam=%v", posDlBeam))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDlBeam.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDlBeam.PosEventHeader.PosSlot]))
							v := TtiDlBeamData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posDlBeam.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posDlBeam.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posDlBeam.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posDlBeam.PosEventHeader.PosPhysCellId],
								},

								SubcellId:          tokens[valStart+posDlBeam.PosSubcellId],
								CurrentBestBeamId:  tokens[valStart+posDlBeam.PosCurrentBestBeamId],
								Current2ndBeamId:   tokens[valStart+posDlBeam.PosCurrent2ndBeamId],
								SelectedBestBeamId: tokens[valStart+posDlBeam.PosSelectedBestBeamId],
								Selected2ndBeamId:  tokens[valStart+posDlBeam.PosSelected2ndBeamId],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "dlPreSchedData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlPreSchedData
							if posDlPreSched.Ready == false {
								posDlPreSched = FindTtiDlPreSchedDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlPreSched=%v", posDlPreSched))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDlPreSched.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDlPreSched.PosEventHeader.PosSlot]))
							v := TtiDlPreSchedData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posDlPreSched.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posDlPreSched.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posDlPreSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posDlPreSched.PosEventHeader.PosPhysCellId],
								},

								CsListEvent:          tokens[valStart+posDlPreSched.PosCsListEvent],
								HighestClassPriority: p.ttiDlPreSchedClassPriority(tokens[valStart+posDlPreSched.PosHighestClassPriority]),
								PrachPreambleIndex: tokens[valStart+posDlPreSched.PosPrachPreambleIndex],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "dlTdSchedSubcellData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlTdSchedSubcellData
							if posDlTdSched.Ready == false {
								posDlTdSched = FindTtiDlTdSchedSubcellDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlTdSched=%v", posDlTdSched))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDlTdSched.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDlTdSched.PosEventHeader.PosSlot]))
							v := TtiDlTdSchedSubcellData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posDlTdSched.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posDlTdSched.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posDlTdSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posDlTdSched.PosEventHeader.PosPhysCellId],
								},

								SubcellId: tokens[valStart+posDlTdSched.PosSubcellId],
								Cs2List:   make([]string, 0),
							}

							for _, rsn := range posDlTdSched.PosRecordSequenceNumber {
								n := 10 // Per UE in CS2 is statically defined for 10UEs
								for k := 0; k < n; k += 1 {
									posRnti := valStart + rsn + 1 + 3*k
									if posRnti > len(tokens) || len(tokens[posRnti]) == 0 {
										break
									}
									v.Cs2List = append(v.Cs2List, tokens[posRnti])
								}
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "dlFdSchedData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlFdSchedData
							if posDlFdSched.Ready == false {
								posDlFdSched = FindTtiDlFdSchedDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlFdSched=%v", posDlFdSched))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDlFdSched.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDlFdSched.PosEventHeader.PosSlot]))
							v := TtiDlFdSchedData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posDlFdSched.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posDlFdSched.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posDlFdSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posDlFdSched.PosEventHeader.PosPhysCellId],
								},

								CellDbIndex:        tokens[valStart+posDlFdSched.PosCellDbIndex],
								SubcellId:        tokens[valStart+posDlFdSched.PosSubcellId],
								TxNumber:           tokens[valStart+posDlFdSched.PosTxNumber],
								DlHarqProcessIndex: tokens[valStart+posDlFdSched.PosDlHarqProcessIndex],
								K1:                 tokens[valStart+posDlFdSched.PosK1],
								AllFields:          make([]string, len(tokens)-valStart),
							}
							copy(v.AllFields, tokens[valStart:])

							// update SLIV field
							slivStr := "("
							// PDSCH mapping type A and normal CP
							sliv := fmt.Sprintf("00_%s", tokens[valStart+posDlFdSched.PosSliv])
							if SL, exist := mapPdschSliv[sliv]; exist {
								slivStr += fmt.Sprintf("TypeA[S=%d;L=%d]", SL[0], SL[1])
							}
							// PDSCH mapping type B and normal CP
							sliv = fmt.Sprintf("10_%s", tokens[valStart+posDlFdSched.PosSliv])
							if SL, exist := mapPdschSliv[sliv]; exist {
								slivStr += fmt.Sprintf(";TypeB[S=%d;L=%d]", SL[0], SL[1])
							}
							slivStr += ")"

							v.AllFields[posDlFdSched.PosSliv] += slivStr

							// update AntPort field
							antPortStr := "(1000+"
							if ports, exist := mapAntPort[p.unsafeAtoi(tokens[valStart+posDlFdSched.PosAntPort])]; exist {
								antPortStr += ports
							}
							antPortStr += ")"

							v.AllFields[posDlFdSched.PosAntPort] += antPortStr

							// update txNumber field
							intTxNum := p.unsafeAtoi(v.AllFields[posDlFdSched.PosTxNumber])
							if intTxNum == 1 {
								v.AllFields[posDlFdSched.PosTxNumber] += fmt.Sprintf("(IniTx)")
							} else {
								v.AllFields[posDlFdSched.PosTxNumber] += fmt.Sprintf("(ReTx%v)", intTxNum-1)
							}

							// update per beaer info:
							// 5G21A = [lcId, scheduledBytesPerBearer, remainingBytesPerBearerInFdEoBuffer, bsrSfn, bsrSlot]
							// 5G20B = [lcId, scheduledBytesPerBearer, remainingBytesPerBearerInFdEoBuffer]
							// there are max 18 bearers
							maxNumBearerPerUe := 18
							sizeSchedBearerRecord := 3
							if p.nok21a {
								sizeSchedBearerRecord = 5
							}
							lcId := make([]string, 0)
							schedBytes := make([]string, 0)
							remainBytes := make([]string, 0)
							bsrSfn := make([]string, 0)
							bsrSlot := make([]string, 0)
							for ib := 0; ib < maxNumBearerPerUe; ib += 1 {
								if p.unsafeAtoi(tokens[valStart+posDlFdSched.PosLcId+sizeSchedBearerRecord*ib]) == 255 {
									break
								} else {
									lcId = append(lcId, tokens[valStart+posDlFdSched.PosLcId+sizeSchedBearerRecord*ib])
									schedBytes = append(schedBytes, tokens[valStart+posDlFdSched.PosLcId+sizeSchedBearerRecord*ib+1])
									remainBytes = append(remainBytes, tokens[valStart+posDlFdSched.PosLcId+sizeSchedBearerRecord*ib+2])
									if p.nok21a {
										bsrSfn = append(bsrSfn, tokens[valStart+posDlFdSched.PosLcId+sizeSchedBearerRecord*ib+3])
										bsrSlot = append(bsrSlot, tokens[valStart+posDlFdSched.PosLcId+sizeSchedBearerRecord*ib+4])
									}
								}
							}

							v.LcIdList = lcId

							var perBearerInfo []string
							if p.nok21a {
								perBearerInfo = []string{fmt.Sprintf("[%s]", strings.Join(lcId, ";")), fmt.Sprintf("[%s]", strings.Join(schedBytes, ";")), fmt.Sprintf("[%s]", strings.Join(remainBytes, ";")),
									fmt.Sprintf("[%s]", strings.Join(bsrSfn, ";")), fmt.Sprintf("[%s]", strings.Join(bsrSlot, ";"))}
							} else {
								perBearerInfo = []string{fmt.Sprintf("[%s]", strings.Join(lcId, ";")), fmt.Sprintf("[%s]", strings.Join(schedBytes, ";")), fmt.Sprintf("[%s]", strings.Join(remainBytes, ";"))}
							}
							v.AllFields = append(append(v.AllFields[:posDlFdSched.PosLcId], perBearerInfo...), v.AllFields[posDlFdSched.PosLcId+sizeSchedBearerRecord*maxNumBearerPerUe:]...)

							// update dlSchedAggFields accordingly only once
							if !dlPerBearerProcessed {
								dlSchedAggFieldsTokens := strings.Split(dlSchedAggFields, ",")
								dlSchedAggFields = strings.Join(append(dlSchedAggFieldsTokens[:1+posDlFdSched.PosLcId+sizeSchedBearerRecord], dlSchedAggFieldsTokens[1+posDlFdSched.PosLcId+sizeSchedBearerRecord*maxNumBearerPerUe:]...), ",")
								dlPerBearerProcessed = true
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if (eventName == "dlHarqRxData" || eventName == "dlHarqRxDataArray") && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlHarqRxData/dlHarqRxDataArray
							if posDlHarq.Ready == false {
								posDlHarq = FindTtiDlHarqRxDataPos(tokens)

								if eventName == "dlHarqRxDataArray" {
									posDlHarq.PosEventHeader.PosRnti += 2
									posDlHarq.PosEventHeader.PosPhysCellId += 2
								}

								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlHarq=%v", posDlHarq))
								}
							}

							if eventName == "dlHarqRxData" {
								k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDlHarq.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDlHarq.PosEventHeader.PosSlot]))
								v := TtiDlHarqRxData{
									// event header
									TtiEventHeader: TtiEventHeader{
										Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
										Sfn:        tokens[valStart+posDlHarq.PosEventHeader.PosSfn],
										Slot:       tokens[valStart+posDlHarq.PosEventHeader.PosSlot],
										Rnti:       tokens[valStart+posDlHarq.PosEventHeader.PosRnti],
										PhysCellId: tokens[valStart+posDlHarq.PosEventHeader.PosPhysCellId],
									},

									HarqSubcellId:      tokens[valStart+posDlHarq.PosHarqSubcellId],
									AckNack:            tokens[valStart+posDlHarq.PosAckNack],
									DlHarqProcessIndex: tokens[valStart+posDlHarq.PosDlHarqProcessIndex],
									PucchFormat: tokens[valStart+posDlHarq.PosPucchFormat],
								}

								mapEventRecord["dlHarqRxData"].Add(strings.Join([]string{strconv.Itoa(k), v.TtiEventHeader.PhysCellId, v.TtiEventHeader.Rnti, v.DlHarqProcessIndex}, "_") , &v)
							} else {
								maxNumHarq := 32	// max 32 HARQ feedbacks per dlHarqRxDataArray
								sizeDlHarqRecord := 14
								for ih := 0; ih < maxNumHarq; ih += 1 {
									posRnti := valStart+posDlHarq.PosEventHeader.PosRnti+ih*sizeDlHarqRecord
									if posRnti >= len(tokens) || len(tokens[posRnti]) == 0 {
										break
									}

									k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDlHarq.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDlHarq.PosEventHeader.PosSlot]))
									v := TtiDlHarqRxData{
										// event header
										TtiEventHeader: TtiEventHeader{
											Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
											Sfn:        tokens[valStart+posDlHarq.PosEventHeader.PosSfn],
											Slot:       tokens[valStart+posDlHarq.PosEventHeader.PosSlot],
											Rnti:       tokens[valStart+posDlHarq.PosEventHeader.PosRnti+ih*sizeDlHarqRecord],
											PhysCellId: tokens[valStart+posDlHarq.PosEventHeader.PosPhysCellId+ih*sizeDlHarqRecord],
										},

										HarqSubcellId:      tokens[valStart+posDlHarq.PosHarqSubcellId+ih*sizeDlHarqRecord],
										AckNack:            tokens[valStart+posDlHarq.PosAckNack+ih*sizeDlHarqRecord],
										DlHarqProcessIndex: tokens[valStart+posDlHarq.PosDlHarqProcessIndex+ih*sizeDlHarqRecord],
										PucchFormat: tokens[valStart+posDlHarq.PosPucchFormat+ih*sizeDlHarqRecord],
									}

									mapEventRecord["dlHarqRxData"].Add(strings.Join([]string{strconv.Itoa(k), v.TtiEventHeader.PhysCellId, v.TtiEventHeader.Rnti, v.DlHarqProcessIndex}, "_") , &v)
								}
							}
						} else if eventName == "dlLaAverageCqi" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlLaAverageCqi
							if posDlLaAvgCqi.Ready == false {
								posDlLaAvgCqi = FindTtiDlLaAverageCqiPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlLaAvgCqi=%v", posDlLaAvgCqi))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDlLaAvgCqi.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDlLaAvgCqi.PosEventHeader.PosSlot]))
							v := TtiDlLaAverageCqi{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posDlLaAvgCqi.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posDlLaAvgCqi.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posDlLaAvgCqi.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posDlLaAvgCqi.PosEventHeader.PosPhysCellId],
								},

								CellDbIndex: tokens[valStart+posDlLaAvgCqi.PosCellDbIndex],
								RrmInstCqi:   tokens[valStart+posDlLaAvgCqi.PosRrmInstCqi],
								Rank:   tokens[valStart+posDlLaAvgCqi.PosRank],
								RrmAvgCqi:   tokens[valStart+posDlLaAvgCqi.PosRrmAvgCqi],
								Mcs:   tokens[valStart+posDlLaAvgCqi.PosMcs],
								RrmDeltaCqi: tokens[valStart+posDlLaAvgCqi.PosRrmDeltaCqi],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "csiSrReportData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - csiSrReportData
							if posCsiSrReport.Ready == false {
								posCsiSrReport = FindTtiCsiSrReportDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posCsiSrReport=%v", posCsiSrReport))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posCsiSrReport.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posCsiSrReport.PosEventHeader.PosSlot]))
							v := TtiCsiSrReportData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posCsiSrReport.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posCsiSrReport.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posCsiSrReport.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posCsiSrReport.PosEventHeader.PosPhysCellId],
								},

								UlChannel: tokens[valStart+posCsiSrReport.PosUlChannel],
								Dtx:   tokens[valStart+posCsiSrReport.PosDtx],
								PucchFormat:   tokens[valStart+posCsiSrReport.PosPucchFormat],
								Cqi:   tokens[valStart+posCsiSrReport.PosCqi],
								PmiRank1:   tokens[valStart+posCsiSrReport.PosPmiRank1],
								PmiRank2: tokens[valStart+posCsiSrReport.PosPmiRank2],
								Ri: tokens[valStart+posCsiSrReport.PosRi],
								Cri: tokens[valStart+posCsiSrReport.PosCri],
								Li: tokens[valStart+posCsiSrReport.PosLi],
								Sr: tokens[valStart+posCsiSrReport.PosSr],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "dlFlowControlData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlFlowControlData
							if posDlFlowControl.Ready == false {
								posDlFlowControl = FindTtiDlFlowControlDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlFlowControl=%v", posDlFlowControl))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDlFlowControl.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDlFlowControl.PosEventHeader.PosSlot]))
							v := TtiDlFlowControlData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posDlFlowControl.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posDlFlowControl.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posDlFlowControl.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posDlFlowControl.PosEventHeader.PosPhysCellId],
								},

								LchId: tokens[valStart+posDlFlowControl.PosLchId],
								ReportType:   tokens[valStart+posDlFlowControl.PosReportType],
								ScheduledBytes:   tokens[valStart+posDlFlowControl.PosScheduledBytes],
								EthAvg:   tokens[valStart+posDlFlowControl.PosEthAvg],
								EthScaled:   tokens[valStart+posDlFlowControl.PosEthScaled],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "dlLaDeltaCqiArray" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlLaDeltaCqiArray
							if posDlLaDeltaCqi.Ready == false {
								posDlLaDeltaCqi = FindTtiDlLaDeltaCqiPos(tokens)

								if eventName == "dlLaDeltaCqiArray" {
									posDlLaDeltaCqi.PosEventHeader.PosRnti += 2
									posDlLaDeltaCqi.PosEventHeader.PosPhysCellId += 2
								}

								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlLaDeltaCqi=%v", posDlLaDeltaCqi))
								}
							}

							if eventName == "dlLaDeltaCqi" {
								k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosSlot]))
								v := TtiDlLaDeltaCqi{
									// event header
									TtiEventHeader: TtiEventHeader{
										Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
										Sfn:        tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosSfn],
										Slot:       tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosSlot],
										Rnti:       tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosRnti],
										PhysCellId: tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosPhysCellId],
									},

									CellDbIndex:      tokens[valStart+posDlLaDeltaCqi.PosCellDbIndex],
									IsDeltaCqiCalculated:            tokens[valStart+posDlLaDeltaCqi.PosIsDeltaCqiCalculated],
									RrmPauseUeInDlScheduling: tokens[valStart+posDlLaDeltaCqi.PosRrmPauseUeInDlScheduling],
									HarqFb: tokens[valStart+posDlLaDeltaCqi.PosHarqFb],
									RrmDeltaCqi: tokens[valStart+posDlLaDeltaCqi.PosRrmDeltaCqi],
									RrmRemainingBucketLevel: tokens[valStart+posDlLaDeltaCqi.PosRrmRemainingBucketLevel],
								}

								mapEventRecord["dlLaDeltaCqi"].Add(strings.Join([]string{strconv.Itoa(k), v.TtiEventHeader.PhysCellId, v.TtiEventHeader.Rnti}, "_") , &v)
							} else {
								maxNumDlOlqc := 64	// max 64 DL LA deltaCqi records per dlHarqRxDataArray
								sizeDlOlqcRecord := 8
								for ih := 0; ih < maxNumDlOlqc; ih += 1 {
									posRnti := valStart+posDlLaDeltaCqi.PosEventHeader.PosRnti+ih*sizeDlOlqcRecord
									if posRnti >= len(tokens) || len(tokens[posRnti]) == 0 {
										break
									}

									k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosSlot]))
									v := TtiDlLaDeltaCqi{
										// event header
										TtiEventHeader: TtiEventHeader{
											Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
											Sfn:        tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosSfn],
											Slot:       tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosSlot],
											Rnti:       tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosRnti+ih*sizeDlOlqcRecord],
											PhysCellId: tokens[valStart+posDlLaDeltaCqi.PosEventHeader.PosPhysCellId+ih*sizeDlOlqcRecord],
										},

										CellDbIndex:      tokens[valStart+posDlLaDeltaCqi.PosCellDbIndex+ih*sizeDlOlqcRecord],
										IsDeltaCqiCalculated:            tokens[valStart+posDlLaDeltaCqi.PosIsDeltaCqiCalculated+ih*sizeDlOlqcRecord],
										RrmPauseUeInDlScheduling: tokens[valStart+posDlLaDeltaCqi.PosRrmPauseUeInDlScheduling+ih*sizeDlOlqcRecord],
										HarqFb: tokens[valStart+posDlLaDeltaCqi.PosHarqFb+ih*sizeDlOlqcRecord],
										RrmDeltaCqi: tokens[valStart+posDlLaDeltaCqi.PosRrmDeltaCqi+ih*sizeDlOlqcRecord],
										RrmRemainingBucketLevel: tokens[valStart+posDlLaDeltaCqi.PosRrmRemainingBucketLevel+ih*sizeDlOlqcRecord],
									}

									mapEventRecord["dlLaDeltaCqi"].Add(strings.Join([]string{strconv.Itoa(k), v.TtiEventHeader.PhysCellId, v.TtiEventHeader.Rnti}, "_") , &v)
								}
							}
						} else if eventName == "ulBsrRxData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulBsrRxData
							if posUlBsr.Ready == false {
								posUlBsr = FindTtiUlBsrRxDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlBsr=%v", posUlBsr))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlBsr.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlBsr.PosEventHeader.PosSlot]))
							v := TtiUlBsrRxData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlBsr.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlBsr.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlBsr.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlBsr.PosEventHeader.PosPhysCellId],
								},

								UlHarqProcessIndex: tokens[valStart+posUlBsr.PosUlHarqProcessIndex],
								BsrFormat:   tokens[valStart+posUlBsr.PosBsrFormat],
								BufferSizeList: make([]string, 0),
							}

							numLcg := 8  // for LCG 0~7 and maxLCG-ID = 7
							for i := 0; i < numLcg; i += 1 {
								// TODO convert bufferSize to a readable string as specified in TS 38.321
								v.BufferSizeList = append(v.BufferSizeList, tokens[valStart+posUlBsr.PosBsrFormat+1+i])
							}

							mapEventRecord[eventName].Add(k, &v)
						}  else if eventName == "ulPreSchedData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulPreSchedData
							if posUlPreSched.Ready == false {
								posUlPreSched = FindTtiUlPreSchedDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlPreSched=%v", posUlPreSched))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlPreSched.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlPreSched.PosEventHeader.PosSlot]))
							v := TtiUlPreSchedData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlPreSched.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlPreSched.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlPreSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlPreSched.PosEventHeader.PosPhysCellId],
								},

								CsListEvent:          tokens[valStart+posUlPreSched.PosCsListEvent],
								HighestClassPriority: p.ttiUlPreSchedClassPriority(tokens[valStart+posUlPreSched.PosHighestClassPriority]),
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "ulTdSchedSubcellData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulTdSchedSubcellData
							if posUlTdSched.Ready == false {
								posUlTdSched = FindTtiUlTdSchedSubcellDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlTdSched=%v", posUlTdSched))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlTdSched.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlTdSched.PosEventHeader.PosSlot]))
							v := TtiUlTdSchedSubcellData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlTdSched.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlTdSched.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlTdSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlTdSched.PosEventHeader.PosPhysCellId],
								},

								SubcellId: tokens[valStart+posUlTdSched.PosSubcellId],
								Cs2List:   make([]string, 0),
							}

							for _, rsn := range posUlTdSched.PosRecordSequenceNumber {
								n := 10 // Per UE in CS2 is statically defined for 10UEs
								for k := 0; k < n; k += 1 {
									posRnti := valStart + rsn + 1 + 3*k
									if posRnti > len(tokens) || len(tokens[posRnti]) == 0 {
										break
									}
									v.Cs2List = append(v.Cs2List, tokens[posRnti])
								}
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "ulFdSchedData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulFdSchedData
							if posUlFdSched.Ready == false {
								posUlFdSched = FindTtiUlFdSchedDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlFdSched=%v", posUlFdSched))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlFdSched.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlFdSched.PosEventHeader.PosSlot]))
							v := TtiUlFdSchedData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlFdSched.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlFdSched.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlFdSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlFdSched.PosEventHeader.PosPhysCellId],
								},

								CellDbIndex:        tokens[valStart+posUlFdSched.PosCellDbIndex],
								SubcellId:        tokens[valStart+posUlFdSched.PosSubcellId],
								TxNumber:           tokens[valStart+posUlFdSched.PosTxNumber],
								UlHarqProcessIndex: tokens[valStart+posUlFdSched.PosUlHarqProcessIndex],
								K2:                 tokens[valStart+posUlFdSched.PosK2],
								AllFields:          make([]string, len(tokens)-valStart),
							}
							copy(v.AllFields, tokens[valStart:])

							// update SLIV field
							slivStr := "("
							// PUSCH mapping type A and normal CP
							sliv := fmt.Sprintf("00_%s", tokens[valStart+posUlFdSched.PosSliv])
							if SL, exist := mapPuschSliv[sliv]; exist {
								slivStr += fmt.Sprintf("TypeA[S=%d;L=%d]", SL[0], SL[1])
							}
							// PDSCH mapping type B and normal CP
							sliv = fmt.Sprintf("10_%s", tokens[valStart+posUlFdSched.PosSliv])
							if SL, exist := mapPuschSliv[sliv]; exist {
								slivStr += fmt.Sprintf(";TypeB[S=%d;L=%d]", SL[0], SL[1])
							}
							slivStr += ")"

							v.AllFields[posUlFdSched.PosSliv] += slivStr

							// update AntPort field
							antPortStr := "("
							if ports, exist := mapAntPort[p.unsafeAtoi(tokens[valStart+posUlFdSched.PosAntPort])]; exist {
								antPortStr += ports
							}
							antPortStr += ")"

							v.AllFields[posUlFdSched.PosAntPort] += antPortStr

							// update txNumber field
							intTxNum := p.unsafeAtoi(v.AllFields[posUlFdSched.PosTxNumber])
							if intTxNum == 1 {
								v.AllFields[posUlFdSched.PosTxNumber] += fmt.Sprintf("(IniTx)")
							} else {
								v.AllFields[posUlFdSched.PosTxNumber] += fmt.Sprintf("(ReTx%v)", intTxNum-1)
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "ulHarqRxData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulHarqRxData
							if posUlHarq.Ready == false {
								posUlHarq = FindTtiUlHarqRxDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlHarq=%v", posUlHarq))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlHarq.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlHarq.PosEventHeader.PosSlot]))
							v := TtiUlHarqRxData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlHarq.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlHarq.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlHarq.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlHarq.PosEventHeader.PosPhysCellId],
								},

								SubcellId: tokens[valStart+posUlHarq.PosSubcellId],
								Dtx: tokens[valStart+posUlHarq.PosDtx],
								CrcResult: tokens[valStart+posUlHarq.PosCrcResult],
								UlHarqProcessIndex: tokens[valStart+posUlHarq.PosUlHarqProcessIndex],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "ulIntraDlToUlDrxSyncDlData" {
							// TODO - event aggregation - ulIntraDlToUlDrxSyncDlData
							if posDrx.Ready == false {
								posDrx = FindTtiUlIntraDlToUlDrxSyncDlDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDrx=%v", posDrx))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posDrx.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posDrx.PosEventHeader.PosSlot]))
							v := TtiUlIntraDlToUlDtxSyncDlData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posDrx.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posDrx.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posDrx.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posDrx.PosEventHeader.PosPhysCellId],
								},

								DrxEnabled: tokens[valStart+posDrx.PosDrxEnabled],
								DlDrxOnDurationTimerOn: tokens[valStart+posDrx.PosDlDrxOnDurationTimerOn],
								DlDrxInactivityTimerOn: tokens[valStart+posDrx.PosDlDrxInactivityTimerOn],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "ulLaDeltaSinr" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulLaDeltaSinr
							if posUlLaDeltaSinr.Ready == false {
								posUlLaDeltaSinr = FindTtiUlLaDeltaSinrPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlLaDeltaSinr=%v", posUlLaDeltaSinr))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlLaDeltaSinr.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlLaDeltaSinr.PosEventHeader.PosSlot]))
							v := TtiUlLaDeltaSinr{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlLaDeltaSinr.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlLaDeltaSinr.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlLaDeltaSinr.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlLaDeltaSinr.PosEventHeader.PosPhysCellId],
								},

								CellDbIndex: tokens[valStart+posUlLaDeltaSinr.PosCellDbIndex],
								IsDeltaSinrCalculated: tokens[valStart+posUlLaDeltaSinr.PosIsDeltaSinrCalculated],
								RrmPauseUeInUlScheduling: tokens[valStart+posUlLaDeltaSinr.PosRrmPauseUeInUlScheduling],
								CrcFb: tokens[valStart+posUlLaDeltaSinr.PosCrcFb],
								RrmDeltaSinr: tokens[valStart+posUlLaDeltaSinr.PosRrmDeltaSinr],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "ulLaAverageSinr" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulLaAverageSinr
							if posUlLaAvgSinr.Ready == false {
								posUlLaAvgSinr = FindTtiUlLaAverageSinrPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlLaAvgSinr=%v", posUlLaAvgSinr))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlLaAvgSinr.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlLaAvgSinr.PosEventHeader.PosSlot]))
							v := TtiUlLaAverageSinr{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlLaAvgSinr.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlLaAvgSinr.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlLaAvgSinr.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlLaAvgSinr.PosEventHeader.PosPhysCellId],
								},

								CellDbIndex: tokens[valStart+posUlLaAvgSinr.PosCellDbIndex],
								RrmInstSinrRank: tokens[valStart+posUlLaAvgSinr.PosRrmInstSinrRank],
								RrmNumOfSinrMeasurements: tokens[valStart+posUlLaAvgSinr.PosRrmNumOfSinrMeasurements],
								RrmInstSinr: tokens[valStart+posUlLaAvgSinr.PosRrmInstSinr],
								RrmAvgSinrUl: tokens[valStart+posUlLaAvgSinr.PosRrmAvgSinrUl],
								RrmSinrCorrection: tokens[valStart+posUlLaAvgSinr.PosRrmSinrCorrection],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "ulLaPhr" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulLaPhr
							if posUlLaPhr.Ready == false {
								posUlLaPhr = FindTtiUlLaPhrPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlLaPhr=%v", posUlLaPhr))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlLaPhr.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlLaPhr.PosEventHeader.PosSlot]))
							v := TtiUlLaPhr{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlLaPhr.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlLaPhr.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlLaPhr.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlLaPhr.PosEventHeader.PosPhysCellId],
								},

								CellDbIndex: tokens[valStart+posUlLaPhr.PosCellDbIndex],
								IsRrmPhrScaledCalculated: tokens[valStart+posUlLaPhr.PosIsRrmPhrScaledCalculated],
								Phr: tokens[valStart+posUlLaPhr.PosPhr],
								RrmNumPuschPrb: tokens[valStart+posUlLaPhr.PosRrmNumPuschPrb],
								RrmPhrScaled: tokens[valStart+posUlLaPhr.PosRrmPhrScaled],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "ulPucchReceiveRespPsData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulPucchReceiveRespPsData
							if posUlPucch.Ready == false {
								posUlPucch = FindTtiUlPucchReceiveRespPsDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlPucch=%v", posUlPucch))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlPucch.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlPucch.PosEventHeader.PosSlot]))
							v := TtiUlPucchReceiveRespPsData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlPucch.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlPucch.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlPucch.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlPucch.PosEventHeader.PosPhysCellId],
								},

								PucchFormat: tokens[valStart+posUlPucch.PosPucchFormat],
								StartPrb: tokens[valStart+posUlPucch.PosStartPrb],
								Rssi: tokens[valStart+posUlPucch.PosRssi],
								SinrLayer0: tokens[valStart+posUlPucch.PosSinrLayer0],
								SinrLayer1: tokens[valStart+posUlPucch.PosSinrLayer1],
								Dtx: tokens[valStart+posUlPucch.PosDtx],
								SrBit: tokens[valStart+posUlPucch.PosSrBit],
								SubcellId: tokens[valStart+posUlPucch.PosSubcellId],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "ulPuschReceiveRespPsData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulPuschReceiveRespPsData
							if posUlPusch.Ready == false {
								posUlPusch = FindTtiUlPuschReceiveRespPsDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlPusch=%v", posUlPusch))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlPusch.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlPusch.PosEventHeader.PosSlot]))
							v := TtiUlPuschReceiveRespPsData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlPusch.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlPusch.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlPusch.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlPusch.PosEventHeader.PosPhysCellId],
								},

								Rssi: tokens[valStart+posUlPusch.PosRssi],
								SinrLayer0: tokens[valStart+posUlPusch.PosSinrLayer0],
								SinrLayer1: tokens[valStart+posUlPusch.PosSinrLayer1],
								Dtx: tokens[valStart+posUlPusch.PosDtx],
								UlRank: tokens[valStart+posUlPusch.PosUlRank],
								UlPmiRank1: tokens[valStart+posUlPusch.PosUlPmiRank1],
								UlPmiRank1Sinr: tokens[valStart+posUlPusch.PosUlPmiRank1Sinr],
								UlPmiRank2: tokens[valStart+posUlPusch.PosUlPmiRank2],
								UlPmiRank2SinrLayer0: tokens[valStart+posUlPusch.PosUlPmiRank2SinrLayer0],
								UlPmiRank2SinrLayer1: tokens[valStart+posUlPusch.PosUlPmiRank2SinrLayer1],
								LongTermRank: tokens[valStart+posUlPusch.PosLongTermRank],
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "ulPduDemuxData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulPduDemuxData
							if posUlPduDemux.Ready == false {
								posUlPduDemux = FindTtiUlPduDemuxDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlPduDemux=%v", posUlPduDemux))
								}
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[valStart+posUlPduDemux.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[valStart+posUlPduDemux.PosEventHeader.PosSlot]))
							v := TtiUlPduDemuxData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[valStart+posUlPduDemux.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlPduDemux.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlPduDemux.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlPduDemux.PosEventHeader.PosPhysCellId],
								},

								HarqId:        tokens[valStart+posUlPduDemux.PosHarqId],
								IsUlCcchData:  tokens[valStart+posUlPduDemux.PosIsUlCcchData],
								IsTcpTraffic:  tokens[valStart+posUlPduDemux.PosIsTcpTraffic],
								TempCrnti:     tokens[valStart+posUlPduDemux.PosTempCrnti],
								LcIdList:      make([]string, 0),
								RcvdBytesList: make([]string, 0),
							}

							numLcId := 32  // for LC 1~32 and maxLC-ID = 32
							for i := 0; i < numLcId; i += 1 {
								posLcId := valStart+posUlPduDemux.PosTempCrnti+1+2*i
								if posLcId >= len(tokens) || len(tokens[posLcId]) == 0 {
									break
								}
								v.LcIdList = append(v.LcIdList, tokens[posLcId])
								v.RcvdBytesList = append(v.RcvdBytesList, tokens[posLcId+1])
							}

							mapEventRecord[eventName].Add(k, &v)
						}
					} else {
						p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Invalid event record detected: %s", line))
					}
				} else {
					p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Invalid event data detected: %s", line))
				}
			}
		}

		fin.Close()
	}

	/*
	if p.debug {
		for _, k := range mapEventRecord["csiSrReportData"].Keys() {
			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("k=%v, v=%v\n", k, mapEventRecord["csiSrReportData"].Val(k)))
		}
	}
	 */

	if p.ttiFilter == "dl" || p.ttiFilter == "both" {
		// update dlSchedAggFields
		// TODO - dlSchedAggFields
		p.writeLog(zapcore.InfoLevel, "updating fields for dlSchedAgg...") //core.QCoreApplication_Instance().ProcessEvents(0)
		if mapEventRecord["dlBeamData"].Len() > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlBeam.hsfn", "dlBeam.sfn", "dlBeam.slot", "dlBeam.currentBestBeamId", "dlBeam.current2ndBeamId", "dlBeam.selectedBestBeamId", "dlBeam.selected2ndBeamid"}, ",")
		}

		if mapEventRecord["dlPreSchedData"].Len() > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlPreSched.hsfn", "dlPreSched.sfn", "dlPreSched.slot", "dlPreSched.csListEvent", "dlPreSched.highestClassPriority", "dlPreSched.prachPreambleIndex"}, ",")
		}

		if mapEventRecord["dlTdSchedSubcellData"].Len() > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlTdSched.hsfn", "dlTdSched.sfn", "dlTdSched.slot", "dlTdSched.cs2List"}, ",")
		}

		if mapEventRecord["dlHarqRxData"].Len() > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlHarq.hsfn", "dlHarq.sfn", "dlHarq.slot", "dlHarq.AckNack", "dlHarq.dlHarqProcessIndex", "dlHarq.pucchFormat"}, ",")
		}

		if mapEventRecord["ulIntraDlToUlDrxSyncDlData"].Len() > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"drx.hsfn", "drx.sfn", "drx.slot", "drx.drxEnabled", "drx.dlDrxOnDurationTimerOn", "drx.dlDrxInactivityTimerOn"}, ",")
		}

		if mapEventRecord["csiSrReportData"].Len() > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"csiSrReport.hsfn", "csiSrReport.sfn", "csiSrReport.slot", "csiSrReport.ulChannel", "csiSrReport.dtx", "csiSrReport.pucchFormat", "csiSrReport.cqi", "csiSrReport.pmiRank1", "csiSrReport.pmiRank2", "csiSrReport.ri", "csiSrReport.cri", "csiSrReport.li", "csiSrReport.sr"}, ",")
		}

		if mapEventRecord["dlLaDeltaCqi"].Len() > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlLaDeltaCqi.hsfn", "dlLaDeltaCqi.sfn", "dlLaDeltaCqi.slot", "dlLaDeltaCqi.isDeltaCqiCalculated", "dlLaDeltaCqi.rrmPauseUeInDlScheduling", "dlLaDeltaCqi.harqFb", "dlLaDeltaCqi.rrmDeltaCqi", "dlLaDeltaCqi.rrmRemainingBucketLevel"}, ",")
		}

		if mapEventRecord["dlLaAverageCqi"].Len() > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlLaAvgCqi.hsfn", "dlLaAvgCqi.sfn", "dlLaAvgCqi.slot", "dlLaAvgCqi.rrmInstCqi", "dlLaAvgCqi.rank", "dlLaAvgCqi.rrmAvgCqi", "dlLaAvgCqi.mcs", "dlLaAvgCqi.rrmDeltaCqi"}, ",")
		}

		if mapEventRecord["dlFlowControlData"].Len() > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlFlowControl.hsfn", "dlFlowControl.sfn", "dlFlowControl.slot", "dlFlowControl.lchId", "dlFlowControl.reportType", "dlFlowControl.scheduledBytes", "dlFlowControl.ethAvg", "dlFlowControl.ethScaled"}, ",")
		}

		dlSchedAggFields += "\n"

		// perform event aggregation
		// TODO - event aggregation with dlFdSchedData
		p.writeLog(zapcore.InfoLevel, "performing event aggregation for dlSchedAgg...[Time-consuming ops which may cause 100% CPU utilization!]")
		wg := &sync.WaitGroup{}
		for p1 := 0; p1 < mapEventRecord["dlFdSchedData"].Len(); p1 += 1 {
			wg.Add(1)
			go func(p1 int) {
				defer wg.Done()

				k1 := mapEventRecord["dlFdSchedData"].Keys()[p1].(int)
				v1 := mapEventRecord["dlFdSchedData"].Val(k1).(*TtiDlFdSchedData)

				// aggregate dlBeamData
				if mapEventRecord["dlBeamData"].Len() > 0 {
					p2 := p.findDlBeam(mapEventRecord["dlFdSchedData"], mapEventRecord["dlBeamData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["dlBeamData"].Keys()[p2].(int)
						v2 := mapEventRecord["dlBeamData"].Val(k2).(*TtiDlBeamData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.CurrentBestBeamId, v2.Current2ndBeamId, v2.SelectedBestBeamId, v2.Selected2ndBeamId}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate dlPreSchedData
				if mapEventRecord["dlPreSchedData"].Len() > 0 {
					p2 := p.findDlPreSched(mapEventRecord["dlFdSchedData"], mapEventRecord["dlTdSchedSubcellData"], mapEventRecord["dlPreSchedData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["dlPreSchedData"].Keys()[p2].(int)
						v2 := mapEventRecord["dlPreSchedData"].Val(k2).(*TtiDlPreSchedData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.CsListEvent, v2.HighestClassPriority, v2.PrachPreambleIndex}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate dlTdSchedSubcellData
				if mapEventRecord["dlTdSchedSubcellData"].Len() > 0 {
					p2 := p.findDlTdSched(mapEventRecord["dlFdSchedData"], mapEventRecord["dlTdSchedSubcellData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["dlTdSchedSubcellData"].Keys()[p2].(int)
						v2 := mapEventRecord["dlTdSchedSubcellData"].Val(k2).(*TtiDlTdSchedSubcellData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, fmt.Sprintf("(%d)[%s]", len(v2.Cs2List), strings.Join(v2.Cs2List, ";"))}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-"}...)
					}
				}

				// aggregate dlHarqRxData
				if mapEventRecord["dlHarqRxData"].Len() > 0 {
					p2 := p.findDlHarq(mapEventRecord["dlFdSchedData"], mapEventRecord["dlHarqRxData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["dlHarqRxData"].Keys()[p2].(string)
						v2 := mapEventRecord["dlHarqRxData"].Val(k2).(*TtiDlHarqRxData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.AckNack, v2.DlHarqProcessIndex, v2.PucchFormat}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate ulIntraDlToUlDrxSyncDlData
				if mapEventRecord["ulIntraDlToUlDrxSyncDlData"].Len() > 0 {
					p2 := p.findDlDrx(mapEventRecord["dlFdSchedData"], mapEventRecord["ulIntraDlToUlDrxSyncDlData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulIntraDlToUlDrxSyncDlData"].Keys()[p2].(int)
						v2 := mapEventRecord["ulIntraDlToUlDrxSyncDlData"].Val(k2).(*TtiUlIntraDlToUlDtxSyncDlData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.DrxEnabled, v2.DlDrxOnDurationTimerOn, v2.DlDrxInactivityTimerOn}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate csiSrReportData
				if mapEventRecord["csiSrReportData"].Len() > 0 {
					p2 := p.findCsiSrReport(mapEventRecord["dlFdSchedData"], mapEventRecord["csiSrReportData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["csiSrReportData"].Keys()[p2].(int)
						v2 := mapEventRecord["csiSrReportData"].Val(k2).(*TtiCsiSrReportData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.UlChannel, v2.Dtx, v2.PucchFormat, v2.Cqi, v2.PmiRank1, v2.PmiRank2, v2.Ri, v2.Cri, v2.Li, v2.Sr}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate dlLaDeltaCqi
				if mapEventRecord["dlLaDeltaCqi"].Len() > 0 {
					p2 := p.findDlLaDeltaCqi(mapEventRecord["dlFdSchedData"], mapEventRecord["dlLaDeltaCqi"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["dlLaDeltaCqi"].Keys()[p2].(string)
						v2 := mapEventRecord["dlLaDeltaCqi"].Val(k2).(*TtiDlLaDeltaCqi)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.IsDeltaCqiCalculated, v2.RrmPauseUeInDlScheduling, v2.HarqFb, v2.RrmDeltaCqi, v2.RrmRemainingBucketLevel}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate dlLaAverageCqi
				if mapEventRecord["dlLaAverageCqi"].Len() > 0 {
					p2 := p.findDlLaAvgCqi(mapEventRecord["dlFdSchedData"], mapEventRecord["dlLaAverageCqi"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["dlLaAverageCqi"].Keys()[p2].(int)
						v2 := mapEventRecord["dlLaAverageCqi"].Val(k2).(*TtiDlLaAverageCqi)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.RrmInstCqi, v2.Rank, v2.RrmAvgCqi, v2.Mcs, v2.RrmDeltaCqi}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate dlFlowControlData
				if mapEventRecord["dlFlowControlData"].Len() > 0 {
					p2 := p.findDlFlowControl(mapEventRecord["dlFdSchedData"], mapEventRecord["dlFlowControlData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["dlFlowControlData"].Keys()[p2].(int)
						v2 := mapEventRecord["dlFlowControlData"].Val(k2).(*TtiDlFlowControlData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.LchId, v2.ReportType, v2.ScheduledBytes, v2.EthAvg, v2.EthScaled}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
					}
				}
			} (p1)
		}
		wg.Wait()

		// output aggregated event: dlSchedAgg
		p.writeLog(zapcore.InfoLevel, "outputting aggregated dlSchedAgg...")
		headerWritten := make(map[string]bool)
		for _, k := range mapEventRecord["dlFdSchedData"].Keys() {
			data := mapEventRecord["dlFdSchedData"].Val(k).(*TtiDlFdSchedData)

			outFn := path.Join(outPath, fmt.Sprintf("dlSchedAgg_pci%s_rnti%s.csv", data.PhysCellId, data.Rnti))
			fout, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
			//defer fout3.Close()
			if err != nil {
				p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", outFn))
				break
			}

			if _, exist := headerWritten[outFn]; !exist {
				fout.WriteString(dlSchedAggFields)
				headerWritten[outFn] = true
			}

			fout.WriteString(fmt.Sprintf("%s,%s\n", data.Hsfn, strings.Join(data.AllFields, ",")))
			fout.Close()
		}
	}

	if p.ttiFilter == "ul" || p.ttiFilter == "both" {
		// update ulSchedAggFields
		// TODO - ulSchedAggFields
		p.writeLog(zapcore.InfoLevel, "updating fields for ulSchedAgg...") //core.QCoreApplication_Instance().ProcessEvents(0)
		if mapEventRecord["ulBsrRxData"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulBsr.hsfn", "ulBsr.sfn", "ulBsr.slot", "ulBsr.bsrFormat", "ulBsr.bufferSizeList"}, ",")
		}

		if mapEventRecord["ulPreSchedData"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulPreSched.hsfn", "ulPreSched.sfn", "ulPreSched.slot", "ulPreSched.csListEvent", "ulPreSched.highestClassPriority"}, ",")
		}

		if mapEventRecord["ulTdSchedSubcellData"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulTdSched.hsfn", "ulTdSched.sfn", "ulTdSched.slot", "ulTdSched.cs2List"}, ",")
		}

		if mapEventRecord["ulHarqRxData"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulHarq.hsfn", "ulHarq.sfn", "ulHarq.slot", "ulHarq.dtx", "ulHarq.crcResult", "ulHarq.ulHarqProcessIndex"}, ",")
		}

		if mapEventRecord["ulIntraDlToUlDrxSyncDlData"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"drx.hsfn", "drx.sfn", "drx.slot", "drx.drxEnabled", "drx.dlDrxOnDurationTimerOn", "drx.dlDrxInactivityTimerOn"}, ",")
		}

		if mapEventRecord["ulLaDeltaSinr"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulLaDeltaSinr.hsfn", "ulLaDeltaSinr.sfn", "ulLaDeltaSinr.slot", "ulLaDeltaSinr.isDeltaSinrCalculated", "ulLaDeltaSinr.rrmPauseUeInUlScheduling", "ulLaDeltaSinr.crcFb", "ulLaDeltaSinr.rrmDeltaSinr"}, ",")
		}

		if mapEventRecord["ulLaAverageSinr"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulLaAvgSinr.hsfn", "ulLaAvgSinr.sfn", "ulLaAvgSinr.slot", "ulLaAvgSinr.rrmInstSinrRank", "ulLaAvgSinr.rrmNumOfSinrMeasurements", "ulLaAvgSinr.rrmInstSinr", "ulLaAvgSinr.rrmSinrCorrection", "ulLaAvgSinr.rrmAvgSinrUl"}, ",")
		}

		if mapEventRecord["ulLaPhr"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulLaPhr.hsfn", "ulLaPhr.sfn", "ulLaPhr.slot", "ulLaPhr.isRrmPhrScaledCalculated", "ulLaPhr.phr", "ulLaPhr.rrmNumPuschPrb", "ulLaPhr.rrmPhrScaled"}, ",")
		}

		if mapEventRecord["ulPucchReceiveRespPsData"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulPucch.hsfn", "ulPucch.sfn", "ulPucch.slot", "ulPucch.pucchFormat", "ulPucch.startPrb", "ulPucch.rssi", "ulPucch.sinrLayer0", "ulPucch.sinrLayer1", "ulPucch.dtx", "ulPucch.srBit"}, ",")
		}

		if mapEventRecord["ulPuschReceiveRespPsData"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulPusch.hsfn", "ulPusch.sfn", "ulPusch.slot", "ulPusch.rssi", "ulPusch.sinrLayer0", "ulPusch.sinrLayer1", "ulPusch.dtx", "ulPusch.ulRank", "ulPusch.ulPmiRank1", "ulPusch.ulPmiRank1Sinr",  "ulPusch.ulPmiRank2", "ulPusch.ulPmiRank2SinrLayer0", "ulPusch.ulPmiRank2SinrLayer1", "ulPusch.longTermRank"}, ",")
		}

		if mapEventRecord["ulPduDemuxData"].Len() > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulPduDemux.hsfn", "ulPduDemux.sfn", "ulPduDemux.slot", "ulPduDemux.harqId", "ulPduDemux.isUlCcchData", "ulPduDemux.isTcpTraffic", "ulPduDemux.tempCrnti", "ulPduDemux.lcId", "ulPduDemux.rcvdBytes"}, ",")
		}

		ulSchedAggFields += "\n"

		// TODO - event aggregation with ulFdSchedData
		p.writeLog(zapcore.InfoLevel, "performing event aggregation for ulSchedAgg...[Time-consuming ops which may cause 100% CPU utilization!]")
		wg2 := &sync.WaitGroup{}
		for p1 := 0; p1 < mapEventRecord["ulFdSchedData"].Len(); p1 += 1 {
			wg2.Add(1)
			go func(p1 int) {
				defer wg2.Done()

				k1 := mapEventRecord["ulFdSchedData"].Keys()[p1].(int)
				v1 := mapEventRecord["ulFdSchedData"].Val(k1).(*TtiUlFdSchedData)

				// aggregate ulBsrRxData
				if mapEventRecord["ulBsrRxData"].Len() > 0 {
					p2 := p.findUlBsr(mapEventRecord["ulFdSchedData"], mapEventRecord["ulBsrRxData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulBsrRxData"].Keys()[p2].(int)
						v2 := mapEventRecord["ulBsrRxData"].Val(k2).(*TtiUlBsrRxData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.BsrFormat, fmt.Sprintf("[%s]", strings.Join(v2.BufferSizeList, ";"))}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate ulPreSchedData
				if mapEventRecord["ulPreSchedData"].Len() > 0 {
					p2 := p.findUlPreSched(mapEventRecord["ulFdSchedData"], mapEventRecord["ulPreSchedData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulPreSchedData"].Keys()[p2].(int)
						v2 := mapEventRecord["ulPreSchedData"].Val(k2).(*TtiUlPreSchedData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.CsListEvent, v2.HighestClassPriority}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate ulTdSchedSubcellData
				if mapEventRecord["ulTdSchedSubcellData"].Len() > 0 {
					p2 := p.findUlTdSched(mapEventRecord["ulFdSchedData"], mapEventRecord["ulTdSchedSubcellData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulTdSchedSubcellData"].Keys()[p2].(int)
						v2 := mapEventRecord["ulTdSchedSubcellData"].Val(k2).(*TtiUlTdSchedSubcellData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, fmt.Sprintf("(%d)[%s]", len(v2.Cs2List), strings.Join(v2.Cs2List, ";"))}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-"}...)
					}
				}

				// aggregate ulHarqRxData
				if mapEventRecord["ulHarqRxData"].Len() > 0 {
					p2 := p.findUlHarq(mapEventRecord["ulFdSchedData"], mapEventRecord["ulHarqRxData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulHarqRxData"].Keys()[p2].(int)
						v2 := mapEventRecord["ulHarqRxData"].Val(k2).(*TtiUlHarqRxData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.Dtx, v2.CrcResult, v2.UlHarqProcessIndex}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate ulIntraDlToUlDrxSyncDlData
				if mapEventRecord["ulIntraDlToUlDrxSyncDlData"].Len() > 0 {
					p2 := p.findUlDrx(mapEventRecord["ulFdSchedData"], mapEventRecord["ulIntraDlToUlDrxSyncDlData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulIntraDlToUlDrxSyncDlData"].Keys()[p2].(int)
						v2 := mapEventRecord["ulIntraDlToUlDrxSyncDlData"].Val(k2).(*TtiUlIntraDlToUlDtxSyncDlData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.DrxEnabled, v2.DlDrxOnDurationTimerOn, v2.DlDrxInactivityTimerOn}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate ulLaDeltaSinr
				if mapEventRecord["ulLaDeltaSinr"].Len() > 0 {
					p2 := p.findUlLaDeltaSinr(mapEventRecord["ulFdSchedData"], mapEventRecord["ulLaDeltaSinr"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulLaDeltaSinr"].Keys()[p2].(int)
						v2 := mapEventRecord["ulLaDeltaSinr"].Val(k2).(*TtiUlLaDeltaSinr)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.IsDeltaSinrCalculated, v2.RrmPauseUeInUlScheduling, v2.CrcFb, v2.RrmDeltaSinr}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate ulLaAverageSinr
				if mapEventRecord["ulLaAverageSinr"].Len() > 0 {
					p2 := p.findUlLaAvgSinr(mapEventRecord["ulFdSchedData"], mapEventRecord["ulLaAverageSinr"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulLaAverageSinr"].Keys()[p2].(int)
						v2 := mapEventRecord["ulLaAverageSinr"].Val(k2).(*TtiUlLaAverageSinr)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.RrmInstSinrRank, v2.RrmNumOfSinrMeasurements, v2.RrmInstSinr, v2.RrmSinrCorrection, v2.RrmAvgSinrUl}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate ulLaPhr
				if mapEventRecord["ulLaPhr"].Len() > 0 {
					p2 := p.findUlLaPhr(mapEventRecord["ulFdSchedData"], mapEventRecord["ulLaPhr"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulLaPhr"].Keys()[p2].(int)
						v2 := mapEventRecord["ulLaPhr"].Val(k2).(*TtiUlLaPhr)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.IsRrmPhrScaledCalculated, v2.Phr, v2.RrmNumPuschPrb, v2.RrmPhrScaled}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate ulPucchReceiveRespPsData
				if mapEventRecord["ulPucchReceiveRespPsData"].Len() > 0 {
					p2 := p.findUlPucch(mapEventRecord["ulFdSchedData"], mapEventRecord["ulPucchReceiveRespPsData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulPucchReceiveRespPsData"].Keys()[p2].(int)
						v2 := mapEventRecord["ulPucchReceiveRespPsData"].Val(k2).(*TtiUlPucchReceiveRespPsData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.PucchFormat, v2.StartPrb, v2.Rssi, v2.SinrLayer0, v2.SinrLayer1, v2.Dtx, v2.SrBit}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate ulPuschReceiveRespPsData
				if mapEventRecord["ulPuschReceiveRespPsData"].Len() > 0 {
					p2 := p.findUlPusch(mapEventRecord["ulFdSchedData"], mapEventRecord["ulPuschReceiveRespPsData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulPuschReceiveRespPsData"].Keys()[p2].(int)
						v2 := mapEventRecord["ulPuschReceiveRespPsData"].Val(k2).(*TtiUlPuschReceiveRespPsData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.Rssi, v2.SinrLayer0, v2.SinrLayer1, v2.Dtx, v2.UlRank, v2.UlPmiRank1, v2.UlPmiRank1Sinr, v2.UlPmiRank2, v2.UlPmiRank2SinrLayer0, v2.UlPmiRank2SinrLayer1, v2.LongTermRank}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
					}
				}

				// aggregate ulPduDemuxData
				if mapEventRecord["ulPduDemuxData"].Len() > 0 {
					p2 := p.findUlPduDemux(mapEventRecord["ulFdSchedData"], mapEventRecord["ulPduDemuxData"], p1)
					if p2 >= 0 {
						k2 := mapEventRecord["ulPduDemuxData"].Keys()[p2].(int)
						v2 := mapEventRecord["ulPduDemuxData"].Val(k2).(*TtiUlPduDemuxData)

						v1.AllFields = append(v1.AllFields, []string{v2.TtiEventHeader.Hsfn, v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.HarqId, v2.IsUlCcchData, v2.IsTcpTraffic, v2.TempCrnti, fmt.Sprintf("[%s]", strings.Join(v2.LcIdList, ";")), fmt.Sprintf("[%s]", strings.Join(v2.RcvdBytesList, ";"))}...)
					} else {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
					}
				}
			} (p1)
		}
		wg2.Wait()

		// output aggregated event: ulSchedAgg
		p.writeLog(zapcore.InfoLevel, "outputting aggregated ulSchedAgg...")
		headerWritten2 := make(map[string]bool)
		for _, k := range mapEventRecord["ulFdSchedData"].Keys() {
			data := mapEventRecord["ulFdSchedData"].Val(k).(*TtiUlFdSchedData)

			outFn := path.Join(outPath, fmt.Sprintf("ulSchedAgg_pci%s_rnti%s.csv", data.PhysCellId, data.Rnti))
			fout, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
			//defer fout3.Close()
			if err != nil {
				p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", outFn))
				break
			}

			if _, exist := headerWritten2[outFn]; !exist {
				fout.WriteString(ulSchedAggFields)
				headerWritten2[outFn] = true
			}

			fout.WriteString(fmt.Sprintf("%s,%s\n", data.Hsfn, strings.Join(data.AllFields, ",")))
			fout.Close()
		}
	}
}

func (p *L2TtiTraceParser) makeTimeStamp(hsfn, sfn, slot int) int {
	return 1024 * p.slotsPerRf * hsfn + p.slotsPerRf * sfn + slot
}

func (p *L2TtiTraceParser) unsafeAtoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func (p *L2TtiTraceParser) ttiDlPreSchedClassPriority(cp string) string {
	// TODO fix classPriority for 5G21A
	classPriority := []string {"rachMsg2", "harqRetxMsg4", "harqRetxSrb1", "harqRetxSrb3", "harqRetxSrb2", "harqRetxVoip", "harqRetxDrb", "dlMacCe", "srb1Traffic", "srb3Traffic", "srb2Traffic", "voipTraffic", "drbTraffic", "deprioritizedVoip", "lastUnUsed"}

	return fmt.Sprintf("%s(%s)", cp, classPriority[p.unsafeAtoi(cp)])
}

func (p *L2TtiTraceParser) ttiUlPreSchedClassPriority(cp string) string {
	// TODO fix classPriority for 5G21A
	classPriority := []string {"rachMsg3", "ulGrantContRes", "harqRetxMsg3", "harqRetxSrb", "harqRetxVoip", "harqRetxDrb", "ulGrantSr", "srbTraffic", "voipTraffic", "ulGrantTa", "drbTraffic", "deprioritizedVoip", "ulProSched", "lastUnUsed", "unknown"}

	return fmt.Sprintf("%s(%s)", cp, classPriority[p.unsafeAtoi(cp)])
}

func (p *L2TtiTraceParser) findDlBeam(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlBeamData)

		if k2 <= k1 {
			//if v1.PhysCellId+v1.Rnti+v1.SubcellId == v2.PhysCellId+v2.Rnti+v2.SubcellId {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findDlPreSched(m1,m2,m3 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	// find possible TD-scheduling opportunity
	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlTdSchedSubcellData)

		if k2 <= k1 {
			//if v1.PhysCellId+v1.SubcellId == v2.PhysCellId+v2.SubcellId  && p.contains(v2.Cs2List, v1.Rnti) {
			if v1.PhysCellId == v2.PhysCellId  && p.contains(v2.Cs2List, v1.Rnti) {
				p2 = i
			}
		} else {
			break
		}
	}

	if p2 < 0 {
		return p2
	}

	// find possible Pre-scheduling opportunity
	k1 = m2.Keys()[p2].(int)
	p3 := -1
	for i := 0; i < m3.Len(); i += 1 {
		k2 := m3.Keys()[i].(int)
		v2 := m3.Val(k2).(*TtiDlPreSchedData)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p3 = i
			}
		} else {
			break
		}
	}

	return p3
}

func (p *L2TtiTraceParser) findDlTdSched(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlTdSchedSubcellData)

		if k2 <= k1 {
			//if v1.PhysCellId+v1.SubcellId == v2.PhysCellId+v2.SubcellId  && p.contains(v2.Cs2List, v1.Rnti) {
			if v1.PhysCellId == v2.PhysCellId  && p.contains(v2.Cs2List, v1.Rnti) {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findDlHarq(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)
	hsfn, sfn, slot := p.incSlot(p.unsafeAtoi(v1.Hsfn), p.unsafeAtoi(v1.Sfn), p.unsafeAtoi(v1.Slot), p.unsafeAtoi(v1.K1))
	harq := p.makeTimeStamp(hsfn, sfn, slot)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		// key = "Timestamp_HarqProcessIndex"
		k2 := m2.Keys()[i].(string)
		k2ts := p.unsafeAtoi(strings.Split(k2, "_")[0])
		v2 := m2.Val(k2).(*TtiDlHarqRxData)

		if k2ts == harq {
			//if v1.PhysCellId+v1.Rnti+v1.SubcellId+v1.DlHarqProcessIndex == v2.PhysCellId+v2.Rnti+v2.HarqSubcellId+v2.DlHarqProcessIndex {
			if v1.PhysCellId+v1.Rnti+v1.DlHarqProcessIndex == v2.PhysCellId+v2.Rnti+v2.DlHarqProcessIndex {
				p2 = i
				break
			}
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findDlLaAvgCqi(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlLaAverageCqi)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti+v1.CellDbIndex == v2.PhysCellId+v2.Rnti+v2.CellDbIndex {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findCsiSrReport(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiCsiSrReportData)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findDlFlowControl(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlFlowControlData)

		if k2 <= k1 {
			if v1.Rnti == v2.Rnti && p.contains(v1.LcIdList, v2.LchId) {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findDlLaDeltaCqi(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(string)
		k2ts := p.unsafeAtoi(strings.Split(k2, "_")[0])
		v2 := m2.Val(k2).(*TtiDlLaDeltaCqi)

		if k2ts <= k1 {
			if v1.PhysCellId+v1.Rnti+v1.CellDbIndex == v2.PhysCellId+v2.Rnti+v2.CellDbIndex {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlBsr(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlBsrRxData)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti+v1.UlHarqProcessIndex == v2.PhysCellId+v2.Rnti+v2.UlHarqProcessIndex {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlHarq(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlHarqRxData)

		//if k2 >= k1 && v1.PhysCellId+v1.Rnti+v1.SubcellId+v1.UlHarqProcessIndex == v2.PhysCellId+v2.Rnti+v2.SubcellId+v2.UlHarqProcessIndex {
		if k2 >= k1 && v1.PhysCellId+v1.Rnti+v1.UlHarqProcessIndex == v2.PhysCellId+v2.Rnti+v2.UlHarqProcessIndex {
			p2 = i
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlDrx(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlIntraDlToUlDtxSyncDlData)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findDlDrx(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlIntraDlToUlDtxSyncDlData)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlLaDeltaSinr(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlLaDeltaSinr)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti+v1.CellDbIndex == v2.PhysCellId+v2.Rnti+v2.CellDbIndex {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlLaAvgSinr(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlLaAverageSinr)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti+v1.CellDbIndex == v2.PhysCellId+v2.Rnti+v2.CellDbIndex {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlLaPhr(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlLaPhr)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti+v1.CellDbIndex == v2.PhysCellId+v2.Rnti+v2.CellDbIndex {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlPucch(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlPucchReceiveRespPsData)

		if k2 <= k1 {
			//if v1.PhysCellId+v1.Rnti+v1.SubcellId == v2.PhysCellId+v2.Rnti+v2.SubcellId {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlPusch(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlPuschReceiveRespPsData)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlPduDemux(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlPduDemuxData)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti+v1.UlHarqProcessIndex == v2.PhysCellId+v2.Rnti+v2.HarqId {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlPreSched(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)
	hsfn, sfn, slot := p.decSlot(p.unsafeAtoi(v1.Hsfn), p.unsafeAtoi(v1.Sfn), p.unsafeAtoi(v1.Slot), p.unsafeAtoi(v1.K2))
	dci := p.makeTimeStamp(hsfn, sfn, slot)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlPreSchedData)

		if k2 == dci {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
				break
			}
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlTdSched(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)
	hsfn, sfn, slot := p.decSlot(p.unsafeAtoi(v1.Hsfn), p.unsafeAtoi(v1.Sfn), p.unsafeAtoi(v1.Slot), p.unsafeAtoi(v1.K2))
	dci := p.makeTimeStamp(hsfn, sfn, slot)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlTdSchedSubcellData)

		if k2 == dci {
			//if v1.PhysCellId+v1.SubcellId == v2.PhysCellId+v2.SubcellId  && p.contains(v2.Cs2List, v1.Rnti) {
			if v1.PhysCellId == v2.PhysCellId  && p.contains(v2.Cs2List, v1.Rnti) {
				p2 = i
				break
			}
		}
	}

	return p2
}

func (p *L2TtiTraceParser) contains(a []string, b string) bool {
	for _, s := range a {
		if s == b {
			return true
		}
	}

	return false
}

func (p *L2TtiTraceParser) incSlot(hsfn, sfn, slot, n int) (int, int, int) {
	slot += n
	if slot >= p.slotsPerRf {
		slot %= p.slotsPerRf
		sfn += 1
	}

	if sfn >= 1024 {
		sfn %= 1024
		hsfn += 1
	}

	return hsfn, sfn, slot
}

func (p *L2TtiTraceParser) decSlot(hsfn, sfn, slot, n int) (int, int, int) {
	slot -= n
	if slot < 0 {
		slot += p.slotsPerRf
		sfn -= 1
	}

	if sfn < 0 {
		sfn += 1024
		hsfn -= 1
	}

	return hsfn, sfn, slot
}

func (p *L2TtiTraceParser) initPdschSliv() map[string][]int {
	// prefix
	// "00": mapping type A + normal cp
	// "10": mapping type B + normal cp
	pdschFromSliv := make(map[string][]int)
	var prefix string

	// case #1: prefix="00"
	prefix = "00"
	for _, S := range utils.PyRange(0, 4, 1) {
		for _, L := range utils.PyRange(3, 15, 1) {
			if S+L >= 3 && S+L < 15 {
				sliv, err := p.makeSliv(S, L)
				if err == nil {
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					pdschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	// case #3: prefix="10"
	prefix = "10"
	for _, S := range utils.PyRange(0, 13, 1) {
		for _, L := range []int{2, 4, 7} {
			if S+L >= 2 && S+L < 15 {
				sliv, err := p.makeSliv(S, L)
				if err == nil {
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					pdschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	return pdschFromSliv
}

func (p *L2TtiTraceParser) initPuschSliv() map[string][]int {
	// prefix
	// "00": mapping type A + normal cp
	// "10": mapping type B + normal cp
	puschFromSliv := make(map[string][]int)
	var prefix string

	// case #1: prefix="00"
	prefix = "00"
	for _, S := range utils.PyRange(0, 1, 1) {
		for _, L := range utils.PyRange(4, 15, 1) {
			if S+L >= 4 && S+L < 15 {
				sliv, err := p.makeSliv(S, L)
				if err == nil {
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					puschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	// case #3: prefix="10"
	prefix = "10"
	for _, S := range utils.PyRange(0, 14, 1) {
		for _, L := range utils.PyRange(1, 15, 1) {
			if S+L >= 1 && S+L < 15 {
				sliv, err := p.makeSliv(S, L)
				if err == nil {
					keyFromSliv := fmt.Sprintf("%s_%d", prefix, sliv)
					puschFromSliv[keyFromSliv] = []int{S, L}
				}
			}
		}
	}

	return puschFromSliv
}

func (p *L2TtiTraceParser) makeSliv(S, L int) (int, error) {
	if L <= 0 || L > 14-S {
		return -1, errors.New(fmt.Sprintf("invalid S/L combination: S=%d, L=%d", S, L))
	}

	sliv := 0
	if (L - 1) <= 7 {
		sliv = 14 * (L-1) + S
	} else {
		sliv = 14 * (14-L+1) + (14-1-S)
	}

	return sliv, nil
}

func (p *L2TtiTraceParser) makeRiv(numPrb, startPrb, bwpSize int) (int, error) {
	if numPrb < 1 || numPrb > (bwpSize-startPrb) {
		return -1, errors.New(fmt.Sprintf("Invalid numPrb/startPrb combination: numPrb=%d, startPrb=%d", numPrb, startPrb))
	}

	riv := 0
	if (numPrb-1) <=  int(math.Floor(float64(bwpSize)/2)) {
		riv = bwpSize * (numPrb-1) + startPrb
	} else {
		riv = bwpSize * (bwpSize-numPrb+1) + (bwpSize-1-startPrb)
	}

	return riv, nil
}

func (p *L2TtiTraceParser) writeLog(level zapcore.Level, s string) {
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
