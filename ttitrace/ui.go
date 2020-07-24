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
	scs2nslots := map[string]int{"15KHz(FDD)":10, "30KHz(TDD-FR1)":20, "120KHz(TDD-FR2)":80}
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
	var mapEventRecord = map[string]*utils.OrderedMap {
		"dlBeamData": utils.NewOrderedMap(),
		"dlPreSchedData": utils.NewOrderedMap(),
		"dlTdSchedSubcellData": utils.NewOrderedMap(),
		"dlFdSchedData": utils.NewOrderedMap(),
		"dlHarqRxData": utils.NewOrderedMap(),
		"dlLaAverageCqi": utils.NewOrderedMap(),
	}

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
							// defer fout.Close()
							if err != nil {
								p.LogEdit.Append(fmt.Sprintf("Fail to open file: %s", outFn))
								break
							}

							// Step-1.1: write HSFN header field
							fout.WriteString("hsfn,")

							row := strings.Join(mapFieldName[key], ",")
							fout.WriteString(fmt.Sprintf("%s\n", row))
							fout.Close()
						} else {
							curSfn, _ := strconv.Atoi(tokens[valStart+posSfn])
							if mapSfnInfo[key].lastSfn > curSfn {
								mapSfnInfo[key] = SfnInfo{curSfn, mapSfnInfo[key].hsfn+1}
							} else {
								mapSfnInfo[key] = SfnInfo{curSfn, mapSfnInfo[key].hsfn}
							}
						}

						// Step-2: write event record
						fout, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
						// defer fout.Close()
						if err != nil {
							p.LogEdit.Append(fmt.Sprintf("Fail to open file: %s", outFn))
							break
						}

						// Step-2.1: write hsfn
						fout.WriteString(fmt.Sprintf("%d,", mapSfnInfo[key].hsfn))

						row := strings.Join(tokens[valStart:], ",")
						fout.WriteString(fmt.Sprintf("%s\n", row))
						fout.Close()


						// Step-3: aggregate events
						if eventName == "dlBeamData" {
							if posDlBeam.Ready == false {
								posDlBeam = FindTtiDlBeamDataPos(tokens)
								p.LogEdit.Append(fmt.Sprintf("posDlBeam=%v", posDlBeam))
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[posDlBeam.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[posDlBeam.PosEventHeader.PosSlot]))
							v := TtiDlBeamData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[posDlBeam.PosEventHeader.PosSfn],
									Slot:       tokens[posDlBeam.PosEventHeader.PosSlot],
									Rnti:       tokens[posDlBeam.PosEventHeader.PosRnti],
									PhysCellId: tokens[posDlBeam.PosEventHeader.PosPhysCellId],
								},

								SubcellId: tokens[posDlBeam.PosSubcellId],
								CurrentBestBeamId: tokens[posDlBeam.PosCurrentBestBeamId],
								Current2ndBeamId: tokens[posDlBeam.PosCurrent2ndBeamId],
								SelectedBestBeamId: tokens[posDlBeam.PosSelectedBestBeamId],
								Selected2ndBeamId: tokens[posDlBeam.PosSelected2ndBeamId],
							}
							mapEventRecord[eventName].Add(k, v)
						} else if eventName == "dlPreSchedData" {
							if posDlPreSched.Ready == false {
								posDlPreSched = FindTtiDlPreSchedDataPos(tokens)
								p.LogEdit.Append(fmt.Sprintf("posDlPreSched=%v", posDlPreSched))
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[posDlPreSched.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[posDlPreSched.PosEventHeader.PosSlot]))
							v := TtiDlPreSchedData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[posDlPreSched.PosEventHeader.PosSfn],
									Slot:       tokens[posDlPreSched.PosEventHeader.PosSlot],
									Rnti:       tokens[posDlPreSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[posDlPreSched.PosEventHeader.PosPhysCellId],
								},

								CsListEvent: tokens[posDlPreSched.PosCsListEvent],
								HighestClassPriority: tokens[posDlPreSched.PosHighestClassPriority],
							}
							mapEventRecord[eventName].Add(k, v)
						} else if eventName == "dlTdSchedSubcellData" {
							if posDlTdSched.Ready == false {
								posDlTdSched = FindTtiDlTdSchedSubcellDataPos(tokens)
								p.LogEdit.Append(fmt.Sprintf("posDlTdSched=%v", posDlTdSched))
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[posDlTdSched.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[posDlTdSched.PosEventHeader.PosSlot]))
							v := TtiDlTdSchedSubcellData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[posDlTdSched.PosEventHeader.PosSfn],
									Slot:       tokens[posDlTdSched.PosEventHeader.PosSlot],
									Rnti:       tokens[posDlTdSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[posDlTdSched.PosEventHeader.PosPhysCellId],
								},

								SubcellId: tokens[posDlTdSched.PosSubcellId],
								Cs2List: make([]string, 0),
							}

							for _, rsn := range posDlTdSched.PosRecordSequenceNumber {
								n := 10	// Per UE in CS2 is statitically defined for 10UEs
								for k := 0; k < n; k += 1 {
									v.Cs2List = append(v.Cs2List, tokens[rsn+1+3*k])
								}
							}

							mapEventRecord[eventName].Add(k, v)
						} else if eventName == "dlFdSchedData" {
							if posDlFdSched.Ready == false {
								posDlFdSched = FindTtiDlFdSchedDataPos(tokens)
								p.LogEdit.Append(fmt.Sprintf("posDlFdSched=%v", posDlFdSched))
							}

							k := p.makeTimeStamp(mapSfnInfo[key].hsfn, p.unsafeAtoi(tokens[posDlFdSched.PosEventHeader.PosSfn]), p.unsafeAtoi(tokens[posDlFdSched.PosEventHeader.PosSlot]))
							v := TtiDlFdSchedData{
								// event header
								TtiEventHeader: TtiEventHeader{
									Hsfn:       strconv.Itoa(mapSfnInfo[key].hsfn),
									Sfn:        tokens[posDlFdSched.PosEventHeader.PosSfn],
									Slot:       tokens[posDlFdSched.PosEventHeader.PosSlot],
									Rnti:       tokens[posDlFdSched.PosEventHeader.PosRnti],
									PhysCellId: tokens[posDlFdSched.PosEventHeader.PosPhysCellId],
								},

								CellDbIndex: tokens[posDlFdSched.PosCellDbIndex],
								TxNumber: tokens[posDlFdSched.PosTxNumber],
								DlHarqProcessIndex: tokens[posDlFdSched.PosDlHarqProcessIndex],
								K1: tokens[posDlFdSched.PosK1],
								AllFields: make([]string, 0),
							}
							copy(v.AllFields, tokens[valStart:])

							mapEventRecord[eventName].Add(k, v)
						} else if eventName == "dlHarqRxData" && posDlHarq.Ready == false {
							// TODO: there is also an event named: dlHarqRxDataArray
							posDlHarq = FindTtiDlHarqRxDataPos(tokens)
							p.LogEdit.Append(fmt.Sprintf("posDlHarq=%v", posDlHarq))
						} else if eventName == "dlLaAverageCqi" && posDlLaAvgCqi.Ready == false {
							posDlLaAvgCqi = FindTtiDlLaAverageCqiPos(tokens)
							p.LogEdit.Append(fmt.Sprintf("posDlLaAvgCqi=%v", posDlLaAvgCqi))
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
