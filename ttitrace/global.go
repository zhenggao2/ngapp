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
	CellDbIndex string
	TxNumber string
	DlHarqProcessIndex string
	K1 string
	AllFields []string
}

type TtiDlFdSchedDataPos struct {
	Ready bool
	PosEventHeader TtiEventHeaderPos
	PosCellDbIndex int
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
	n := 9
	p := TtiDlFdSchedDataPos{
		Ready: false,
		PosEventHeader: FindTtiEventHeaderPos(tokens),
		PosCellDbIndex: -1,
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
	PosRrmInstCqi int
	PosRank int
	PosCellDbIndex int
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
