package nrgrid

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"go.uber.org/zap"
)

type ScsSpecificCarrierUi struct{
	ScsLabel *widgets.QLabel
	Scs *widgets.QComboBox
	BwLabel *widgets.QLabel
	Bw *widgets.QComboBox
	NRbLabel *widgets.QLabel
	NRb *widgets.QLineEdit
	OffsetToCarrierLabel *widgets.QLabel
	OffsetToCarrier *widgets.QLineEdit
	widget *widgets.QGroupBox
}

type ScsSpecificCarrier struct{
	Scs string
	Bw string
	NumRbs int
	OffsetToCarrier int
}

type SsbGridUi struct{
	ScsLabel *widgets.QLabel
	Scs *widgets.QComboBox
	PatternLabel *widgets.QLabel
	Pattern *widgets.QLineEdit
	KSsbLabel *widgets.QLabel
	KSsb *widgets.QLineEdit
	NCrbSsbLabel *widgets.QLabel
	NCrbSsb *widgets.QLineEdit
	widget *widgets.QWidget
}

type SsbGrid struct{
	Scs string
	Pattern string
	KSsb int
	NCrbSsb int
}

type SsbBurstUi struct{
	InOneGroupLabel *widgets.QLabel
	InOneGroup *widgets.QLineEdit
	GroupPresenceLabel *widgets.QLabel
	GroupPresence *widgets.QLineEdit
	PeriodLabel *widgets.QLabel
	Period *widgets.QComboBox
	widget *widgets.QWidget
}

type SsbBurst struct{
	InOneGroup string
	GroupPresence string
	Period string
}

type MibUi struct{
	SfnLabel *widgets.QLabel
	Sfn *widgets.QLineEdit
	HrfLabel *widgets.QLabel
	Hrf *widgets.QLineEdit
	DmrsTypeAPosLabel *widgets.QLabel
	DmrsTypeAPos *widgets.QComboBox
	ScsCommonLabel *widgets.QLabel
	ScsCommon *widgets.QComboBox
	Coreset0Label *widgets.QLabel
	Coreset0 *widgets.QLineEdit
	Coreset0InfoLabel *widgets.QLabel
	Css0Label *widgets.QLabel
	Css0 *widgets.QLineEdit
	widget *widgets.QGroupBox
}

type Mib struct{
	Sfn int
	Hrf int
	DmrsTypeAPos string
	ScsCommon string
	Coreset0 int
	Coreset0Details *Coreset0Info
	Css0 int
}

type GridSettingsUi struct{
	OpBandLabel *widgets.QLabel
	OpBand *widgets.QComboBox
	OpBandInfoLabel *widgets.QLabel
	SsbGridUi *SsbGridUi
	SsbBurstUi *SsbBurstUi
	MibUi *MibUi
	ScsSpecificCarrierUi *ScsSpecificCarrierUi
	widget *widgets.QWidget
}

type GridSettings struct{
	OpBand string
	SsbGrid *SsbGrid
	SsbBurst *SsbBurst
	Mib *Mib
	ScsSpecificCarrier *ScsSpecificCarrier
}

type NrGridUi struct{
	Debug bool
	Logger *zap.Logger
	LogEdit *widgets.QTextEdit
	Args map[string]interface{}

	gridSettingsUi *GridSettingsUi
	okBtn *widgets.QPushButton
	cancelBtn *widgets.QPushButton
	widget *widgets.QDialog
}

func (p *NrGridUi) InitUi() {
	// notes on Qt::Alignment:
	// void QGridLayout::addWidget(QWidget *widget, int fromRow, int fromColumn, int rowSpan, int columnSpan, Qt::Alignment alignment = Qt::Alignment())
	// The alignment is specified by alignment. The default alignment is 0, which means that the widget fills the entire cell.

	// notes on Qt::WindowsFlags:
	// QLabel::QLabel(const QString &text, QWidget *parent = nullptr, Qt::WindowFlags f = Qt::WindowFlags())
	// Qt::Widget	0x00000000	This is the default type for QWidget. Widgets of this type are child widgets if they have a parent, and independent windows if they have no parent. See also Qt::Window and Qt::SubWindow.

	p.okBtn = widgets.NewQPushButton2("OK", nil)
	p.cancelBtn = widgets.NewQPushButton2("Cancel", nil)

	p.gridSettingsUi = p.initGridSettingsUi()

	tabWidget := widgets.NewQTabWidget(nil)
	tabWidget.AddTab(p.gridSettingsUi.widget, "Grid Settings")

	hboxLayout := widgets.NewQHBoxLayout()
	hboxLayout.AddStretch(0)
	hboxLayout.AddWidget(p.okBtn, 0, 0)
	hboxLayout.AddWidget(p.cancelBtn, 0, 0)

	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(tabWidget, 0, 0)
	layout.AddLayout(hboxLayout, 0)

	p.widget = widgets.NewQDialog(nil, core.Qt__Widget)
	p.widget.SetLayout(layout)
	p.widget.SetWindowTitle("5G Resource Grid")

	p.initSlots()

	p.widget.Exec()
}

func (p *NrGridUi) initGridSettingsUi() *GridSettingsUi {
	ui := new(GridSettingsUi)

	ui.OpBandLabel = widgets.NewQLabel2("Operating band:", nil, core.Qt__Widget)
	ui.OpBand = widgets.NewQComboBox(nil)
	ui.OpBandInfoLabel = widgets.NewQLabel(nil, core.Qt__Widget)

	ui.SsbGridUi = p.initSsbGridUi()
	ui.SsbBurstUi = p.initSsbBurstUi()
	ui.MibUi = p.initMibUi()
	ui.ScsSpecificCarrierUi = p.initScsSpecificCarrierUi()

	tabWidget := widgets.NewQTabWidget(nil)
	tabWidget.AddTab(ui.SsbGridUi.widget, "SSB Grid")
	tabWidget.AddTab(ui.SsbBurstUi.widget, "SSB Burst(ServingCellConfigCommonSIB)")

	gridLayout := widgets.NewQGridLayout(nil)
	gridLayout.AddWidget2(ui.OpBandLabel, 0, 0, 0)
	gridLayout.AddWidget2(ui.OpBand, 0, 1, 0)
	gridLayout.AddWidget3(ui.OpBandInfoLabel, 1, 0, 1, 2, 0)
	gridLayout.AddWidget3(tabWidget, 2, 0, 1, 2, 0)
	gridLayout.AddWidget3(ui.MibUi.widget, 3, 0, 1, 2, 0)
	gridLayout.AddWidget3(ui.ScsSpecificCarrierUi.widget, 4, 0, 1, 2, 0)

	layout := widgets.NewQVBoxLayout()
	layout.AddLayout(gridLayout, 0)
	layout.AddStretch(0)

	ui.widget = widgets.NewQWidget(nil, core.Qt__Widget)
	ui.widget.SetLayout(layout)

	return ui
}

func (p *NrGridUi) initSsbGridUi() *SsbGridUi {
	ui := new(SsbGridUi)

	ui.ScsLabel = widgets.NewQLabel2("Subcarrier spacing:", nil, core.Qt__Widget)
	ui.Scs = widgets.NewQComboBox(nil)
	ui.PatternLabel = widgets.NewQLabel2("SSB pattern:", nil, core.Qt__Widget)
	ui.Pattern = widgets.NewQLineEdit(nil)
	ui.Pattern.SetEnabled(false)
	ui.KSsbLabel = widgets.NewQLabel2("k_SSB[0-23]:", nil, core.Qt__Widget)
	ui.KSsb = widgets.NewQLineEdit2("0", nil)
	ui.NCrbSsbLabel = widgets.NewQLabel2("n_CRB_SSB:", nil, core.Qt__Widget)
	ui.NCrbSsb = widgets.NewQLineEdit(nil)
	ui.NCrbSsb.SetEnabled(false)

	gridLayout := widgets.NewQGridLayout(nil)
	gridLayout.AddWidget2(ui.ScsLabel, 0, 0, 0)
	gridLayout.AddWidget2(ui.Scs, 0, 1, 0)
	gridLayout.AddWidget2(ui.PatternLabel, 1, 0, 0)
	gridLayout.AddWidget2(ui.Pattern, 1, 1, 0)
	gridLayout.AddWidget2(ui.KSsbLabel, 2, 0, 0)
	gridLayout.AddWidget2(ui.KSsb, 2, 1, 0)
	gridLayout.AddWidget2(ui.NCrbSsbLabel, 3, 0, 0)
	gridLayout.AddWidget2(ui.NCrbSsb, 3, 1, 0)

	layout := widgets.NewQVBoxLayout()
	layout.AddLayout(gridLayout, 0)
	layout.AddStretch(0)

	ui.widget = widgets.NewQWidget(nil, core.Qt__Widget)
	ui.widget.SetLayout(layout)

	return ui
}

func (p *NrGridUi) initSsbBurstUi() *SsbBurstUi {
	ui := new(SsbBurstUi)

	ui.InOneGroupLabel = widgets.NewQLabel2("inOneGroup(ssb-PositionsInBurst):", nil, core.Qt__Widget)
	ui.InOneGroup = widgets.NewQLineEdit(nil)
	ui.InOneGroup.SetPlaceholderText("11111111")
	ui.GroupPresenceLabel = widgets.NewQLabel2("groupPresence(ssb-PositionsInBurst):", nil, core.Qt__Widget)
	ui.GroupPresence = widgets.NewQLineEdit(nil)
	ui.GroupPresence.SetPlaceholderText("11111111")
	ui.PeriodLabel = widgets.NewQLabel2("ssb-PeriodicityServingCell:", nil, core.Qt__Widget)
	ui.Period = widgets.NewQComboBox(nil)
	ui.Period.AddItems([]string{"5ms", "10ms", "20ms", "40ms", "80ms", "160ms"})
	// refer to 3GPP 38.213 4.1
	// For initial cell selection, a UE may assume that half frames with SS/PBCH blocks occur with a periodicity of 2 frames.
	ui.Period.SetCurrentText("20ms")

	gridLayout := widgets.NewQGridLayout(nil)
	gridLayout.AddWidget2(ui.InOneGroupLabel, 0, 0, 0)
	gridLayout.AddWidget2(ui.InOneGroup, 0, 1, 0)
	gridLayout.AddWidget2(ui.GroupPresenceLabel, 1, 0, 0)
	gridLayout.AddWidget2(ui.GroupPresence, 1, 1, 0)
	gridLayout.AddWidget2(ui.PeriodLabel, 2, 0, 0)
	gridLayout.AddWidget2(ui.Period, 2, 1, 0)

	layout := widgets.NewQVBoxLayout()
	layout.AddLayout(gridLayout, 0)
	layout.AddStretch(0)

	ui.widget = widgets.NewQWidget(nil, core.Qt__Widget)
	ui.widget.SetLayout(layout)

	return ui
}

func (p *NrGridUi) initMibUi() *MibUi{
	ui := new(MibUi)

	ui.SfnLabel = widgets.NewQLabel2("SFN[0-1023]:", nil, core.Qt__Widget)
	ui.Sfn = widgets.NewQLineEdit2("0", nil)
	ui.HrfLabel = widgets.NewQLabel2("Half frame bit[0/1]:", nil, core.Qt__Widget)
	ui.Hrf = widgets.NewQLineEdit2("0", nil)
	ui.DmrsTypeAPosLabel = widgets.NewQLabel2("dmrs-TypeA-Position:", nil, core.Qt__Widget)
	ui.DmrsTypeAPos = widgets.NewQComboBox(nil)
	ui.DmrsTypeAPos.AddItems([]string{"pos2", "pos3"})
	ui.ScsCommonLabel = widgets.NewQLabel2("subCarrierSpacingCommon:", nil, core.Qt__Widget)
	ui.ScsCommon = widgets.NewQComboBox(nil)
	ui.Coreset0Label = widgets.NewQLabel2("coresetZero(PDCCH-ConfigSIB1)[0-15]:", nil, core.Qt__Widget)
	ui.Coreset0 = widgets.NewQLineEdit2("0", nil)
	ui.Coreset0InfoLabel = widgets.NewQLabel(nil, core.Qt__Widget)
	ui.Css0Label = widgets.NewQLabel2("searchSpaceZero(PDCCH-ConfigSIB1)[0-15]:", nil, core.Qt__Widget)
	ui.Css0 = widgets.NewQLineEdit2("0", nil)

	gridLayout := widgets.NewQGridLayout(nil)
	gridLayout.AddWidget2(ui.SfnLabel, 0, 0, 0)
	gridLayout.AddWidget2(ui.Sfn, 0, 1, 0)
	gridLayout.AddWidget2(ui.HrfLabel, 1, 0, 0)
	gridLayout.AddWidget2(ui.Hrf, 1, 1, 0)
	gridLayout.AddWidget2(ui.DmrsTypeAPosLabel, 2, 0, 0)
	gridLayout.AddWidget2(ui.DmrsTypeAPos, 2, 1, 0)
	gridLayout.AddWidget2(ui.ScsCommonLabel, 3, 0, 0)
	gridLayout.AddWidget2(ui.ScsCommon, 3, 1, 0)
	gridLayout.AddWidget2(ui.Coreset0Label, 4, 0, 0)
	gridLayout.AddWidget2(ui.Coreset0, 4, 1, 0)
	gridLayout.AddWidget3(ui.Coreset0InfoLabel, 5, 0, 1, 2, 0)
	gridLayout.AddWidget2(ui.Css0Label, 6, 0, 0)
	gridLayout.AddWidget2(ui.Css0, 6, 1, 0)

	ui.widget = widgets.NewQGroupBox2("MIB", nil)
	ui.widget.SetLayout(gridLayout)

	return ui
}

func (p *NrGridUi) initScsSpecificCarrierUi() *ScsSpecificCarrierUi{
	ui := new(ScsSpecificCarrierUi)

	ui.ScsLabel = widgets.NewQLabel2("subcarrierSpacing:", nil, core.Qt__Widget)
	ui.Scs = widgets.NewQComboBox(nil)
	ui.BwLabel = widgets.NewQLabel2("Transmission bandwidth:", nil, core.Qt__Widget)
	ui.Bw = widgets.NewQComboBox(nil)
	ui.NRbLabel = widgets.NewQLabel2("N_RB(carrierBandwidth):", nil, core.Qt__Widget)
	ui.NRb = widgets.NewQLineEdit(nil)
	ui.NRb.SetEnabled(false)
	ui.OffsetToCarrierLabel = widgets.NewQLabel2("offsetToCarrier:", nil, core.Qt__Widget)
	ui.OffsetToCarrier = widgets.NewQLineEdit2("0", nil)

	gridLayout := widgets.NewQGridLayout(nil)
	gridLayout.AddWidget2(ui.ScsLabel, 0, 0, 0)
	gridLayout.AddWidget2(ui.Scs, 0, 1, 0)
	gridLayout.AddWidget2(ui.BwLabel, 1, 0, 0)
	gridLayout.AddWidget2(ui.Bw, 1, 1, 0)
	gridLayout.AddWidget2(ui.NRbLabel, 2, 0, 0)
	gridLayout.AddWidget2(ui.NRb, 2, 1, 0)
	gridLayout.AddWidget2(ui.OffsetToCarrierLabel, 3, 0, 0)
	gridLayout.AddWidget2(ui.OffsetToCarrier, 3, 1, 0)

	ui.widget = widgets.NewQGroupBox2("Carrier Grid(SCS-SpecificCarrier)", nil)
	ui.widget.SetLayout(gridLayout)

	return ui
}

func (p *NrGridUi) initSlots() {
	p.okBtn.ConnectClicked(p.onOkBtnClicked)
	p.cancelBtn.ConnectClicked(p.onCancelBtnClicked)
}

func (p *NrGridUi) onOkBtnClicked(checked bool) {
	//p.Logger.Info("general info", zap.String("test", "something"))
	//p.Logger.Error("error raised", zap.String("test", "something"))
	p.widget.Accept()
}

func (p *NrGridUi) onCancelBtnClicked(checked bool) {
	p.widget.Reject()
}



