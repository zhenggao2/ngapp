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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type L2TtiTraceParser struct {
	log          *zap.Logger
	ttiTracePath string
	ttiPattern   string
	ttiRat       string
	ttiScs       string
	ttiFilter    string
	maxgo        int
	debug        bool
	nok21a       bool

	slotsPerRf int
	ttiFiles []string
}

type SfnInfo struct {
	lastSfn int
	hsfn int
}

func (p *L2TtiTraceParser) Init(log *zap.Logger, trace, pattern, rat, scs, filter string, maxgo int, debug bool) {
	p.log = log
	p.ttiTracePath = trace
	p.ttiPattern = strings.ToLower(pattern)
	p.ttiRat = strings.ToLower(rat)
	p.ttiScs = strings.ToLower(scs)
	p.ttiFilter = strings.ToLower(filter)
	p.maxgo = maxgo
	p.debug = debug
	p.nok21a = true

	fileInfo, err := ioutil.ReadDir(p.ttiTracePath)
	if err != nil {
		p.writeLog(zapcore.FatalLevel, fmt.Sprintf("Fail to read directory: %s.", p.ttiTracePath))
		return
	}

	for _, file := range fileInfo {
	    if !file.IsDir() && filepath.Ext(file.Name()) == p.ttiPattern {
			p.ttiFiles = append(p.ttiFiles, filepath.Join(p.ttiTracePath, file.Name()))
		}
	}

	p.writeLog(zapcore.InfoLevel, fmt.Sprintf("Initializing L2 TTI trace parser...(working dir: %v)", trace))
}

// func (p *L2TtiTraceParser) onOkBtnClicked(checked bool) {
func (p *L2TtiTraceParser) Exec() {
	scs2nslots := map[string]int{"15k": 10, "30k": 20, "120k": 80}
	p.slotsPerRf = scs2nslots[strings.ToLower(p.ttiScs)]

	// recreate dir for parsed l2 tti trace
	outPath := filepath.Join(p.ttiTracePath, "parsed_tti")
	outPath2 := filepath.Join(p.ttiTracePath, "parsed_tti", "per_events")
	os.RemoveAll(outPath)
	if err := os.MkdirAll(outPath2, 0775); err != nil {
		panic(fmt.Sprintf("Fail to create directory: %v", err))
	}

	// key=EventName_PCI_RNTI or EventName for both mapFieldName and mapSfnInfo
	mapFieldName := make(map[string][]string)
	eventId := 0

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
	var mapEventRecord = map[string]map[string]*utils.OrderedMap {
		"dlBeamData":           make(map[string]*utils.OrderedMap),
		"dlPreSchedData":       make(map[string]*utils.OrderedMap),
		"dlTdSchedSubcellData": make(map[string]*utils.OrderedMap),
		"dlFdSchedData":        make(map[string]*utils.OrderedMap),
		"dlHarqRxData":         make(map[string]*utils.OrderedMap),
		"dlLaAverageCqi":       make(map[string]*utils.OrderedMap),
		"csiSrReportData":       make(map[string]*utils.OrderedMap),
		"dlFlowControlData":       make(map[string]*utils.OrderedMap),
		"dlLaDeltaCqi":       make(map[string]*utils.OrderedMap),
		"ulBsrRxData":       make(map[string]*utils.OrderedMap),
		"ulPreSchedData":       make(map[string]*utils.OrderedMap),
		"ulTdSchedSubcellData": make(map[string]*utils.OrderedMap),
		"ulFdSchedData":       make(map[string]*utils.OrderedMap),
		"ulHarqRxData":       make(map[string]*utils.OrderedMap),
		"ulIntraDlToUlDrxSyncDlData":       make(map[string]*utils.OrderedMap),
		"ulLaDeltaSinr":       make(map[string]*utils.OrderedMap),
		"ulLaAverageSinr":       make(map[string]*utils.OrderedMap),
		"ulLaPhr":       make(map[string]*utils.OrderedMap),
		"ulPucchReceiveRespPsData":       make(map[string]*utils.OrderedMap),
		"ulPuschReceiveRespPsData":       make(map[string]*utils.OrderedMap),
		"ulPduDemuxData":       make(map[string]*utils.OrderedMap),
	}
	var dlSchedAggFields string
	var ulSchedAggFields string
	var dlPerBearerProcessed bool
	var mapPdschSliv = p.initPdschSliv()
	var mapPuschSliv = p.initPuschSliv()

	var mapUlFdUes = make(map[string][]string)
	var mapDlFdUes = make(map[string][]string)

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
		p.writeLog(zapcore.InfoLevel, fmt.Sprintf("  parsing: %s", fn))

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

						outFn := filepath.Join(outPath2, fmt.Sprintf("%s.csv", key))
						if _, exist := mapFieldName[key]; !exist {
							mapFieldName[key] = make([]string, valStart)
							copy(mapFieldName[key], tokens[:valStart])

							// Step-1: write event header only once
							fout, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
							//defer fout.Close()
							if err != nil {
								p.writeLog(zapcore.ErrorLevel, fmt.Sprintf("Fail to open file: %s", outFn))
								break
							}

							// Step-1.1: write eventId header field
							fout.WriteString("eventId,")

							row := strings.Join(mapFieldName[key], ",")
							fout.WriteString(fmt.Sprintf("%s\n", row))
							fout.Close()

							// update dlSchedAggFields
							if len(dlSchedAggFields) == 0 && eventName == "dlFdSchedData" {
								dlSchedAggFields = "eventId," + row
							}

							// update ulSchedAggFields
							if len(ulSchedAggFields) == 0 && eventName == "ulFdSchedData" {
								ulSchedAggFields = "eventId," + row
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
						fout2.WriteString(fmt.Sprintf("%v,", eventId))

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

							v := TtiDlBeamData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:   eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "dlPreSchedData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlPreSchedData
							if posDlPreSched.Ready == false {
								posDlPreSched = FindTtiDlPreSchedDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlPreSched=%v", posDlPreSched))
								}
							}

							v := TtiDlPreSchedData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
									Sfn:        tokens[valStart+posDlPreSched.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posDlPreSched.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posDlPreSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posDlPreSched.PosEventHeader.PosPhysCellId],
								},

								CsListEvent:          tokens[valStart+posDlPreSched.PosCsListEvent],
								HighestClassPriority: p.ttiDlPreSchedClassPriority(tokens[valStart+posDlPreSched.PosHighestClassPriority]),
								PrachPreambleIndex: tokens[valStart+posDlPreSched.PosPrachPreambleIndex],
							}

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "dlTdSchedSubcellData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlTdSchedSubcellData
							if posDlTdSched.Ready == false {
								posDlTdSched = FindTtiDlTdSchedSubcellDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlTdSched=%v", posDlTdSched))
								}
							}

							v := TtiDlTdSchedSubcellData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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
									posRnti := valStart + rsn + 1 + 3 * k
									if posRnti > len(tokens) || len(tokens[posRnti]) == 0 {
										break
									}
									v.Cs2List = append(v.Cs2List, tokens[posRnti])
								}
							}

							k2 := v.TtiEventHeader.PhysCellId
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "dlFdSchedData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlFdSchedData
							if posDlFdSched.Ready == false {
								posDlFdSched = FindTtiDlFdSchedDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlFdSched=%v", posDlFdSched))
								}
							}

							v := TtiDlFdSchedData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							// count number of FD-Scheduled UEs
							kfu := strings.Join([]string{v.Sfn, v.Slot, v.PhysCellId}, "_")
							if _, e := mapDlFdUes[kfu]; !e {
								mapDlFdUes[kfu] = []string{fmt.Sprintf("%v_%v", eventId, v.Rnti)}
							} else {
								mapDlFdUes[kfu] = append(mapDlFdUes[kfu], fmt.Sprintf("%v_%v", eventId, v.Rnti))
							}

							if len(mapDlFdUes[kfu]) > 4 {
								p.writeLog(zapcore.DebugLevel, fmt.Sprintf("key=%v, %v_%v_%v_%v: size of mapDlFdUes=%v", key, v.eventId, v.Sfn, v.Slot, v.PhysCellId, len(mapDlFdUes[kfu])))
							}

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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
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
								v := TtiDlHarqRxData{
									// event header
									TtiEventHeader: TtiEventHeader{
										eventId:    eventId,
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

								k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
								if _, e := mapEventRecord["dlHarqRxData"][k2]; !e {
									mapEventRecord["dlHarqRxData"][k2] = utils.NewOrderedMap()
								}
								mapEventRecord["dlHarqRxData"][k2].Add(eventId, &v)
							} else {
								maxNumHarq := 32	// max 32 HARQ feedbacks per dlHarqRxDataArray
								sizeDlHarqRecord := 14
								for ih := 0; ih < maxNumHarq; ih += 1 {
									posRnti := valStart+posDlHarq.PosEventHeader.PosRnti+ih*sizeDlHarqRecord
									if posRnti >= len(tokens) || len(tokens[posRnti]) == 0 {
										break
									}

									v := TtiDlHarqRxData{
										// event header
										TtiEventHeader: TtiEventHeader{
											//eventId:    strconv.Itoa(mapSfnInfo[key].hsfn),
											eventId:   eventId,
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

									k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
									if _, e := mapEventRecord["dlHarqRxData"][k2]; !e {
										mapEventRecord["dlHarqRxData"][k2] = utils.NewOrderedMap()
									}
									mapEventRecord["dlHarqRxData"][k2].Add(eventId, &v)
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

							v := TtiDlLaAverageCqi{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "csiSrReportData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - csiSrReportData
							if posCsiSrReport.Ready == false {
								posCsiSrReport = FindTtiCsiSrReportDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posCsiSrReport=%v", posCsiSrReport))
								}
							}

							v := TtiCsiSrReportData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "dlFlowControlData" && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
							// TODO - event aggregation - dlFlowControlData
							if posDlFlowControl.Ready == false {
								posDlFlowControl = FindTtiDlFlowControlDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDlFlowControl=%v", posDlFlowControl))
								}
							}

							v := TtiDlFlowControlData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if (eventName == "dlLaDeltaCqi" || eventName == "dlLaDeltaCqiArray") && (p.ttiFilter == "dl" || p.ttiFilter == "both") {
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
								v := TtiDlLaDeltaCqi{
									// event header
									TtiEventHeader: TtiEventHeader{
										eventId:    eventId,
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

								k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
								if _, e := mapEventRecord["dlLaDeltaCqi"][k2]; !e {
									mapEventRecord["dlLaDeltaCqi"][k2] = utils.NewOrderedMap()
								}
								mapEventRecord["dlLaDeltaCqi"][k2].Add(eventId, &v)
							} else {
								maxNumDlOlqc := 64	// max 64 DL LA deltaCqi records per dlHarqRxDataArray
								sizeDlOlqcRecord := 8
								for ih := 0; ih < maxNumDlOlqc; ih += 1 {
									posRnti := valStart+posDlLaDeltaCqi.PosEventHeader.PosRnti+ih*sizeDlOlqcRecord
									if posRnti >= len(tokens) || len(tokens[posRnti]) == 0 {
										break
									}

									v := TtiDlLaDeltaCqi{
										// event header
										TtiEventHeader: TtiEventHeader{
											eventId:    eventId,
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

									k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
									if _, e := mapEventRecord["dlLaDeltaCqi"][k2]; !e {
										mapEventRecord["dlLaDeltaCqi"][k2] = utils.NewOrderedMap()
									}
									mapEventRecord["dlLaDeltaCqi"][k2].Add(eventId, &v)
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

							v := TtiUlBsrRxData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						}  else if eventName == "ulPreSchedData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulPreSchedData
							if posUlPreSched.Ready == false {
								posUlPreSched = FindTtiUlPreSchedDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlPreSched=%v", posUlPreSched))
								}
							}

							v := TtiUlPreSchedData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
									Sfn:        tokens[valStart+posUlPreSched.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posUlPreSched.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posUlPreSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posUlPreSched.PosEventHeader.PosPhysCellId],
								},

								CsListEvent:          tokens[valStart+posUlPreSched.PosCsListEvent],
								HighestClassPriority: p.ttiUlPreSchedClassPriority(tokens[valStart+posUlPreSched.PosHighestClassPriority]),
							}

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "ulTdSchedSubcellData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulTdSchedSubcellData
							if posUlTdSched.Ready == false {
								posUlTdSched = FindTtiUlTdSchedSubcellDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlTdSched=%v", posUlTdSched))
								}
							}

							v := TtiUlTdSchedSubcellData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "ulFdSchedData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulFdSchedData
							if posUlFdSched.Ready == false {
								posUlFdSched = FindTtiUlFdSchedDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlFdSched=%v", posUlFdSched))
								}
							}

							v := TtiUlFdSchedData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							// count number of FD-Scheduled UEs
							kfu := strings.Join([]string{v.Sfn, v.Slot, v.PhysCellId}, "_")
							if _, e := mapUlFdUes[kfu]; !e {
								mapUlFdUes[kfu] = []string{fmt.Sprintf("%v_%v", eventId, v.Rnti)}
							} else {
								mapUlFdUes[kfu] = append(mapUlFdUes[kfu], fmt.Sprintf("%v_%v", eventId, v.Rnti))
							}

							if len(mapUlFdUes[kfu]) > 4 {
								p.writeLog(zapcore.DebugLevel, fmt.Sprintf("key=%v, %v_%v_%v_%v: size of mapUlFdUes=%v", key, v.eventId, v.Sfn, v.Slot, v.PhysCellId, len(mapUlFdUes[kfu])))
							}

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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "ulHarqRxData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulHarqRxData
							if posUlHarq.Ready == false {
								posUlHarq = FindTtiUlHarqRxDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlHarq=%v", posUlHarq))
								}
							}

							v := TtiUlHarqRxData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "ulIntraDlToUlDrxSyncDlData" {
							// TODO - event aggregation - ulIntraDlToUlDrxSyncDlData
							if posDrx.Ready == false {
								posDrx = FindTtiUlIntraDlToUlDrxSyncDlDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posDrx=%v", posDrx))
								}
							}

							v := TtiUlIntraDlToUlDtxSyncDlData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
									Sfn:        tokens[valStart+posDrx.PosEventHeader.PosSfn],
									Slot:       tokens[valStart+posDrx.PosEventHeader.PosSlot],
									Rnti:       tokens[valStart+posDrx.PosEventHeader.PosRnti],
									PhysCellId: tokens[valStart+posDrx.PosEventHeader.PosPhysCellId],
								},

								DrxEnabled: tokens[valStart+posDrx.PosDrxEnabled],
								DlDrxOnDurationTimerOn: tokens[valStart+posDrx.PosDlDrxOnDurationTimerOn],
								DlDrxInactivityTimerOn: tokens[valStart+posDrx.PosDlDrxInactivityTimerOn],
							}

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "ulLaDeltaSinr" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulLaDeltaSinr
							if posUlLaDeltaSinr.Ready == false {
								posUlLaDeltaSinr = FindTtiUlLaDeltaSinrPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlLaDeltaSinr=%v", posUlLaDeltaSinr))
								}
							}

							v := TtiUlLaDeltaSinr{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "ulLaAverageSinr" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulLaAverageSinr
							if posUlLaAvgSinr.Ready == false {
								posUlLaAvgSinr = FindTtiUlLaAverageSinrPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlLaAvgSinr=%v", posUlLaAvgSinr))
								}
							}

							v := TtiUlLaAverageSinr{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "ulLaPhr" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulLaPhr
							if posUlLaPhr.Ready == false {
								posUlLaPhr = FindTtiUlLaPhrPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlLaPhr=%v", posUlLaPhr))
								}
							}

							v := TtiUlLaPhr{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "ulPucchReceiveRespPsData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulPucchReceiveRespPsData
							if posUlPucch.Ready == false {
								posUlPucch = FindTtiUlPucchReceiveRespPsDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlPucch=%v", posUlPucch))
								}
							}

							v := TtiUlPucchReceiveRespPsData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "ulPuschReceiveRespPsData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulPuschReceiveRespPsData
							if posUlPusch.Ready == false {
								posUlPusch = FindTtiUlPuschReceiveRespPsDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlPusch=%v", posUlPusch))
								}
							}

							v := TtiUlPuschReceiveRespPsData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						} else if eventName == "ulPduDemuxData" && (p.ttiFilter == "ul" || p.ttiFilter == "both") {
							// TODO - event aggregation - ulPduDemuxData
							if posUlPduDemux.Ready == false {
								posUlPduDemux = FindTtiUlPduDemuxDataPos(tokens)
								if p.debug {
									p.writeLog(zapcore.DebugLevel, fmt.Sprintf("posUlPduDemux=%v", posUlPduDemux))
								}
							}

							v := TtiUlPduDemuxData{
								// event header
								TtiEventHeader: TtiEventHeader{
									eventId:    eventId,
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

							k2 := v.TtiEventHeader.PhysCellId + "_" + v.TtiEventHeader.Rnti
							if _, e := mapEventRecord[eventName][k2]; !e {
								mapEventRecord[eventName][k2] = utils.NewOrderedMap()
							}
							mapEventRecord[eventName][k2].Add(eventId, &v)
						}

						eventId++
					} else {
						p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Invalid event record detected: %s", line))
					}
				} else {
					p.writeLog(zapcore.DebugLevel, fmt.Sprintf("Invalid event data detected: %s", line))
				}
			}
		}

		fin.Close()
	}

	/*
	for event := range mapEventRecord {
		for dn := range mapEventRecord[event] {
			for _, ts := range mapEventRecord[event][dn].Keys() {
				p.writeLog(zapcore.DebugLevel, fmt.Sprintf("event=%v,dn=%v,k=%v, v=%v\n", event, dn, ts, mapEventRecord[event][dn].Val(ts)))
			}
		}
	}
	 */

	if p.ttiFilter == "dl" || p.ttiFilter == "both" {
		// update dlSchedAggFields
		// TODO - dlSchedAggFields
		p.writeLog(zapcore.InfoLevel, "updating fields for dlSchedAgg...") //core.QCoreApplication_Instance().ProcessEvents(0)

		// Print number of FD-scheduled UEs
		dlSchedAggFields += ",nbrFdUes,fdUeRntis"

		if len(mapEventRecord["dlBeamData"]) > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlBeam.eventId", "dlBeam.sfn", "dlBeam.slot", "dlBeam.currentBestBeamId", "dlBeam.current2ndBeamId", "dlBeam.selectedBestBeamId", "dlBeam.selected2ndBeamid"}, ",")
		}

		if len(mapEventRecord["dlPreSchedData"]) > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlPreSched.eventId", "dlPreSched.sfn", "dlPreSched.slot", "dlPreSched.csListEvent", "dlPreSched.highestClassPriority", "dlPreSched.prachPreambleIndex"}, ",")
		}

		if len(mapEventRecord["dlTdSchedSubcellData"]) > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlTdSched.eventId", "dlTdSched.sfn", "dlTdSched.slot", "dlTdSched.cs2List"}, ",")
		}

		if len(mapEventRecord["dlHarqRxData"]) > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlHarq.eventId", "dlHarq.sfn", "dlHarq.slot", "dlHarq.AckNack", "dlHarq.dlHarqProcessIndex", "dlHarq.pucchFormat"}, ",")
		}

		if len(mapEventRecord["ulIntraDlToUlDrxSyncDlData"]) > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"drx.eventId", "drx.sfn", "drx.slot", "drx.drxEnabled", "drx.dlDrxOnDurationTimerOn", "drx.dlDrxInactivityTimerOn"}, ",")
		}

		if len(mapEventRecord["csiSrReportData"]) > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"csiSrReport.eventId", "csiSrReport.sfn", "csiSrReport.slot", "csiSrReport.ulChannel", "csiSrReport.dtx", "csiSrReport.pucchFormat", "csiSrReport.cqi", "csiSrReport.pmiRank1", "csiSrReport.pmiRank2", "csiSrReport.ri", "csiSrReport.cri", "csiSrReport.li", "csiSrReport.sr"}, ",")
		}

		if len(mapEventRecord["dlLaDeltaCqi"]) > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlLaDeltaCqi.eventId", "dlLaDeltaCqi.sfn", "dlLaDeltaCqi.slot", "dlLaDeltaCqi.isDeltaCqiCalculated", "dlLaDeltaCqi.rrmPauseUeInDlScheduling", "dlLaDeltaCqi.harqFb", "dlLaDeltaCqi.rrmDeltaCqi", "dlLaDeltaCqi.rrmRemainingBucketLevel"}, ",")
		}

		if len(mapEventRecord["dlLaAverageCqi"]) > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlLaAvgCqi.eventId", "dlLaAvgCqi.sfn", "dlLaAvgCqi.slot", "dlLaAvgCqi.rrmInstCqi", "dlLaAvgCqi.rank", "dlLaAvgCqi.rrmAvgCqi", "dlLaAvgCqi.mcs", "dlLaAvgCqi.rrmDeltaCqi"}, ",")
		}

		if len(mapEventRecord["dlFlowControlData"]) > 0 {
			dlSchedAggFields += ","
			dlSchedAggFields += strings.Join([]string{"dlFlowControl.eventId", "dlFlowControl.sfn", "dlFlowControl.slot", "dlFlowControl.lchId", "dlFlowControl.reportType", "dlFlowControl.scheduledBytes", "dlFlowControl.ethAvg", "dlFlowControl.ethScaled"}, ",")
		}

		dlSchedAggFields += "\n"

		// perform event aggregation
		// TODO - event aggregation with dlFdSchedData
		p.writeLog(zapcore.InfoLevel, "performing event aggregation for dlSchedAgg...[Time-consuming ops which may cause 100% CPU utilization!]")
		wg := &sync.WaitGroup{}
		// for p1 := 0; p1 < mapEventRecord["dlFdSchedData"].Len(); p1 += 1 {
		for dn := range mapEventRecord["dlFdSchedData"] {
			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("  processing DL UE(PCI_RNTI) = %v, nbrRecord=%v, please wait...", dn, mapEventRecord["dlFdSchedData"][dn].Len()))
			for p1 := 0; p1 < mapEventRecord["dlFdSchedData"][dn].Len(); p1 += 1 {
				/*
				for {
					if runtime.NumGoroutine() >= p.maxgo {
						time.Sleep(1 * time.Second)
					} else {
						break
					}
				}
				 */

				wg.Add(1)

				go func(dn string, p1 int) {
					defer wg.Done()

					k1 := mapEventRecord["dlFdSchedData"][dn].Keys()[p1].(int)
					v1 := mapEventRecord["dlFdSchedData"][dn].Val(k1).(*TtiDlFdSchedData)
					dnPci := strings.Split(dn, "_")[0]
					dnRnti := strings.Split(dn, "_")[1]

					// aggregate mapDlFdUes
					ku := fmt.Sprintf("%v_%v_%v", v1.Sfn, v1.Slot, dnPci)
					dlFdUes := make([]string, 0)
					for _, ue := range mapDlFdUes[ku] {
						ueEventId := p.unsafeAtoi(strings.Split(ue, "_")[0])
						if math.Abs(float64(ueEventId - v1.eventId)) <= 16 {
							dlFdUes = append(dlFdUes, ue)
						}
					}
					v1.AllFields = append(v1.AllFields, fmt.Sprintf("%v,[%v]", len(dlFdUes), strings.Join(dlFdUes, ";")))

					// aggregate dlBeamData
					if _, e := mapEventRecord["dlBeamData"][dn]; e {
						p2 := p.findDlBeam(mapEventRecord["dlFdSchedData"][dn], mapEventRecord["dlBeamData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["dlBeamData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["dlBeamData"][dn].Val(k2).(*TtiDlBeamData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.CurrentBestBeamId, v2.Current2ndBeamId, v2.SelectedBestBeamId, v2.Selected2ndBeamId}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["dlBeamData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate dlPreSchedData
					if _, e := mapEventRecord["dlPreSchedData"][dn]; e {
						p2 := p.findDlPreSched(mapEventRecord["dlFdSchedData"][dn], mapEventRecord["dlPreSchedData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["dlPreSchedData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["dlPreSchedData"][dn].Val(k2).(*TtiDlPreSchedData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.CsListEvent, v2.HighestClassPriority, v2.PrachPreambleIndex}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["dlPreSchedData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate dlTdSchedSubcellData
					if _, e := mapEventRecord["dlTdSchedSubcellData"][dnPci]; e {
						p2 := p.findDlTdSched(mapEventRecord["dlFdSchedData"][dn], mapEventRecord["dlTdSchedSubcellData"][dnPci], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["dlTdSchedSubcellData"][dnPci].Keys()[p2].(int)
							v2 := mapEventRecord["dlTdSchedSubcellData"][dnPci].Val(k2).(*TtiDlTdSchedSubcellData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, fmt.Sprintf("(%d)[%s]", len(v2.Cs2List), strings.Join(v2.Cs2List, ";"))}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["dlTdSchedSubcellData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-"}...)
					}

					// aggregate dlHarqRxData
					if _, e := mapEventRecord["dlHarqRxData"][dn]; e {
						p2 := p.findDlHarq(mapEventRecord["dlFdSchedData"][dn], mapEventRecord["dlHarqRxData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["dlHarqRxData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["dlHarqRxData"][dn].Val(k2).(*TtiDlHarqRxData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.AckNack, v2.DlHarqProcessIndex, v2.PucchFormat}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["dlHarqRxData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate ulIntraDlToUlDrxSyncDlData
					if _, e := mapEventRecord["ulIntraDlToUlDrxSyncDlData"][dn]; e {
						p2 := p.findDlDrx(mapEventRecord["dlFdSchedData"][dn], mapEventRecord["ulIntraDlToUlDrxSyncDlData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulIntraDlToUlDrxSyncDlData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulIntraDlToUlDrxSyncDlData"][dn].Val(k2).(*TtiUlIntraDlToUlDtxSyncDlData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.DrxEnabled, v2.DlDrxOnDurationTimerOn, v2.DlDrxInactivityTimerOn}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulIntraDlToUlDrxSyncDlData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate csiSrReportData
					if _, e := mapEventRecord["csiSrReportData"][dn]; e {
						p2 := p.findCsiSrReport(mapEventRecord["dlFdSchedData"][dn], mapEventRecord["csiSrReportData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["csiSrReportData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["csiSrReportData"][dn].Val(k2).(*TtiCsiSrReportData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.UlChannel, v2.Dtx, v2.PucchFormat, v2.Cqi, v2.PmiRank1, v2.PmiRank2, v2.Ri, v2.Cri, v2.Li, v2.Sr}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["csiSrReportData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate dlLaDeltaCqi
					if _, e := mapEventRecord["dlLaDeltaCqi"][dn]; e {
						p2 := p.findDlLaDeltaCqi(mapEventRecord["dlFdSchedData"][dn], mapEventRecord["dlLaDeltaCqi"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["dlLaDeltaCqi"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["dlLaDeltaCqi"][dn].Val(k2).(*TtiDlLaDeltaCqi)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.IsDeltaCqiCalculated, v2.RrmPauseUeInDlScheduling, v2.HarqFb, v2.RrmDeltaCqi, v2.RrmRemainingBucketLevel}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["dlLaDeltaCqi"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate dlLaAverageCqi
					if _, e := mapEventRecord["dlLaAverageCqi"][dn]; e {
						p2 := p.findDlLaAvgCqi(mapEventRecord["dlFdSchedData"][dn], mapEventRecord["dlLaAverageCqi"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["dlLaAverageCqi"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["dlLaAverageCqi"][dn].Val(k2).(*TtiDlLaAverageCqi)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.RrmInstCqi, v2.Rank, v2.RrmAvgCqi, v2.Mcs, v2.RrmDeltaCqi}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["dlLaAverageCqi"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate dlFlowControlData
					if _, e := mapEventRecord["dlFlowControlData"][dnRnti]; e {
						p2 := p.findDlFlowControl(mapEventRecord["dlFdSchedData"][dn], mapEventRecord["dlFlowControlData"][dnRnti], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["dlFlowControlData"][dnRnti].Keys()[p2].(int)
							v2 := mapEventRecord["dlFlowControlData"][dnRnti].Val(k2).(*TtiDlFlowControlData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.LchId, v2.ReportType, v2.ScheduledBytes, v2.EthAvg, v2.EthScaled}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["dlFlowControlData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
					}
				}(dn, p1)
			}
		}
		wg.Wait()

		// output aggregated event: dlSchedAgg
		p.writeLog(zapcore.InfoLevel, "outputting aggregated dlSchedAgg...")
		headerWritten := make(map[string]bool)
		for dn := range mapEventRecord["dlFdSchedData"] {
			for _, k := range mapEventRecord["dlFdSchedData"][dn].Keys() {
				data := mapEventRecord["dlFdSchedData"][dn].Val(k).(*TtiDlFdSchedData)

				outFn := filepath.Join(outPath, fmt.Sprintf("dlSchedAgg_pci%s_rnti%s.csv", data.PhysCellId, data.Rnti))
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

				fout.WriteString(fmt.Sprintf("%v,%v\n", data.eventId, strings.Join(data.AllFields, ",")))
				fout.Close()
			}
		}
	}

	if p.ttiFilter == "ul" || p.ttiFilter == "both" {
		// update ulSchedAggFields
		// TODO - ulSchedAggFields
		p.writeLog(zapcore.InfoLevel, "updating fields for ulSchedAgg...") //core.QCoreApplication_Instance().ProcessEvents(0)
		// Print number of FD-scheduled UEs
		ulSchedAggFields += ",nbrFdUes,fdUeRntis"

		if len(mapEventRecord["ulBsrRxData"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulBsr.eventId", "ulBsr.sfn", "ulBsr.slot", "ulBsr.bsrFormat", "ulBsr.bufferSizeList"}, ",")
		}

		if len(mapEventRecord["ulPreSchedData"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulPreSched.eventId", "ulPreSched.sfn", "ulPreSched.slot", "ulPreSched.csListEvent", "ulPreSched.highestClassPriority"}, ",")
		}

		if len(mapEventRecord["ulTdSchedSubcellData"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulTdSched.eventId", "ulTdSched.sfn", "ulTdSched.slot", "ulTdSched.cs2List"}, ",")
		}

		if len(mapEventRecord["ulHarqRxData"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulHarq.eventId", "ulHarq.sfn", "ulHarq.slot", "ulHarq.dtx", "ulHarq.crcResult", "ulHarq.ulHarqProcessIndex"}, ",")
		}

		if len(mapEventRecord["ulIntraDlToUlDrxSyncDlData"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"drx.eventId", "drx.sfn", "drx.slot", "drx.drxEnabled", "drx.dlDrxOnDurationTimerOn", "drx.dlDrxInactivityTimerOn"}, ",")
		}

		if len(mapEventRecord["ulLaDeltaSinr"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulLaDeltaSinr.eventId", "ulLaDeltaSinr.sfn", "ulLaDeltaSinr.slot", "ulLaDeltaSinr.isDeltaSinrCalculated", "ulLaDeltaSinr.rrmPauseUeInUlScheduling", "ulLaDeltaSinr.crcFb", "ulLaDeltaSinr.rrmDeltaSinr"}, ",")
		}

		if len(mapEventRecord["ulLaAverageSinr"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulLaAvgSinr.eventId", "ulLaAvgSinr.sfn", "ulLaAvgSinr.slot", "ulLaAvgSinr.rrmInstSinrRank", "ulLaAvgSinr.rrmNumOfSinrMeasurements", "ulLaAvgSinr.rrmInstSinr", "ulLaAvgSinr.rrmSinrCorrection", "ulLaAvgSinr.rrmAvgSinrUl"}, ",")
		}

		if len(mapEventRecord["ulLaPhr"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulLaPhr.eventId", "ulLaPhr.sfn", "ulLaPhr.slot", "ulLaPhr.isRrmPhrScaledCalculated", "ulLaPhr.phr", "ulLaPhr.rrmNumPuschPrb", "ulLaPhr.rrmPhrScaled"}, ",")
		}

		if len(mapEventRecord["ulPucchReceiveRespPsData"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulPucch.eventId", "ulPucch.sfn", "ulPucch.slot", "ulPucch.pucchFormat", "ulPucch.startPrb", "ulPucch.rssi", "ulPucch.sinrLayer0", "ulPucch.sinrLayer1", "ulPucch.dtx", "ulPucch.srBit"}, ",")
		}

		if len(mapEventRecord["ulPuschReceiveRespPsData"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulPusch.eventId", "ulPusch.sfn", "ulPusch.slot", "ulPusch.rssi", "ulPusch.sinrLayer0", "ulPusch.sinrLayer1", "ulPusch.dtx", "ulPusch.ulRank", "ulPusch.ulPmiRank1", "ulPusch.ulPmiRank1Sinr",  "ulPusch.ulPmiRank2", "ulPusch.ulPmiRank2SinrLayer0", "ulPusch.ulPmiRank2SinrLayer1", "ulPusch.longTermRank"}, ",")
		}

		if len(mapEventRecord["ulPduDemuxData"]) > 0 {
			ulSchedAggFields += ","
			ulSchedAggFields += strings.Join([]string{"ulPduDemux.eventId", "ulPduDemux.sfn", "ulPduDemux.slot", "ulPduDemux.harqId", "ulPduDemux.isUlCcchData", "ulPduDemux.isTcpTraffic", "ulPduDemux.tempCrnti", "ulPduDemux.lcId", "ulPduDemux.rcvdBytes"}, ",")
		}

		ulSchedAggFields += "\n"

		// TODO - event aggregation with ulFdSchedData
		p.writeLog(zapcore.InfoLevel, "performing event aggregation for ulSchedAgg...[Time-consuming ops which may cause 100% CPU utilization!]")
		wg2 := &sync.WaitGroup{}
		for dn := range mapEventRecord["ulFdSchedData"] {
			p.writeLog(zapcore.InfoLevel, fmt.Sprintf("  processing UL UE(PCI_RNTI) = %v, nbrRecord=%v, please wait...", dn, mapEventRecord["ulFdSchedData"][dn].Len()))
			for p1 := 0; p1 < mapEventRecord["ulFdSchedData"][dn].Len(); p1 += 1 {
				wg2.Add(1)

				go func(dn string, p1 int) {
					defer wg2.Done()

					k1 := mapEventRecord["ulFdSchedData"][dn].Keys()[p1].(int)
					v1 := mapEventRecord["ulFdSchedData"][dn].Val(k1).(*TtiUlFdSchedData)
					dnPci := strings.Split(dn, "_")[0]

					// aggregate mapUlFdUes
					ku := fmt.Sprintf("%v_%v_%v", v1.Sfn, v1.Slot, dnPci)
					ulFdUes := make([]string, 0)
					for _, ue := range mapUlFdUes[ku] {
						ueEventId := p.unsafeAtoi(strings.Split(ue, "_")[0])
						if math.Abs(float64(ueEventId - v1.eventId)) <= 16 {
							ulFdUes = append(ulFdUes, ue)
						}
					}
					v1.AllFields = append(v1.AllFields, fmt.Sprintf("%v,[%v]", len(ulFdUes), strings.Join(ulFdUes, ";")))

					// aggregate ulBsrRxData
					if _, e := mapEventRecord["ulBsrRxData"][dn]; e {
						p2 := p.findUlBsr(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulBsrRxData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulBsrRxData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulBsrRxData"][dn].Val(k2).(*TtiUlBsrRxData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.BsrFormat, fmt.Sprintf("[%s]", strings.Join(v2.BufferSizeList, ";"))}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulBsrRxData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-"}...)
					}

					// aggregate ulPreSchedData
					if _, e := mapEventRecord["ulPreSchedData"][dn]; e {
						p2 := p.findUlPreSched(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulPreSchedData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulPreSchedData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulPreSchedData"][dn].Val(k2).(*TtiUlPreSchedData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.CsListEvent, v2.HighestClassPriority}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulPreSchedData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-"}...)
					}

					// aggregate ulTdSchedSubcellData
					if _, e := mapEventRecord["ulTdSchedSubcellData"][dnPci]; e {
						p2 := p.findUlTdSched(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulTdSchedSubcellData"][dnPci], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulTdSchedSubcellData"][dnPci].Keys()[p2].(int)
							v2 := mapEventRecord["ulTdSchedSubcellData"][dnPci].Val(k2).(*TtiUlTdSchedSubcellData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, fmt.Sprintf("(%d)[%s]", len(v2.Cs2List), strings.Join(v2.Cs2List, ";"))}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulTdSchedSubcellData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-"}...)
					}

					// aggregate ulHarqRxData
					if _, e := mapEventRecord["ulHarqRxData"][dn]; e {
						p2 := p.findUlHarq(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulHarqRxData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulHarqRxData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulHarqRxData"][dn].Val(k2).(*TtiUlHarqRxData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.Dtx, v2.CrcResult, v2.UlHarqProcessIndex}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulHarqRxData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate ulIntraDlToUlDrxSyncDlData
					if _, e := mapEventRecord["ulIntraDlToUlDrxSyncDlData"][dn]; e {
						p2 := p.findUlDrx(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulIntraDlToUlDrxSyncDlData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulIntraDlToUlDrxSyncDlData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulIntraDlToUlDrxSyncDlData"][dn].Val(k2).(*TtiUlIntraDlToUlDtxSyncDlData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.DrxEnabled, v2.DlDrxOnDurationTimerOn, v2.DlDrxInactivityTimerOn}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulIntraDlToUlDrxSyncDlData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate ulLaDeltaSinr
					if _, e := mapEventRecord["ulLaDeltaSinr"][dn]; e {
						p2 := p.findUlLaDeltaSinr(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulLaDeltaSinr"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulLaDeltaSinr"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulLaDeltaSinr"][dn].Val(k2).(*TtiUlLaDeltaSinr)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.IsDeltaSinrCalculated, v2.RrmPauseUeInUlScheduling, v2.CrcFb, v2.RrmDeltaSinr}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulLaDeltaSinr"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate ulLaAverageSinr
					if _, e := mapEventRecord["ulLaAverageSinr"][dn]; e {
						p2 := p.findUlLaAvgSinr(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulLaAverageSinr"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulLaAverageSinr"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulLaAverageSinr"][dn].Val(k2).(*TtiUlLaAverageSinr)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.RrmInstSinrRank, v2.RrmNumOfSinrMeasurements, v2.RrmInstSinr, v2.RrmSinrCorrection, v2.RrmAvgSinrUl}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulLaAverageSinr"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate ulLaPhr
					if _, e := mapEventRecord["ulLaPhr"][dn]; e {
						p2 := p.findUlLaPhr(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulLaPhr"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulLaPhr"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulLaPhr"][dn].Val(k2).(*TtiUlLaPhr)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.IsRrmPhrScaledCalculated, v2.Phr, v2.RrmNumPuschPrb, v2.RrmPhrScaled}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulLaPhr"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate ulPucchReceiveRespPsData
					if _, e := mapEventRecord["ulPucchReceiveRespPsData"][dn]; e {
						p2 := p.findUlPucchRcvResp(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulPucchReceiveRespPsData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulPucchReceiveRespPsData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulPucchReceiveRespPsData"][dn].Val(k2).(*TtiUlPucchReceiveRespPsData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.PucchFormat, v2.StartPrb, v2.Rssi, v2.SinrLayer0, v2.SinrLayer1, v2.Dtx, v2.SrBit}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulPucchReceiveRespPsData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate ulPuschReceiveRespPsData
					if _, e := mapEventRecord["ulPuschReceiveRespPsData"][dn]; e {
						p2 := p.findUlPuschRcvResp(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulPuschReceiveRespPsData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulPuschReceiveRespPsData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulPuschReceiveRespPsData"][dn].Val(k2).(*TtiUlPuschReceiveRespPsData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.Rssi, v2.SinrLayer0, v2.SinrLayer1, v2.Dtx, v2.UlRank, v2.UlPmiRank1, v2.UlPmiRank1Sinr, v2.UlPmiRank2, v2.UlPmiRank2SinrLayer0, v2.UlPmiRank2SinrLayer1, v2.LongTermRank}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulPuschReceiveRespPsData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
					}

					// aggregate ulPduDemuxData
					if _, e := mapEventRecord["ulPduDemuxData"][dn]; e {
						p2 := p.findUlPduDemux(mapEventRecord["ulFdSchedData"][dn], mapEventRecord["ulPduDemuxData"][dn], p1)
						if p2 >= 0 {
							k2 := mapEventRecord["ulPduDemuxData"][dn].Keys()[p2].(int)
							v2 := mapEventRecord["ulPduDemuxData"][dn].Val(k2).(*TtiUlPduDemuxData)

							v1.AllFields = append(v1.AllFields, []string{strconv.Itoa(v2.TtiEventHeader.eventId), v2.TtiEventHeader.Sfn, v2.TtiEventHeader.Slot, v2.HarqId, v2.IsUlCcchData, v2.IsTcpTraffic, v2.TempCrnti, fmt.Sprintf("[%s]", strings.Join(v2.LcIdList, ";")), fmt.Sprintf("[%s]", strings.Join(v2.RcvdBytesList, ";"))}...)
						} else {
							v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
						}
					} else if len(mapEventRecord["ulPduDemuxData"]) > 0 {
						v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-"}...)
					}
				}(dn, p1)
			}
		}
		wg2.Wait()

		// output aggregated event: ulSchedAgg
		p.writeLog(zapcore.InfoLevel, "outputting aggregated ulSchedAgg...")
		headerWritten2 := make(map[string]bool)
		for dn := range mapEventRecord["ulFdSchedData"] {
			for _, k := range mapEventRecord["ulFdSchedData"][dn].Keys() {
				data := mapEventRecord["ulFdSchedData"][dn].Val(k).(*TtiUlFdSchedData)

				outFn := filepath.Join(outPath, fmt.Sprintf("ulSchedAgg_pci%s_rnti%s.csv", data.PhysCellId, data.Rnti))
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

				fout.WriteString(fmt.Sprintf("%v,%v\n", data.eventId, strings.Join(data.AllFields, ",")))
				fout.Close()
			}
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

		if v2.eventId <= v1.eventId {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findDlPreSched(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlPreSchedData)

		//if v2.Sfn == v1.Sfn && v2.Slot == v1.Slot && v2.eventId <= v1.eventId {
		if v2.eventId <= v1.eventId {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findDlTdSched(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlTdSchedSubcellData)

		//if v2.Sfn == v1.Sfn && v2.Slot == v1.Slot && v2.eventId <= v1.eventId {
		if v2.eventId <= v1.eventId {
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
	_, sfn, slot := p.incSlot(1, p.unsafeAtoi(v1.Sfn), p.unsafeAtoi(v1.Slot), p.unsafeAtoi(v1.K1))

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlHarqRxData)

		if v2.Sfn == strconv.Itoa(sfn) && v2.Slot == strconv.Itoa(slot) && v2.eventId >= v1.eventId {
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

		if v2.eventId <= v1.eventId {
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

		if v2.eventId <= v1.eventId {
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

		if v2.eventId <= v1.eventId {
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
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlLaDeltaCqi)

		if v2.eventId <= v1.eventId {
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

		if v2.eventId <= v1.eventId {
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

		//if v2.eventId >= v1.eventId {
		if v1.Sfn == v2.Sfn && v1.Slot == v2.Slot && math.Abs(float64(v2.eventId - v1.eventId)) <= 100 {
			if v1.PhysCellId+v1.Rnti+v1.UlHarqProcessIndex == v2.PhysCellId+v2.Rnti+v2.UlHarqProcessIndex {
				p2 = i
				break
			}
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

		if v2.eventId <= v1.eventId {
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

		if v2.eventId <= v1.eventId {
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

		if v2.eventId <= v1.eventId {
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

		if v2.eventId <= v1.eventId {
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

		if v2.eventId <= v1.eventId {
			if v1.PhysCellId+v1.Rnti+v1.CellDbIndex == v2.PhysCellId+v2.Rnti+v2.CellDbIndex {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlPucchRcvResp(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlPucchReceiveRespPsData)

		if v2.eventId <= v1.eventId {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *L2TtiTraceParser) findUlPuschRcvResp(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiUlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlPuschReceiveRespPsData)

		//if v2.eventId <= v1.eventId {
		if v1.Sfn == v2.Sfn && v1.Slot == v2.Slot && math.Abs(float64(v2.eventId - v1.eventId)) <= 100 {
			if v1.PhysCellId+v1.Rnti == v2.PhysCellId+v2.Rnti {
				p2 = i
				break
			}
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

		if v2.eventId <= v1.eventId {
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
	_, sfn, slot := p.decSlot(1, p.unsafeAtoi(v1.Sfn), p.unsafeAtoi(v1.Slot), p.unsafeAtoi(v1.K2))

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlPreSchedData)

		// if v2.Sfn == strconv.Itoa(sfn) && v2.Slot == strconv.Itoa(slot) && v2.eventId <= v1.eventId {
		if v2.Sfn == strconv.Itoa(sfn) && v2.Slot == strconv.Itoa(slot) && math.Abs(float64(v2.eventId - v1.eventId)) <= 100 {
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
	_, sfn, slot := p.decSlot(1, p.unsafeAtoi(v1.Sfn), p.unsafeAtoi(v1.Slot), p.unsafeAtoi(v1.K2))

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiUlTdSchedSubcellData)

		// if v2.Sfn == strconv.Itoa(sfn) && v2.Slot == strconv.Itoa(slot) && v2.eventId <= v1.eventId {
		if v2.Sfn == strconv.Itoa(sfn) && v2.Slot == strconv.Itoa(slot) && math.Abs(float64(v2.eventId - v1.eventId)) <= 100 {
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
