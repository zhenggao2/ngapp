package ttitrace

import (
	"bufio"
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"github.com/zhenggao2/ngapp/utils"
	"go.uber.org/zap"
	"os"
	"path"
	"strings"
	"strconv"
	"errors"
	"math"
)

type TtiTraceUi struct {
	Debug   bool
	Logger  *zap.Logger
	LogEdit *widgets.QTextEdit
	Args    map[string]interface{}

	ratComb   *widgets.QComboBox
	scsComb *widgets.QComboBox
	chooseBtn *widgets.QPushButton
	okBtn     *widgets.QPushButton
	cancelBtn *widgets.QPushButton
	widget    *widgets.QDialog

	slotsPerRf int
	ttiFiles []string
	prevDir  string
}

type SfnInfo struct {
	lastSfn int
	hsfn int
}

func (p *TtiTraceUi) InitUi() {
	// notes on Qt::Alignment:
	// void QGridLayout::addWidget(QWidget *widget, int fromRow, int fromColumn, int rowSpan, int columnSpan, Qt::Alignment alignment = Qt::Alignment())
	// The alignment is specified by alignment. The default alignment is 0, which means that the widget fills the entire cell.

	// notes on Qt::WindowsFlags:
	// QLabel::QLabel(const QString &text, QWidget *parent = nullptr, Qt::WindowFlags f = Qt::WindowFlags())
	// Qt::Widget	0x00000000	This is the default type for QWidget. Widgets of this type are child widgets if they have a parent, and independent windows if they have no parent. See also Qt::Window and Qt::SubWindow.

	ratLabel := widgets.NewQLabel2("Select RAT:", nil, core.Qt__Widget)
	p.ratComb = widgets.NewQComboBox(nil)
	p.ratComb.AddItems([]string{"5G"})

	scsLabel := widgets.NewQLabel2("NRCELLGRP-scs:", nil, core.Qt__Widget)
	p.scsComb = widgets.NewQComboBox(nil)
	p.scsComb.AddItems([]string{"15KHz(FDD)", "30KHz(TDD-FR1)", "120KHz(TDD-FR2)"})
	p.scsComb.SetCurrentIndex(1)

	chooseLabel := widgets.NewQLabel2("Select TTI files:", nil, core.Qt__Widget)
	p.chooseBtn = widgets.NewQPushButton2("...", nil)

	p.okBtn = widgets.NewQPushButton2("OK", nil)
	p.cancelBtn = widgets.NewQPushButton2("Cancel", nil)

	hboxLayout1 := widgets.NewQHBoxLayout()
	hboxLayout1.AddWidget(chooseLabel, 0, 0)
	hboxLayout1.AddWidget(p.chooseBtn, 0, 0)
	hboxLayout1.AddStretch(0)

	gridLayout := widgets.NewQGridLayout(nil)
	gridLayout.AddWidget2(ratLabel, 0, 0, 0)
	gridLayout.AddWidget2(p.ratComb, 0, 1, 0)
	gridLayout.AddWidget2(scsLabel, 1, 0, 0)
	gridLayout.AddWidget2(p.scsComb, 1, 1, 0)
	gridLayout.AddLayout2(hboxLayout1, 2, 0, 1, 2, 0)

	hboxLayout2 := widgets.NewQHBoxLayout()
	hboxLayout2.AddStretch(0)
	hboxLayout2.AddWidget(p.okBtn, 0, 0)
	hboxLayout2.AddWidget(p.cancelBtn, 0, 0)

	layout := widgets.NewQVBoxLayout()
	layout.AddLayout(gridLayout, 0)
	layout.AddStretch(0)
	layout.AddLayout(hboxLayout2, 0)

	p.widget = widgets.NewQDialog(nil, core.Qt__Widget)
	p.widget.SetLayout(layout)
	p.widget.SetWindowTitle("TTI Trace Tool")

	p.initSlots()

	p.widget.Exec()
}

func (p *TtiTraceUi) initSlots() {
	p.chooseBtn.ConnectClicked(p.onChooseBtnClicked)
	p.okBtn.ConnectClicked(p.onOkBtnClicked)
	p.cancelBtn.ConnectClicked(p.onCancelBtnClicked)
}

func (p *TtiTraceUi) onChooseBtnClicked(checked bool) {
	var curDir string
	if len(p.prevDir) == 0 {
		curDir = "."
	} else {
		curDir = p.prevDir
	}

	p.ttiFiles = widgets.QFileDialog_GetOpenFileNames(p.widget, "Choose TTI Files", curDir, "Data files (*.dat);;CSV files (*.csv);;All files (*.*)", "All files (*.*)", widgets.QFileDialog__DontResolveSymlinks)
	if len(p.ttiFiles) > 0 {
		p.prevDir = path.Dir(p.ttiFiles[0])
	}

	/*
	p.LogEdit.Append("List of TTI files:")
	for _, fn := range p.ttiFiles {
		p.LogEdit.Append(fn)
	}
	*/
}

func (p *TtiTraceUi) onOkBtnClicked(checked bool) {
	scs2nslots := map[string]int{"15KHz(FDD)": 10, "30KHz(TDD-FR1)": 20, "120KHz(TDD-FR2)": 80}
	p.slotsPerRf = scs2nslots[p.scsComb.CurrentText()]

	// recreate dir for parsed ttis
	outPath := path.Join(p.prevDir, "parsed_ttis")
	os.RemoveAll(outPath)
	if err := os.MkdirAll(outPath, 0775); err != nil {
		panic(fmt.Sprintf("Fail to create directory: %v", err))
	}

	// key=EventName_PCI_RNTI or EventName for both mapFieldName and mapSfnInfo
	mapFieldName := make(map[string][]string)
	mapSfnInfo := make(map[string]SfnInfo)

	// Field positions per event
	var posDlBeam TtiDlBeamDataPos
	var posDlPreSched TtiDlPreSchedDataPos
	var posDlTdSched TtiDlTdSchedSubcellDataPos
	var posDlFdSched TtiDlFdSchedDataPos
	var posDlHarq TtiDlHarqRxDataPos
	var posDlLaAvgCqi TtiDlLaAverageCqiPos
	var mapEventRecord = map[string]*utils.OrderedMap{
		"dlBeamData":           utils.NewOrderedMap(),
		"dlPreSchedData":       utils.NewOrderedMap(),
		"dlTdSchedSubcellData": utils.NewOrderedMap(),
		"dlFdSchedData":        utils.NewOrderedMap(),
		"dlHarqRxData":         utils.NewOrderedMap(),
		"dlLaAverageCqi":       utils.NewOrderedMap(),
	}
	var dlSchedAggFields string
	var dlPerBearerProcessed bool
	var mapPdschSliv map[string][]int = p.initPdschSliv()

	for _, fn := range p.ttiFiles {
		p.LogEdit.Append(fmt.Sprintf("Parsing tti file: %s", fn))
		/*
			//QEventLoop::ProcessEventsFlag
			type QEventLoop__ProcessEventsFlag int64

			const (
				QEventLoop__AllEvents              QEventLoop__ProcessEventsFlag = QEventLoop__ProcessEventsFlag(0x00)
				QEventLoop__ExcludeUserInputEvents QEventLoop__ProcessEventsFlag = QEventLoop__ProcessEventsFlag(0x01)
				QEventLoop__ExcludeSocketNotifiers QEventLoop__ProcessEventsFlag = QEventLoop__ProcessEventsFlag(0x02)
				QEventLoop__WaitForMoreEvents      QEventLoop__ProcessEventsFlag = QEventLoop__ProcessEventsFlag(0x04)
				QEventLoop__X11ExcludeTimers       QEventLoop__ProcessEventsFlag = QEventLoop__ProcessEventsFlag(0x08)
				QEventLoop__EventLoopExec          QEventLoop__ProcessEventsFlag = QEventLoop__ProcessEventsFlag(0x20)
				QEventLoop__DialogExec             QEventLoop__ProcessEventsFlag = QEventLoop__ProcessEventsFlag(0x40)
			)
		*/
		core.QCoreApplication_Instance().ProcessEvents(0)

		fin, err := os.Open(fn)
		defer fin.Close()
		if err != nil {
			p.LogEdit.Append(err.Error())
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
							mapSfnInfo[key] = SfnInfo{sfn, 0}

							// Step-1: write event header only once
							fout, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
							defer fout.Close()
							if err != nil {
								p.LogEdit.Append(fmt.Sprintf("Fail to open file: %s", outFn))
								break
							}

							// Step-1.1: write HSFN header field
							fout.WriteString("hsfn,")

							row := strings.Join(mapFieldName[key], ",")
							fout.WriteString(fmt.Sprintf("%s\n", row))
							//fout.Close()

							// update dlSchedAggFields
							if len(dlSchedAggFields) == 0 && eventName == "dlFdSchedData" {
								dlSchedAggFields = "hsfn," + row
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
						defer fout2.Close()
						if err != nil {
							p.LogEdit.Append(fmt.Sprintf("Fail to open file: %s", outFn))
							break
						}

						// Step-2.1: write hsfn
						fout2.WriteString(fmt.Sprintf("%d,", mapSfnInfo[key].hsfn))

						row := strings.Join(tokens[valStart:], ",")
						fout2.WriteString(fmt.Sprintf("%s\n", row))
						//fout2.Close()

						// Step-3: aggregate events
						if eventName == "dlBeamData" {
							if posDlBeam.Ready == false {
								posDlBeam = FindTtiDlBeamDataPos(tokens)
								if p.Debug {
									p.LogEdit.Append(fmt.Sprintf("posDlBeam=%v", posDlBeam))
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
						} else if eventName == "dlPreSchedData" {
							if posDlPreSched.Ready == false {
								posDlPreSched = FindTtiDlPreSchedDataPos(tokens)
								if p.Debug {
									p.LogEdit.Append(fmt.Sprintf("posDlPreSched=%v", posDlPreSched))
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
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "dlTdSchedSubcellData" {
							if posDlTdSched.Ready == false {
								posDlTdSched = FindTtiDlTdSchedSubcellDataPos(tokens)
								if p.Debug {
									p.LogEdit.Append(fmt.Sprintf("posDlTdSched=%v", posDlTdSched))
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
								n := 10 // Per UE in CS2 is statitically defined for 10UEs
								for k := 0; k < n; k += 1 {
									posRnti := valStart + rsn + 1 + 3*k
									if posRnti > len(tokens) || len(tokens[posRnti]) == 0 {
										break
									}
									v.Cs2List = append(v.Cs2List, tokens[posRnti])
								}
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "dlFdSchedData" {
							if posDlFdSched.Ready == false {
								posDlFdSched = FindTtiDlFdSchedDataPos(tokens)
								if p.Debug {
									p.LogEdit.Append(fmt.Sprintf("posDlFdSched=%v", posDlFdSched))
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

							// update RIV field
							rivStr := "(RIV="
							numPrb := p.unsafeAtoi(tokens[valStart+posDlFdSched.PosNumOfPrb])
							startPrb := p.unsafeAtoi(tokens[valStart+posDlFdSched.PosStartPrb])
							// refer to:
							//  12	Bandwidth part operation of 3GPP 38.213
							//  5.1.2.2.2	Downlink resource allocation type 1 of 3GPP 38.214
							bwpSize := 275
							riv, err := p.makeRiv(numPrb, startPrb, bwpSize)
							if err == nil {
								rivStr += strconv.Itoa(riv)
							} else {
								rivStr += "-1"
							}
							rivStr += ")"

							v.AllFields[posDlFdSched.PosStartPrb] += rivStr

							// update AntPort field
							mapPdschDmrsType1AntPort := map[int]string {
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
							antPortStr := "("
							if ports, exist := mapPdschDmrsType1AntPort[p.unsafeAtoi(tokens[valStart+posDlFdSched.PosAntPort])]; exist {
								antPortStr += ports
							}
							antPortStr += ")"

							v.AllFields[posDlFdSched.PosAntPort] += antPortStr

							// update per beaer info [lcId, scheduledBytesPerBearer, remainingBytesPerBearerInFdEoBuffer]
							// there are max 18 bearers
							maxNumBearerPerUe := 18
							lcId := make([]string, 0)
							schedBytes := make([]string, 0)
							remainBytes := make([]string, 0)
							for ib := 0; ib < maxNumBearerPerUe; ib += 1 {
								if p.unsafeAtoi(tokens[valStart+posDlFdSched.PosLcId+3*ib]) == 255 {
									break
								} else {
									lcId = append(lcId, tokens[valStart+posDlFdSched.PosLcId+3*ib])
									schedBytes = append(schedBytes, tokens[valStart+posDlFdSched.PosLcId+3*ib+1])
									remainBytes = append(remainBytes, tokens[valStart+posDlFdSched.PosLcId+3*ib+2])
								}
							}

							perBearerInfo := []string{fmt.Sprintf("[%s]", strings.Join(lcId, ";")), fmt.Sprintf("[%s]", strings.Join(schedBytes, ";")), fmt.Sprintf("[%s]", strings.Join(remainBytes, ";"))}
							v.AllFields = append(append(v.AllFields[:posDlFdSched.PosLcId], perBearerInfo...), v.AllFields[posDlFdSched.PosLcId+3*maxNumBearerPerUe:]...)

							// update dlSchedAggFields accordingly only once
							if !dlPerBearerProcessed {
								dlSchedAggFieldsTokens := strings.Split(dlSchedAggFields, ",")
								dlSchedAggFields = strings.Join(append(dlSchedAggFieldsTokens[:1+posDlFdSched.PosLcId+3], dlSchedAggFieldsTokens[1+posDlFdSched.PosLcId+3*maxNumBearerPerUe:]...), ",")
								dlPerBearerProcessed = true
							}

							mapEventRecord[eventName].Add(k, &v)
						} else if eventName == "dlHarqRxData" || eventName == "dlHarqRxDataArray" {
							if posDlHarq.Ready == false {
								posDlHarq = FindTtiDlHarqRxDataPos(tokens)

								if eventName == "dlHarqRxDataArray" {
									posDlHarq.PosEventHeader.PosRnti += 2
									posDlHarq.PosEventHeader.PosPhysCellId += 2
								}

								if p.Debug {
									p.LogEdit.Append(fmt.Sprintf("posDlHarq=%v", posDlHarq))
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
								}

								mapEventRecord[eventName].Add(strconv.Itoa(k)+"_"+v.DlHarqProcessIndex, &v)
							} else {
								maxNumHarq := 32	// max 32 HARQ feedbacks per dlHarqRxDataArray
								for ih := 0; ih < maxNumHarq; ih += 1 {
									posRnti := valStart+posDlHarq.PosEventHeader.PosRnti+ih*9
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
											Rnti:       tokens[valStart+posDlHarq.PosEventHeader.PosRnti+ih*9],
											PhysCellId: tokens[valStart+posDlHarq.PosEventHeader.PosPhysCellId+ih*9],
										},

										HarqSubcellId:      tokens[valStart+posDlHarq.PosHarqSubcellId+ih*9],
										AckNack:            tokens[valStart+posDlHarq.PosAckNack+ih*9],
										DlHarqProcessIndex: tokens[valStart+posDlHarq.PosDlHarqProcessIndex+ih*9],
									}

									mapEventRecord["dlHarqRxData"].Add(strconv.Itoa(k)+"_"+v.DlHarqProcessIndex, &v)
								}
							}
						} else if eventName == "dlLaAverageCqi" {
							if posDlLaAvgCqi.Ready == false {
								posDlLaAvgCqi = FindTtiDlLaAverageCqiPos(tokens)
								if p.Debug {
									p.LogEdit.Append(fmt.Sprintf("posDlLaAvgCqi=%v", posDlLaAvgCqi))
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
								RrmAvgCqi:   tokens[valStart+posDlLaAvgCqi.PosRrmAvgCqi],
								RrmDeltaCqi: tokens[valStart+posDlLaAvgCqi.PosRrmDeltaCqi],
							}

							mapEventRecord[eventName].Add(k, &v)
						}
					} else {
						p.LogEdit.Append(fmt.Sprintf("Invalid event record detected: %s", line))
					}
				} else {
					p.LogEdit.Append(fmt.Sprintf("Invalid event data detected: %s", line))
				}
			}
		}
	}

	// update dlSchedAggFields
	p.LogEdit.Append("updating fields for dlSchedAgg...")
	core.QCoreApplication_Instance().ProcessEvents(0)

	if mapEventRecord["dlBeamData"].Len() > 0 {
		dlSchedAggFields += ","
		dlSchedAggFields += strings.Join([]string{"currentBestBeamId", "current2ndBeamId", "selectedBestBeamId", "selected2ndBeamid"}, ",")
	}

	if mapEventRecord["dlPreSchedData"].Len() > 0 {
		dlSchedAggFields += ","
		dlSchedAggFields += strings.Join([]string{"csListEvent", "highestClassPriority"}, ",")
	}

	if mapEventRecord["dlTdSchedSubcellData"].Len() > 0 {
		dlSchedAggFields += ",cs2List"
	}

	if mapEventRecord["dlHarqRxData"].Len() > 0 {
		dlSchedAggFields += ",AckNack"
	}

	if mapEventRecord["dlLaAverageCqi"].Len() > 0 {
		dlSchedAggFields += ","
		dlSchedAggFields += strings.Join([]string{"rrmAvgCqi", "rrmDeltaCqi"}, ",")
	}

	dlSchedAggFields += "\n"

	// perform event aggregation
	p.LogEdit.Append("performing event aggregation for dlSchedAgg...")
	core.QCoreApplication_Instance().ProcessEvents(0)

	for p1 := 0; p1 < mapEventRecord["dlFdSchedData"].Len(); p1 += 1 {
		k1 := mapEventRecord["dlFdSchedData"].Keys()[p1].(int)
		v1 := mapEventRecord["dlFdSchedData"].Val(k1).(*TtiDlFdSchedData)

		// aggregate dlBeamData
		p2 := p.findDlBeam(mapEventRecord["dlFdSchedData"], mapEventRecord["dlBeamData"], p1)
		if p2 >= 0 {
			k2 := mapEventRecord["dlBeamData"].Keys()[p2].(int)
			v2 := mapEventRecord["dlBeamData"].Val(k2).(*TtiDlBeamData)

			v1.AllFields = append(v1.AllFields, []string{v2.CurrentBestBeamId, v2.Current2ndBeamId, v2.SelectedBestBeamId, v2.Selected2ndBeamId}...)
		} else {
			v1.AllFields = append(v1.AllFields, []string{"-", "-", "-", "-"}...)
		}

		// aggregate dlPreSchedData
		p3 := p.findDlPreSched(mapEventRecord["dlFdSchedData"], mapEventRecord["dlPreSchedData"], p1)
		if p3 >= 0 {
			k3 := mapEventRecord["dlPreSchedData"].Keys()[p3].(int)
			v3 := mapEventRecord["dlPreSchedData"].Val(k3).(*TtiDlPreSchedData)

			v1.AllFields = append(v1.AllFields, []string{v3.CsListEvent, v3.HighestClassPriority}...)
		} else {
			v1.AllFields = append(v1.AllFields, []string{"-", "-"}...)
		}

		// aggregate dlTdSchedSubcellData
		p4 := p.findDlTdSched(mapEventRecord["dlFdSchedData"], mapEventRecord["dlTdSchedSubcellData"], p1)
		if p4 >= 0 {
			k4 := mapEventRecord["dlTdSchedSubcellData"].Keys()[p4].(int)
			v4 := mapEventRecord["dlTdSchedSubcellData"].Val(k4).(*TtiDlTdSchedSubcellData)

			v1.AllFields = append(v1.AllFields, fmt.Sprintf("(%d)[%s]", len(v4.Cs2List), strings.Join(v4.Cs2List, ";")))
		} else {
			v1.AllFields = append(v1.AllFields, "-")
		}

		// aggregate dlHarqRxData
		p5 := p.findDlHarq(mapEventRecord["dlFdSchedData"], mapEventRecord["dlHarqRxData"], p1)
		if p5 >= 0 {
			k5 := mapEventRecord["dlHarqRxData"].Keys()[p5].(string)
			v5 := mapEventRecord["dlHarqRxData"].Val(k5).(*TtiDlHarqRxData)

			v1.AllFields = append(v1.AllFields, v5.AckNack)
		} else {
			v1.AllFields = append(v1.AllFields, "-")
		}

		// aggregate dlLaAverageCqi
		p6 := p.findDlLaAvgCqi(mapEventRecord["dlFdSchedData"], mapEventRecord["dlLaAverageCqi"], p1)
		if p6 >= 0 {
			k6 := mapEventRecord["dlLaAverageCqi"].Keys()[p6].(int)
			v6 := mapEventRecord["dlLaAverageCqi"].Val(k6).(*TtiDlLaAverageCqi)

			v1.AllFields = append(v1.AllFields, []string{v6.RrmAvgCqi, v6.RrmDeltaCqi}...)
		} else {
			v1.AllFields = append(v1.AllFields, []string{"-", "-"}...)
		}

	}

	if p.Debug {
		for k, v := range mapEventRecord {
			p.LogEdit.Append(fmt.Sprintf("Event=%q, EventData=%v\n", k, v))
		}
	}

	// output aggregated event: dlSchedAgg
	p.LogEdit.Append("outputing aggregated dlSchedAgg...")
	core.QCoreApplication_Instance().ProcessEvents(0)

	headerWritten := make(map[string]bool)
	for _, k := range mapEventRecord["dlFdSchedData"].Keys() {
		data := mapEventRecord["dlFdSchedData"].Val(k).(*TtiDlFdSchedData)

		outFn := path.Join(outPath, fmt.Sprintf("dlSchedAgg_pci%s_rnti%s.csv", data.PhysCellId, data.Rnti))
		fout, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
		defer fout.Close()
		if err != nil {
			p.LogEdit.Append(fmt.Sprintf("Fail to open file: %s", outFn))
			break
		}

		if _, exist := headerWritten[outFn]; !exist {
			fout.WriteString(dlSchedAggFields)
			headerWritten[outFn] = true
		}

		fout.WriteString(fmt.Sprintf("%s,%s\n", data.Hsfn, strings.Join(data.AllFields, ",")))
		//fout.Close()
	}

	p.widget.Accept()
}

func (p *TtiTraceUi) onCancelBtnClicked(checked bool) {
	p.widget.Reject()
}

func (p *TtiTraceUi) makeTimeStamp(hsfn, sfn, slot int) int {
	return 1024 * p.slotsPerRf * hsfn + p.slotsPerRf * sfn + slot
}

func (p *TtiTraceUi) unsafeAtoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func (p *TtiTraceUi) ttiDlPreSchedClassPriority(cp string) string {
	classPriority := []string {"rachMsg2", "harqRetxMsg4", "harqRetxSrb1", "harqRetxSrb3", "harqRetxSrb2", "harqRetxDrb", "dlMacCe", "srb1Traffic", "srb3Traffic", "srb2Traffic", "drbTraffic", "lastUnUsed"}

	return fmt.Sprintf("%s(%s)", cp, classPriority[p.unsafeAtoi(cp)])
}

func (p *TtiTraceUi) findDlBeam(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlBeamData)

		if k2 <= k1 {
			if v1.PhysCellId+v1.Rnti+v1.CellDbIndex == v2.PhysCellId+v2.Rnti+v2.SubcellId {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *TtiTraceUi) findDlPreSched(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlPreSchedData)

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

func (p *TtiTraceUi) findDlTdSched(m1,m2 *utils.OrderedMap, p1 int) int {
	k1 := m1.Keys()[p1].(int)
	v1 := m1.Val(k1).(*TtiDlFdSchedData)

	p2 := -1
	for i := 0; i < m2.Len(); i += 1 {
		k2 := m2.Keys()[i].(int)
		v2 := m2.Val(k2).(*TtiDlTdSchedSubcellData)

		if k2 <= k1 {
			if v1.PhysCellId+v1.CellDbIndex == v2.PhysCellId+v2.SubcellId  && p.contains(v2.Cs2List, v1.Rnti) {
				p2 = i
			}
		} else {
			break
		}
	}

	return p2
}

func (p *TtiTraceUi) findDlHarq(m1,m2 *utils.OrderedMap, p1 int) int {
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
			if v1.PhysCellId+v1.Rnti+v1.CellDbIndex == v2.PhysCellId+v2.Rnti+v2.HarqSubcellId && v1.DlHarqProcessIndex == v2.DlHarqProcessIndex {
				p2 = i
				return p2
			}
		}
	}

	return p2
}

func (p *TtiTraceUi) findDlLaAvgCqi(m1,m2 *utils.OrderedMap, p1 int) int {
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

func (p *TtiTraceUi) contains(a []string, b string) bool {
	for _, s := range a {
		if s == b {
			return true
		}
	}

	return false
}

func (p *TtiTraceUi) incSlot(hsfn, sfn, slot, n int) (int, int, int) {
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

func (p *TtiTraceUi) initPdschSliv() map[string][]int {
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

func (p *TtiTraceUi) makeSliv(S, L int) (int, error) {
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

func (p *TtiTraceUi) makeRiv(numPrb, startPrb, bwpSize int) (int, error) {
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
