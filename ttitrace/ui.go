package ttitrace

import (
	"bufio"
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
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
	chooseBtn *widgets.QPushButton
	okBtn     *widgets.QPushButton
	cancelBtn *widgets.QPushButton
	widget    *widgets.QDialog

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
	gridLayout.AddLayout2(hboxLayout1, 1, 0, 1, 2, 0)

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
	// recreate dir for parsed ttis
	outPath := path.Join(p.prevDir, "parsed_ttis")
	os.RemoveAll(outPath)
	if err := os.MkdirAll(outPath, 0775); err != nil {
		panic(fmt.Sprintf("Fail to create directory: %v", err))
	}

	mapFieldName := make(map[string][]string)
	mapSfnInfo := make(map[string]SfnInfo)
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

							// write event header only once
							fout, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
							defer fout.Close()
							if err != nil {
								p.LogEdit.Append(fmt.Sprintf("Fail to open file: %s", outFn))
								break
							}

							// write HSFN header field
							fout.WriteString("hsfn,")

							row := strings.Join(mapFieldName[key], ",")
							fout.WriteString(fmt.Sprintf("%s\n", row))
						} else {
							curSfn, _ := strconv.Atoi(tokens[valStart+posSfn])
							if mapSfnInfo[key].lastSfn > curSfn {
								mapSfnInfo[key] = SfnInfo{curSfn, mapSfnInfo[key].hsfn+1}
							} else {
								mapSfnInfo[key] = SfnInfo{curSfn, mapSfnInfo[key].hsfn}
							}
						}

						// write event record
						fout, err := os.OpenFile(outFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
						defer fout.Close()
						if err != nil {
							p.LogEdit.Append(fmt.Sprintf("Fail to open file: %s", outFn))
							break
						}

						// write hsfn
						fout.WriteString(fmt.Sprintf("%d,", mapSfnInfo[key].hsfn))

						row := strings.Join(tokens[valStart:], ",")
						fout.WriteString(fmt.Sprintf("%s\n", row))
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
