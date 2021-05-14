package ttitrace

import (
	"strings"
)

type TtiEventHeader struct {
	Hsfn string
	Sfn string
	Slot string
	Rnti string
	PhysCellId string
}

type TtiEventHeaderPos struct {
	PosHsfn int
	PosSfn  int
	PosSlot int
	PosRnti int
	PosPhysCellId int
}

func FindTtiEventHeaderPos(tokens []string) TtiEventHeaderPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 5
	p := TtiEventHeaderPos{
		PosHsfn: -1,
		PosSfn: -1,
		PosSlot: -1,
		PosRnti: -1,
		PosPhysCellId: -1,
	}

	// Field hsfn is user-defined!
	i := 1
	for pos, item := range tokens {
		if strings.ToLower(item) == "sfn" && p.PosSfn < 0 {
			p.PosSfn = pos
			i += 1
		} else if strings.ToLower(item) == "slot" && p.PosSlot < 0 {
			p.PosSlot = pos
			i += 1
		} else if strings.ToLower(item) == "rnti" && p.PosRnti < 0 {
			p.PosRnti = pos
			i += 1
		} else if strings.ToLower(item) == "physcellid" && p.PosPhysCellId < 0 {
			p.PosPhysCellId = pos
			i += 1
		}

		if i >= n {
			break
		}
	}

	return p
}

type TtiDlBeamData struct {
	TtiEventHeader
	SubcellId          string
	CurrentBestBeamId  string
	Current2ndBeamId   string
	SelectedBestBeamId string
	Selected2ndBeamId  string
}

type TtiDlBeamDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosSubcellId          int
	PosCurrentBestBeamId  int
	PosCurrent2ndBeamId   int
	PosSelectedBestBeamId int
	PosSelected2ndBeamId  int
}


func FindTtiDlBeamDataPos(tokens []string) TtiDlBeamDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 5
	p := TtiDlBeamDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosSubcellId: -1,
		PosCurrentBestBeamId: -1,
		PosCurrent2ndBeamId: -1,
		PosSelectedBestBeamId: -1,
		PosSelected2ndBeamId: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "subcellid" && p.PosSubcellId < 0 {
			p.PosSubcellId = pos
			i += 1
		} else if strings.ToLower(item) == "currentbestbeamid" && p.PosCurrentBestBeamId < 0 {
			p.PosCurrentBestBeamId = pos
			i += 1
		} else if strings.ToLower(item) == "current2ndbeamid" && p.PosCurrent2ndBeamId < 0 {
			p.PosCurrent2ndBeamId = pos
			i += 1
		} else if strings.ToLower(item) == "selectedbestbeamid" && p.PosSelectedBestBeamId < 0 {
			p.PosSelectedBestBeamId = pos
			i += 1
		} else if strings.ToLower(item) == "selected2ndbeamid" && p.PosSelected2ndBeamId < 0 {
			p.PosSelected2ndBeamId = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiDlPreSchedData struct {
	TtiEventHeader
	CsListEvent string
	HighestClassPriority string
	PrachPreambleIndex string
}

type TtiDlPreSchedDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosCsListEvent int
	PosHighestClassPriority int
	PosPrachPreambleIndex int
}

func FindTtiDlPreSchedDataPos(tokens []string) TtiDlPreSchedDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 3
	p := TtiDlPreSchedDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosCsListEvent: -1,
		PosHighestClassPriority: -1,
		PosPrachPreambleIndex: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "cslistevent" && p.PosCsListEvent < 0 {
			p.PosCsListEvent = pos
			i += 1
		} else if strings.ToLower(item) == "highestclasspriority" && p.PosHighestClassPriority < 0 {
			p.PosHighestClassPriority = pos
			i += 1
		} else if strings.ToLower(item) == "prachpreambleindex" && p.PosPrachPreambleIndex< 0 {
			p.PosPrachPreambleIndex= pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiDlTdSchedSubcellData struct {
	TtiEventHeader
	SubcellId string
	Cs2List []string
}

type TtiDlTdSchedSubcellDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosSubcellId int
	PosRecordSequenceNumber []int
}

func FindTtiDlTdSchedSubcellDataPos(tokens []string) TtiDlTdSchedSubcellDataPos {
	p := TtiDlTdSchedSubcellDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosSubcellId: -1,
		PosRecordSequenceNumber: make([]int, 0),
	}

	for pos, item := range tokens {
		if strings.ToLower(item) == "subcellid" && p.PosSubcellId < 0 {
			p.PosSubcellId = pos
		} else if strings.ToLower(item) == "recordsequencenumber" {
			p.PosRecordSequenceNumber = append(p.PosRecordSequenceNumber, pos)
		}
	}

	p.Ready = true
	return p
}

type TtiDlFdSchedData struct {
	TtiEventHeader
	CellDbIndex        string
	SubcellId string
	TxNumber           string
	DlHarqProcessIndex string
	K1                 string
	LcIdList           []string
	AllFields          []string
}

type TtiDlFdSchedDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosCellDbIndex int
	PosSubcellId int
	PosTxNumber int
	PosDlHarqProcessIndex int
	PosK1 int

	// additional position for RIV/SLIV/AntPort/per-Bearer post-processing
	PosNumOfPrb int
	PosStartPrb int
	PosSliv int
	PosAntPort int
	PosLcId int
}

func FindTtiDlFdSchedDataPos(tokens []string) TtiDlFdSchedDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 10
	p := TtiDlFdSchedDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosCellDbIndex: -1,
		PosSubcellId: -1,
		PosTxNumber: -1,
		PosDlHarqProcessIndex: -1,
		PosK1: -1,
		PosNumOfPrb: -1,
		PosStartPrb: -1,
		PosSliv: -1,
		PosAntPort: -1,
		PosLcId: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "celldbindex" && p.PosCellDbIndex < 0 {
			p.PosCellDbIndex = pos
			i += 1
		} else if strings.ToLower(item) == "subcellid" && p.PosSubcellId < 0 {
			p.PosSubcellId = pos
			i += 1
		} else if strings.ToLower(item) == "txnumber" && p.PosTxNumber < 0 {
			p.PosTxNumber = pos
			i += 1
		} else if strings.ToLower(item) == "dlharqprocessindex" && p.PosDlHarqProcessIndex < 0 {
			p.PosDlHarqProcessIndex = pos
			i += 1
		} else if strings.ToLower(item) == "k1" && p.PosK1 < 0 {
			p.PosK1 = pos
			i += 1
		} else if strings.ToLower(item) == "numofprb" && p.PosNumOfPrb < 0 {
			p.PosNumOfPrb= pos
			i += 1
		} else if strings.ToLower(item) == "startprb" && p.PosStartPrb < 0 {
			p.PosStartPrb = pos
			i += 1
		} else if strings.ToLower(item) == "sliv" && p.PosSliv < 0 {
			p.PosSliv = pos
			i += 1
		} else if strings.ToLower(item) == "antport" && p.PosAntPort < 0 {
			p.PosAntPort = pos
			i += 1
		} else if strings.ToLower(item) == "lcid" && p.PosLcId < 0 {
			p.PosLcId = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiDlHarqRxData struct {
	TtiEventHeader
	HarqSubcellId string
	AckNack string
	DlHarqProcessIndex string
	PucchFormat string
}

type TtiDlHarqRxDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosHarqSubcellId int
	PosAckNack int
	PosDlHarqProcessIndex int
	PosPucchFormat int
}

func FindTtiDlHarqRxDataPos(tokens []string) TtiDlHarqRxDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 4
	p := TtiDlHarqRxDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosHarqSubcellId: -1,
		PosAckNack: -1,
		PosDlHarqProcessIndex: -1,
		PosPucchFormat: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "harqsubcellid" && p.PosHarqSubcellId < 0 {
			p.PosHarqSubcellId = pos
			i += 1
		} else if strings.ToLower(item) == "acknack" && p.PosAckNack < 0 {
			p.PosAckNack = pos
			i += 1
		} else if strings.ToLower(item) == "dlharqprocessindex" && p.PosDlHarqProcessIndex < 0 {
			p.PosDlHarqProcessIndex = pos
			i += 1
		} else if strings.ToLower(item) == "pucchformat" && p.PosPucchFormat < 0 {
			p.PosPucchFormat = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiDlLaAverageCqi struct {
	TtiEventHeader
	CellDbIndex string
	RrmInstCqi string
	Rank string
	RrmAvgCqi string
	Mcs string
	RrmDeltaCqi string
}

type TtiDlLaAverageCqiPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosCellDbIndex int
	PosRrmInstCqi int
	PosRank int
	PosRrmAvgCqi int
	PosMcs int
	PosRrmDeltaCqi int
}

func FindTtiDlLaAverageCqiPos(tokens []string) TtiDlLaAverageCqiPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 6
	p := TtiDlLaAverageCqiPos{
		Ready:                 false,
		PosEventHeader:        FindTtiEventHeaderPos(tokens),
		PosCellDbIndex:      -1,
		PosRrmInstCqi: -1,
		PosRank: -1,
		PosRrmAvgCqi:            -1,
		PosMcs: -1,
		PosRrmDeltaCqi: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "celldbindex" && p.PosCellDbIndex < 0 {
			p.PosCellDbIndex = pos
			i += 1
		} else if strings.ToLower(item) == "rrminstcqi" && p.PosRrmInstCqi < 0 {
			p.PosRrmInstCqi = pos
			i += 1
		} else if strings.ToLower(item) == "rank" && p.PosRank < 0 {
			p.PosRank = pos
			i += 1
		} else if strings.ToLower(item) == "rrmavgcqi" && p.PosRrmAvgCqi < 0 {
			p.PosRrmAvgCqi = pos
			i += 1
		} else if strings.ToLower(item) == "mcs" && p.PosMcs < 0 {
			p.PosMcs = pos
			i += 1
		} else if strings.ToLower(item) == "rrmdeltacqi" && p.PosRrmDeltaCqi < 0 {
			p.PosRrmDeltaCqi = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiCsiSrReportData struct {
	TtiEventHeader
	UlChannel string
	Dtx string
	PucchFormat string
	Cqi string
	PmiRank1 string
	PmiRank2 string
	Ri string
	Cri string
	Li string
	Sr string
}

type TtiCsiSrReportDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosUlChannel int
	PosDtx int
	PosPucchFormat int
	PosCqi int
	PosPmiRank1 int
	PosPmiRank2 int
	PosRi int
	PosCri int
	PosLi int
	PosSr int
}

func FindTtiCsiSrReportDataPos(tokens []string) TtiCsiSrReportDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 10
	p := TtiCsiSrReportDataPos{
		Ready:                 false,
		PosEventHeader:        FindTtiEventHeaderPos(tokens),
		PosUlChannel:      -1,
		PosDtx:      -1,
		PosPucchFormat:      -1,
		PosCqi:      -1,
		PosPmiRank1:      -1,
		PosPmiRank2:      -1,
		PosRi:      -1,
		PosCri:      -1,
		PosLi:      -1,
		PosSr:      -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "ulchannel" && p.PosUlChannel < 0 {
			p.PosUlChannel = pos
			i += 1
		} else if strings.ToLower(item) == "dtx" && p.PosDtx < 0 {
			p.PosDtx = pos
			i += 1
		} else if strings.ToLower(item) == "pucchformat" && p.PosPucchFormat < 0 {
			p.PosPucchFormat = pos
			i += 1
		} else if strings.ToLower(item) == "cqi" && p.PosCqi < 0 {
			p.PosCqi = pos
			i += 1
		} else if strings.ToLower(item) == "pmirank1" && p.PosPmiRank1 < 0 {
			p.PosPmiRank1 = pos
			i += 1
		} else if strings.ToLower(item) == "pmirank2" && p.PosPmiRank2 < 0 {
			p.PosPmiRank2 = pos
			i += 1
		} else if strings.ToLower(item) == "ri" && p.PosRi < 0 {
			p.PosRi = pos
			i += 1
		} else if strings.ToLower(item) == "cri" && p.PosCri < 0 {
			p.PosCri = pos
			i += 1
		} else if strings.ToLower(item) == "li" && p.PosLi < 0 {
			p.PosLi = pos
			i += 1
		} else if strings.ToLower(item) == "sr" && p.PosSr < 0 {
			p.PosSr = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiDlFlowControlData struct {
	TtiEventHeader
	LchId string
	ReportType string
	ScheduledBytes string
	EthAvg string
	EthScaled string
}

type TtiDlFlowControlDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosLchId int
	PosReportType int
	PosScheduledBytes int
	PosEthAvg int
	PosEthScaled int
}

func FindTtiDlFlowControlDataPos(tokens []string) TtiDlFlowControlDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 5
	p := TtiDlFlowControlDataPos{
		Ready:                 false,
		PosEventHeader:        FindTtiEventHeaderPos(tokens),
		PosLchId:      -1,
		PosReportType:      -1,
		PosScheduledBytes:      -1,
		PosEthAvg:      -1,
		PosEthScaled:      -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "lchid" && p.PosLchId < 0 {
			p.PosLchId = pos
			i += 1
		} else if strings.ToLower(item) == "reporttype" && p.PosReportType < 0 {
			p.PosReportType = pos
			i += 1
		} else if strings.ToLower(item) == "scheduledbytes" && p.PosScheduledBytes < 0 {
			p.PosScheduledBytes = pos
			i += 1
		} else if strings.ToLower(item) == "ethavg" && p.PosEthAvg < 0 {
			p.PosEthAvg = pos
			i += 1
		} else if strings.ToLower(item) == "ethscaled" && p.PosEthScaled < 0 {
			p.PosEthScaled = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiDlLaDeltaCqi struct {
	TtiEventHeader
	CellDbIndex string
	IsDeltaCqiCalculated string
	RrmPauseUeInDlScheduling string
	HarqFb string
	RrmDeltaCqi string
	RrmRemainingBucketLevel string
}

type TtiDlLaDeltaCqiPos struct {
	Ready                       bool
	PosEventHeader              TtiEventHeaderPos
	PosCellDbIndex              int
	PosIsDeltaCqiCalculated     int
	PosRrmPauseUeInDlScheduling int
	PosHarqFb                   int
	PosRrmDeltaCqi              int
	PosRrmRemainingBucketLevel  int
}

func FindTtiDlLaDeltaCqiPos(tokens []string) TtiDlLaDeltaCqiPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 6
	p := TtiDlLaDeltaCqiPos{
		Ready:                       false,
		PosEventHeader:              FindTtiEventHeaderPos(tokens),
		PosCellDbIndex:              -1,
		PosIsDeltaCqiCalculated:     -1,
		PosRrmPauseUeInDlScheduling: -1,
		PosHarqFb:                   -1,
		PosRrmDeltaCqi:              -1,
		PosRrmRemainingBucketLevel:  -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "celldbindex" && p.PosCellDbIndex < 0 {
			p.PosCellDbIndex = pos
			i += 1
		} else if strings.ToLower(item) == "isdeltacqicalculated" && p.PosIsDeltaCqiCalculated < 0 {
			p.PosIsDeltaCqiCalculated = pos
			i += 1
		} else if strings.ToLower(item) == "rrmpauseueindlscheduling" && p.PosRrmPauseUeInDlScheduling < 0 {
			p.PosRrmPauseUeInDlScheduling = pos
			i += 1
		} else if strings.ToLower(item) == "harqfb" && p.PosHarqFb < 0 {
			p.PosHarqFb = pos
			i += 1
		} else if strings.ToLower(item) == "rrmdeltacqi" && p.PosRrmDeltaCqi < 0 {
			p.PosRrmDeltaCqi= pos
			i += 1
		} else if strings.ToLower(item) == "rrmremainingbucketlevel" && p.PosRrmRemainingBucketLevel < 0 {
			p.PosRrmRemainingBucketLevel = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlBsrRxData struct {
	TtiEventHeader
	UlHarqProcessIndex string
	BsrFormat string
	BufferSizeList []string
}

type TtiUlBsrRxDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosUlHarqProcessIndex int
	PosBsrFormat int
}

func FindTtiUlBsrRxDataPos(tokens []string) TtiUlBsrRxDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 2
	p := TtiUlBsrRxDataPos{
		Ready:                       false,
		PosEventHeader:              FindTtiEventHeaderPos(tokens),
		PosUlHarqProcessIndex:              -1,
		PosBsrFormat:     -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "ulharqprocessindex" && p.PosUlHarqProcessIndex < 0 {
			p.PosUlHarqProcessIndex = pos
			i += 1
		} else if strings.ToLower(item) == "bsrformat" && p.PosBsrFormat < 0 {
			p.PosBsrFormat = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlFdSchedData struct {
	TtiEventHeader
	CellDbIndex        string
	SubcellId	string
	TxNumber           string
	UlHarqProcessIndex string
	K2                 string
	AllFields          []string
}

type TtiUlFdSchedDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosCellDbIndex int
	PosSubcellId int
	PosTxNumber int
	PosUlHarqProcessIndex int
	PosK2 int

	// additional position for RIV/SLIV/AntPort/per-Bearer post-processing
	PosNumOfPrb int
	PosStartPrb int
	PosSliv int
	PosAntPort int
}

func FindTtiUlFdSchedDataPos(tokens []string) TtiUlFdSchedDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 9
	p := TtiUlFdSchedDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosCellDbIndex: -1,
		PosSubcellId: -1,
		PosTxNumber: -1,
		PosUlHarqProcessIndex: -1,
		PosK2: -1,
		PosNumOfPrb: -1,
		PosStartPrb: -1,
		PosSliv: -1,
		PosAntPort: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "celldbindex" && p.PosCellDbIndex < 0 {
			p.PosCellDbIndex = pos
			i += 1
		} else if strings.ToLower(item) == "subcellid" && p.PosSubcellId < 0 {
			p.PosSubcellId = pos
			i += 1
		} else if strings.ToLower(item) == "txnumber" && p.PosTxNumber < 0 {
			p.PosTxNumber = pos
			i += 1
		} else if strings.ToLower(item) == "ulharqprocessindex" && p.PosUlHarqProcessIndex < 0 {
			p.PosUlHarqProcessIndex = pos
			i += 1
		} else if strings.ToLower(item) == "k2" && p.PosK2 < 0 {
			p.PosK2 = pos
			i += 1
		} else if strings.ToLower(item) == "numofprb" && p.PosNumOfPrb < 0 {
			p.PosNumOfPrb= pos
			i += 1
		} else if strings.ToLower(item) == "startprb" && p.PosStartPrb < 0 {
			p.PosStartPrb = pos
			i += 1
		} else if strings.ToLower(item) == "sliv" && p.PosSliv < 0 {
			p.PosSliv = pos
			i += 1
		} else if strings.ToLower(item) == "antport" && p.PosAntPort < 0 {
			p.PosAntPort = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlHarqRxData struct {
	TtiEventHeader
	SubcellId string
	Dtx string
	CrcResult string
	UlHarqProcessIndex string
}

type TtiUlHarqRxDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosSubcellId int
	PosDtx int
	PosCrcResult int
	PosUlHarqProcessIndex int
}

func FindTtiUlHarqRxDataPos(tokens []string) TtiUlHarqRxDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 4
	p := TtiUlHarqRxDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosSubcellId: -1,
		PosDtx: -1,
		PosCrcResult: -1,
		PosUlHarqProcessIndex: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "subcellid" && p.PosSubcellId < 0 {
			p.PosSubcellId = pos
			i += 1
		} else if strings.ToLower(item) == "dtx" && p.PosDtx < 0 {
			p.PosDtx = pos
			i += 1
		} else if strings.ToLower(item) == "crcresult" && p.PosCrcResult < 0 {
			p.PosCrcResult = pos
			i += 1
		} else if strings.ToLower(item) == "ulharqprocessindex" && p.PosUlHarqProcessIndex < 0 {
			p.PosUlHarqProcessIndex = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlIntraDlToUlDtxSyncDlData struct {
	TtiEventHeader
	DrxEnabled string
	DlDrxOnDurationTimerOn string
	DlDrxInactivityTimerOn string
}

type TtiUlIntraDlToUlDrxSyncDlDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosDrxEnabled int
	PosDlDrxOnDurationTimerOn int
	PosDlDrxInactivityTimerOn int
}

func FindTtiUlIntraDlToUlDrxSyncDlDataPos(tokens []string) TtiUlIntraDlToUlDrxSyncDlDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 3
	p := TtiUlIntraDlToUlDrxSyncDlDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosDrxEnabled: -1,
		PosDlDrxOnDurationTimerOn: -1,
		PosDlDrxInactivityTimerOn: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "drxenabled" && p.PosDrxEnabled< 0 {
			p.PosDrxEnabled = pos
			i += 1
		} else if strings.ToLower(item) == "dldrxondurationtimeron" && p.PosDlDrxOnDurationTimerOn < 0 {
			p.PosDlDrxOnDurationTimerOn = pos
			i += 1
		} else if strings.ToLower(item) == "dldrxinactivitytimeron" && p.PosDlDrxInactivityTimerOn < 0 {
			p.PosDlDrxInactivityTimerOn = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlLaDeltaSinr struct {
	TtiEventHeader
	CellDbIndex string
	IsDeltaSinrCalculated string
	RrmPauseUeInUlScheduling string
	CrcFb string
	RrmDeltaSinr string
}

type TtiUlLaDeltaSinrPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosCellDbIndex int
	PosIsDeltaSinrCalculated int
	PosRrmPauseUeInUlScheduling int
	PosCrcFb int
	PosRrmDeltaSinr int
}


func FindTtiUlLaDeltaSinrPos(tokens []string) TtiUlLaDeltaSinrPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 5
	p := TtiUlLaDeltaSinrPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosCellDbIndex: -1,
		PosIsDeltaSinrCalculated: -1,
		PosRrmPauseUeInUlScheduling: -1,
		PosCrcFb: -1,
		PosRrmDeltaSinr: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "celldbindex" && p.PosCellDbIndex < 0 {
			p.PosCellDbIndex = pos
			i += 1
		} else if strings.ToLower(item) == "isdeltasinrcalculated" && p.PosIsDeltaSinrCalculated < 0 {
			p.PosIsDeltaSinrCalculated = pos
			i += 1
		} else if strings.ToLower(item) == "rrmpauseueinulscheduling" && p.PosRrmPauseUeInUlScheduling < 0 {
			p.PosRrmPauseUeInUlScheduling = pos
			i += 1
		} else if strings.ToLower(item) == "crcfb" && p.PosCrcFb < 0 {
			p.PosCrcFb = pos
			i += 1
		} else if strings.ToLower(item) == "rrmdeltasinr" && p.PosRrmDeltaSinr < 0 {
			p.PosRrmDeltaSinr = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlLaAverageSinr struct {
	TtiEventHeader
	CellDbIndex string
	RrmInstSinrRank string
	RrmNumOfSinrMeasurements string
	RrmInstSinr string
	RrmAvgSinrUl string
	RrmSinrCorrection string
}

type TtiUlLaAverageSinrPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosCellDbIndex int
	PosRrmInstSinrRank int
	PosRrmNumOfSinrMeasurements int
	PosRrmInstSinr int
	PosRrmAvgSinrUl int
	PosRrmSinrCorrection int
}

func FindTtiUlLaAverageSinrPos(tokens []string) TtiUlLaAverageSinrPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 6
	p := TtiUlLaAverageSinrPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosCellDbIndex: -1,
		PosRrmInstSinrRank: -1,
		PosRrmNumOfSinrMeasurements: -1,
		PosRrmInstSinr: -1,
		PosRrmAvgSinrUl: -1,
		PosRrmSinrCorrection: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "celldbindex" && p.PosCellDbIndex < 0 {
			p.PosCellDbIndex = pos
			i += 1
		} else if strings.ToLower(item) == "rrminstsinrrank" && p.PosRrmInstSinrRank < 0 {
			p.PosRrmInstSinrRank = pos
			i += 1
		} else if strings.ToLower(item) == "rrmnumofsinrmeasurements" && p.PosRrmNumOfSinrMeasurements < 0 {
			p.PosRrmNumOfSinrMeasurements = pos
			i += 1
		} else if strings.ToLower(item) == "rrminstsinr" && p.PosRrmInstSinr < 0 {
			p.PosRrmInstSinr = pos
			i += 1
		} else if strings.ToLower(item) == "rrmavgsinrul" && p.PosRrmAvgSinrUl < 0 {
			p.PosRrmAvgSinrUl = pos
			i += 1
		} else if strings.ToLower(item) == "rrmsinrcorrection" && p.PosRrmSinrCorrection < 0 {
			p.PosRrmSinrCorrection = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlLaPhr struct {
	TtiEventHeader
	CellDbIndex string
	IsRrmPhrScaledCalculated string
	Phr string
	RrmNumPuschPrb string
	RrmPhrScaled string
}

type TtiUlLaPhrPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosCellDbIndex int
	PosIsRrmPhrScaledCalculated int
	PosPhr int
	PosRrmNumPuschPrb int
	PosRrmPhrScaled int
}

func FindTtiUlLaPhrPos(tokens []string) TtiUlLaPhrPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 5
	p := TtiUlLaPhrPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosCellDbIndex: -1,
		PosIsRrmPhrScaledCalculated: -1,
		PosPhr: -1,
		PosRrmNumPuschPrb: -1,
		PosRrmPhrScaled: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "celldbindex" && p.PosCellDbIndex < 0 {
			p.PosCellDbIndex = pos
			i += 1
		} else if strings.ToLower(item) == "isrrmphrscaledcalculated" && p.PosIsRrmPhrScaledCalculated < 0 {
			p.PosIsRrmPhrScaledCalculated = pos
			i += 1
		} else if strings.ToLower(item) == "phr" && p.PosPhr < 0 {
			p.PosPhr = pos
			i += 1
		} else if strings.ToLower(item) == "rrmnumpuschprb" && p.PosRrmNumPuschPrb < 0 {
			p.PosRrmNumPuschPrb = pos
			i += 1
		} else if strings.ToLower(item) == "rrmphrscaled" && p.PosRrmPhrScaled < 0 {
			p.PosRrmPhrScaled = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlPucchReceiveRespPsData struct {
	TtiEventHeader
	PucchFormat string
	StartPrb string
	Rssi string
	SinrLayer0 string //sinr_[0]
	SinrLayer1 string //sinr_[1]
	Dtx string
	SrBit string
	SubcellId string
}

type TtiUlPucchReceiveRespPsDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosPucchFormat int
	PosStartPrb int
	PosRssi int
	PosSinrLayer0 int
	PosSinrLayer1 int
	PosDtx int
	PosSrBit int
	PosSubcellId int
}

func FindTtiUlPucchReceiveRespPsDataPos(tokens []string) TtiUlPucchReceiveRespPsDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 8
	p := TtiUlPucchReceiveRespPsDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosPucchFormat: -1,
		PosStartPrb: -1,
		PosRssi: -1,
		PosSinrLayer0: -1,
		PosSinrLayer1: -1,
		PosDtx: -1,
		PosSrBit: -1,
		PosSubcellId: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "pucchformat" && p.PosPucchFormat < 0 {
			p.PosPucchFormat = pos
			i += 1
		} else if strings.ToLower(item) == "startprb" && p.PosStartPrb < 0 {
			p.PosStartPrb = pos
			i += 1
		} else if strings.ToLower(item) == "rssi" && p.PosRssi < 0 {
			p.PosRssi = pos
			i += 1
		} else if strings.ToLower(item) == "sinr_[0]" && p.PosSinrLayer0 < 0 {
			p.PosSinrLayer0 = pos
			i += 1
		} else if strings.ToLower(item) == "sinr_[1]" && p.PosSinrLayer1 < 0 {
			p.PosSinrLayer1 = pos
			i += 1
		} else if strings.ToLower(item) == "dtx" && p.PosDtx < 0 {
			p.PosDtx = pos
			i += 1
		} else if strings.ToLower(item) == "srbit" && p.PosSrBit < 0 {
			p.PosSrBit = pos
			i += 1
		} else if strings.ToLower(item) == "subcellid" && p.PosSubcellId < 0 {
			p.PosSubcellId = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlPuschReceiveRespPsData struct {
	TtiEventHeader
	Rssi string
	SinrLayer0 string // sinr_[0]
	SinrLayer1 string // sinr_[1]
	Dtx string
	UlRank string
	UlPmiRank1 string
	UlPmiRank1Sinr string
	UlPmiRank2 string
	UlPmiRank2SinrLayer0 string //ulPmiRank2Sinr_[0]
	UlPmiRank2SinrLayer1 string //ulPmiRank2Sinr_[1]
	LongTermRank string
}

type TtiUlPuschReceiveRespPsDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosRssi int
	PosSinrLayer0 int
	PosSinrLayer1 int
	PosDtx int
	PosUlRank int
	PosUlPmiRank1 int
	PosUlPmiRank1Sinr int
	PosUlPmiRank2 int
	PosUlPmiRank2SinrLayer0 int
	PosUlPmiRank2SinrLayer1 int
	PosLongTermRank int
}

func FindTtiUlPuschReceiveRespPsDataPos(tokens []string) TtiUlPuschReceiveRespPsDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 11
	p := TtiUlPuschReceiveRespPsDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosRssi: -1,
		PosSinrLayer0: -1,
		PosSinrLayer1: -1,
		PosDtx: -1,
		PosUlRank: -1,
		PosUlPmiRank1: -1,
		PosUlPmiRank1Sinr: -1,
		PosUlPmiRank2: -1,
		PosUlPmiRank2SinrLayer0: -1,
		PosUlPmiRank2SinrLayer1: -1,
		PosLongTermRank: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "rssi" && p.PosRssi < 0 {
			p.PosRssi = pos
			i += 1
		} else if strings.ToLower(item) == "sinr_[0]" && p.PosSinrLayer0 < 0 {
			p.PosSinrLayer0 = pos
			i += 1
		} else if strings.ToLower(item) == "sinr_[1]" && p.PosSinrLayer1 < 0 {
			p.PosSinrLayer1 = pos
			i += 1
		} else if strings.ToLower(item) == "dtx" && p.PosDtx < 0 {
			p.PosDtx = pos
			i += 1
		} else if strings.ToLower(item) == "ulrank" && p.PosUlRank < 0 {
			p.PosUlRank = pos
			i += 1
		} else if strings.ToLower(item) == "ulpmirank1" && p.PosUlPmiRank1 < 0 {
			p.PosUlPmiRank1 = pos
			i += 1
		} else if strings.ToLower(item) == "ulpmirank1sinr" && p.PosUlPmiRank1Sinr < 0 {
			p.PosUlPmiRank1Sinr = pos
			i += 1
		} else if strings.ToLower(item) == "ulpmirank2" && p.PosUlPmiRank2 < 0 {
			p.PosUlPmiRank2 = pos
			i += 1
		} else if strings.ToLower(item) == "ulpmirank2sinr_[0]" && p.PosUlPmiRank2SinrLayer0 < 0 {
			p.PosUlPmiRank2SinrLayer0 = pos
			i += 1
		} else if strings.ToLower(item) == "ulpmirank2sinr_[1]" && p.PosUlPmiRank2SinrLayer1 < 0 {
			p.PosUlPmiRank2SinrLayer1 = pos
			i += 1
		} else if strings.ToLower(item) == "longtermrank" && p.PosLongTermRank < 0 {
			p.PosLongTermRank = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlPduDemuxData struct {
	TtiEventHeader
	HarqId        string
	IsUlCcchData  string
	IsTcpTraffic  string
	TempCrnti     string
	LcIdList      []string
	RcvdBytesList []string
}

type TtiUlPduDemuxDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosHarqId int
	PosIsUlCcchData int
	PosIsTcpTraffic int
	PosTempCrnti int
}

func FindTtiUlPduDemuxDataPos(tokens []string) TtiUlPduDemuxDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 4
	p := TtiUlPduDemuxDataPos{
		Ready:                       false,
		PosEventHeader:              FindTtiEventHeaderPos(tokens),
		PosHarqId:              -1,
		PosIsUlCcchData:     -1,
		PosIsTcpTraffic:     -1,
		PosTempCrnti:     -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "harqid" && p.PosHarqId< 0 {
			p.PosHarqId = pos
			i += 1
		} else if strings.ToLower(item) == "isulccchdata" && p.PosIsUlCcchData < 0 {
			p.PosIsUlCcchData = pos
			i += 1
		} else if strings.ToLower(item) == "istcptraffic" && p.PosIsTcpTraffic < 0 {
			p.PosIsTcpTraffic = pos
			i += 1
		} else if strings.ToLower(item) == "tempcrnti" && p.PosTempCrnti < 0 {
			p.PosTempCrnti = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlPreSchedData struct {
	TtiEventHeader
	CsListEvent string
	HighestClassPriority string
}

type TtiUlPreSchedDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosCsListEvent int
	PosHighestClassPriority int
}

func FindTtiUlPreSchedDataPos(tokens []string) TtiUlPreSchedDataPos {
	// n is the total number of interested fields, make sure to update n if any field is added or removed.
	n := 2
	p := TtiUlPreSchedDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosCsListEvent: -1,
		PosHighestClassPriority: -1,
	}

	i := 0
	for pos, item := range tokens {
		if strings.ToLower(item) == "cslistevent" && p.PosCsListEvent < 0 {
			p.PosCsListEvent = pos
			i += 1
		} else if strings.ToLower(item) == "highestclasspriority" && p.PosHighestClassPriority < 0 {
			p.PosHighestClassPriority = pos
			i += 1
		}

		if i >= n {
			p.Ready = true
			break
		}
	}

	return p
}

type TtiUlTdSchedSubcellData struct {
	TtiEventHeader
	SubcellId string
	Cs2List []string
}

type TtiUlTdSchedSubcellDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosSubcellId int
	PosRecordSequenceNumber []int
}

func FindTtiUlTdSchedSubcellDataPos(tokens []string) TtiUlTdSchedSubcellDataPos {
	p := TtiUlTdSchedSubcellDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosSubcellId: -1,
		PosRecordSequenceNumber: make([]int, 0),
	}

	for pos, item := range tokens {
		if strings.ToLower(item) == "subcellid" && p.PosSubcellId < 0 {
			p.PosSubcellId = pos
		} else if strings.ToLower(item) == "recordsequencenumber" {
			p.PosRecordSequenceNumber = append(p.PosRecordSequenceNumber, pos)
		}
	}

	p.Ready = true
	return p
}
