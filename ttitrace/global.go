package ttitrace

type TtiEventHeader struct {
	Hsfn int
	Sfn  int
	Slot int
	Rnti int
	PhysCellId int
}

type TtiEventHeaderPos struct {
	PosHsfn int
	PosSfn  int
	PosSlot int
	PosRnti int
	PosPhysCellId int
}

type TtiDlBeamData struct {
	TtiEventHeader
	SubcellId          int
	CurrentBestBeamId  int
	Current2ndBeamId   int
	SelectedBestBeamId int
	Selected2ndBeamId  int
}

type TtiDlBeamDataPos struct {
	Ready bool
	TtiEventHeaderPos
	PosSubcellId          int
	PosCurrentBestBeamId  int
	PosCurrent2ndBeamId   int
	PosSelectedBestBeamId int
	PosSelected2ndBeamId  int
}

type TtiDlPreSchedData struct {
	TtiEventHeader
	CsListEvent string
	HighestClassPriority string
}

type TtiDlPreSchedDataPos struct {
	Ready bool
	TtiEventHeaderPos
	PosCsListEvent int
	PosHighestClassPriority int
}

type TtiDlTdSchedSubcellData struct {
	TtiEventHeader
	SubcellId int
	Cs2List []int
}

type TtiDlTdSchedSubcellDataPos struct {
	Ready bool
	TtiEventHeaderPos
	PosSubcellId int
	PosRecordSequenceNumber int
}

type TtiDlFdSchedData struct {
	TtiEventHeader
	CellDbIndex int
	TxNumber int
	DlHarqProcessIndex int
	K1 int
	AllFields []string
}

type TtiDlFdSchedDataPos struct {
	Ready bool
	TtiEventHeaderPos
	PosCellDbIndex int
	PosTxNumber int
	PosDlHarqProcessIndex int
	PosK1 int
}

type TtiDlHarqRxData struct {
	TtiEventHeader
	HarqSubcellId int
	AckNack string
	DlHarqProcessIndex int
}

type TtiDlHarqRxDataPos struct {
	Ready bool
	TtiEventHeaderPos
	PosHarqSubcellId int
	PosAckNack int
	PosDlHarqProcessIndex int
}

type TtiDlLaAverageCqi struct {
	TtiEventHeader
	CellDbIndex int
	RrmAvgCqi float64
	RrmDeltaCqi float64
}

type TtiDlLaAverageCqiPos struct {
	Ready bool
	TtiEventHeaderPos
	PosCellDbIndex int
	PosRrmAvgCqi int
	PosRrmDeltaCqi int
}


