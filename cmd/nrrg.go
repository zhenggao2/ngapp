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
package cmd

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/zhenggao2/ngapp/nrgrid"
	"github.com/zhenggao2/ngapp/utils"
	"math"
	"strconv"
	"strings"
	"github.com/Knetic/govaluate"
)

var (
    flags NrrgFlags
    minChBw int
	boldRed = color.New(color.FgRed).Add(color.Bold).SprintFunc()
	regRed = color.New(color.FgRed)
	boldGreen = color.New(color.FgGreen).Add(color.Bold).SprintFunc()
	regGreen = color.New(color.FgGreen)
	boldBlue = color.New(color.FgBlue).Add(color.Bold).SprintFunc()
	regBlue = color.New(color.FgBlue)
	boldYellow = color.New(color.FgYellow).Add(color.Bold).SprintFunc()
	regYellow = color.New(color.FgYellow)
)


type NrrgFlags struct {
	gridsetting GridSettingFlags
	mib MibFlags
	commonSetting CommonSettingFlags
	css0 Css0Flags
	coreset1 Coreset1Flags
	uss UssFlags
	dci10 Dci10Flags
	dci11 Dci11Flags
	msg3 Msg3Flags
	dci01 Dci01Flags
	bwp BwpFlags
	rach RachFlags
	dmrsCommon DmrsCommonFlags
	dmrsPdsch DmrsPdschFlags
	ptrsPdsch PtrsPdschFlags
	dmrsPusch DmrsPuschFlags
	ptrsPusch PtrsPuschFlags
	pdsch PdschFlags
	pusch PuschFlags
	nzpCsiRs NzpCsiRsFlags
	trs TrsFlags
	csiIm CsiImFlags
	csiReport CsiReportFlags
	srs SrsFlags
	pucch PucchFlags
	advanced AdvancedFlags
}

// flags for grid settings
type GridSettingFlags struct {
	// 2023-2-19: For simplicity, assume the same SCS for SSB/RMSI/Carrier/BWP etc.
	scs string

	// freqBand settings
	band string
	_duplexMode    string
	_maxDlFreq    int
	_freqRange    string
	_unlicensed bool

	// ssbGrid settings
	_ssbScs      string
	gscn int
	_ssbPattern string
	_kSsbScs float64
	_kSsb        int
	_nCrbSsbScs float64
	_nCrbSsb    int
	ssbPeriod   string
	_maxLBar       int
	_maxL	int
	candSsbIndex []int

	// carrierGrid settings
	_carrierScs       string
	bw               string
	dlArfcn int
	_carrierNumRbs   int
	_offsetToCarrier int

	// MIB settings
	_mibCommonScs                string
	rmsiCoreset0             int
	rmsiCss0                 int
	_coreset0MultiplexingPat int
	_coreset0NumRbs          int
	_coreset0NumSymbs        int
	_coreset0OffsetList      []int
	_coreset0Offset          int
}

// mib
type MibFlags struct {
	sfn                      int
	hrf                      int
	dmrsTypeAPos             string
	//_mibCommonScs                string
	//rmsiCoreset0             int
	//rmsiCss0                 int
	//_coreset0MultiplexingPat int
	//_coreset0NumRbs          int
	//_coreset0NumSymbs        int
	//_coreset0OffsetList      []int
	//_coreset0Offset          int
}

// common setting
type CommonSettingFlags struct {
	pci     int
	numUeAp string
	// common setting - tdd ul/dl config common
	_refScs       string
	patPeriod     []string
	patNumDlSlots []int
	patNumDlSymbs []int
	patNumUlSymbs []int
	patNumUlSlots []int
}

// CSS0
type Css0Flags struct {
	css0AggLevel      int
	css0NumCandidates string
}

// CORESET1
type Coreset1Flags struct {
	coreset1FreqRes string
	// TODO: rename coreset1NumSymbs to coreset1Duration
	// coreset1NumSymbs        int
	coreset1Duration int
	coreset1CceRegMap       string
	coreset1RegBundleSize   string
	coreset1InterleaverSize string
	coreset1ShiftInd        int
	// coreset1PrecoderGranularity string
}

// USS
type UssFlags struct {
	ussPeriod        string
	ussOffset        int
	ussDuration      int
	ussFirstSymbs    string
	ussAggLevel      int
	ussNumCandidates string
}

// DCI 1_0 scheduling Sib1/Msg2/Msg4 with SI-RNTI/RA-RNTI/TC-RNTI
type Dci10Flags struct {
	_rnti                    []string
	_muPdcch                 []int
	_muPdsch                 []int
	dci10TdRa                []int
	_tdMappingType           []string
	_tdK0                    []int
	_tdSliv                  []int
	_tdStartSymb             []int
	_tdNumSymbs              []int
	_fdRaType                []string
	_fdBitwidthRaType1 []int
	_fdRa                    []string
	dci10FdStartRb           []int
	dci10FdNumRbs            []int
	dci10FdVrbPrbMappingType []string
	_fdBundleSize            []string
	dci10McsCw0              []int
	_tbs                     []int
	dci10Msg2TbScaling       int
	dci10Msg4DeltaPri        int
	dci10Msg4TdK1            int
}

// DCI 1_1 scheduling PDSCH with C-RNTI
type Dci11Flags struct {
	_rnti                    string
	_muPdcch                 int
	_muPdsch                 int
	_actBwp                  int
	_indicatedBwp            int
	dci11TdRa                int
	dci11TdMappingType       string
	dci11TdK0                int
	dci11TdSliv              int
	dci11TdStartSymb         int
	dci11TdNumSymbs          int
	dci11FdRaType            string
	_dci11FdBitwidthRaType0 int
	_dci11FdBitwidthRaType1 int
	dci11FdRa                string
	dci11FdStartRb           int
	dci11FdNumRbs            int
	dci11FdVrbPrbMappingType string
	dci11FdBundleSize        string
	dci11McsCw0              int
	dci11McsCw1              int
	_tbs                     []int
	dci11DeltaPri            int
	dci11TdK1                int
	dci11AntPorts            int
}

// Msg3 PUSCH scheduled by UL grant in RAR(Msg2)
type Msg3Flags struct {
	_muPusch            int
	msg3TdRa            int
	_tdMappingType      string
	_tdK2               int
	_tdDelta            int
	_tdSliv             int
	_tdStartSymb        int
	_tdNumSymbs         int
	_fdRaType           string
	msg3FdFreqHop       string
	msg3FdRa            string
	msg3FdStartRb       int
	msg3FdNumRbs        int
	_fdSecondHopFreqOff int
	msg3McsCw0          int
	_tbs                int
}


// DCI 0_1 scheduling PUSCH with C-RNTI
type Dci01Flags struct {
	_rnti                string
	_muPdcch             int
	_muPusch             int
	_actBwp              int
	_indicatedBwp        int
	dci01TdRa            int
	dci01TdMappingType   string
	dci01TdK2            int
	dci01TdSliv          int
	dci01TdStartSymb     int
	dci01TdNumSymbs      int
	dci01FdRaType        string
	dci01FdFreqHop       string
	dci01FdRa            string
	dci01FdStartRb       int
	dci01FdNumRbs        int
	dci01McsCw0          int
	_tbs                 int
	dci01CbTpmiNumLayers int
	dci01Sri             string
	dci01AntPorts        int
	dci01PtrsDmrsMap     int
}

// initial/dedicated UL/DL BWP
type BwpFlags struct {
	_bwpType    []string
	_bwpId      []int
	_bwpScs     []string
	_bwpCp      []string
	_bwpLocAndBw []int
	_bwpStartRb  []int
	_bwpNumRbs   []int
}

const (
	INI_DL_BWP int = 0
	DED_DL_BWP int = 1
	INI_UL_BWP int = 2
	DED_UL_BWP int = 3
	N_SC_RB int = 12
	DMRS_SIB1 int = 0
	DMRS_MSG2 int = 1
	DMRS_MSG4 int = 2
	DMRS_MSG3 int = 3
)

// random access
type RachFlags struct {
	prachConfId                        int
	_raFormat                          string
	_raX                               int
	_raY                               []int
	_raSubfNumFr1SlotNumFr2            []int
	_raStartingSymb                    int
	_raNumSlotsPerSubfFr1Per60KSlotFr2 int
	_raNumOccasionsPerSlot             int
	_raDuration                        int
	msg1Scs                            string
	msg1Fdm                            int
	msg1FreqStart                      int
	raRespWin                          string
	totNumPreambs                      int
	ssbPerRachOccasion                 string
	cbPreambsPerSsb                    int
	contResTimer                       string
	msg3Tp                             string
	_raLen                             int
	_raNumRbs                          int
	_raKBar                            int
}
	
// DMRS for SIB1/Msg2/Msg4/Msg3
type DmrsCommonFlags struct {
	_schInfo           []string
	_dmrsType          []string
	_dmrsAddPos        []string
	_maxLength         []string
	_dmrsPorts         []int
	_cdmGroupsWoData   []int
	_numFrontLoadSymbs []int
	_tdL               []string
	_fdK               []string
}

// DMRS for PDSCH
type DmrsPdschFlags struct {
	pdschDmrsType      string
	pdschDmrsAddPos    string
	pdschMaxLength     string
	_dmrsPorts         []int
	_cdmGroupsWoData   int
	_numFrontLoadSymbs int
	_tdL               string
	_fdK               string
}

// PTRS for PDSCH
type PtrsPdschFlags struct {
	pdschPtrsEnabled     bool
	pdschPtrsTimeDensity int
	pdschPtrsFreqDensity int
	pdschPtrsReOffset    string
	_dmrsPorts           int
}

// DMRS for PUSCH
type DmrsPuschFlags struct {
	puschDmrsType      string
	puschDmrsAddPos    string
	puschMaxLength     string
	_dmrsPorts         []int
	_cdmGroupsWoData   int
	_numFrontLoadSymbs int
	_tdL               string
	_fdK               string
}

// PTRS for PUSCH
type PtrsPuschFlags struct {
	puschPtrsEnabled       bool
	puschPtrsTimeDensity   int
	puschPtrsFreqDensity   int
	puschPtrsReOffset      string
	puschPtrsMaxNumPorts   string
	_dmrsPorts             []int
	puschPtrsTimeDensityTp int
	puschPtrsGrpPatternTp  string
	_numGrpsTp             int
	_samplesPerGrpTp       int
	_dmrsPortsTp           []int
}

// PDSCH-config and PDSCH-ServingCellConfig
type PdschFlags struct {
	pdschAggFactor string
	pdschRbgCfg    string
	_rbgSize       int
	pdschMcsTable  string
	pdschXOh       string
}

// PUSCH-config and PUSCH-ServingCellConfig
type PuschFlags struct {
	puschTxCfg                   string
	puschCbSubset                string
	puschCbMaxRankNonCbMaxLayers int
	puschFreqHopOffset           int
	puschTp                      string
	puschAggFactor               string
	puschRbgCfg                  string
	_rbgSize                     int
	puschMcsTable                string
	puschXOh                     string
}

// NZP-CSI-RS resource
type NzpCsiRsFlags struct {
	_resSetId             int
	_trsInfo              bool
	_resId                int
	nzpCsiRsFreqAllocRow  string
	nzpCsiRsFreqAllocBits string
	nzpCsiRsNumPorts      string
	nzpCsiRsCdmType       string
	nzpCsiRsDensity       string
	nzpCsiRsFirstSymb     int
	nzpCsiRsFirstSymb2    int
	nzpCsiRsStartRb       int
	nzpCsiRsNumRbs        int
	nzpCsiRsPeriod        string
	nzpCsiRsOffset        int
	_row                  int
	_kBarLBar             []string
	_ki                   []int
	_li                   []int
	_cdmGrpIndj           []int
	_kap                  []int
	_lap                  []int
}

// TRS resource
type TrsFlags struct {
	_resSetId        int
	_trsInfo         bool
	_firstResId      int
	_freqAllocRow    string
	trsFreqAllocBits string
	_numPorts        string
	_cdmType         string
	_density         string
	trsFirstSymbs    []int
	trsStartRb       int
	trsNumRbs        int
	trsPeriod        string
	// TRS occupies two NZP-CSI-RS resources in one slot or four NZP-CSI-RS resources in two consecutive slots
	trsOffset   []int
	_row        int
	_kBarLBar   []string
	_ki         []int
	_li         []int
	_cdmGrpIndj []int
	_kap        []int
	_lap        []int
}

// CSI-IM resource
type CsiImFlags struct {
	_resSetId      int
	_resId         int
	csiImRePattern string
	csiImScLoc     string
	csiImSymbLoc   int
	csiImStartRb   int
	csiImNumRbs    int
	csiImPeriod    string
	csiImOffset    int
}

// CSI-ResourceConfig and CSI-ReportConfig
type CsiReportFlags struct {
	_resCfgType        []string
	_resCfgId          []int
	_resSetId          []int
	_resBwpId          []int
	_resType           []string
	_repCfgId          int
	_resCfgIdChnMeas   int
	_resCfgIdCsiImIntf int
	_repCfgType        string
	csiRepPeriod       string
	csiRepOffset       int
	_ulBwpId           int
	csiRepPucchRes     int
	_quantity          string
}

// SRS resource
type SrsFlags struct {
	_resId           []int
	srsNumPorts      []string
	srsNonCbPtrsPort []string
	srsNumCombs      []string
	srsCombOff       []int
	srsCs            []int
	srsStartPos      []int
	srsNumSymbs      []string
	srsRepetition    []string
	srsFreqPos       []int
	srsFreqShift     []int
	srsCSrs          []int
	srsBSrs          []int
	srsBHop          []int
	_type            []string
	srsPeriod        []string
	srsOffset        []int
	_mSRSb           []string
	_Nb              []string
	// SRS resource set
	_resSetId          []int
	srsSetResIdList    []string
	_resType           []string
	_usage             []string
	_dci01NonCbSrsList []string
}

// PUCCH-FormatConfig, PUCCH resource and DSR resource
type PucchFlags struct {
	// PUCCH-FormatConfig
	pucchFmtCfgNumSlots string
	pucchFmtCfgInterSlotFreqHop string
	pucchFmtCfgAddDmrs bool
	pucchFmtCfgSimAckCsi bool

	// PUCCH resource
	_pucchResId []int
	_pucchFormat []string
	_pucchResSetId []int
	pucchStartRb []int
	pucchIntraSlotFreqHop []string
	pucchSecondHopPrb []int
	pucchNumRbs []int
	pucchStartSymb []int
	pucchNumSymbs []int

	// DSR resource
	_dsrResId []int
	_dsrPucchRes []int
	dsrPeriod []string
	dsrOffset []int
}

// Advanced settings
type AdvancedFlags struct {
	bestSsb int
	pdcchSlotSib1 int
	prachOccMsg1 int
	pdcchOccMsg2 int
	pdcchOccMsg4 int
	dsrRes int
}

// nrrgCmd represents the nrrg command
var nrrgCmd = &cobra.Command{
	Use:   "nrrg",
	Short: "NR resource grid tool",
	Long: `CMD "nrrg" generates NR resource grid according to configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		viper.WriteConfig()
	},
}

// confGridSettingCmd represents the "nrrg conf gridsetting" command
var confGridSettingCmd = &cobra.Command{
	Use:   "gridsetting",
	Short: "",
	Long: `CMD "nrrg gridsetting" can be used to get/set following grid settings.
-freqBand: frequency band related configurations
-ssbGrid: SSB grid related configurations
-carrierGrid: Carrier grid related configurations`,
	PreRun: func(cmd *cobra.Command, args []string) {
	    loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()

		// process gridsetting.band
		if cmd.Flags().Lookup("band").Changed {
			regGreen.Printf("[INFO]: Processing gridSetting.band...\n")
			band := flags.gridsetting.band
			p, exist := nrgrid.OpBands[band]
			if !exist {
				regRed.Printf("[ERR]: Invalid frequency band(FreqBandIndicatorNR): %v\n", band)
				return
			}

			v, _ := strconv.Atoi(band[1:])
			var fr string
			if v >= 1 && v <= 256 {
				fr = "FR1"
			} else if v >= 257 && v <= 262 {
				fr = "FR2-1" // FR2-1	24250 MHz – 52600 MHz
			} else {
				fr = "FR2-2" // FR2-2	52600 MHz – 71000 MHz
			}

			if p.DuplexMode == "TDD" {
				fmt.Printf("Frequency Band Info [%v]: UL/DL: %v, %v, %v\n", band, p.UlBand, p.DuplexMode, fr)
			} else if p.DuplexMode == "FDD" {
				fmt.Printf("Frequency Band Info [%v]: UL: %v, DL: %v, %v, %v\n", band, p.UlBand, p.DlBand, p.DuplexMode, fr)
			} else if p.DuplexMode == "SDL" {
				fmt.Printf("Frequency Band Info [%v]: DL: %v, %v, %v\n", band, p.DlBand, p.DuplexMode, fr)
			} else {
				fmt.Printf("Frequency Band Info [%v]: UL: %v, %v, %v\n", band, p.UlBand, p.DuplexMode, fr)
			}

			// update band info
			flags.gridsetting._duplexMode = p.DuplexMode
			flags.gridsetting._maxDlFreq = p.MaxDlFreq
			flags.gridsetting._freqRange = fr
			if v == 46 || v == 96 || v == 102 {
				flags.gridsetting._unlicensed = true
			} else {
				flags.gridsetting._unlicensed = false
			}

			// FR2-1 and FR2-2 are not supported!
			if v > 256 {
				regRed.Printf("[ERR]: FR2-1 and FR2-2 are not supported!\n")
				return
			}

			// SDL and SUL are not supported!
			if p.DuplexMode == "SDL" || p.DuplexMode == "SUL" {
				regRed.Printf("[ERR]: %v is not supported!\n", p.DuplexMode)
				return
			}

			// NR-U is not supported!
			if band == "n46" || band == "n96" || band == "n102" {
				regRed.Printf("[ERR]: NR-U [Band n46/n96/n102] is not supported!\n")
				return
			}

			// get available SSB scs
			var ssbScsSet []string
			for _, v := range nrgrid.SsbRasters[band] {
				ssbScsSet = append(ssbScsSet, v[0])
			}
			fmt.Printf("Available SSB SCS: %v\n", ssbScsSet)

			// get available RMSI scs and carrier scs
			var rmsiScsSet []string
			var carrierScsSet []string
			if flags.gridsetting._freqRange == "FR1" {
				rmsiScsSet = append(rmsiScsSet, []string{"15KHz", "30KHz"}...)

				scsFr1 := []int{15, 30, 60}
				for _, scs := range scsFr1 {
					key := fmt.Sprintf("%v_%v", band, scs)
					valid := false
					for _, i := range nrgrid.BandScs2BwFr1[key] {
					    if i > 0 {
					    	valid = true
					    	break
						}
					}
					if valid {
						carrierScsSet = append(carrierScsSet, fmt.Sprintf("%vKHz", scs))
					}
				}
			} else if flags.gridsetting._freqRange == "FR2-1" {
				rmsiScsSet = append(rmsiScsSet, []string{"60KHz", "120KHz"}...)
				carrierScsSet = append(carrierScsSet, []string{"60KHz", "120KHz"}...)
			} else {
				rmsiScsSet = ssbScsSet
				carrierScsSet = append(carrierScsSet, []string{"120KHz", "480KHz", "960KHz"}...)
			}
			fmt.Printf("Available RMSI SCS(subcarrierSpacingCommon of MIB): %v\n", rmsiScsSet)
			fmt.Printf("Available Carrier SCS(subcarrierSpacing of SCS-SpecificCarrier): %v\n", carrierScsSet)

			// update rach info
			err := updateRach()
			if err != nil {
				regRed.Printf("[ERR]: %v\n", err.Error())
				return
			}
		}

		// process gridsetting.scs
		if cmd.Flags().Lookup("scs").Changed {
			regGreen.Printf("[INFO]: Processing gridSetting.scs...\n")

			// set SCS for SSB/RMSI/Carrier
			flags.gridsetting._ssbScs = flags.gridsetting.scs
			flags.gridsetting._carrierScs = flags.gridsetting.scs
			flags.gridsetting._mibCommonScs = flags.gridsetting.scs

			// update SSB pattern
			band := flags.gridsetting.band
			scs := flags.gridsetting._ssbScs
			for _, v := range nrgrid.SsbRasters[band] {
				if v[0] == scs {
					fmt.Printf("SSB Raster Info: %v\n", v)
					flags.gridsetting._ssbPattern = v[1]
				}
			}

			// update SSB burst (refer to 3GPP TS 38.213 vh40: 4.1	Cell search)
			pat := flags.gridsetting._ssbPattern
			dm := flags.gridsetting._duplexMode
			freq := flags.gridsetting._maxDlFreq
			nru := flags.gridsetting._unlicensed
			if (pat == "Case A" && !nru) || pat == "Case B" || (pat == "Case C" && !nru && dm == "FDD") {
				if freq <= 3000 {
					flags.gridsetting._maxLBar = 4
					flags.gridsetting._maxL = 4
				} else {
					flags.gridsetting._maxLBar = 8
					flags.gridsetting._maxL = 8
				}
			} else if pat == "Case C" && !nru && dm == "TDD" {
				if freq < 1880 {
					flags.gridsetting._maxLBar = 4
					flags.gridsetting._maxL = 4
				} else {
					flags.gridsetting._maxLBar = 8
					flags.gridsetting._maxL = 8
				}
			} else if pat == "Case A" && nru {
				flags.gridsetting._maxLBar = 10
				flags.gridsetting._maxL = 8
			} else if pat == "Case C" && nru {
				flags.gridsetting._maxLBar = 20
				flags.gridsetting._maxL = 8
			} else {
				// pat == "Case D/E or Case F/G(for FR2-2)
				flags.gridsetting._maxLBar = 64
				flags.gridsetting._maxL = 64
			}

			// update refScs of TDD-UL-DL-Config
			flags.commonSetting._refScs = scs
			// update u_PDCCH/u_PDSCH/u_PUSCH in DCI 0_1/1_1 and Msg3 PUSCH
			u := nrgrid.Scs2Mu[flags.gridsetting._carrierScs]
			flags.dci01._muPdcch = u
			flags.dci01._muPusch = u
			flags.dci11._muPdcch = u
			flags.dci11._muPdsch = u
			flags.msg3._muPusch = u
			flags.msg3._tdDelta = nrgrid.PuschTimeAllocMsg3K2Delta[flags.gridsetting._carrierScs]
			// update SCS of initial UL BWP and dedicated UL/DL BWP
			flags.bwp._bwpScs[DED_DL_BWP] = flags.gridsetting._carrierScs
			flags.bwp._bwpScs[INI_UL_BWP] = flags.gridsetting._carrierScs
			flags.bwp._bwpScs[DED_UL_BWP] = flags.gridsetting._carrierScs
			// get SR periodicity and offset(38.331 vh30 periodicityAndOffset and periodicityAndOffset-r17 of SchedulingRequestResourceConfig)
			fmt.Printf("Available SR periodicity: %v\n", nrgrid.SrPeriodSet[flags.gridsetting._carrierScs])
			// TODO: validate USS first symbols
			// update TRS periodicity (2023/2/20: For simplicity, TRS is not supported!)
			fmt.Printf("Available TRS periodicity: %v\n", []string{"slots10", "slots20", "slots40", "slots80", "slots160", "slots320", "slots640"}[u:u+4])

			// update u_PDCCH/u_PDSCH for SIB1/Msg2/Msg4
			u = nrgrid.Scs2Mu[flags.gridsetting._mibCommonScs]
			flags.dci10._muPdcch = []int{u, u, u}
			flags.dci10._muPdsch = []int{u, u, u}
			// update SCS for initial DL BWP
			// refer to 3GPP TS 38.331 vh30: subcarrierSpacing of BWP
			// For the initial DL BWP and operation in licensed spectrum this field has the same value as the field subCarrierSpacingCommon in MIB of the same serving cell.
			flags.bwp._bwpScs[INI_DL_BWP] = flags.gridsetting._mibCommonScs
			// update ra-ResponseWindow
			// refer to 3GPP TS 38.331 vh30: ra-ResponseWindow of RACH-ConfigGeneric
			// The network configures a value lower than or equal to 10 ms when Msg2 is transmitted in licensed spectrum and a value lower than or equal to 40 ms when Msg2 is transmitted with shared spectrum channel access (see TS 38.321 [3], clause 5.1.4).
			var rarWinSet []string
			switch flags.gridsetting._mibCommonScs {
			case "15KHz":
				rarWinSet = append(rarWinSet, []string{"sl1", "sl2", "sl4", "sl8", "sl10"}...)
			case "30KHz":
				rarWinSet = append(rarWinSet, []string{"sl1", "sl2", "sl4", "sl8", "sl10", "sl20"}...)
			case "60KHz":
				rarWinSet = append(rarWinSet, []string{"sl1", "sl2", "sl4", "sl8", "sl10", "sl20", "sl40"}...)
			case "120KHz":
				rarWinSet = append(rarWinSet, []string{"sl1", "sl2", "sl4", "sl8", "sl10", "sl20", "sl40", "sl80"}...)
			}
			fmt.Printf("Available ra-ResponseWindow: %v\n", rarWinSet)
		}

		// process gridsetting.bw
		if cmd.Flags().Lookup("bw").Changed {
			regGreen.Printf("[INFO]: Processing gridSetting.bw...\n")

			// update N_RB of carrier and initial DL BWP
			fr := flags.gridsetting._freqRange
			bw := flags.gridsetting.bw
			carrierScsVal, _ := strconv.Atoi(flags.gridsetting._carrierScs[:len(flags.gridsetting._carrierScs)-3])
			var idx int
			if fr == "FR1" {
				idx = utils.IndexStr(nrgrid.BwSetFr1, bw)
			} else if fr == "FR2-1" {
				idx = utils.IndexStr(nrgrid.BwSetFr21, bw)
			} else {
				idx = utils.IndexStr(nrgrid.BwSetFr22, bw)
			}
			if idx < 0 {
				regRed.Printf("Invalid carrier bandwidth for %v: carrierBw=%v\n", fr, bw)
				return
			}
			flags.gridsetting._carrierNumRbs = nrgrid.NrbFr1[carrierScsVal][idx]

			// update RB_Start and L_RB for initial UL BWP and dedicated UL/DL BWP
			flags.bwp._bwpStartRb[DED_DL_BWP] = 0
			flags.bwp._bwpNumRbs[DED_DL_BWP] = flags.gridsetting._carrierNumRbs
			flags.bwp._bwpLocAndBw[DED_DL_BWP] = makeRiv(flags.gridsetting._carrierNumRbs, 0, 275)
			flags.bwp._bwpStartRb[INI_UL_BWP] = 0
			flags.bwp._bwpNumRbs[INI_UL_BWP] = flags.gridsetting._carrierNumRbs
			flags.bwp._bwpLocAndBw[INI_UL_BWP] = makeRiv(flags.gridsetting._carrierNumRbs, 0, 275)
			flags.bwp._bwpStartRb[DED_UL_BWP] = 0
			flags.bwp._bwpNumRbs[DED_UL_BWP] = flags.gridsetting._carrierNumRbs
			flags.bwp._bwpLocAndBw[DED_UL_BWP] = makeRiv(flags.gridsetting._carrierNumRbs, 0, 275)
		}

		regGreen.Printf("[INFO]: Post-processing...\n")
		// update n_CRB_SSB/k_SSB
		updateKSsbAndNCrbSsb()

		// validate CORESET0
		err := validateCoreset0()
		if err != nil {
			regRed.Printf("[ERR]: %s\n", err.Error())
			return
		}

		// validate CSS0
		err = validateCss0()
		if err != nil {
			regRed.Printf("[ERR]: %s\n", err.Error())
			return
		}

		// determine SC#0RB#0 of CORESET0 based on coreset0Offset
		rmsiScsVal, _ := strconv.Atoi(flags.gridsetting._mibCommonScs[:len(flags.gridsetting._mibCommonScs)-3])
		iscSsbSc0Rb0 := flags.gridsetting._offsetToCarrier * 12 + (flags.gridsetting._nCrbSsb * 12 * int(flags.gridsetting._nCrbSsbScs) + flags.gridsetting._kSsb * int(flags.gridsetting._kSsbScs)) / rmsiScsVal
		nscCoreset0Offset := flags.gridsetting._coreset0Offset * 12
		iscCoreset0Sc0Rb0 := iscSsbSc0Rb0 - nscCoreset0Offset
		fmt.Printf("offsetToCarrier=%v, nCrbSsb=%v(SCS=%.0fKHz), kSsb=%v(SCS=%.0fKHz) -> iscSsbSc0Rb0=%v\n", flags.gridsetting._offsetToCarrier, flags.gridsetting._nCrbSsb, flags.gridsetting._nCrbSsbScs, flags.gridsetting._kSsb, flags.gridsetting._kSsbScs, iscSsbSc0Rb0)
		fmt.Printf("coreset0Offset=%v -> nscCoreset0Offset=%v\n", flags.gridsetting._coreset0Offset, nscCoreset0Offset)
		fmt.Printf("iscCoreset0Sc0Rb0=%v\n", iscCoreset0Sc0Rb0)

	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

func updateRach() error {
	regYellow.Printf("calling updateRach\n")

    var p *nrgrid.RachInfo
    var exist bool
	if flags.gridsetting._freqRange == "FR1"{
		if flags.gridsetting._duplexMode == "FDD" {
			p, exist = nrgrid.RaCfgFr1FddSUl[flags.rach.prachConfId]
		} else {
			p, exist = nrgrid.RaCfgFr1Tdd[flags.rach.prachConfId]
		}
	} else {
		p, exist = nrgrid.RaCfgFr2Tdd[flags.rach.prachConfId]
	}

	if !exist {
		return  errors.New(fmt.Sprintf("Invalid configurations for PRACH: %v,%v with prach-ConfigurationIndex=%v\n",
			flags.gridsetting._freqRange, flags.gridsetting._duplexMode, flags.rach.prachConfId))
	}

	fmt.Printf("RACH Info: %v\n", *p)

	flags.rach._raFormat = p.Format
	flags.rach._raX = p.X
	flags.rach._raY = p.Y
	flags.rach._raSubfNumFr1SlotNumFr2 = p.SubfNumFr1SlotNumFr2
	flags.rach._raNumSlotsPerSubfFr1Per60KSlotFr2 = p.NumSlotsPerSubfFr1Per60KSlotFr2
	flags.rach._raNumOccasionsPerSlot = p.NumOccasionsPerSlot
	flags.rach._raDuration = p.Duration

	var raScsSet []string
	if flags.rach._raFormat == "0" || flags.rach._raFormat == "1" || flags.rach._raFormat == "2" || flags.rach._raFormat == "3" {
		raScsSet = append(raScsSet, nrgrid.ScsRaLongPrach["839_"+flags.rach._raFormat])
	} else {
		if flags.gridsetting._freqRange == "FR1" {
			raScsSet = append(raScsSet, []string{"15KHz", "30KHz"}...)
		} else if flags.gridsetting._freqRange == "FR2-1" {
			raScsSet = append(raScsSet, []string{"60KHz", "120KHz"}...)
		} else {
			raScsSet = append(raScsSet, []string{"120KHz", "480KHz", "960KHz"}...)
		}
	}
	fmt.Printf("Available short PRACH SCS(msg1-SubcarrierSpacing of RACH-ConfigCommon): %v\n", raScsSet)

	return nil
}

// convert ARFCN to F_REF(MHz) (refer to 38.104 vh80)
//  Table 5.4.2.1-1: NR-ARFCN parameters for the global frequency raster
func arfcn2Fref(arfcn int, maxFreq int) float64 {
	if maxFreq < 3000 {
		return float64(arfcn) * 0.005
	} else if maxFreq < 24250 {
		return 3000 + 0.015 * (float64(arfcn) - 600000)
	} else {
		return 24250.08 + 0.06 * (float64(arfcn) - 2016667)
	}
}

// convert GSCN to SS_REF(MHz) (refer to 38.104 vh80)
//  Table 5.4.3.1-1: GSCN parameters for the global frequency raster
func gscn2Ssref(gscn int, maxFreq int) float64 {
	if maxFreq < 3000 {
		N := math.Floor((float64(gscn) + 1.5) / 3)
		M := math.Mod(float64(gscn) + 1.5, 3) * 2
		ssRef := 1.2 * N + 0.05 * M

		fmt.Printf("GSCN=%v, N=%v, M=%v, SS_REF=%vMHz\n", gscn, N, M, ssRef)
		return ssRef
	} else if maxFreq < 24250 {
		N := gscn - 7499
		return 3000 + 1.44 * float64(N)
	} else {
		N := gscn - 22256
		return 24250.08 + 17.28 * float64(N)
	}
}

// calculate N_CRB_SSB and k_SSB given GSCN and DL ARFCN
func updateKSsbAndNCrbSsb() error {
	regYellow.Printf("calling updateKSsbAndNCrbSsb\n")

	ssbScs, _ := strconv.ParseFloat(flags.gridsetting._ssbScs[:len(flags.gridsetting._ssbScs)-3], 64)
	carrierScs, _ := strconv.ParseFloat(flags.gridsetting._carrierScs[:len(flags.gridsetting._carrierScs)-3], 64)
	rmsiScs, _ := strconv.ParseFloat(flags.gridsetting._mibCommonScs[:len(flags.gridsetting._mibCommonScs)-3], 64)

	if flags.gridsetting._freqRange == "FR1" {
		flags.gridsetting._kSsbScs = 15
		flags.gridsetting._nCrbSsbScs = 15
	} else if flags.gridsetting._freqRange == "FR2-1" {
		flags.gridsetting._kSsbScs = rmsiScs
		flags.gridsetting._nCrbSsbScs = 60
	} else {
		flags.gridsetting._kSsbScs = ssbScs
		flags.gridsetting._nCrbSsbScs = 60
	}

	ssFreq := gscn2Ssref(flags.gridsetting.gscn, flags.gridsetting._maxDlFreq)
	ssFreqSc0Rb0 := ssFreq - 120 * ssbScs / 1000

	dlFreq := arfcn2Fref(flags.gridsetting.dlArfcn, flags.gridsetting._maxDlFreq)
	var dlFreqPointA float64
	if flags.gridsetting._carrierNumRbs % 2 == 0 {
		dlFreqPointA = dlFreq - 12 * float64(flags.gridsetting._carrierNumRbs / 2) * carrierScs / 1000
	} else {
		dlFreqPointA = dlFreq - 12 * (math.Floor(float64(flags.gridsetting._carrierNumRbs) / 2) + 6) * carrierScs / 1000
	}

	nCrbSsb := math.Floor((ssFreqSc0Rb0 - dlFreqPointA) / (12 * flags.gridsetting._nCrbSsbScs / 1000))
	kSsb := (ssFreqSc0Rb0 - dlFreqPointA - 12 * flags.gridsetting._nCrbSsbScs / 1000 * nCrbSsb) / (flags.gridsetting._kSsbScs / 1000)

	fmt.Printf("%v: nCrbSsb SCS=%.0fKHz, kSsb SCS=%.0fKHz\n", flags.gridsetting._freqRange, flags.gridsetting._nCrbSsbScs, flags.gridsetting._kSsbScs)
	fmt.Printf("ssFreq=%vMHz, ssFreqSc0Rb0=%vMHz, dlFreq=%vMHz, dlFreqPointA=%vMHz, nCrbSsb=%v, kSsb=%v\n",
		ssFreq, ssFreqSc0Rb0, dlFreq, dlFreqPointA, nCrbSsb, kSsb)

	flags.gridsetting._nCrbSsb = int(nCrbSsb)
	flags.gridsetting._kSsb = int(math.Ceil(kSsb))

	return nil
}

func validateCoreset0() error {
	regYellow.Printf("calling validateCoreset0\n")

	band := flags.gridsetting.band
	fr := flags.gridsetting._freqRange
	ssbScs := flags.gridsetting._ssbScs
	rmsiScs := flags.gridsetting._mibCommonScs

	var ssbScsSet []string
	var rmsiScsSet []string
	for _, v := range nrgrid.SsbRasters[band] {
		ssbScsSet = append(ssbScsSet, v[0])
	}
	if fr == "FR1" {
		rmsiScsSet = append(rmsiScsSet, []string{"15KHz", "30KHz"}...)
	} else if fr == "FR2-1" {
		rmsiScsSet = append(rmsiScsSet, []string{"60KHz", "120KHz"}...)
	} else {
		rmsiScsSet = ssbScsSet
	}

	if !(utils.ContainsStr(ssbScsSet, ssbScs) && utils.ContainsStr(rmsiScsSet, rmsiScs)) {
		return errors.New(fmt.Sprintf("Invalid SSB SCS and/or RMSI SCS settings!\nSSB SCS range: %v and ssbScs=%v\nRMSI SCS range: %v and rmsiScs=%v\n", ssbScsSet, ssbScs, rmsiScsSet, rmsiScs))
	}

	// calculate minChBw
	key := fmt.Sprintf("%v_%v", band, rmsiScs[:len(rmsiScs)-3])
	var bwSubset []string
	if fr == "FR1" {
		for i, v := range nrgrid.BandScs2BwFr1[key] {
			if v == 1 {
				bwSubset = append(bwSubset, nrgrid.BwSetFr1[i])
			}
		}
	} else if fr == "FR2-1" {
		for i, v := range nrgrid.BandScs2BwFr21[key] {
			if v == 1 {
				bwSubset = append(bwSubset, nrgrid.BwSetFr21[i])
			}
		}
	} else {
		for i, v := range nrgrid.BandScs2BwFr22[key] {
			if v == 1 {
				bwSubset = append(bwSubset, nrgrid.BwSetFr22[i])
			}
		}
	}

	if len(bwSubset) > 0 {
		minChBw, _ = strconv.Atoi(bwSubset[0][:len(bwSubset[0])-3])
		fmt.Printf("Available transmission bandwidth: %v\n", bwSubset)
		fmt.Printf("Minimum transmission bandwidth is %v\n", bwSubset[0])
	} else {
		minChBw = -1
		return errors.New(fmt.Sprintf("Invalid configurations for minChBw calculation: band=%v, freqRange=%v, rmsiScs=%v\n", band, fr, rmsiScs))
	}

	// validate coresetZero
	key = fmt.Sprintf("%v_%v_%v", ssbScs[:len(ssbScs)-3], rmsiScs[:len(rmsiScs)-3], flags.gridsetting.rmsiCoreset0)
	var p *nrgrid.Coreset0Info
	var exist bool
	if (band == "n79" || band == "n104") && ssbScs[:len(ssbScs)-3] == "30" && (rmsiScs[:len(rmsiScs)-3] == "15" || rmsiScs[:len(rmsiScs)-3] == "30") {
		// 38.101-1 vh80 Table 5.2-1: NR operating bands in FR1
		// NOTE 17: For this band, CORESET#0 values from Table 13-5 or Table 13-6 in [8, TS 38.213] are applied regardless of the minimum channel bandwidth.
		p, exist = nrgrid.Coreset0Fr1MinChBw40m[key]
	} else if fr == "FR1" && utils.ContainsInt([]int{5, 10}, minChBw) {
		p, exist = nrgrid.Coreset0Fr1MinChBw5m10m[key]
	} else if fr == "FR1" && minChBw == 40 {
		p, exist = nrgrid.Coreset0Fr1MinChBw40m[key]
	} else {
		// FR2-1
		p, exist = nrgrid.Coreset0Fr21[key]
	}
	if !exist || p == nil {
		return errors.New(fmt.Sprintf("Invalid configurations for CORESET0: fr=%v, ssbScs=%v, rmsiScs=%v, minChBw=%vMHz, coresetZero=%v", fr, ssbScs, rmsiScs, minChBw, flags.gridsetting.rmsiCoreset0))
	}
	fmt.Printf("CORESET0 Info: %v\n", *p)
	flags.gridsetting._coreset0MultiplexingPat = p.MultiplexingPat
	flags.gridsetting._coreset0NumRbs = p.NumRbs
	flags.gridsetting._coreset0NumSymbs = p.NumSymbs
	flags.gridsetting._coreset0OffsetList = p.OffsetLst

	// validate CORESET0 bw against carrier bw
	carrierBw := flags.gridsetting.bw
	rmsiScsVal, _ := strconv.Atoi(rmsiScs[:len(rmsiScs)-3])
	var numRbsRmsiScs int
	var idx int
	if fr == "FR1" {
		idx = utils.IndexStr(nrgrid.BwSetFr1, carrierBw)
	} else if fr == "FR2-1" {
		idx = utils.IndexStr(nrgrid.BwSetFr21, carrierBw)
	} else {
		idx = utils.IndexStr(nrgrid.BwSetFr22, carrierBw)
	}
	if idx < 0 {
		return errors.New(fmt.Sprintf("Invalid carrier bandwidth for %v: carrierBw=%v\n", fr, carrierBw))
	}
	numRbsRmsiScs = nrgrid.NrbFr1[rmsiScsVal][idx]

	if numRbsRmsiScs < flags.gridsetting._coreset0NumRbs {
		return errors.New(fmt.Sprintf("Invalid configurations for CORESET0: numRbsRmsiScs=%v, coreset0NumRbs=%v\n", numRbsRmsiScs, flags.gridsetting._coreset0NumRbs))
	}

	// update coreset0Offset w.r.t k_SSB
	kssb := flags.gridsetting._kSsb
	if len(flags.gridsetting._coreset0OffsetList) == 2 {
		if kssb == 0 {
			flags.gridsetting._coreset0Offset = flags.gridsetting._coreset0OffsetList[0]
		} else {
			flags.gridsetting._coreset0Offset = flags.gridsetting._coreset0OffsetList[1]
		}
	} else {
		flags.gridsetting._coreset0Offset = flags.gridsetting._coreset0OffsetList[0]
	}

	fmt.Printf("CORESET0: multiplexingPattern=%v, numRbs=%v, numSymbs=%v, offset=%v\n", flags.gridsetting._coreset0MultiplexingPat, flags.gridsetting._coreset0NumRbs, flags.gridsetting._coreset0NumSymbs, flags.gridsetting._coreset0Offset)

	// Basic assumptions: If offset >= 0, then 1st RB of CORESET0 aligns with the carrier edge; if offset < 0, then 1st RB of SSB aligns with the carrier edge.
	// if offset >= 0, min bw = max(coreset0NumRbs, offset + 20 * scsSsb / scsRmsi), and n_CRB_SSB needs update w.r.t to offset
	// if offset < 0, min bw = coreset0NumRbs - offset, and don't have to update n_CRB_SSB
	ssbScsVal, _ := strconv.Atoi(ssbScs[:len(ssbScs)-3])
	var minBw int
	if flags.gridsetting._coreset0Offset >= 0 {
		minBw = utils.MaxInt([]int{flags.gridsetting._coreset0NumRbs, flags.gridsetting._coreset0Offset + 20 * ssbScsVal / rmsiScsVal})
	} else {
		minBw = flags.gridsetting._coreset0NumRbs - flags.gridsetting._coreset0Offset
	}
	if numRbsRmsiScs < minBw {
		return errors.New(fmt.Sprintf("Invalid configurations for CORESET0: numRbsRmsiScs=%v, minBw=%v(coreset0NumRbs=%v,offset=%v)\n", numRbsRmsiScs, minBw, flags.gridsetting._coreset0NumRbs, flags.gridsetting._coreset0Offset))
	}

	// validate coreste0NumSymbs against dmrs-pointA-Position
	// refer to 3GPP TS 38.211 vf80: 7.3.2.2	Control-resource set (CORESET)
	// N_CORESET_symb = 3 is supported only if the higher-layer parameter dmrs-TypeA-Position equals 3;
	if flags.gridsetting._coreset0NumSymbs == 3 && flags.mib.dmrsTypeAPos != "pos3" {
		return errors.New(fmt.Sprintf("coreset0NumSymb can be 3 only if dmrs-TypeA-Position is pos3! (corest0NumSymbs=%v,dmrsTypeAPos=%v)\n", flags.gridsetting._coreset0NumSymbs, flags.mib.dmrsTypeAPos))
	}

	// update info of initial dl bwp
	if flags.gridsetting._coreset0Offset >= 0 {
		upper := utils.MinInt([]int{numRbsRmsiScs - flags.gridsetting._coreset0NumRbs, numRbsRmsiScs - (flags.gridsetting._coreset0NumRbs + 20 * ssbScsVal / rmsiScsVal)})
		fmt.Printf("Available RB_Start for Initial DL BWP: [%v..%v]\n", 0, upper)
	} else {
		upper := utils.MinInt([]int{numRbsRmsiScs - flags.gridsetting._coreset0NumRbs, numRbsRmsiScs - (flags.gridsetting._coreset0NumRbs + 20 * ssbScsVal / rmsiScsVal)})
		fmt.Printf("Available RB_Start for Initial DL BWP: [%v..%v]\n", -flags.gridsetting._coreset0Offset, upper)
	}
	fmt.Printf("Available L_RBs for Initial DL BWP: [%v]\n", flags.gridsetting._coreset0NumRbs)

	// update info of 'frquency domain assignment' bitwidth of SIB1/Msg2/Msg4
	nrb := float64(flags.gridsetting._coreset0NumRbs)
	bitwidth := utils.CeilInt(math.Log2(nrb * (nrb + 1) / 2))
	flags.dci10._fdBitwidthRaType1 = []int{bitwidth, bitwidth, bitwidth}
	fmt.Printf("Bitwidth of the 'frequency domain assignment' field of DCI 1_0 scheduling SIB1/Msg2/Msg4: %v bits\n", bitwidth)

	return nil
}

func validateCss0() error {
	regYellow.Printf("calling validateCss0\n")

    fr := flags.gridsetting._freqRange
    pat := flags.gridsetting._coreset0MultiplexingPat
    css0 := flags.gridsetting.rmsiCss0

    switch pat {
	case 1:
		if fr == "FR1" || ((fr == "FR2-1" || fr == "FR2-2") && css0 >= 0 && css0 <= 13) {
			return nil
		} else {
			return errors.New(fmt.Sprintf("Invalid CSS0 setting!\nsearchSpaceZero range is [0, 13] for SSB/CORESET0 multiplexing pattern 1 and FR2\n"))
		}
	case 2, 3:
		if css0 != 0 {
			return errors.New(fmt.Sprintf("Invalid CSS0 setting!\nsearchSpaceZero range is [0] for SSB/CORESET0 multiplexing pattern %v\n", css0))
		}
	}

	return nil
}

// calculate RIV (refer to 38.214 vh40)
//  5.1.2.2.2	Downlink resource allocation type 1
func makeRiv(L_RBs, RB_start, N_BWP_size int) int {
	if L_RBs < 1 || L_RBs > (N_BWP_size - RB_start) {
		regRed.Printf("[ERR]: Invalid combination of L_RBs=%d, RB_start=%d, N_BWP_size=%d", L_RBs, RB_start, N_BWP_size)
		return -1
	}

	var riv int
	if (L_RBs - 1) <= int(math.Floor(float64(N_BWP_size) / 2)) {
		riv = N_BWP_size * (L_RBs - 1) + RB_start
	} else {
		riv = N_BWP_size * (N_BWP_size - L_RBs + 1) + (N_BWP_size - 1 - RB_start)
	}

	return riv
}

// confMibCmd represents the nrrg mib command
var confMibCmd = &cobra.Command{
	Use:   "mib",
	Short: "",
	Long: `CMD "nrrg mib" can be used to get/set MIB related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()

		if cmd.Flags().Lookup("dmrsTypeAPos").Changed {
		    dmrsTypeAPos := flags.mib.dmrsTypeAPos

			// validate CORESET duration
			// refer to 3GPP TS 38.211 vf80: 7.3.2.2	Control-resource set (CORESET)
			// N_CORESET_symb = 3 is supported only if the higher-layer parameter dmrs-TypeA-Position equals 3;
			if flags.gridsetting._coreset0NumSymbs == 3 && dmrsTypeAPos != "pos3" {
				fmt.Printf("coreset0NumSymb can be 3 only if dmrs-TypeA-Position is pos3! (corest0NumSymbs=%v,dmrsTypeAPos=%v)\n", flags.gridsetting._coreset0NumSymbs, flags.mib.dmrsTypeAPos)
				return
			}
			if flags.coreset1.coreset1Duration == 3 && dmrsTypeAPos != "pos3" {
				fmt.Printf("coreset1Duration can be 3 only if dmrs-TypeA-Position is pos3! (coreset1Duration=%v,dmrsTypeAPos=%v)\n", flags.coreset1.coreset1Duration, flags.mib.dmrsTypeAPos)
				return
			}

			// validate 'Time domain resource assignment" field of DCI 1_0
			err := validateDci10TdRa()
			if err != nil {
				fmt.Print(err.Error())
				return
			}

			// validate 'Time domain resource assignment" field of DCI 1_1
			err2 := validateDci11TdRa()
			if err2 != nil {
				fmt.Print(err2.Error())
				return
			}

			// validate DCI 0_1 scheduled PUSCH
			err3 := validatePuschAntPorts()
			if err3 != nil {
				fmt.Print(err3.Error())
				return
			}

			// update TBS of Msg3 PUSCH scheduled by RAR Msg2
			err4 := updateMsg3PuschTbs()
			if err4 != nil {
				fmt.Print(err4.Error())
				return
			}
		}

	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

/*
validateDci10TdRa validates the "Time domain resource assignment" field of DCI 1_0 scheduling SIB1/Msg2/Msg4, updates associated DMRS, and calculate TBS.
 */
func validateDci10TdRa() error {
	fmt.Printf("\n-->%s\n", "calling validateDci10TdRa")

	// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1-1: Valid S and L combinations
	// Note 1:	S = 3 is applicable only if dmrs-TypeA-Position = 3
	// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1.1-1: Applicable PDSCH time domain resource allocation
	// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1.1-2: Default PDSCH time domain resource allocation A for normal CP
	// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1.1-3: Default PDSCH time domain resource allocation A for extended CP
	// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1.1-4: Default PDSCH time domain resource allocation B
	// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1.1-5: Default PDSCH time domain resource allocation C
	dmrsTypeAPos := flags.mib.dmrsTypeAPos
	for i, rnti := range flags.dci10._rnti {
		row := flags.dci10.dci10TdRa[i] + 1
		key := fmt.Sprintf("%v_%v", row, dmrsTypeAPos[3:])
		var p *nrgrid.TimeAllocInfo
		var exist bool

		switch rnti {
		case "SI-RNTI":
			switch flags.gridsetting._coreset0MultiplexingPat {
			case 1:
				p, exist = nrgrid.PdschTimeAllocDefANormCp[key]
			case 2:
				// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1.1-4: Default PDSCH time domain resource allocation B
				// Note 1: If the PDSCH was scheduled with SI-RNTI in PDCCH Type0 common search space, the UE may assume that this PDSCH resource allocation is not applied.
				if utils.ContainsInt(nrgrid.PdschTimeAllocDefBNote1Set, flags.dci10.dci10TdRa[i] + 1) {
					return errors.New(fmt.Sprintf("Row %v is invalid for SIB1 (refer to 'Note 1' of Table 5.1.2.1.1-4 of TS 38.214).", flags.dci10.dci10TdRa[i] + 1))
				}
				p, exist = nrgrid.PdschTimeAllocDefB[key]
			case 3:
				// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1.1-5: Default PDSCH time domain resource allocation C
				// Note 1: The UE may assume that this PDSCH resource allocation is not used, if the PDSCH was scheduled with SI-RNTI in PDCCH Type0 common search space.
				if utils.ContainsInt(nrgrid.PdschTimeAllocDefCNote1Set, flags.dci10.dci10TdRa[i] + 1) {
					return errors.New(fmt.Sprintf("Row %v is invalid for SIB1 (refer to 'Note 1' of Table 5.1.2.1.1-5 of TS 38.214).", flags.dci10.dci10TdRa[i] + 1))
				}
				p, exist = nrgrid.PdschTimeAllocDefC[key]
			}

		case "RA-RNTI", "TC-RNTI":
			if flags.bwp._bwpCp[INI_DL_BWP] == "normal" {
				p, exist = nrgrid.PdschTimeAllocDefANormCp[key]
			} else {
				p, exist = nrgrid.PdschTimeAllocDefAExtCp[key]
			}
		}

		if !exist || p == nil {
			return errors.New(fmt.Sprintf("Invalid PDSCH time domain allocation: dci10TdRa=%v, dmrsTypeAPos=%v\n", flags.dci10.dci10TdRa[i], flags.mib.dmrsTypeAPos))
		} else {
			// update dci10 info
			fmt.Printf("nrgrid.TimeAllocInfo(DCI 1_0, rnti=%v, coreset0MultiplexingPat=%v): %v\n", rnti, flags.gridsetting._coreset0MultiplexingPat, *p)
			flags.dci10._tdMappingType[i] = p.MappingType
			flags.dci10._tdK0[i] = p.K0
			flags.dci10._tdStartSymb[i] = p.S
			flags.dci10._tdNumSymbs[i] = p.L
			sliv, _ := nrgrid.ToSliv(p.S, p.L, "PDSCH", p.MappingType, "normal")
			flags.dci10._tdSliv[i] = sliv

			// update dmrs info
			// refer to 3GPP TS 38.214 vfa0: 5.1.6.2	DM-RS reception procedure
			// When receiving PDSCH scheduled by DCI format 1_0, the UE shall assume the number of DM-RS CDM groups without data is 1 which corresponds to CDM group 0 for the case of PDSCH with allocation duration of 2 symbols, and the UE shall assume that the number of DM-RS CDM groups without data is 2 which corresponds to CDM group {0,1} for all other cases.
			if p.L == 2 {
				flags.dmrsCommon._cdmGroupsWoData[i] = 1
			} else {
				flags.dmrsCommon._cdmGroupsWoData[i] = 2
			}

			// -For PDSCH with mapping type A, the UE shall assume dmrs-AdditionalPosition='pos2' and up to two additional single-symbol DM-RS present in a slot according to the PDSCH duration indicated in the DCI as defined in Clause 7.4.1.1 of [4, TS 38.211], and
			// -For PDSCH with allocation duration of 7 symbols for normal CP or 6 symbols for extended CP with mapping type B, the UE shall assume one additional single-symbol DM-RS present in the 5th or 6th symbol when the front-loaded DM-RS symbol is in the 1st or 2nd symbol respectively of the PDSCH allocation duration, otherwise the UE shall assume that the additional DM-RS symbol is not present, and
			// -For PDSCH with allocation duration of 4 symbols with mapping type B, the UE shall assume that no additional DM-RS are present, and
			// -For PDSCH with allocation duration of 2 symbols with mapping type B, the UE shall assume that no additional DM-RS are present, and the UE shall assume that the PDSCH is present in the symbol carrying DM-RS.
			if p.MappingType == "typeA" {
				flags.dmrsCommon._dmrsAddPos[i] = "pos2"
			} else {
				// Initial DL bwp always use normal CP.
				if flags.bwp._bwpCp[INI_DL_BWP] == "normal" && p.L == 7 {
					flags.dmrsCommon._dmrsAddPos[i] = "pos1"
				} else {
					flags.dmrsCommon._dmrsAddPos[i] = "pos0"
				}
			}

			// update TBS info
			err := updateDci10Tbs(i)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

/*
updateDci10Tbs updates the TBS field of DCI 1_0 scheduling Sib1/Msg2/Msg4.
i: index of the flags.dci10 slices
 */
func updateDci10Tbs(i int) error {
	fmt.Printf("\n-->%s\n", "calling updateDci10Tbs")

	td := flags.dci10._tdNumSymbs[i]
	ld := 0
	fd := flags.dci10.dci10FdNumRbs[i]
	mcs := flags.dci10.dci10McsCw0[i]

	// refer to 3GPP TS 38.211 vf80: 7.4.1.1.2	Mapping to physical resources (DMRS for PDSCH)
	// -for PDSCH mapping type A, ld is the duration between the first OFDM symbol of the slot and the last OFDM symbol of the scheduled PDSCH resources in the slot
	// -for PDSCH mapping type B, ld is the duration of the scheduled PDSCH resources
	if flags.dci10._tdMappingType[i] == "typeA" {
		ld = flags.dci10._tdStartSymb[i] + flags.dci10._tdNumSymbs[i]
	} else {
		ld = td
	}

	key2 := fmt.Sprintf("%v_%v_%v", ld, flags.dci10._tdMappingType[i], flags.dmrsCommon._dmrsAddPos[i])
	// refer to 3GPP TS 38.214 vfa0:
	// When receiving PDSCH scheduled by DCI format 1_0 or ..., and a single symbol front-loaded DM-RS of configuration type 1 on DM-RS port 1000 is transmitted, and ...
	dmrs, exist2 := nrgrid.DmrsPdschPosOneSymb[key2]
	if !exist2 || dmrs == nil {
		return errors.New(fmt.Sprintf("Invalid DMRS for PDSCH settings: rnti=%v, numFrontLoadSymbs=%v, key=%v\n", flags.dci10._rnti[i], 1, key2))
	}

	// refer to 3GPP TS 38.211 vf80: 7.4.1.1.2	Mapping to physical resources (DMRS for PDSCH)
	// The case dmrs-AdditionalPosition equals to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2'.
	// For PDSCH mapping type A, l_d = 3 and l_d = 4 symbols in Tables 7.4.1.1.2-3 and 7.4.1.1.2-4 respectively is only applicable when dmrs-TypeA-Position is equal to 'pos2'.
	// TODO: For PDSCH mapping type A single-symbol DMRS, l1 = 11 except if ...
	// For PDSCH mapping type B
	//  -if the PDSCH duration ld is 2 or 4 OFDM symbols, only single-symbol DM-RS is supported.
	// TODO: For PDSCH mapping type B, when PDSCH allocation collides with CORESET ...
	dmrsTypeAPos := flags.mib.dmrsTypeAPos
	if flags.dci10._tdMappingType[i] == "typeA" {
		if (ld == 3 || ld == 4) && dmrsTypeAPos != "pos2" {
			return errors.New(fmt.Sprintf("For PDSCH mapping type A, ld = 3 and ld = 4 symbols in Tables 7.4.1.1.2-3 and 7.4.1.1.2-4 respectively is only applicable when dmrs-TypeA-Position is equal to 'pos2'.\nld=%v, dmrsTypeAPos=%v\n", ld, dmrsTypeAPos))
		}
	}

	dmrsOh := (2 * flags.dmrsCommon._cdmGroupsWoData[i]) * len(dmrs)
	fmt.Printf("PDSCH(%v) DMRS overhead: cdmGroupsWoData=%v, key=%v, dmrs=%v\n", flags.dci10._rnti[i], flags.dmrsCommon._cdmGroupsWoData[i], key2, dmrs)

	tbs, err := getTbs("PDSCH", false, flags.dci10._rnti[i], "qam64", td, fd, mcs, 1, dmrsOh, 0, 1)
	if err != nil {
		return err
	} else {
		fmt.Printf("CW0 TBS=%v bits\n", tbs)
		flags.dci10._tbs[i] = tbs
	}

	return nil
}

/*
validateDci11TdRa validates the "Time domain resource assignment" field of DCI 1_1 scheduling PDSCH.
 */
func validateDci11TdRa() error {
	fmt.Printf("\n-->%s\n", "calling validateDci11TdRa")

	dmrsTypeAPos := flags.mib.dmrsTypeAPos

	if flags.dci11.dci11TdRa >= 0 && flags.dci11.dci11TdRa <= 15 {
		// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1.1-1: Applicable PDSCH time domain resource allocation
		// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1.1-2: Default PDSCH time domain resource allocation A for normal CP
		// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1.1-3: Default PDSCH time domain resource allocation A for extended CP
		row := flags.dci11.dci11TdRa + 1
		key := fmt.Sprintf("%v_%v", row, dmrsTypeAPos[3:])
		var p *nrgrid.TimeAllocInfo
		var exist bool

		if flags.bwp._bwpCp[DED_DL_BWP] == "normal" {
			p, exist = nrgrid.PdschTimeAllocDefANormCp[key]
		} else {
			p, exist = nrgrid.PdschTimeAllocDefAExtCp[key]
		}

		if !exist {
			return errors.New(fmt.Sprintf("Invalid PDSCH time domain allocation: dci11TdRa=%v, dmrsTypeAPos=%v\n", flags.dci11.dci11TdRa, flags.mib.dmrsTypeAPos))
		} else {
			// update dci11 info
			fmt.Printf("nrgrid.TimeAllocInfo(DCI 1_1, rnti=C-RNTI): %v\n", *p)
			flags.dci11.dci11TdMappingType = p.MappingType
			flags.dci11.dci11TdK0 = p.K0
			flags.dci11.dci11TdStartSymb = p.S
			flags.dci11.dci11TdNumSymbs = p.L
			sliv, _ := nrgrid.ToSliv(p.S, p.L, "PDSCH", p.MappingType, "normal")
			flags.dci11.dci11TdSliv = sliv
		}
	} else {
		// refer to 3GPP TS 38.214 vfa0: Table 5.1.2.1-1: Valid S and L combinations
		// Note 1:	S = 3 is applicable only if dmrs-TypeA-Position = 3
		if flags.dci11.dci11TdMappingType == "typeA" && flags.dci11.dci11TdStartSymb == 3 && dmrsTypeAPos != "pos3" {
			return errors.New(fmt.Sprintf("S = 3 is applicable only if dmrs-TypeA-Position = 3 when PDSCH mapping type is typeA.\ndci11TdStartSymb=%v,dmrsTypeAPos=%v\n", flags.dci11.dci11TdStartSymb, dmrsTypeAPos))
		}
	}

	// validate 'Antenna port(s)' field of DCI 1_1 scheduling PDSCH
	err := validatePdschAntPorts()
	if err != nil {
		return err
	}

	return nil
}

/*
validatePdschAntPorts validates PDSCH configurations, updates DMRS/PTRS for PDSCH and updates PDSCH TBS.
 */
func validatePdschAntPorts() error {
	fmt.Printf("\n-->%s\n", "calling validatePdschAntPorts")

	dmrsType := flags.dmrsPdsch.pdschDmrsType
	maxLength := flags.dmrsPdsch.pdschMaxLength
	var mcsSet []int
	if flags.dci11.dci11McsCw0 >= 0 {
		mcsSet = append(mcsSet, flags.dci11.dci11McsCw0)
	}
	if flags.dci11.dci11McsCw1 >= 0 {
		mcsSet = append(mcsSet, flags.dci11.dci11McsCw1)
	}

	ap := flags.dci11.dci11AntPorts
	// update DMRS for PDSCH
	var tokens []string
	var p *nrgrid.AntPortsInfo
	var exist bool
	if dmrsType == "type1" && maxLength == "len1" && len(mcsSet) == 1 {
		tokens = strings.Split(nrgrid.Dci11AntPortsDmrsType1MaxLen1OneCwValid, "-")
	    p, exist = nrgrid.Dci11AntPortsDmrsType1MaxLen1OneCw[ap]
	} else if dmrsType == "type1" && maxLength == "len2" && len(mcsSet) == 1 {
		tokens = strings.Split(nrgrid.Dci11AntPortsDmrsType1MaxLen2OneCwValid, "-")
		p, exist = nrgrid.Dci11AntPortsDmrsType1MaxLen2OneCw[ap]
	} else if dmrsType == "type1" && maxLength == "len2" && len(mcsSet) == 2 {
		tokens = strings.Split(nrgrid.Dci11AntPortsDmrsType1MaxLen2TwoCwsValid, "-")
		p, exist = nrgrid.Dci11AntPortsDmrsType1MaxLen2TwoCws[ap]
	} else if dmrsType == "type2" && maxLength == "len1" && len(mcsSet) == 1 {
		tokens = strings.Split(nrgrid.Dci11AntPortsDmrsType2MaxLen1OneCwValid, "-")
		p, exist = nrgrid.Dci11AntPortsDmrsType2MaxLen1OneCw[ap]
	} else if dmrsType == "type2" && maxLength == "len1" && len(mcsSet) == 2 {
		tokens = strings.Split(nrgrid.Dci11AntPortsDmrsType2MaxLen1TwoCwsValid, "-")
		p, exist = nrgrid.Dci11AntPortsDmrsType2MaxLen1TwoCws[ap]
	} else if dmrsType == "type2" && maxLength == "len2" && len(mcsSet) == 1 {
		tokens = strings.Split(nrgrid.Dci11AntPortsDmrsType2MaxLen2OneCwValid, "-")
		p, exist = nrgrid.Dci11AntPortsDmrsType2MaxLen2OneCw[ap]
	} else if dmrsType == "type2" && maxLength == "len2" && len(mcsSet) == 2 {
		tokens = strings.Split(nrgrid.Dci11AntPortsDmrsType2MaxLen2TwoCwsValid, "-")
		p, exist = nrgrid.Dci11AntPortsDmrsType2MaxLen2TwoCws[ap]
	} else {
		return errors.New(fmt.Sprintf("Invalid settings for DCI 1_1 'Antenna port(s)'.\ndmrsType=%v, maxLength=%v, len(mcsSet)=%v\n", dmrsType, maxLength, len(mcsSet)))
	}

	minVal, _ := strconv.Atoi(tokens[0])
	maxVal, _ := strconv.Atoi(tokens[1])
	fmt.Printf("'Antenna port(s)' field of DCI 1_1 range: [%v, %v]\n", minVal, maxVal)

	if !exist || p == nil {
		return errors.New(fmt.Sprintf("Invalid settings for DCI 1_1 'Antenna port(s)'.\ndmrsType=%v, maxLength=%v, len(mcsSet)=%v, dci11AntPorts=%v\n", dmrsType, maxLength, len(mcsSet), ap))
	}


	fmt.Printf("nrgrid.AntPortsInfo(PDSCH): %v\n", *p)
	for i := range p.DmrsPorts {
		p.DmrsPorts[i] += 1000
	}

	flags.dmrsPdsch._cdmGroupsWoData = p.CdmGroups
	flags.dmrsPdsch._dmrsPorts = p.DmrsPorts
	flags.dmrsPdsch._numFrontLoadSymbs = p.NumDmrsSymbs

	// update PTRS for PDSCH
	maxDmrsPorts := utils.MaxInt(flags.dmrsPdsch._dmrsPorts)
	noPtrs := false
	// refer to 3GPP TS 38.214 vfa0: 5.1.6.2	DM-RS reception procedure
	// If a UE receiving PDSCH is configured with the higher layer parameter phaseTrackingRS in DMRS-DownlinkConfig, the UE may assume that the following configurations are not occurring simultaneously for the received PDSCH:
	//-	any DM-RS ports among 1004-1007 or 1006-1011 for DM-RS configurations type 1 and type 2, respectively are scheduled for the UE and the other UE(s) sharing the DM-RS REs on the same CDM group(s), and
	//-	PT-RS is transmitted to the UE.
	if (dmrsType == "type1" && maxDmrsPorts >= 1004) || (dmrsType == "type2" && maxDmrsPorts >= 1006) {
		noPtrs = true
	}

	fmt.Printf("PDSCH noPtrs=%v\n", noPtrs)

	if noPtrs {
		flags.ptrsPdsch.pdschPtrsEnabled = false
	} else {
	    // refer to 3GPP TS 38.214 vfa0: 5.1.6.3	PT-RS reception procedure
		// If a UE is scheduled with one codeword, the PT-RS antenna port is associated with the lowest indexed DM-RS antenna port among the DM-RS antenna ports assigned for the PDSCH.
		// If a UE is scheduled with two codewords, the PT-RS antenna port is associated with the lowest indexed DM-RS antenna port among the DM-RS antenna ports assigned for the codeword with the higher MCS. If the MCS indices of the two codewords are the same, the PT-RS antenna port is associated with the lowest indexed DM-RS antenna port assigned for codeword 0.
	    if len(mcsSet) == 1 {
	    	flags.ptrsPdsch._dmrsPorts = flags.dmrsPdsch._dmrsPorts[0]
		} else {
		    // refer to 3GPP TS 38.211 vf80: Table 7.3.1.3-1: Codeword-to-layer mapping for spatial multiplexing.
			// refer to 3GPP TS 38.211 vf80: 7.3.1.4	Antenna port mapping
		    numLayersCw0 := utils.FloorInt(float64(len(flags.dmrsPdsch._dmrsPorts)) / 2)
		    if mcsSet[0] >= mcsSet[1] {
		    	flags.ptrsPdsch._dmrsPorts = flags.dmrsPdsch._dmrsPorts[0]
			} else {
				flags.ptrsPdsch._dmrsPorts = flags.dmrsPdsch._dmrsPorts[numLayersCw0]
			}
		}
	}

	// update PDSCH TBS
	fdRaType := flags.dci11.dci11FdRaType
	fdRa := flags.dci11.dci11FdRa
	if (fdRaType == "raType0" && len(fdRa) != flags.dci11._dci11FdBitwidthRaType0) || (fdRaType == "raType1" && len(fdRa) != flags.dci11._dci11FdBitwidthRaType1) {
		return errors.New(fmt.Sprintf("Invalid 'Frequency domain resource assignment' field of DCI 1_1: fdRaType=%v, fdRa=%v, len(fdRa)=%v, bitwidthRaType0=%v, bitwidthRaType1=%v\n", fdRaType, fdRa, len(fdRa), flags.dci11._dci11FdBitwidthRaType0, flags.dci11._dci11FdBitwidthRaType1))
	}

	fd := 0
	if fdRaType == "raType0" {
		rbgs := getRaType0Rbgs(flags.bwp._bwpStartRb[DED_DL_BWP], flags.bwp._bwpNumRbs[DED_DL_BWP], flags.pdsch._rbgSize)
		for i, c := range fdRa {
			if c == '1' {
				fd += rbgs[i]
			}
		}
	} else {
		fd = flags.dci11.dci11FdNumRbs
	}

	// calculate DMRS overhead
	td := flags.dci11.dci11TdNumSymbs
	ld := 0
	tdMappingType := flags.dci11.dci11TdMappingType
	dmrsAddPos := flags.dmrsPdsch.pdschDmrsAddPos

	// refer to 3GPP TS 38.211 vf80: 7.4.1.1.2	Mapping to physical resources (DMRS for PDSCH)
	// -for PDSCH mapping type A, ld is the duration between the first OFDM symbol of the slot and the last OFDM symbol of the scheduled PDSCH resources in the slot
	// -for PDSCH mapping type B, ld is the duration of the scheduled PDSCH resources
	if tdMappingType == "typeA" {
		ld = flags.dci11.dci11TdStartSymb + flags.dci11.dci11TdNumSymbs
	} else {
		ld = td
	}

	key := fmt.Sprintf("%v_%v_%v", ld, tdMappingType, dmrsAddPos)
	var dmrs []int
	// var exist bool
	if flags.dmrsPdsch._numFrontLoadSymbs == 1 {
		dmrs, exist = nrgrid.DmrsPdschPosOneSymb[key]
	} else {
		dmrs, exist = nrgrid.DmrsPdschPosTwoSymbs[key]
	}

	if !exist || dmrs == nil {
	    return errors.New(fmt.Sprintf("Invalid DMRS for PDSCH settings: rnti=%v, numFrontLoadSymbs=%v, key=%v\n", flags.dci11._rnti, flags.dmrsPdsch._numFrontLoadSymbs, key))
	}

	// refer to 3GPP TS 38.211 vf80: 7.4.1.1.2	Mapping to physical resources (DMRS for PDSCH)
	// The case dmrs-AdditionalPosition equals to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2'.
	// For PDSCH mapping type A, l_d = 3 and l_d = 4 symbols in Tables 7.4.1.1.2-3 and 7.4.1.1.2-4 respectively is only applicable when dmrs-TypeA-Position is equal to 'pos2'.
	// TODO: For PDSCH mapping type A single-symbol DMRS, l1 = 11 except if ...
	// For PDSCH mapping type B
	//  -if the PDSCH duration ld is 2 or 4 OFDM symbols, only single-symbol DM-RS is supported.
	// TODO: For PDSCH mapping type B, when PDSCH allocation collides with CORESET ...
	dmrsTypeAPos := flags.mib.dmrsTypeAPos
	if dmrsAddPos == "pos3" && dmrsTypeAPos != "pos2" {
		return errors.New(fmt.Sprintf("The case dmrs-AdditionalPosition equals to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2'.\npdschDmrsAddPos=%v,dmrsTypeAPos=%v\n", flags.dmrsPdsch.pdschDmrsAddPos, dmrsTypeAPos))
	}
	if tdMappingType == "typeA" {
		if (ld == 3 || ld == 4) && dmrsTypeAPos != "pos2" {
			return errors.New(fmt.Sprintf("For PDSCH mapping type A, ld = 3 and ld = 4 symbols in Tables 7.4.1.1.2-3 and 7.4.1.1.2-4 respectively is only applicable when dmrs-TypeA-Position is equal to 'pos2'.\nld=%v, dmrsTypeAPos=%v\n", ld, dmrsTypeAPos))
		}
	}
	if tdMappingType == "typeB" && (ld == 2 || ld == 4) && flags.dmrsPdsch._numFrontLoadSymbs != 1 {
	    return errors.New(fmt.Sprintf("For PDSCH mapping type B, if the PDSCH duration ld is 2 or 4 OFDM symbols, only single-symbol DM-RS is supported.\n tdMappingType=%v, ld=%v, numFrontLoadSymbs=%v\n", tdMappingType, ld, flags.dmrsPdsch._numFrontLoadSymbs))
	}


	dmrsOh := (2 * flags.dmrsPdsch._cdmGroupsWoData) * len(dmrs)
	fmt.Printf("PDSCH(C-RNTI) DMRS overhead: cdmGroupsWoData=%v, key=%v, dmrs=%v\n", flags.dmrsPdsch._cdmGroupsWoData, key, dmrs)

	xoh, _ := strconv.Atoi(flags.pdsch.pdschXOh[3:])
	if flags.dci11.dci11McsCw0 >= 0 {
		tbs, err := getTbs("PDSCH", false, "C-RNTI", flags.pdsch.pdschMcsTable, td, fd, flags.dci11.dci11McsCw0, len(flags.dmrsPdsch._dmrsPorts), dmrsOh, xoh, 1)
		if err != nil {
			return err
		} else {
			fmt.Printf("CW0 TBS=%v bits\n", tbs)
			flags.dci11._tbs[0] = tbs
		}
	}

	if flags.dci11.dci11McsCw1 >= 0 {
		tbs, err := getTbs("PDSCH", false, "C-RNTI", flags.pdsch.pdschMcsTable, td, fd, flags.dci11.dci11McsCw1, len(flags.dmrsPdsch._dmrsPorts), dmrsOh, xoh, 1)
		if err != nil {
			return err
		} else {
			fmt.Printf("CW1 TBS=%v bits\n", tbs)
			flags.dci11._tbs[1] = tbs
		}
	}

	return nil
}

/*
updateMsg3PuschTbs updates the TBS field of Msg3 PUSCH scheduled by RAR Msg2.
 */
func updateMsg3PuschTbs() error {
	fmt.Printf("\n-->%s\n", "calling updateMsg3PuschTbs")

	td := flags.msg3._tdNumSymbs
	fd := flags.msg3.msg3FdNumRbs

	// update 'cdm groups without data' of DMRS for Msg3 PUSCH
	if td <= 2 && flags.rach.msg3Tp == "disabled" {
		flags.dmrsCommon._cdmGroupsWoData[DMRS_MSG3] = 1
	} else {
		flags.dmrsCommon._cdmGroupsWoData[DMRS_MSG3] = 2
	}

	// calculate DMRS overhead
	mappingType := flags.msg3._tdMappingType
	freqHop := flags.msg3.msg3FdFreqHop
	var v1, v2, v []int
	var e1, e2, e bool
	var key1, key2, key string
	// refer to 3GPP TS 38.214 vfb0: 6.2.2	UE DM-RS transmission procedure
	// If frequency hopping is disabled:
	// -	The UE shall assume dmrs-AdditionalPosition equals to 'pos2' and up to two additional DM-RS can be transmitted according to PUSCH duration, or
	// If frequency hopping is enabled:
	// -	The UE shall assume dmrs-AdditionalPosition equals to 'pos1' and up to one additional DM-RS can be transmitted according to PUSCH duration.
	if freqHop == "enabled" {
		// refer to 3GPP TS 38.211 vf80: 6.4.1.1.3	Precoding and mapping to physical resources
		// -	ld is the duration per hop according to Table 6.4.1.1.3-6 if intra-slot frequency hopping is used.
		// -	if the higher-layer parameter dmrs-AdditionalPosition is not set to 'pos0' and intra-slot frequency hopping is enabled according to clause 7.3.1.1.2 in [4, TS 38.212] and by higher layer, Tables 6.4.1.1.3-6 shall be used assuming dmrs-AdditionalPosition is equal to 'pos1' for each hop.

		// refer to 3GPP TS 38.214 vfb0: 6.3	UE PUSCH frequency hopping procedure
		// In case of intra-slot frequency hopping,
		// -	The number of symbols in the first hop is given by floor(N_PUSCH,s_symb/2) , the number of symbols in the second hop is given by N_PUSCH,s_symb - floor(N_PUSCH,s_symb/2) , where N_PUSCH,s_symb is the length of the PUSCH transmission in OFDM symbols in one slot.
		ld1 := utils.FloorInt(float64(td) / 2)
		ld2 := td - ld1
		var strDmrsTypeAPos string
		var strDmrsMsg3AddPos string
		if mappingType == "typeA" {
			strDmrsTypeAPos = flags.mib.dmrsTypeAPos[3:]
		} else {
			strDmrsTypeAPos = "0"
		}
		if flags.dmrsCommon._dmrsAddPos[DMRS_MSG3] != "pos0" {
			strDmrsMsg3AddPos = "pos1"
		} else {
			strDmrsMsg3AddPos = "pos0"
		}

		key1 = fmt.Sprintf("%v_%v_%v_%v_1st", ld1, mappingType, strDmrsTypeAPos, strDmrsMsg3AddPos)
		key2 = fmt.Sprintf("%v_%v_%v_%v_2nd", ld2, mappingType, strDmrsTypeAPos, strDmrsMsg3AddPos)
		v1, e1 = nrgrid.DmrsPuschPosOneSymbWithIntraSlotFh[key1]
		v2, e2 = nrgrid.DmrsPuschPosOneSymbWithIntraSlotFh[key2]
		if !e1 || v1 == nil || !e2 || v2 == nil {
			return errors.New(fmt.Sprintf("Invalid key(key_1stHop=\"%v\", key_2ndHop=\"%v\") when referring DmrsPuschPosOneSymbWithIntraSlotFh!", key1, key2))
		}
	} else {
		// refer to 3GPP TS 38.211 vf80: 6.4.1.1.3	Precoding and mapping to physical resources
		// -	ld is the duration between the first OFDM symbol of the slot and the last OFDM symbol of the scheduled PUSCH resources in the slot for PUSCH mapping type A according to Tables 6.4.1.1.3-3 and 6.4.1.1.3-4 if intra-slot frequency hopping is not used, or
		// -	ld is the duration of scheduled PUSCH resources for PUSCH mapping type B according to Tables 6.4.1.1.3-3 and 6.4.1.1.3-4 if intra-slot frequency hopping is not used, or
		var ld int
		if mappingType == "typeA" {
			ld = flags.msg3._tdStartSymb + td
		} else {
			ld = td
		}
		key = fmt.Sprintf("%v_%v_%v", ld, mappingType, flags.dmrsCommon._dmrsAddPos[DMRS_MSG3])
		v, e = nrgrid.DmrsPuschPosOneSymbWoIntraSlotFh[key]
		if !e || v == nil {
			return errors.New(fmt.Sprintf("Invalid key(=\"%v\") when referring DmrsPuschPosOneSymbWoIntraSlotFh!", key))
		}
	}

	// refer to 3GPP TS 38.214 vfb0: 6.2.2	UE DM-RS transmission procedure
	// When transmitted PUSCH is neither scheduled by DCI format 0_1 with CRC scrambled by C-RNTI, CS-RNTI, SP-CSI-RNTI or MCS-C-RNTI, nor corresponding to a configured grant, the UE shall use single symbol front-loaded DM-RS of configuration type 1 on DM-RS port 0 and...
	// refer to 3GPP TS 38.211 vf80: 6.4.1.1.3	Precoding and mapping to physical resources
	// For PUSCH mapping type A, the case dmrs-AdditionalPosition equal to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2'.
	// For PUSCH mapping type A, ld = 4 symbols in Table 6.4.1.1.3-4 is only applicable when dmrs-TypeA-Position is equal to 'pos2'.
	//
	// Rationale: dmrsAddPos of Msg3 is either 'pos1' or 'pos2', so restriction #1 is not relevant. Table 6.4.1.1.3-4 is for double symbols front-loaded DMRS, so restriction #2 is not relevant.

	var dmrsOh int
	if freqHop == "enabled" {
		dmrsOh = (2 * flags.dmrsCommon._cdmGroupsWoData[DMRS_MSG3]) * (len(v1) + len(v2))
		fmt.Printf("Msg3 DMRS overhead: cdmGroupsWoData=%v, key1=\"%v\", val1=%v, key2=\"%v\", val2=%v\n", flags.dmrsCommon._cdmGroupsWoData[DMRS_MSG3], key1, v1, key2, v2)
	} else {
		dmrsOh = (2 * flags.dmrsCommon._cdmGroupsWoData[DMRS_MSG3]) * len(v)
		fmt.Printf("Msg3 DMRS overhead: cdmGroupsWoData=%v, key=\"%v\", val=%v\n", flags.dmrsCommon._cdmGroupsWoData[DMRS_MSG3], key, v)
	}

	var tp bool
	if flags.rach.msg3Tp == "enabled" {
		tp = true
	} else {
		tp = false
	}

	tbs, err := getTbs("PUSCH", tp, "MSG3", "qam64", td, fd, flags.msg3.msg3McsCw0, 1, dmrsOh, 0, 1)
	if err != nil {
		return err
	} else {
		fmt.Printf("CW0 TBS=%v bits\n", tbs)
		flags.msg3._tbs = tbs
	}

	return nil
}

/*
validatePuschAntPorts validates PUSCH configurations, updates DMRS/PTRS for PUSCH and updates PUSCH TBS.
 */
func validatePuschAntPorts() error {
	fmt.Printf("\n-->%s\n", "calling validatePuschAntPorts")
	//TODO

	return nil
}


/*
getRaType0Rbgs return RBGs for PDSCH/PUSCH resource allocation Type 0.
 */
func getRaType0Rbgs(bwpStart, bwpSize, P int) []int {
	fmt.Printf("\n-->%s\n", "calling getRaType0Rbgs")

	bitwidth := utils.CeilInt((float64(bwpSize) + float64(bwpStart % P)) / float64(P))
	rbgs := make([]int, bitwidth)
	for i := 0; i < bitwidth; i++ {
		if i == 0 {
		    rbgs[i] = P - bwpStart % P
		} else if i == bitwidth - 1 {
			if (bwpStart + bwpSize) % P > 0 {
				rbgs[i] = (bwpStart + bwpSize) % P
			} else {
				rbgs[i] = P
			}
		} else {
			rbgs[i] = P
		}
	}

	return rbgs
}

func validateDci01TdRa() error {
	fmt.Printf("\n-->%s\n", "calling validateDci01TdRa")

	dmrsTypeAPos := flags.mib.dmrsTypeAPos

	// refer to 3GPP TS 38.211 vf80: 6.4.1.1.3	Precoding and mapping to physical resources (DMRS for PUSCH)
	// For PUSCH mapping type A, the case dmrs-AdditionalPosition equal to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2'.
	// For PUSCH mapping type A, l_d = 4 symbols in Table 6.4.1.1.3-4 is only applicable when dmrs-TypeA-Position is equal to 'pos2'.
	//
	//- ld is the duration between the first OFDM symbol of the slot and the last OFDM symbol of the scheduled PUSCH resources in the slot for PUSCH mapping type A according to Tables 6.4.1.1.3-3 and 6.4.1.1.3-4 if intra-slot frequency hopping is not used, or
	//-	ld is the duration of scheduled PUSCH resources for PUSCH mapping type B according to Tables 6.4.1.1.3-3 and 6.4.1.1.3-4 if intra-slot frequency hopping is not used, or
	//-	ld is the duration per hop according to Table 6.4.1.1.3-6 if intra-slot frequency hopping is used.
	if flags.dci01.dci01TdMappingType == "typeA" && flags.dmrsPusch.puschDmrsAddPos == "pos3" && dmrsTypeAPos != "pos2" {
		return errors.New(fmt.Sprintf("For PUSCH mapping type A, the case dmrs-AdditionalPosition equal to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2'.\ndci01TdMappingType=%v, puschDmrsAddPos=%v, dmrsTypeAPos=%v\n", flags.dci01.dci01TdMappingType, flags.dmrsPusch.puschDmrsAddPos, dmrsTypeAPos))
	}

	// update PUSCH TBS
	// TODO

	return nil
}

func validateMsg3TdRa() error {
	fmt.Printf("\n-->%s\n", "calling validateMsg3TdRa")

	// dmrsTypeAPos := flags.mib.dmrsTypeAPos


	// update Msg3 TBS
	// TODO

	return nil
}

/*
getTbs calculates TBS for PUSCH/PDSCH.
	sch: PUSCH or PDSCH
	tp: PUSCH transform percoding flag
	rnti: C-RNTI, SI-RNTI, RA-RNTI, TC-RNTI
	mcsTab: qam64, qam64LowSE or qam256
	td: number of symbols
	fd: number of PRBs
	mcs: MCS
	layer: number of spatial multiplexing layers
	dmrs: overhead of DMRS
	xoh: the xOverhead
	scale: TB scaling for Msg2
 */
func getTbs(sch string, tp bool, rnti string, mcsTab string, td int, fd int, mcs int, layer int, dmrs int, xoh int, scale float64) (int, error) {
	fmt.Printf("\n-->%s\n", "calling getTbs")

	rntiSet := []string{"C-RNTI", "SI-RNTI", "RA-RNTI", "TC-RNTI", "MSG3"}
	mcsTabSet := []string{"qam256", "qam64", "qam64LowSE"}

	if !utils.ContainsStr(rntiSet, rnti) || !utils.ContainsStr(mcsTabSet, mcsTab) {
		return 0, errors.New(fmt.Sprintf("Invalid RNTI or MCS table!\n"))
	}

	// refer to 3GPP TS 38.214 vfa0
	// 5.1.3	Modulation order, target code rate, redundancy version and transport block size determination
	// 6.1.4	Modulation order, redundancy version and transport block size determination
	// 1st step: get Qm and R(x1024)
	var p *nrgrid.McsInfo
	if sch == "PDSCH" || (sch == "PUSCH" && !tp) {
		if rnti == "C-RNTI" && mcsTab == "qam256" {
			p = nrgrid.PdschMcsTabQam256[mcs]
		} else if rnti == "C-RNTI" &&  mcsTab == "qam64LowSE" {
			p = nrgrid.PdschMcsTabQam64LowSE[mcs]
		} else {
			p = nrgrid.PdschMcsTabQam64[mcs]
		}
	} else if sch == "PUSCH" && tp {
		if rnti == "C-RNTI" && mcsTab == "qam256" {
			p = nrgrid.PdschMcsTabQam256[mcs]
		} else if rnti == "C-RNTI" &&  mcsTab == "qam64LowSE" {
			p = nrgrid.PuschTpMcsTabQam64LowSE[mcs]
		} else {
			p = nrgrid.PuschTpMcsTabQam64[mcs]
		}
	}

	if p == nil {
		return 0, errors.New(fmt.Sprintf("Invalid MCS: sch=%v, tp=%v, rnti=%v, mcsTab=%v, mcs=%v", sch, tp, rnti, mcsTab, mcs))
	}
	Qm, R := p.ModOrder, p.CodeRate

	// The UE is not expected to decode a PDSCH scheduled with P-RNTI, RA-RNTI, SI-RNTI and Qm > 2.
	// FIXME: assume PDSCH scheduled with TC-RNTI has the same restraint.
	if (rnti == "RA-RNTI" || rnti == "SI-RNTI" || rnti == "TC-RNTI") && Qm > 2 {
		return 0, errors.New(fmt.Sprintf("The UE is not expected to decode a PDSCH scheduled with P-RNTI, RA-RNTI, SI-RNTI and Qm > 2.\nnrgrid.McsInfo=%v\n", *p))
	}

	// 2nd step: get N_RE
	N_RE_ap := N_SC_RB * td - dmrs - xoh
	min := utils.MinInt([]int{156, N_RE_ap})
	N_RE := min * fd

	// 3rd step: get N_info
	N_info := scale * float64(N_RE) * (R / 1024) * float64(Qm) * float64(layer)

	// 4th step: get TBS
	var tbs int
	if N_info <= 3824 {
		n := utils.MaxInt([]int{3, utils.FloorInt(math.Log2(N_info)) - 6})
		n2 := 1 << n
		N_info_ap := utils.MaxInt([]int{24, n2 * utils.FloorInt(N_info / float64(n2))})
		for _, v := range nrgrid.TbsTabLessThan3824 {
			if v >= N_info_ap {
				tbs = v
				break
			}
		}
	} else {
		n := utils.FloorInt(math.Log2(N_info-24)) - 5
		n2 := 1 << n
		N_info_ap := utils.MaxInt([]int{3840, n2 * utils.RoundInt((N_info-24) / float64(n2))})
		if R <= 256 {
			C := utils.CeilInt(float64(N_info_ap+24) / 3816)
			tbs = 8*C*utils.CeilInt(float64(N_info_ap+24) / float64(8*C)) - 24
		} else {
			if N_info_ap > 8424 {
				C := utils.CeilInt(float64(N_info_ap+24) / 8424)
				tbs = 8*C*utils.CeilInt(float64(N_info_ap+24) / float64(8*C)) - 24
			} else {
				tbs = 8*utils.CeilInt(float64(N_info_ap+24)/8) - 24
			}
		}
	}

	// The UE is not expected to receive a PDSCH assigned by a PDCCH with CRC scrambled by SI-RNTI with a TBS exceeding 2976 bits.
	if rnti == "SI-RNTI" && tbs > 2976 {
		return 0, errors.New(fmt.Sprintf("The UE is not expected to receive a PDSCH assigned by a PDCCH with CRC scrambled by SI-RNTI with a TBS exceeding 2976 bits.\nCalculated TBS=%v bits\n", tbs))
	}

	return tbs, nil
}

// confCommonSettingCmd represents the nrrg conf commonsetting command
var confCommonSettingCmd = &cobra.Command{
	Use:   "commonsetting",
	Short: "",
	Long: `nrrg conf commonsetting can be used to get/set common-setting related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confCss0Cmd represents the nrrg conf css0 command
var confCss0Cmd = &cobra.Command{
	Use:   "css0",
	Short: "",
	Long: `nrrg conf css0 can be used to get/set Common search space(CSS0) related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confCoreset1Cmd represents the nrrg conf coreset1 command
var confCoreset1Cmd = &cobra.Command{
	Use:   "coreset1",
	Short: "",
	Long: `nrrg conf coreset1 can be used to get/set CORESET1 related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confUssCmd represents the nrrg conf uss command
var confUssCmd = &cobra.Command{
	Use:   "uss",
	Short: "",
	Long: `nrrg conf uss can be used to get/set UE-specific search space related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confDci10Cmd represents the nrrg conf dci10 command
var confDci10Cmd = &cobra.Command{
	Use:   "dci10",
	Short: "",
	Long: `nrrg conf dci10 can be used to get/set DCI 1_0 (scheduling SIB1/Msg2/Msg4) related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confDci11Cmd represents the nrrg conf dci11 command
var confDci11Cmd = &cobra.Command{
	Use:   "dci11",
	Short: "",
	Long: `nrrg conf dci11 can be used to get/set DCI 1_1(scheduling PDSCH with C-RNTI) related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confMsg3Cmd represents the nrrg conf msg3 command
var confMsg3Cmd = &cobra.Command{
	Use:   "msg3",
	Short: "",
	Long: `nrrg conf msg3 can be used to get/set Msg3(scheduled by UL grant in RAR) related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confDci01Cmd represents the nrrg conf dci01 command
var confDci01Cmd = &cobra.Command{
	Use:   "dci01",
	Short: "",
	Long: `nrrg conf dci01 can be used to get/set DCI 0_1(scheduling PUSCH with C-RNTI) related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confBwpCmd represents the nrrg conf bwp command
var confBwpCmd = &cobra.Command{
	Use:   "bwp",
	Short: "",
	Long: `nrrg conf bwp can be used to get/set generic BWP related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confRachCmd represents the nrrg conf rach command
var confRachCmd = &cobra.Command{
	Use:   "rach",
	Short: "",
	Long: `nrrg conf rach can be used to get/set random access related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confDmrsCommonCmd represents the nrrg conf dmrscommon command
var confDmrsCommonCmd = &cobra.Command{
	Use:   "dmrscommon",
	Short: "",
	Long: `nrrg conf dmrscommon can be used to get/set DMRS of SIB1/Msg2/Msg4/Msg3 related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confDmrsPdschCmd represents the nrrg conf dmrspdsch command
var confDmrsPdschCmd = &cobra.Command{
	Use:   "dmrspdsch",
	Short: "",
	Long: `nrrg conf dmrspdsch can be used to get/set DMRS of PDSCH related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confPtrsPdschCmd represents the nrrg conf ptrspdsch command
var confPtrsPdschCmd = &cobra.Command{
	Use:   "ptrspdsch",
	Short: "",
	Long: `nrrg conf ptrspdsch can be used to get/set PTRS of PDSCH related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confDmrsPuschCmd represents the nrrg conf dmrspusch command
var confDmrsPuschCmd = &cobra.Command{
	Use:   "dmrspusch",
	Short: "",
	Long: `nrrg conf dmrspusch can be used to get/set DMRS of PUSCH related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confPtrsPuschCmd represents the nrrg conf ptrspusch command
var confPtrsPuschCmd = &cobra.Command{
	Use:   "ptrspusch",
	Short: "",
	Long: `nrrg conf ptrspusch can be used to get/set PTRS of PUSCH related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confPdschCmd represents the nrrg conf pdsch command
var confPdschCmd = &cobra.Command{
	Use:   "pdsch",
	Short: "",
	Long: `nrrg conf pdsch can be used to get/set PDSCH-config or PDSCH-ServingCellConfig related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
	    laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confPuschCmd represents the nrrg conf pusch command
var confPuschCmd = &cobra.Command{
	Use:   "pusch",
	Short: "",
	Long: `nrrg conf pusch can be used to get/set PUSCH-config or PUSCH-ServingCellConfig related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confNzpCsiRsCmd represents the nrrg conf nzpcsirs command
var confNzpCsiRsCmd = &cobra.Command{
	Use:   "nzpcsirs",
	Short: "",
	Long: `nrrg conf nzpcsirs can be used to get/set NZP-CSI-RS resource related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confTrsCmd represents the nrrg conf trs command
var confTrsCmd = &cobra.Command{
	Use:   "trs",
	Short: "",
	Long: `nrrg conf trs can be used to get/set TRS resources related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confCsiImCmd represents the nrrg conf csiim command
var confCsiImCmd = &cobra.Command{
	Use:   "csiim",
	Short: "",
	Long: `nrrg conf csiim can be used to get/set CSI-IM resource related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confCsiReportCmd represents the nrrg conf csireport command
var confCsiReportCmd = &cobra.Command{
	Use:   "csireport",
	Short: "",
	Long: `nrrg conf csireport can be used to get/set CSI-ResourceConfig and CSI-ReportConfig related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confSrscmd represents the nrrg conf srs command
var confSrsCmd = &cobra.Command{
	Use:   "srs",
	Short: "",
	Long: `nrrg conf srs can be used to get/set SRS-Resource and SRS-ResourceSet related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confPucchcmd represents the nrrg conf pucch command
var confPucchCmd = &cobra.Command{
	Use:   "pucch",
	Short: "",
	Long: `nrrg conf pucch can be used to get/set PUCCH-FormatConfig/PUCCH-Resource/SchedulingRequestResourceConfig related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// confAdvancedCmd represents the nrrg conf advanced command
var confAdvancedCmd = &cobra.Command{
	Use:   "advanced",
	Short: "",
	Long: `nrrg conf advanced can be used to get/set advanced-settings related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// TODO: add more subcmd here!!!

// nrrgSimCmd represents the nrrg sim command
var nrrgSimCmd = &cobra.Command{
	Use:   "sim",
	Short: "",
	Long: `nrrg sim can be used to perform NR-Uu simulation.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg sim called")
		viper.WriteConfig()

		// Examples of 'expression evaluation'
		// KPI = Formula
		expression, err := govaluate.NewEvaluableExpression("(f1+f2-f3)/f4*f5/(f6+f7)/100")
		fmt.Println(expression.Vars())
		fmt.Println(expression.Tokens())
		fmt.Println(expression.String())
		sql, _ := expression.ToSQLQuery()
		fmt.Println(sql)

		parameters := make(map[string]interface{})
		parameters["f1"] = 2
		parameters["f2"] = 2
		parameters["f3"] = 3
		parameters["f4"] = 1
		parameters["f5"] = 2
		parameters["f6"] = 2
		parameters["f7"] = 2

		result, err := expression.Evaluate(parameters)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			if math.IsNaN(result.(float64)) {
				fmt.Println("NaN")
			} else if math.IsInf(result.(float64), 0) {
				fmt.Println("Inf")
			} else {
				fmt.Println(result)
			}
		}
	},
}

func init() {
	nrrgCmd.AddCommand(confGridSettingCmd)
	nrrgCmd.AddCommand(confMibCmd)
	nrrgCmd.AddCommand(confCommonSettingCmd)
	nrrgCmd.AddCommand(confCss0Cmd)
	nrrgCmd.AddCommand(confCoreset1Cmd)
	nrrgCmd.AddCommand(confUssCmd)
	nrrgCmd.AddCommand(confDci10Cmd)
	nrrgCmd.AddCommand(confDci11Cmd)
	nrrgCmd.AddCommand(confMsg3Cmd)
	nrrgCmd.AddCommand(confDci01Cmd)
	nrrgCmd.AddCommand(confBwpCmd)
	nrrgCmd.AddCommand(confRachCmd)
	nrrgCmd.AddCommand(confDmrsCommonCmd)
	nrrgCmd.AddCommand(confDmrsPdschCmd)
	nrrgCmd.AddCommand(confPtrsPdschCmd)
	nrrgCmd.AddCommand(confDmrsPuschCmd)
	nrrgCmd.AddCommand(confPtrsPuschCmd)
	nrrgCmd.AddCommand(confPdschCmd)
	nrrgCmd.AddCommand(confPuschCmd)
	nrrgCmd.AddCommand(confNzpCsiRsCmd)
	nrrgCmd.AddCommand(confTrsCmd)
	nrrgCmd.AddCommand(confCsiImCmd)
	nrrgCmd.AddCommand(confCsiReportCmd)
	nrrgCmd.AddCommand(confSrsCmd)
	nrrgCmd.AddCommand(confPucchCmd)
	nrrgCmd.AddCommand(confAdvancedCmd)

	nrrgCmd.AddCommand(nrrgSimCmd)
	rootCmd.AddCommand(nrrgCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nrrgCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	initConfGridSettingCmd()
	//initConfSsbGridCmd()
	//initConfSsbBurstCmd()
	initConfMibCmd()
	//initConfCarrierGridCmd()
	initConfCommonSettingCmd()
	initConfCss0Cmd()
	initConfCoreset1Cmd()
	initConfUssCmd()
	initConfDci10Cmd()
	initConfDci11Cmd()
	initConfMsg3Cmd()
	initConfDci01Cmd()
	initConfBwpCmd()
	initConfRachCmd()
	initConfDmrsCommonCmd()
	initConfDmrsPdschCmd()
	initConfPtrsPdschCmd()
	initConfDmrsPuschCmd()
	initConfPtrsPuschCmd()
	initConfPdschCmd()
	initConfPuschCmd()
	initConfNzpCsiRsCmd()
	initConfTrsCmd()
	initConfCsiImCmd()
	initConfCsiReportCmd()
	initConfSrsCmd()
	initConfPucchCmd()
	initConfAdvancedCmd()
}

func initConfGridSettingCmd() {
	// freqBand part
	confGridSettingCmd.Flags().StringVar(&flags.gridsetting.band, "band", "n28", "Operating band")
	confGridSettingCmd.Flags().StringVar(&flags.gridsetting._duplexMode, "_duplexMode", "FDD", "Duplex mode")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._maxDlFreq, "_maxDlFreq", 803, "Maximum DL frequency(MHz)")
	confGridSettingCmd.Flags().StringVar(&flags.gridsetting._freqRange, "_freqRange", "FR1", "Frequency range(FR1/FR2)")
	confGridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting.band", confGridSettingCmd.Flags().Lookup("band"))
	viper.BindPFlag("nrrg.gridsetting._duplexMode", confGridSettingCmd.Flags().Lookup("_duplexMode"))
	viper.BindPFlag("nrrg.gridsetting._maxDlFreq", confGridSettingCmd.Flags().Lookup("_maxDlFreq"))
	viper.BindPFlag("nrrg.gridsetting._freqRange", confGridSettingCmd.Flags().Lookup("_freqRange"))
	confGridSettingCmd.Flags().MarkHidden("_duplexMode")
	confGridSettingCmd.Flags().MarkHidden("_maxDlFreq")
	confGridSettingCmd.Flags().MarkHidden("_freqRange")

	// SCS
	confGridSettingCmd.Flags().StringVar(&flags.gridsetting.scs, "scs", "15KHz", "Subcarrier spacing for SSB/RMSI/Carrier/BWP etc.")
	confGridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting.scs", confGridSettingCmd.Flags().Lookup("scs"))

	// ssbGrid part and ssbBurst part
	confGridSettingCmd.Flags().StringVar(&flags.gridsetting._ssbScs, "_ssbScs", "15KHz", "SSB subcarrier spacing")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting.gscn, "gscn", 1931, "SSB GSCN")
	confGridSettingCmd.Flags().StringVar(&flags.gridsetting._ssbPattern, "_ssbPattern", "Case A", "SSB pattern")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._kSsb, "_kSsb", 2, "k_SSB[0..23]")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._nCrbSsb, "_nCrbSsb", 69, "n_CRB_SSB")
	confGridSettingCmd.Flags().StringVar(&flags.gridsetting.ssbPeriod, "ssbPeriod", "20ms", "ssb-PeriodicityServingCell[5ms,10ms,20ms,40ms,80ms,160ms]")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._maxLBar, "_maxLBar", 4, "L_max_bar as specified in 38.213")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._maxL, "_maxL", 4, "L_max as specified in 38.213")
	confGridSettingCmd.Flags().IntSliceVar(&flags.gridsetting.candSsbIndex, "candSsbIndex", []int{0,1,2,3}, "List of candidate SSB index")
	confGridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting._ssbScs", confGridSettingCmd.Flags().Lookup("_ssbScs"))
	viper.BindPFlag("nrrg.gridsetting.gscn", confGridSettingCmd.Flags().Lookup("gscn"))
	viper.BindPFlag("nrrg.gridsetting._ssbPattern", confGridSettingCmd.Flags().Lookup("_ssbPattern"))
	viper.BindPFlag("nrrg.gridsetting._kSsb", confGridSettingCmd.Flags().Lookup("_kSsb"))
	viper.BindPFlag("nrrg.gridsetting._nCrbSsb", confGridSettingCmd.Flags().Lookup("_nCrbSsb"))
	viper.BindPFlag("nrrg.ssbBurst.ssbPeriod", confGridSettingCmd.Flags().Lookup("ssbPeriod"))
	viper.BindPFlag("nrrg.ssbBurst._maxLBar", confGridSettingCmd.Flags().Lookup("_maxLBar"))
	viper.BindPFlag("nrrg.ssbBurst._maxL", confGridSettingCmd.Flags().Lookup("_maxL"))
	viper.BindPFlag("nrrg.ssbBurst.candSsbIndex", confGridSettingCmd.Flags().Lookup("candSsbIndex"))
	confGridSettingCmd.Flags().MarkHidden("_ssbScs")
	confGridSettingCmd.Flags().MarkHidden("_ssbPattern")
	confGridSettingCmd.Flags().MarkHidden("_kSsb")
	confGridSettingCmd.Flags().MarkHidden("_nCrbSsb")
	confGridSettingCmd.Flags().MarkHidden("_maxLBar")
	confGridSettingCmd.Flags().MarkHidden("_maxL")

	// carrierGrid part and MIB-subCarrierSpacingCommon
	confGridSettingCmd.Flags().StringVar(&flags.gridsetting._carrierScs, "_carrierScs", "15KHz", "subcarrierSpacing of SCS-SpecificCarrier")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting.dlArfcn, "dlArfcn", 154600, "DL ARFCN")
	confGridSettingCmd.Flags().StringVar(&flags.gridsetting.bw, "bw", "30MHz", "Transmission bandwidth(MHz)")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._carrierNumRbs, "_carrierNumRbs", 160, "carrierBandwidth(N_RB) of SCS-SpecificCarrier")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._offsetToCarrier, "_offsetToCarrier", 0, "_offsetToCarrier of SCS-SpecificCarrier")
	confGridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting._carrierScs", confGridSettingCmd.Flags().Lookup("_carrierScs"))
	viper.BindPFlag("nrrg.gridsetting.dlArfcn", confGridSettingCmd.Flags().Lookup("dlArfcn"))
	viper.BindPFlag("nrrg.gridsetting.bw", confGridSettingCmd.Flags().Lookup("bw"))
	viper.BindPFlag("nrrg.gridsetting._carrierNumRbs", confGridSettingCmd.Flags().Lookup("_carrierNumRbs"))
	viper.BindPFlag("nrrg.gridsetting._offsetToCarrier", confGridSettingCmd.Flags().Lookup("_offsetToCarrier"))
	confGridSettingCmd.Flags().MarkHidden("_carrierScs")
	confGridSettingCmd.Flags().MarkHidden("_carrierNumRbs")
	confGridSettingCmd.Flags().MarkHidden("_offsetToCarrier")

	// MIB part
	confGridSettingCmd.Flags().StringVar(&flags.gridsetting._mibCommonScs, "_mibCommonScs", "15KHz", "subCarrierSpacingCommon of MIB")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting.rmsiCoreset0, "rmsiCoreset0", 7, "coresetZero of PDCCH-ConfigSIB1[0..15]")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting.rmsiCss0, "rmsiCss0", 4, "searchSpaceZero of PDCCH-ConfigSIB1[0..15]")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._coreset0MultiplexingPat, "_coreset0MultiplexingPat", 1, "Multiplexing pattern of CORESET0")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._coreset0NumRbs, "_coreset0NumRbs", 48, "Number of PRBs of CORESET0")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._coreset0NumSymbs, "_coreset0NumSymbs", 1, "Number of OFDM symbols of CORESET0")
	confGridSettingCmd.Flags().IntSliceVar(&flags.gridsetting._coreset0OffsetList, "_coreset0OffsetList", []int{16}, "List of offset of CORESET0")
	confGridSettingCmd.Flags().IntVar(&flags.gridsetting._coreset0Offset, "_coreset0Offset", 16, "Offset of CORESET0")
	confGridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting._mibCommonScs", confGridSettingCmd.Flags().Lookup("_mibCommonScs"))
	viper.BindPFlag("nrrg.gridsetting.rmsiCoreset0", confGridSettingCmd.Flags().Lookup("rmsiCoreset0"))
	viper.BindPFlag("nrrg.gridsetting.rmsiCss0", confGridSettingCmd.Flags().Lookup("rmsiCss0"))
	viper.BindPFlag("nrrg.gridsetting._coreset0MultiplexingPat", confGridSettingCmd.Flags().Lookup("_coreset0MultiplexingPat"))
	viper.BindPFlag("nrrg.gridsetting._coreset0NumRbs", confGridSettingCmd.Flags().Lookup("_coreset0NumRbs"))
	viper.BindPFlag("nrrg.gridsetting._coreset0NumSymbs", confGridSettingCmd.Flags().Lookup("_coreset0NumSymbs"))
	viper.BindPFlag("nrrg.gridsetting._coreset0OffsetList", confGridSettingCmd.Flags().Lookup("_coreset0OffsetList"))
	viper.BindPFlag("nrrg.gridsetting._coreset0Offset", confGridSettingCmd.Flags().Lookup("_coreset0Offset"))
	confGridSettingCmd.Flags().MarkHidden("_mibCommonScs")
	confGridSettingCmd.Flags().MarkHidden("_coreset0MultiplexingPat")
	confGridSettingCmd.Flags().MarkHidden("_coreset0NumRbs")
	confGridSettingCmd.Flags().MarkHidden("_coreset0NumSymbs")
	confGridSettingCmd.Flags().MarkHidden("_coreset0OffsetList")
	confGridSettingCmd.Flags().MarkHidden("_coreset0Offset")
}

func initConfMibCmd() {
	confMibCmd.Flags().IntVar(&flags.mib.sfn, "sfn", 0, "System frame number(SFN)[0..1023]")
	confMibCmd.Flags().IntVar(&flags.mib.hrf, "hrf", 0, "Half frame bit[0,1]")
	confMibCmd.Flags().StringVar(&flags.mib.dmrsTypeAPos, "dmrsTypeAPos", "pos2", "dmrs-TypeA-Position[pos2,pos3]")

	confMibCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.mib.sfn", confMibCmd.Flags().Lookup("sfn"))
	viper.BindPFlag("nrrg.mib.hrf", confMibCmd.Flags().Lookup("hrf"))
	viper.BindPFlag("nrrg.mib.dmrsTypeAPos", confMibCmd.Flags().Lookup("dmrsTypeAPos"))

}

func initConfCommonSettingCmd() {
	confCommonSettingCmd.Flags().IntVar(&flags.commonSetting.pci, "pci", 0, "Physical cell identity[0..1007]")
	confCommonSettingCmd.Flags().StringVar(&flags.commonSetting.numUeAp, "numUeAp", "2T", "Number of UE antennas[1T,2T,4T]")
	confCommonSettingCmd.Flags().StringVar(&flags.commonSetting._refScs, "_refScs", "30KHz", "referenceSubcarrierSpacing of TDD-UL-DL-ConfigCommon")
	confCommonSettingCmd.Flags().StringSliceVar(&flags.commonSetting.patPeriod, "patPeriod", []string{"5ms"}, "dl-UL-TransmissionPeriodicity of TDD-UL-DL-ConfigCommon[0.5ms,0.625ms,1ms,1.25ms,2ms,2.5ms,3ms,4ms,5ms,10ms]")
	confCommonSettingCmd.Flags().IntSliceVar(&flags.commonSetting.patNumDlSlots, "patNumDlSlots", []int{7}, "nrofDownlinkSlot of TDD-UL-DL-ConfigCommon[0..80]")
	confCommonSettingCmd.Flags().IntSliceVar(&flags.commonSetting.patNumDlSymbs, "patNumDlSymbs", []int{6}, "nrofDownlinkSymbols of TDD-UL-DL-ConfigCommon[0..13]")
	confCommonSettingCmd.Flags().IntSliceVar(&flags.commonSetting.patNumUlSymbs, "patNumUlSymbs", []int{4}, "nrofUplinkSymbols of TDD-UL-DL-ConfigCommon[0..13]")
	confCommonSettingCmd.Flags().IntSliceVar(&flags.commonSetting.patNumUlSlots, "patNumUlSlots", []int{2}, "nrofUplinkSlots of TDD-UL-DL-ConfigCommon[0..80]")
	confCommonSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.commonsetting.pci", confCommonSettingCmd.Flags().Lookup("pci"))
	viper.BindPFlag("nrrg.commonsetting.numUeAp", confCommonSettingCmd.Flags().Lookup("numUeAp"))
	viper.BindPFlag("nrrg.commonsetting._refScs", confCommonSettingCmd.Flags().Lookup("_refScs"))
	viper.BindPFlag("nrrg.commonsetting.patPeriod", confCommonSettingCmd.Flags().Lookup("patPeriod"))
	viper.BindPFlag("nrrg.commonsetting.patNumDlSlots", confCommonSettingCmd.Flags().Lookup("patNumDlSlots"))
	viper.BindPFlag("nrrg.commonsetting.patNumDlSymbs", confCommonSettingCmd.Flags().Lookup("patNumDlSymbs"))
	viper.BindPFlag("nrrg.commonsetting.patNumUlSymbs", confCommonSettingCmd.Flags().Lookup("patNumUlSymbs"))
	viper.BindPFlag("nrrg.commonsetting.patNumUlSlots", confCommonSettingCmd.Flags().Lookup("patNumUlSlots"))
	confCommonSettingCmd.Flags().MarkHidden("_refScs")
}

func initConfCss0Cmd() {
	confCss0Cmd.Flags().IntVar(&flags.css0.css0AggLevel, "css0AggLevel", 4, "CCE aggregation level of CSS0[4,8,16]")
	confCss0Cmd.Flags().StringVar(&flags.css0.css0NumCandidates, "css0NumCandidates", "n4", "Number of PDCCH candidates of CSS0[n1,n2,n4]")
	confCss0Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.css0.css0AggLevel", confCss0Cmd.Flags().Lookup("css0AggLevel"))
	viper.BindPFlag("nrrg.css0.css0NumCandidates", confCss0Cmd.Flags().Lookup("css0NumCandidates"))
}

func initConfCoreset1Cmd() {
	confCoreset1Cmd.Flags().StringVar(&flags.coreset1.coreset1FreqRes, "coreset1FreqRes", "111111111111111111111111111111111111111111111", "frequencyDomainResources of ControlResourceSet")
	// confCoreset1Cmd.Flags().IntVar(&flags.coreset1.coreset1NumSymbs, "coreset1Duration", 1, "duration of ControlResourceSet[1..3]")
	confCoreset1Cmd.Flags().IntVar(&flags.coreset1.coreset1Duration, "coreset1Duration", 1, "duration of ControlResourceSet[1..3]")
	confCoreset1Cmd.Flags().StringVar(&flags.coreset1.coreset1CceRegMap, "coreset1CceRegMap", "interleaved", "cce-REG-MappingType of ControlResourceSet[1..3]")
	confCoreset1Cmd.Flags().StringVar(&flags.coreset1.coreset1RegBundleSize, "coreset1RegBundleSize", "n2", "reg-BundleSize of ControlResourceSet[n2,n6]")
	confCoreset1Cmd.Flags().StringVar(&flags.coreset1.coreset1InterleaverSize, "coreset1InterleaverSize", "n2", "interleaverSize of ControlResourceSet[n2,n3,n6]")
	confCoreset1Cmd.Flags().IntVar(&flags.coreset1.coreset1ShiftInd, "coreset1ShiftInd", 0, "shiftIndex of ControlResourceSet[0..274]")
	confCoreset1Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.coreset1.coreset1FreqRes", confCoreset1Cmd.Flags().Lookup("coreset1FreqRes"))
	viper.BindPFlag("nrrg.coreset1.coreset1Duration", confCoreset1Cmd.Flags().Lookup("coreset1Duration"))
	viper.BindPFlag("nrrg.coreset1.coreset1CceRegMap", confCoreset1Cmd.Flags().Lookup("coreset1CceRegMap"))
	viper.BindPFlag("nrrg.coreset1.coreset1RegBundleSize", confCoreset1Cmd.Flags().Lookup("coreset1RegBundleSize"))
	viper.BindPFlag("nrrg.coreset1.coreset1InterleaverSize", confCoreset1Cmd.Flags().Lookup("coreset1InterleaverSize"))
	viper.BindPFlag("nrrg.coreset1.coreset1ShiftInd", confCoreset1Cmd.Flags().Lookup("coreset1ShiftInd"))
}

func initConfUssCmd() {
	confUssCmd.Flags().StringVar(&flags.uss.ussPeriod, "ussPeriod", "sl1", "monitoringSlotPeriodicity of SearchSpace[sl1,sl2,sl4,sl5,sl8,sl10,sl16,sl20,sl40,sl80,sl160,sl320,sl640,sl1280,sl2560]")
	confUssCmd.Flags().IntVar(&flags.uss.ussOffset, "ussOffset", 0, "monitoringSlotOffset of SearchSpace[0..ussPeriod-1]")
	confUssCmd.Flags().IntVar(&flags.uss.ussDuration, "ussDuration", 1, "duration of SearchSpace[1 or 2..ussPeriod-1]")
	confUssCmd.Flags().StringVar(&flags.uss.ussFirstSymbs, "ussFirstSymbs", "10101010101010", "monitoringSymbolsWithinSlot of SearchSpace")
	confUssCmd.Flags().IntVar(&flags.uss.ussAggLevel, "ussAggLevel", 4, "aggregationLevel of SearchSpace[1,2,4,8,16]")
	confUssCmd.Flags().StringVar(&flags.uss.ussNumCandidates, "ussNumCandidates", "n1", "nrofCandidates of SearchSpace[n1,n2,n3,n4,n5,n6,n8]")
	confUssCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.uss.ussPeriod", confUssCmd.Flags().Lookup("ussPeriod"))
	viper.BindPFlag("nrrg.uss.ussOffset", confUssCmd.Flags().Lookup("ussOffset"))
	viper.BindPFlag("nrrg.uss.ussDuration", confUssCmd.Flags().Lookup("ussDuration"))
	viper.BindPFlag("nrrg.uss.ussFirstSymbs", confUssCmd.Flags().Lookup("ussFirstSymbs"))
	viper.BindPFlag("nrrg.uss.ussAggLevel", confUssCmd.Flags().Lookup("ussAggLevel"))
	viper.BindPFlag("nrrg.uss.ussNumCandidates", confUssCmd.Flags().Lookup("ussNumCandidates"))
}

func initConfDci10Cmd() {
	confDci10Cmd.Flags().StringSliceVar(&flags.dci10._rnti, "_rnti", []string{"SI-RNTI", "RA-RNTI", "TC-RNTI"}, "RNTI for DCI 1_0")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10._muPdcch, "_muPdcch", []int{1, 1, 1}, "Subcarrier spacing of PDCCH[0..3]")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10._muPdsch, "_muPdsch", []int{1, 1, 1}, "Subcarrier spacing of PDSCH[0..3]")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10.dci10TdRa, "dci10TdRa", []int{10, 10, 10}, "Time-domain-resource-assignment field of DCI 1_0[0..15]")
	confDci10Cmd.Flags().StringSliceVar(&flags.dci10._tdMappingType, "_tdMappingType", []string{"typeB", "typeB", "typeB"}, "Mapping type for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10._tdK0, "_tdK0", []int{0, 0, 0}, "Slot offset K0 for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10._tdSliv, "_tdSliv", []int{26, 26, 26}, "SLIV for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10._tdStartSymb, "_tdStartSymb", []int{12, 12, 12}, "Starting symbol S for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10._tdNumSymbs, "_tdNumSymbs", []int{2, 2, 2}, "Number of OFDM symbols L for PDSCH time-domain allocation")
	confDci10Cmd.Flags().StringSliceVar(&flags.dci10._fdRaType, "_fdRaType", []string{"raType1", "raType1", "raType1"}, "resourceAllocation for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10._fdBitwidthRaType1, "_fdBitwidthRaType1", []int{11, 11, 11}, "Bitwidth of PDSCH frequency-domain allocation for RA Type 1")
	confDci10Cmd.Flags().StringSliceVar(&flags.dci10._fdRa, "_fdRa", []string{"00001011111", "00001011111", "00001011111"}, "Frequency-domain-resource-assignment field of DCI 1_0")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10.dci10FdStartRb, "dci10FdStartRb", []int{0, 0, 0}, "RB_start of RIV for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10.dci10FdNumRbs, "dci10FdNumRbs", []int{48, 48, 48}, "L_RBs of RIV for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().StringSliceVar(&flags.dci10.dci10FdVrbPrbMappingType, "dci10FdVrbPrbMappingType", []string{"interleaved", "interleaved", "interleaved"}, "VRB-to-PRB-mapping field of DCI 1_0")
	confDci10Cmd.Flags().StringSliceVar(&flags.dci10._fdBundleSize, "_fdBundleSize", []string{"n2", "n2", "n2"}, "L(vrb-ToPRB-Interleaver) for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10.dci10McsCw0, "dci10McsCw0", []int{2, 2, 2}, "Modulation-and-coding-scheme field of DCI 1_0[0..9]")
	confDci10Cmd.Flags().IntSliceVar(&flags.dci10._tbs, "_tbs", []int{408, 408, 408}, "Transport block size(bits) for PDSCH")
	confDci10Cmd.Flags().IntVar(&flags.dci10.dci10Msg2TbScaling, "dci10Msg2TbScaling", 0, "TB-scaling field of DCI 1_0 scheduling Msg2[0..2]")
	confDci10Cmd.Flags().IntVar(&flags.dci10.dci10Msg4DeltaPri, "dci10Msg4DeltaPri", 1, "PUCCH-resource-indicator field of DCI 1_0 scheduling Msg4[0..7]")
	confDci10Cmd.Flags().IntVar(&flags.dci10.dci10Msg4TdK1, "dci10Msg4TdK1", 6, "PDSCH-to-HARQ_feedback-timing-indicator(K1) field of DCI 1_0 scheduling Msg4[0..7]")
	confDci10Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dci10._rnti", confDci10Cmd.Flags().Lookup("_rnti"))
	viper.BindPFlag("nrrg.dci10._muPdcch", confDci10Cmd.Flags().Lookup("_muPdcch"))
	viper.BindPFlag("nrrg.dci10._muPdsch", confDci10Cmd.Flags().Lookup("_muPdsch"))
	viper.BindPFlag("nrrg.dci10.dci10TdRa", confDci10Cmd.Flags().Lookup("dci10TdRa"))
	viper.BindPFlag("nrrg.dci10._tdMappingType", confDci10Cmd.Flags().Lookup("_tdMappingType"))
	viper.BindPFlag("nrrg.dci10._tdK0", confDci10Cmd.Flags().Lookup("_tdK0"))
	viper.BindPFlag("nrrg.dci10._tdSliv", confDci10Cmd.Flags().Lookup("_tdSliv"))
	viper.BindPFlag("nrrg.dci10._tdStartSymb", confDci10Cmd.Flags().Lookup("_tdStartSymb"))
	viper.BindPFlag("nrrg.dci10._tdNumSymbs", confDci10Cmd.Flags().Lookup("_tdNumSymbs"))
	viper.BindPFlag("nrrg.dci10._fdRaType", confDci10Cmd.Flags().Lookup("_fdRaType"))
	viper.BindPFlag("nrrg.dci10._fdBitwidthRaType1", confDci10Cmd.Flags().Lookup("_fdBitwidthRaType1"))
	viper.BindPFlag("nrrg.dci10._fdRa", confDci10Cmd.Flags().Lookup("_fdRa"))
	viper.BindPFlag("nrrg.dci10.dci10FdStartRb", confDci10Cmd.Flags().Lookup("dci10FdStartRb"))
	viper.BindPFlag("nrrg.dci10.dci10FdNumRbs", confDci10Cmd.Flags().Lookup("dci10FdNumRbs"))
	viper.BindPFlag("nrrg.dci10.dci10FdVrbPrbMappingType", confDci10Cmd.Flags().Lookup("dci10FdVrbPrbMappingType"))
	viper.BindPFlag("nrrg.dci10._fdBundleSize", confDci10Cmd.Flags().Lookup("_fdBundleSize"))
	viper.BindPFlag("nrrg.dci10.dci10McsCw0", confDci10Cmd.Flags().Lookup("dci10McsCw0"))
	viper.BindPFlag("nrrg.dci10._tbs", confDci10Cmd.Flags().Lookup("_tbs"))
	viper.BindPFlag("nrrg.dci10.dci10Msg2TbScaling", confDci10Cmd.Flags().Lookup("dci10Msg2TbScaling"))
	viper.BindPFlag("nrrg.dci10.dci10Msg4DeltaPri", confDci10Cmd.Flags().Lookup("dci10Msg4DeltaPri"))
	viper.BindPFlag("nrrg.dci10.dci10Msg4TdK1", confDci10Cmd.Flags().Lookup("dci10Msg4TdK1"))
	confDci10Cmd.Flags().MarkHidden("_rnti")
	confDci10Cmd.Flags().MarkHidden("_muPdcch")
	confDci10Cmd.Flags().MarkHidden("_muPdsch")
	confDci10Cmd.Flags().MarkHidden("_tdMappingType")
	confDci10Cmd.Flags().MarkHidden("_tdK0")
	confDci10Cmd.Flags().MarkHidden("_tdSliv")
	confDci10Cmd.Flags().MarkHidden("_tdStartSymb")
	confDci10Cmd.Flags().MarkHidden("_tdNumSymbs")
	confDci10Cmd.Flags().MarkHidden("_fdRaType")
	confDci10Cmd.Flags().MarkHidden("_fdBitwidthRaType1")
	confDci10Cmd.Flags().MarkHidden("_fdRa")
	confDci10Cmd.Flags().MarkHidden("_fdBundleSize")
	confDci10Cmd.Flags().MarkHidden("_tbs")
}

func initConfDci11Cmd() {
	confDci11Cmd.Flags().StringVar(&flags.dci11._rnti, "_rnti", "C-RNTI", "RNTI for DCI 1_1")
	confDci11Cmd.Flags().IntVar(&flags.dci11._muPdcch, "_muPdcch", 1, "Subcarrier spacing of PDCCH[0..3]")
	confDci11Cmd.Flags().IntVar(&flags.dci11._muPdsch, "_muPdsch", 1, "Subcarrier spacing of PDSCH[0..3]")
	confDci11Cmd.Flags().IntVar(&flags.dci11._actBwp, "_actBwp", 1, "Active DL bandwidth part of PDSCH[0..1]")
	confDci11Cmd.Flags().IntVar(&flags.dci11._indicatedBwp, "_indicatedBwp", 1, "Bandwidth-part-indicator field of DCI 1_1[0..1]")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11TdRa, "dci11TdRa", 16, "Time-domain-resource-assignment field of DCI 1_1[0..15 or 16]")
	confDci11Cmd.Flags().StringVar(&flags.dci11.dci11TdMappingType, "dci11TdMappingType", "typeA", "Mapping type for PDSCH time-domain allocation[typeA,typeB]")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11TdK0, "dci11TdK0", 0, "Slot offset K0 for PDSCH time-domain allocation")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11TdSliv, "dci11TdSliv", 27, "SLIV for PDSCH time-domain allocation")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11TdStartSymb, "dci11TdStartSymb", 0, "Starting symbol S for PDSCH time-domain allocation")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11TdNumSymbs, "dci11TdNumSymbs", 14, "Number of OFDM symbols L for PDSCH time-domain allocation")
	confDci11Cmd.Flags().StringVar(&flags.dci11.dci11FdRaType, "dci11FdRaType", "raType1", "resourceAllocation for PDSCH frequency-domain allocation[raType0,raType1]")
	confDci11Cmd.Flags().IntVar(&flags.dci11._dci11FdBitwidthRaType0, "_dci11FdBitwidthRaType0", 18, "Bitwidth of PDSCH frequency-domain allocation for RA Type 0")
	confDci11Cmd.Flags().IntVar(&flags.dci11._dci11FdBitwidthRaType1, "_dci11FdBitwidthRaType1", 16, "Bitwidth of PDSCH frequency-domain allocation for RA Type 1")
	confDci11Cmd.Flags().StringVar(&flags.dci11.dci11FdRa, "dci11FdRa", "0000001000100001", "Frequency-domain-resource-assignment field of DCI 1_1")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11FdStartRb, "dci11FdStartRb", 0, "RB_start of RIV for PDSCH frequency-domain allocation")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11FdNumRbs, "dci11FdNumRbs", 273, "L_RBs of RIV for PDSCH frequency-domain allocation")
	confDci11Cmd.Flags().StringVar(&flags.dci11.dci11FdVrbPrbMappingType, "dci11FdVrbPrbMappingType", "interleaved", "VRB-to-PRB-mapping field of DCI 1_1[nonInterleaved,interleaved]")
	confDci11Cmd.Flags().StringVar(&flags.dci11.dci11FdBundleSize, "dci11FdBundleSize", "n2", "L(vrb-ToPRB-Interleaver) for PDSCH frequency-domain allocation[n2,n4]")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11McsCw0, "dci11McsCw0", 27, "Modulation-and-coding-scheme-cw0 field of DCI 1_1[-1 or 0..28]")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11McsCw1, "dci11McsCw1", -1, "Modulation-and-coding-scheme-cw1 field of DCI 1_1[-1 or 0..28]")
	confDci11Cmd.Flags().IntSliceVar(&flags.dci11._tbs, "_tbs", []int{1277992, -1}, "Transport block size(bits) for PDSCH")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11DeltaPri, "dci11DeltaPri", 1, "PUCCH-resource-indicator field of DCI 1_1[0..4]")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11TdK1, "dci11TdK1", 2, "PDSCH-to-HARQ_feedback-timing-indicator(K1) field of DCI 1_1[0..7]")
	confDci11Cmd.Flags().IntVar(&flags.dci11.dci11AntPorts, "dci11AntPorts", 10, "Antenna_port(s) field of DCI 1_1[0..15]")
	confDci11Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dci11._rnti", confDci11Cmd.Flags().Lookup("_rnti"))
	viper.BindPFlag("nrrg.dci11._muPdcch", confDci11Cmd.Flags().Lookup("_muPdcch"))
	viper.BindPFlag("nrrg.dci11._muPdsch", confDci11Cmd.Flags().Lookup("_muPdsch"))
	viper.BindPFlag("nrrg.dci11._actBwp", confDci11Cmd.Flags().Lookup("_actBwp"))
	viper.BindPFlag("nrrg.dci11._indicatedBwp", confDci11Cmd.Flags().Lookup("_indicatedBwp"))
	viper.BindPFlag("nrrg.dci11.dci11TdRa", confDci11Cmd.Flags().Lookup("dci11TdRa"))
	viper.BindPFlag("nrrg.dci11.dci11TdMappingType", confDci11Cmd.Flags().Lookup("dci11TdMappingType"))
	viper.BindPFlag("nrrg.dci11.dci11TdK0", confDci11Cmd.Flags().Lookup("dci11TdK0"))
	viper.BindPFlag("nrrg.dci11.dci11TdSliv", confDci11Cmd.Flags().Lookup("dci11TdSliv"))
	viper.BindPFlag("nrrg.dci11.dci11TdStartSymb", confDci11Cmd.Flags().Lookup("dci11TdStartSymb"))
	viper.BindPFlag("nrrg.dci11.dci11TdNumSymbs", confDci11Cmd.Flags().Lookup("dci11TdNumSymbs"))
	viper.BindPFlag("nrrg.dci11.dci11FdRaType", confDci11Cmd.Flags().Lookup("dci11FdRaType"))
	viper.BindPFlag("nrrg.dci11._dci11FdBitwidthRaType0", confDci11Cmd.Flags().Lookup("_dci11FdBitwidthRaType0"))
	viper.BindPFlag("nrrg.dci11._dci11FdBitwidthRaType1", confDci11Cmd.Flags().Lookup("_dci11FdBitwidthRaType1"))
	viper.BindPFlag("nrrg.dci11.dci11FdRa", confDci11Cmd.Flags().Lookup("dci11FdRa"))
	viper.BindPFlag("nrrg.dci11.dci11FdStartRb", confDci11Cmd.Flags().Lookup("dci11FdStartRb"))
	viper.BindPFlag("nrrg.dci11.dci11FdNumRbs", confDci11Cmd.Flags().Lookup("dci11FdNumRbs"))
	viper.BindPFlag("nrrg.dci11.dci11FdVrbPrbMappingType", confDci11Cmd.Flags().Lookup("dci11FdVrbPrbMappingType"))
	viper.BindPFlag("nrrg.dci11.dci11FdBundleSize", confDci11Cmd.Flags().Lookup("dci11FdBundleSize"))
	viper.BindPFlag("nrrg.dci11.dci11McsCw0", confDci11Cmd.Flags().Lookup("dci11McsCw0"))
	viper.BindPFlag("nrrg.dci11.dci11McsCw1", confDci11Cmd.Flags().Lookup("dci11McsCw1"))
	viper.BindPFlag("nrrg.dci11._tbs", confDci11Cmd.Flags().Lookup("_tbs"))
	viper.BindPFlag("nrrg.dci11.dci11DeltaPri", confDci11Cmd.Flags().Lookup("dci11DeltaPri"))
	viper.BindPFlag("nrrg.dci11.dci11TdK1", confDci11Cmd.Flags().Lookup("dci11TdK1"))
	viper.BindPFlag("nrrg.dci11.dci11AntPorts", confDci11Cmd.Flags().Lookup("dci11AntPorts"))
	confDci11Cmd.Flags().MarkHidden("_rnti")
	confDci11Cmd.Flags().MarkHidden("_muPdcch")
	confDci11Cmd.Flags().MarkHidden("_muPdsch")
	confDci11Cmd.Flags().MarkHidden("_actBwp")
	confDci11Cmd.Flags().MarkHidden("_indicatedBwp")
	confDci11Cmd.Flags().MarkHidden("_dci11FdBitwidthRaType0")
	confDci11Cmd.Flags().MarkHidden("_dci11FdBitwidthRaType1")
	confDci11Cmd.Flags().MarkHidden("_tbs")
}

func initConfMsg3Cmd() {
	confMsg3Cmd.Flags().IntVar(&flags.msg3._muPusch, "_muPusch", 1, "Subcarrier spacing of PUSCH[0..3]")
	confMsg3Cmd.Flags().IntVar(&flags.msg3.msg3TdRa, "msg3TdRa", 6, "PUSCH-time-resource-allocation field of RAR UL grant scheduling Msg3[0..15]")
	confMsg3Cmd.Flags().StringVar(&flags.msg3._tdMappingType, "_tdMappingType", "typeB", "Mapping type for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().IntVar(&flags.msg3._tdK2, "_tdK2", 1, "Slot offset K2 for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().IntVar(&flags.msg3._tdDelta, "_tdDelta", 3, "Slot offset delta for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().IntVar(&flags.msg3._tdSliv, "_tdSliv", 74, "SLIV for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().IntVar(&flags.msg3._tdStartSymb, "_tdStartSymb", 4, "Starting symbol S for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().IntVar(&flags.msg3._tdNumSymbs, "_tdNumSymbs", 6, "Number of OFDM symbols L for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().StringVar(&flags.msg3._fdRaType, "_fdRaType", "raType1", "resourceAllocation for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().StringVar(&flags.msg3.msg3FdFreqHop, "msg3FdFreqHop", "enabled", "Frequency-hopping-flag field of RAR UL grant scheduling Msg3[disabled,enabled]")
	confMsg3Cmd.Flags().StringVar(&flags.msg3.msg3FdRa, "msg3FdRa", "0100000100001101", "PUSCH-frequency-resource-allocation field of RAR UL grant scheduling Msg3")
	confMsg3Cmd.Flags().IntVar(&flags.msg3.msg3FdStartRb, "msg3FdStartRb", 0, "RB_start of RIV for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().IntVar(&flags.msg3.msg3FdNumRbs, "msg3FdNumRbs", 62, "L_RBs of RIV for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().IntVar(&flags.msg3._fdSecondHopFreqOff, "_fdSecondHopFreqOff", 68, "Frequency offset of second hop for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().IntVar(&flags.msg3.msg3McsCw0, "msg3McsCw0", 2, "MCS field of RAR UL grant scheduling Msg3[0..28]")
	confMsg3Cmd.Flags().IntVar(&flags.msg3._tbs, "_tbs", 1544, "Transport block size(bits) for PUSCH")
	confMsg3Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.msg3._muPusch", confMsg3Cmd.Flags().Lookup("_muPusch"))
	viper.BindPFlag("nrrg.msg3.msg3TdRa", confMsg3Cmd.Flags().Lookup("msg3TdRa"))
	viper.BindPFlag("nrrg.msg3._tdMappingType", confMsg3Cmd.Flags().Lookup("_tdMappingType"))
	viper.BindPFlag("nrrg.msg3._tdK2", confMsg3Cmd.Flags().Lookup("_tdK2"))
	viper.BindPFlag("nrrg.msg3._tdDelta", confMsg3Cmd.Flags().Lookup("_tdDelta"))
	viper.BindPFlag("nrrg.msg3._tdSliv", confMsg3Cmd.Flags().Lookup("_tdSliv"))
	viper.BindPFlag("nrrg.msg3._tdStartSymb", confMsg3Cmd.Flags().Lookup("_tdStartSymb"))
	viper.BindPFlag("nrrg.msg3._tdNumSymbs", confMsg3Cmd.Flags().Lookup("_tdNumSymbs"))
	viper.BindPFlag("nrrg.msg3._fdRaType", confMsg3Cmd.Flags().Lookup("_fdRaType"))
	viper.BindPFlag("nrrg.msg3.msg3FdFreqHop", confMsg3Cmd.Flags().Lookup("msg3FdFreqHop"))
	viper.BindPFlag("nrrg.msg3.msg3FdRa", confMsg3Cmd.Flags().Lookup("msg3FdRa"))
	viper.BindPFlag("nrrg.msg3.msg3FdStartRb", confMsg3Cmd.Flags().Lookup("msg3FdStartRb"))
	viper.BindPFlag("nrrg.msg3.msg3FdNumRbs", confMsg3Cmd.Flags().Lookup("msg3FdNumRbs"))
	viper.BindPFlag("nrrg.msg3._fdSecondHopFreqOff", confMsg3Cmd.Flags().Lookup("_fdSecondHopFreqOff"))
	viper.BindPFlag("nrrg.msg3.msg3McsCw0", confMsg3Cmd.Flags().Lookup("msg3McsCw0"))
	viper.BindPFlag("nrrg.msg3._tbs", confMsg3Cmd.Flags().Lookup("_tbs"))
	confMsg3Cmd.Flags().MarkHidden("_muPusch")
	confMsg3Cmd.Flags().MarkHidden("_tdMappingType")
	confMsg3Cmd.Flags().MarkHidden("_tdK2")
	confMsg3Cmd.Flags().MarkHidden("_tdDelta")
	confMsg3Cmd.Flags().MarkHidden("_tdSliv")
	confMsg3Cmd.Flags().MarkHidden("_tdStartSymb")
	confMsg3Cmd.Flags().MarkHidden("_tdNumSymbs")
	confMsg3Cmd.Flags().MarkHidden("_fdRaType")
	confMsg3Cmd.Flags().MarkHidden("_fdSecondHopFreqOff")
	confMsg3Cmd.Flags().MarkHidden("_tbs")
}

func initConfDci01Cmd() {
	confDci01Cmd.Flags().StringVar(&flags.dci01._rnti, "_rnti", "C-RNTI", "RNTI for DCI 0_1")
	confDci01Cmd.Flags().IntVar(&flags.dci01._muPdcch, "_muPdcch", 1, "Subcarrier spacing of PDCCH[0..3]")
	confDci01Cmd.Flags().IntVar(&flags.dci01._muPusch, "_muPusch", 1, "Subcarrier spacing of PUSCH[0..3]")
	confDci01Cmd.Flags().IntVar(&flags.dci01._actBwp, "_actBwp", 1, "Active UL bandwidth part of PUSCH[0..1]")
	confDci01Cmd.Flags().IntVar(&flags.dci01._indicatedBwp, "_indicatedBwp", 1, "Bandwidth-part-indicator field of DCI 0_1[0..1]")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01TdRa, "dci01TdRa", 16, "Time-domain-resource-assignment field of DCI 0_1[0..15 or 16]")
	confDci01Cmd.Flags().StringVar(&flags.dci01.dci01TdMappingType, "dci01TdMappingType", "typeA", "Mapping type for PUSCH time-domain allocation[typeA,typeB]")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01TdK2, "dci01TdK2", 1, "Slot offset K2 for PUSCH time-domain allocation[0..32]")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01TdSliv, "dci01TdSliv", 27, "SLIV for PUSCH time-domain allocation")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01TdStartSymb, "dci01TdStartSymb", 0, "Starting symbol S for PUSCH time-domain allocation")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01TdNumSymbs, "dci01TdNumSymbs", 14, "Number of OFDM symbols L for PUSCH time-domain allocation")
	confDci01Cmd.Flags().StringVar(&flags.dci01.dci01FdRaType, "dci01FdRaType", "raType1", "resourceAllocation for PUSCH frequency-domain allocation[raType0,raType1]")
	confDci01Cmd.Flags().StringVar(&flags.dci01.dci01FdFreqHop, "dci01FdFreqHop", "disabled", "Frequency-hopping-flag field for DCI 0_1[disabled,intraSlot,interSlot]")
	confDci01Cmd.Flags().StringVar(&flags.dci01.dci01FdRa, "dci01FdRa", "0000001000100001", "Frequency-domain-resource-assignment field of DCI 0_1")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01FdStartRb, "dci01FdStartRb", 0, "RB_start of RIV for PUSCH frequency-domain allocation")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01FdNumRbs, "dci01FdNumRbs", 273, "L_RBs of RIV for PUSCH frequency-domain allocation")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01McsCw0, "dci01McsCw0", 28, "Modulation-and-coding-scheme-cw0 field of DCI 0_1[0..28]")
	confDci01Cmd.Flags().IntVar(&flags.dci01._tbs, "_tbs", 475584, "Transport block size(bits) for PUSCH")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01CbTpmiNumLayers, "dci01CbTpmiNumLayers", 2, "Precoding-information-and-number-of-layers field of DCI 0_1[0..63]")
	confDci01Cmd.Flags().StringVar(&flags.dci01.dci01Sri, "dci01Sri", "", "SRS-resource-indicator field of DCI 0_1")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01AntPorts, "dci01AntPorts", 0, "Antenna_port(s) field of DCI 0_1[0..7]")
	confDci01Cmd.Flags().IntVar(&flags.dci01.dci01PtrsDmrsMap, "dci01PtrsDmrsMap", 0, "PTRS-DMRS-association field of DCI 0_1[0..3]")
	confDci01Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dci01._rnti", confDci01Cmd.Flags().Lookup("_rnti"))
	viper.BindPFlag("nrrg.dci01._muPdcch", confDci01Cmd.Flags().Lookup("_muPdcch"))
	viper.BindPFlag("nrrg.dci01._muPusch", confDci01Cmd.Flags().Lookup("_muPusch"))
	viper.BindPFlag("nrrg.dci01._actBwp", confDci01Cmd.Flags().Lookup("_actBwp"))
	viper.BindPFlag("nrrg.dci01._indicatedBwp", confDci01Cmd.Flags().Lookup("_indicatedBwp"))
	viper.BindPFlag("nrrg.dci01.dci01TdRa", confDci01Cmd.Flags().Lookup("dci01TdRa"))
	viper.BindPFlag("nrrg.dci01.dci01TdMappingType", confDci01Cmd.Flags().Lookup("dci01TdMappingType"))
	viper.BindPFlag("nrrg.dci01.dci01TdK2", confDci01Cmd.Flags().Lookup("dci01TdK2"))
	viper.BindPFlag("nrrg.dci01.dci01TdSliv", confDci01Cmd.Flags().Lookup("dci01TdSliv"))
	viper.BindPFlag("nrrg.dci01.dci01TdStartSymb", confDci01Cmd.Flags().Lookup("dci01TdStartSymb"))
	viper.BindPFlag("nrrg.dci01.dci01TdNumSymbs", confDci01Cmd.Flags().Lookup("dci01TdNumSymbs"))
	viper.BindPFlag("nrrg.dci01.dci01FdRaType", confDci01Cmd.Flags().Lookup("dci01FdRaType"))
	viper.BindPFlag("nrrg.dci01.dci01FdFreqHop", confDci01Cmd.Flags().Lookup("dci01FdFreqHop"))
	viper.BindPFlag("nrrg.dci01.dci01FdRa", confDci01Cmd.Flags().Lookup("dci01FdRa"))
	viper.BindPFlag("nrrg.dci01.dci01FdStartRb", confDci01Cmd.Flags().Lookup("dci01FdStartRb"))
	viper.BindPFlag("nrrg.dci01.dci01FdNumRbs", confDci01Cmd.Flags().Lookup("dci01FdNumRbs"))
	viper.BindPFlag("nrrg.dci01.dci01McsCw0", confDci01Cmd.Flags().Lookup("dci01McsCw0"))
	viper.BindPFlag("nrrg.dci01._tbs", confDci01Cmd.Flags().Lookup("_tbs"))
	viper.BindPFlag("nrrg.dci01.dci01CbTpmiNumLayers", confDci01Cmd.Flags().Lookup("dci01CbTpmiNumLayers"))
	viper.BindPFlag("nrrg.dci01.dci01Sri", confDci01Cmd.Flags().Lookup("dci01Sri"))
	viper.BindPFlag("nrrg.dci01.dci01AntPorts", confDci01Cmd.Flags().Lookup("dci01AntPorts"))
	viper.BindPFlag("nrrg.dci01.dci01PtrsDmrsMap", confDci01Cmd.Flags().Lookup("dci01PtrsDmrsMap"))
	confDci01Cmd.Flags().MarkHidden("_rnti")
	confDci01Cmd.Flags().MarkHidden("_muPdcch")
	confDci01Cmd.Flags().MarkHidden("_muPusch")
	confDci01Cmd.Flags().MarkHidden("_actBwp")
	confDci01Cmd.Flags().MarkHidden("_indicatedBwp")
	confDci01Cmd.Flags().MarkHidden("_tbs")
}

func initConfBwpCmd() {
	confBwpCmd.Flags().StringSliceVar(&flags.bwp._bwpType, "_bwpType", []string{"iniDlBwp", "dedDlBwp", "iniUlBwp", "dedUlBwp"}, "BWP type")
	confBwpCmd.Flags().IntSliceVar(&flags.bwp._bwpId, "_bwpId", []int{0, 1, 0, 1}, "bwp-Id of BWP-Uplink or BWP-Downlink")
	confBwpCmd.Flags().StringSliceVar(&flags.bwp._bwpScs, "_bwpScs", []string{"30KHz", "30KHz", "30KHz", "30KHz"}, "subcarrierSpacing of BWP")
	confBwpCmd.Flags().StringSliceVar(&flags.bwp._bwpCp, "_bwpCp", []string{"normal", "normal", "normal", "normal"}, "cyclicPrefix of BWP")
	confBwpCmd.Flags().IntSliceVar(&flags.bwp._bwpLocAndBw, "_bwpLocAndBw", []int{12925, 1099, 1099, 1099}, "locationAndBandwidth of BWP")
	confBwpCmd.Flags().IntSliceVar(&flags.bwp._bwpStartRb, "_bwpStartRb", []int{0, 0, 0, 0}, "RB_start of BWP")
	confBwpCmd.Flags().IntSliceVar(&flags.bwp._bwpNumRbs, "_bwpNumRbs", []int{48, 273, 273, 273}, "L_RBs of BWP")
	confBwpCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.bwp._bwpType", confBwpCmd.Flags().Lookup("_bwpType"))
	viper.BindPFlag("nrrg.bwp._bwpId", confBwpCmd.Flags().Lookup("_bwpId"))
	viper.BindPFlag("nrrg.bwp._bwpScs", confBwpCmd.Flags().Lookup("_bwpScs"))
	viper.BindPFlag("nrrg.bwp._bwpCp", confBwpCmd.Flags().Lookup("_bwpCp"))
	viper.BindPFlag("nrrg.bwp._bwpLocAndBw", confBwpCmd.Flags().Lookup("_bwpLocAndBw"))
	viper.BindPFlag("nrrg.bwp._bwpStartRb", confBwpCmd.Flags().Lookup("_bwpStartRb"))
	viper.BindPFlag("nrrg.bwp._bwpNumRbs", confBwpCmd.Flags().Lookup("_bwpNumRbs"))
	confBwpCmd.Flags().MarkHidden("_bwpType")
	confBwpCmd.Flags().MarkHidden("_bwpId")
	confBwpCmd.Flags().MarkHidden("_bwpScs")
	confBwpCmd.Flags().MarkHidden("_bwpCp")
	confBwpCmd.Flags().MarkHidden("_bwpLocAndBw")
	confBwpCmd.Flags().MarkHidden("_bwpStartRb")
	confBwpCmd.Flags().MarkHidden("_bwpNumRbs")
}

func initConfRachCmd() {
	confRachCmd.Flags().IntVar(&flags.rach.prachConfId, "prachConfId", 148, "prach-ConfigurationIndex of RACH-ConfigGeneric[0..255]")
	confRachCmd.Flags().StringVar(&flags.rach._raFormat, "_raFormat", "B4", "Preamble format")
	confRachCmd.Flags().IntVar(&flags.rach._raX, "_raX", 2, "The x in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().IntSliceVar(&flags.rach._raY, "_raY", []int{1}, "The y in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().IntSliceVar(&flags.rach._raSubfNumFr1SlotNumFr2, "_raSubfNumFr1SlotNumFr2", []int{9}, "The Subframe-number in 3GPP TS 38.211 Table 6.3.3.2-2 and Table 6.3.3.2-3, or the Slot-number in Table 6.3.3.2-4")
	confRachCmd.Flags().IntVar(&flags.rach._raStartingSymb, "_raStartingSymb", 0, "The Starting-symbol in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().IntVar(&flags.rach._raNumSlotsPerSubfFr1Per60KSlotFr2, "_raNumSlotsPerSubfFr1Per60KSlotFr2", 1, "The Number-of-PRACH-slots-within-a-subframe in 3GPP TS 38.211 Table 6.3.3.2-2 and Table 6.3.3.2-3, or the Number-of-PRACH-slots-within-a-60-kHz-slot in Table 6.3.3.2-4")
	confRachCmd.Flags().IntVar(&flags.rach._raNumOccasionsPerSlot, "_raNumOccasionsPerSlot", 1, "The Number-of-time-domain-PRACH-occasions-within-a-PRACH-slot in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().IntVar(&flags.rach._raDuration, "_raDuration", 12, "The PRACH-duration in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().StringVar(&flags.rach.msg1Scs, "msg1Scs", "30KHz", "msg1-SubcarrierSpacing of RACH-ConfigCommon")
	confRachCmd.Flags().IntVar(&flags.rach.msg1Fdm, "msg1Fdm", 1, "msg1-FDM of RACH-ConfigGeneric[1,2,4,8]")
	confRachCmd.Flags().IntVar(&flags.rach.msg1FreqStart, "msg1FreqStart", 0, "msg1-FrequencyStart of RACH-ConfigGeneric[0..274]")
	confRachCmd.Flags().StringVar(&flags.rach.raRespWin, "raRespWin", "sl20", "ra-ResponseWindow of RACH-ConfigGeneric[sl1,sl2,sl4,sl8,sl10,sl20,sl40,sl80]")
	confRachCmd.Flags().IntVar(&flags.rach.totNumPreambs, "totNumPreambs", 64, "totalNumberOfRA-Preambles of RACH-ConfigCommon[1..64]")
	confRachCmd.Flags().StringVar(&flags.rach.ssbPerRachOccasion, "ssbPerRachOccasion", "one", "ssb-perRACH-Occasion of RACH-ConfigGeneric[oneEighth,oneFourth,oneHalf,one,two,four,eight,sixteen]")
	confRachCmd.Flags().IntVar(&flags.rach.cbPreambsPerSsb, "cbPreambsPerSsb", 64, "cb-PreamblesPerSSB of RACH-ConfigCommon[depends on ssbPerRachOccasion]")
	confRachCmd.Flags().StringVar(&flags.rach.contResTimer, "contResTimer", "sf64", "ra-ContentionResolutionTimer of RACH-ConfigGeneric[sf8,sf16,sf24,sf32,sf40,sf48,sf56,sf64]")
	confRachCmd.Flags().StringVar(&flags.rach.msg3Tp, "msg3Tp", "disabled", "msg3-transformPrecoder of RACH-ConfigGeneric[disabled,enabled]")
	confRachCmd.Flags().IntVar(&flags.rach._raLen, "_raLen", 139, "L_RA of 3GPP TS 38.211 Table 6.3.3.1-1 and Table 6.3.3.1-2")
	confRachCmd.Flags().IntVar(&flags.rach._raNumRbs, "_raNumRbs", 12, "Allocation-expressed-in-number-of-RBs-for-PUSCH of 3GPP TS 38.211 Table 6.3.3.2-1")
	confRachCmd.Flags().IntVar(&flags.rach._raKBar, "_raKBar", 2, "k_bar of 3GPP TS 38.211 Table 6.3.3.2-1")
	confRachCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.rach.prachConfId", confRachCmd.Flags().Lookup("prachConfId"))
	viper.BindPFlag("nrrg.rach._raFormat", confRachCmd.Flags().Lookup("_raFormat"))
	viper.BindPFlag("nrrg.rach._raX", confRachCmd.Flags().Lookup("_raX"))
	viper.BindPFlag("nrrg.rach._raY", confRachCmd.Flags().Lookup("_raY"))
	viper.BindPFlag("nrrg.rach._raSubfNumFr1SlotNumFr2", confRachCmd.Flags().Lookup("_raSubfNumFr1SlotNumFr2"))
	viper.BindPFlag("nrrg.rach._raStartingSymb", confRachCmd.Flags().Lookup("_raStartingSymb"))
	viper.BindPFlag("nrrg.rach._raNumSlotsPerSubfFr1Per60KSlotFr2", confRachCmd.Flags().Lookup("_raNumSlotsPerSubfFr1Per60KSlotFr2"))
	viper.BindPFlag("nrrg.rach._raNumOccasionsPerSlot", confRachCmd.Flags().Lookup("_raNumOccasionsPerSlot"))
	viper.BindPFlag("nrrg.rach._raDuration", confRachCmd.Flags().Lookup("_raDuration"))
	viper.BindPFlag("nrrg.rach.msg1Scs", confRachCmd.Flags().Lookup("msg1Scs"))
	viper.BindPFlag("nrrg.rach.msg1Fdm", confRachCmd.Flags().Lookup("msg1Fdm"))
	viper.BindPFlag("nrrg.rach.msg1FreqStart", confRachCmd.Flags().Lookup("msg1FreqStart"))
	viper.BindPFlag("nrrg.rach.raRespWin", confRachCmd.Flags().Lookup("raRespWin"))
	viper.BindPFlag("nrrg.rach.totNumPreambs", confRachCmd.Flags().Lookup("totNumPreambs"))
	viper.BindPFlag("nrrg.rach.ssbPerRachOccasion", confRachCmd.Flags().Lookup("ssbPerRachOccasion"))
	viper.BindPFlag("nrrg.rach.cbPreambsPerSsb", confRachCmd.Flags().Lookup("cbPreambsPerSsb"))
	viper.BindPFlag("nrrg.rach.contResTimer", confRachCmd.Flags().Lookup("contResTimer"))
	viper.BindPFlag("nrrg.rach.msg3Tp", confRachCmd.Flags().Lookup("msg3Tp"))
	viper.BindPFlag("nrrg.rach._raLen", confRachCmd.Flags().Lookup("_raLen"))
	viper.BindPFlag("nrrg.rach._raNumRbs", confRachCmd.Flags().Lookup("_raNumRbs"))
	viper.BindPFlag("nrrg.rach._raKBar", confRachCmd.Flags().Lookup("_raKBar"))
	confRachCmd.Flags().MarkHidden("_raFormat")
	confRachCmd.Flags().MarkHidden("_raX")
	confRachCmd.Flags().MarkHidden("_raY")
	confRachCmd.Flags().MarkHidden("_raSubfNumFr1SlotNumFr2")
	confRachCmd.Flags().MarkHidden("_raStartingSymb")
	confRachCmd.Flags().MarkHidden("_raNumSlotsPerSubfFr1Per60KSlotFr2")
	confRachCmd.Flags().MarkHidden("_raNumOccasionsPerSlot")
	confRachCmd.Flags().MarkHidden("_raDuration")
	confRachCmd.Flags().MarkHidden("_raLen")
	confRachCmd.Flags().MarkHidden("_raNumRbs")
	confRachCmd.Flags().MarkHidden("_raKBar")
}

func initConfDmrsCommonCmd() {
	confDmrsCommonCmd.Flags().StringSliceVar(&flags.dmrsCommon._schInfo, "_schInfo", []string{"SIB1", "Msg2", "Msg4", "Msg3"}, "Information of UL/DL-SCH")
	confDmrsCommonCmd.Flags().StringSliceVar(&flags.dmrsCommon._dmrsType, "_dmrsType", []string{"type1", "type1", "type1", "type1"}, "dmrs-Type as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	confDmrsCommonCmd.Flags().StringSliceVar(&flags.dmrsCommon._dmrsAddPos, "_dmrsAddPos", []string{"pos0", "pos0", "pos0", "pos1"}, "dmrs-AdditionalPosition as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	confDmrsCommonCmd.Flags().StringSliceVar(&flags.dmrsCommon._maxLength, "_maxLength", []string{"len1", "len1", "len1", "len1"}, "maxLength as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	confDmrsCommonCmd.Flags().IntSliceVar(&flags.dmrsCommon._dmrsPorts, "_dmrsPorts", []int{1000, 1000, 1000, 0}, "DMRS antenna ports")
	confDmrsCommonCmd.Flags().IntSliceVar(&flags.dmrsCommon._cdmGroupsWoData, "_cdmGroupsWoData", []int{1, 1, 1, 2}, "CDM group(s) without data")
	confDmrsCommonCmd.Flags().IntSliceVar(&flags.dmrsCommon._numFrontLoadSymbs, "_numFrontLoadSymbs", []int{1, 1, 1, 1}, "Number of front-load DMRS symbols")
	// _tdL for SIB1/Msg2/Msg4 is underscore(_) separated
	// _tdL for Msg3 is underscore(_) separated if msg3FreqHop is disabled, otherwise, _tdL is semicolon(;) separated for each hop
	confDmrsCommonCmd.Flags().StringSliceVar(&flags.dmrsCommon._tdL, "_tdL", []string{"0", "0", "0", "0;0"}, "Time-domain locations for DMRS")
	// _fdK indicates REs per PRB for DMRS
	confDmrsCommonCmd.Flags().StringSliceVar(&flags.dmrsCommon._fdK, "_fdK", []string{"101010101010", "101010101010", "101010101010", "111111111111"}, "Frequency-domain locations of DMRS")
	confDmrsCommonCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dmrscommon._schInfo", confDmrsCommonCmd.Flags().Lookup("_schInfo"))
	viper.BindPFlag("nrrg.dmrscommon._dmrsType", confDmrsCommonCmd.Flags().Lookup("_dmrsType"))
	viper.BindPFlag("nrrg.dmrscommon._dmrsAddPos", confDmrsCommonCmd.Flags().Lookup("_dmrsAddPos"))
	viper.BindPFlag("nrrg.dmrscommon._maxLength", confDmrsCommonCmd.Flags().Lookup("_maxLength"))
	viper.BindPFlag("nrrg.dmrscommon._dmrsPorts", confDmrsCommonCmd.Flags().Lookup("_dmrsPorts"))
	viper.BindPFlag("nrrg.dmrscommon._cdmGroupsWoData", confDmrsCommonCmd.Flags().Lookup("_cdmGroupsWoData"))
	viper.BindPFlag("nrrg.dmrscommon._numFrontLoadSymbs", confDmrsCommonCmd.Flags().Lookup("_numFrontLoadSymbs"))
	viper.BindPFlag("nrrg.dmrscommon._tdL", confDmrsCommonCmd.Flags().Lookup("_tdL"))
	viper.BindPFlag("nrrg.dmrscommon._fdK", confDmrsCommonCmd.Flags().Lookup("_fdK"))
	confDmrsCommonCmd.Flags().MarkHidden("_schInfo")
	confDmrsCommonCmd.Flags().MarkHidden("_dmrsType")
	confDmrsCommonCmd.Flags().MarkHidden("_dmrsAddPos")
	confDmrsCommonCmd.Flags().MarkHidden("_maxLength")
	confDmrsCommonCmd.Flags().MarkHidden("_dmrsPorts")
	confDmrsCommonCmd.Flags().MarkHidden("_cdmGroupsWoData")
	confDmrsCommonCmd.Flags().MarkHidden("_numFrontLoadSymbs")
	confDmrsCommonCmd.Flags().MarkHidden("_tdL")
	confDmrsCommonCmd.Flags().MarkHidden("_fdK")
}

func initConfDmrsPdschCmd() {
	confDmrsPdschCmd.Flags().StringVar(&flags.dmrsPdsch.pdschDmrsType, "pdschDmrsType", "type1", "dmrs-Type as in DMRS-DownlinkConfig[type1,type2]")
	confDmrsPdschCmd.Flags().StringVar(&flags.dmrsPdsch.pdschDmrsAddPos, "pdschDmrsAddPos", "pos0", "dmrs-additionalPosition as in DMRS-DownlinkConfig[pos0,pos1,pos2,pos3]")
	confDmrsPdschCmd.Flags().StringVar(&flags.dmrsPdsch.pdschMaxLength, "pdschMaxLength", "len1", "maxLength as in DMRS-DownlinkConfig[len1,len2]")
	confDmrsPdschCmd.Flags().IntSliceVar(&flags.dmrsPdsch._dmrsPorts, "_dmrsPorts", []int{1000, 1001, 1002, 1003}, "DMRS antenna ports")
	confDmrsPdschCmd.Flags().IntVar(&flags.dmrsPdsch._cdmGroupsWoData, "_cdmGroupsWoData", 2, "CDM group(s) without data")
	confDmrsPdschCmd.Flags().IntVar(&flags.dmrsPdsch._numFrontLoadSymbs, "_numFrontLoadSymbs", 1, "Number of front-load DMRS symbols")
	// _tdL is underscore(_) separated
	confDmrsPdschCmd.Flags().StringVar(&flags.dmrsPdsch._tdL, "_tdL", "2", "Time-domain locations for DMRS")
	// _fdK indicates REs per PRB for DMRS
	confDmrsPdschCmd.Flags().StringVar(&flags.dmrsPdsch._fdK, "_fdK", "111111111111", "Frequency-domain locations of DMRS")
	confDmrsPdschCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dmrspdsch.pdschDmrsType", confDmrsPdschCmd.Flags().Lookup("pdschDmrsType"))
	viper.BindPFlag("nrrg.dmrspdsch.pdschDmrsAddPos", confDmrsPdschCmd.Flags().Lookup("pdschDmrsAddPos"))
	viper.BindPFlag("nrrg.dmrspdsch.pdschMaxLength", confDmrsPdschCmd.Flags().Lookup("pdschMaxLength"))
	viper.BindPFlag("nrrg.dmrspdsch._dmrsPorts", confDmrsPdschCmd.Flags().Lookup("_dmrsPorts"))
	viper.BindPFlag("nrrg.dmrspdsch._cdmGroupsWoData", confDmrsPdschCmd.Flags().Lookup("_cdmGroupsWoData"))
	viper.BindPFlag("nrrg.dmrspdsch._numFrontLoadSymbs", confDmrsPdschCmd.Flags().Lookup("_numFrontLoadSymbs"))
	viper.BindPFlag("nrrg.dmrspdsch._tdL", confDmrsPdschCmd.Flags().Lookup("_tdL"))
	viper.BindPFlag("nrrg.dmrspdsch._fdK", confDmrsPdschCmd.Flags().Lookup("_fdK"))
	confDmrsPdschCmd.Flags().MarkHidden("_dmrsPorts")
	confDmrsPdschCmd.Flags().MarkHidden("_cdmGroupsWoData")
	confDmrsPdschCmd.Flags().MarkHidden("_numFrontLoadSymbs")
	confDmrsPdschCmd.Flags().MarkHidden("_tdL")
	confDmrsPdschCmd.Flags().MarkHidden("_fdK")
}

func initConfPtrsPdschCmd() {
	confPtrsPdschCmd.Flags().BoolVar(&flags.ptrsPdsch.pdschPtrsEnabled, "pdschPtrsEnabled", true, "Enable PTRS of PDSCH[false,true]")
	confPtrsPdschCmd.Flags().IntVar(&flags.ptrsPdsch.pdschPtrsTimeDensity, "pdschPtrsTimeDensity", 1, "The L_PTRS deduced from timeDensity of PTRS-DownlinkConfig[1,2,4]")
	confPtrsPdschCmd.Flags().IntVar(&flags.ptrsPdsch.pdschPtrsFreqDensity, "pdschPtrsFreqDensity", 2, "The K_PTRS deduced from frequencyDensity of PTRS-DownlinkConfig[2,4]")
	confPtrsPdschCmd.Flags().StringVar(&flags.ptrsPdsch.pdschPtrsReOffset, "pdschPtrsReOffset", "offset00", "resourceElementOffset of PTRS-DownlinkConfig[offset00,offset01,offset10,offset11]")
	confPtrsPdschCmd.Flags().IntVar(&flags.ptrsPdsch._dmrsPorts, "_dmrsPorts", 1000, "Associated DMRS antenna port")
	confPtrsPdschCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.ptrspdsch.pdschPtrsEnabled", confPtrsPdschCmd.Flags().Lookup("pdschPtrsEnabled"))
	viper.BindPFlag("nrrg.ptrspdsch.pdschPtrsTimeDensity", confPtrsPdschCmd.Flags().Lookup("pdschPtrsTimeDensity"))
	viper.BindPFlag("nrrg.ptrspdsch.pdschPtrsFreqDensity", confPtrsPdschCmd.Flags().Lookup("pdschPtrsFreqDensity"))
	viper.BindPFlag("nrrg.ptrspdsch.pdschPtrsReOffset", confPtrsPdschCmd.Flags().Lookup("pdschPtrsReOffset"))
	viper.BindPFlag("nrrg.ptrspdsch._dmrsPorts", confPtrsPdschCmd.Flags().Lookup("_dmrsPorts"))
	confPtrsPdschCmd.Flags().MarkHidden("_dmrsPorts")
}

func initConfDmrsPuschCmd() {
	confDmrsPuschCmd.Flags().StringVar(&flags.dmrsPusch.puschDmrsType, "puschDmrsType", "type1", "dmrs-Type as in DMRS-UplinkConfig[type1,type2]")
	confDmrsPuschCmd.Flags().StringVar(&flags.dmrsPusch.puschDmrsAddPos, "puschDmrsAddPos", "pos0", "dmrs-additionalPosition as in DMRS-UplinkConfig[pos0,pos1,pos2,pos3]")
	confDmrsPuschCmd.Flags().StringVar(&flags.dmrsPusch.puschMaxLength, "puschMaxLength", "len1", "maxLength as in DMRS-UplinkConfig[len1,len2]")
	confDmrsPuschCmd.Flags().IntSliceVar(&flags.dmrsPusch._dmrsPorts, "_dmrsPorts", []int{0, 1}, "DMRS antenna ports")
	confDmrsPuschCmd.Flags().IntVar(&flags.dmrsPusch._cdmGroupsWoData, "_cdmGroupsWoData", 1, "CDM group(s) without data")
	confDmrsPuschCmd.Flags().IntVar(&flags.dmrsPusch._numFrontLoadSymbs, "_numFrontLoadSymbs", 1, "Number of front-load DMRS symbols")
	// _tdL is underscore(_) separated
	confDmrsPuschCmd.Flags().StringVar(&flags.dmrsPusch._tdL, "_tdL", "2", "Time-domain locations for DMRS")
	// _fdK indicates REs per PRB for DMRS
	confDmrsPuschCmd.Flags().StringVar(&flags.dmrsPusch._fdK, "_fdK", "101010101010", "Frequency-domain locations of DMRS")
	confDmrsPuschCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dmrspusch.puschDmrsType", confDmrsPuschCmd.Flags().Lookup("puschDmrsType"))
	viper.BindPFlag("nrrg.dmrspusch.puschDmrsAddPos", confDmrsPuschCmd.Flags().Lookup("puschDmrsAddPos"))
	viper.BindPFlag("nrrg.dmrspusch.puschMaxLength", confDmrsPuschCmd.Flags().Lookup("puschMaxLength"))
	viper.BindPFlag("nrrg.dmrspusch._dmrsPorts", confDmrsPuschCmd.Flags().Lookup("_dmrsPorts"))
	viper.BindPFlag("nrrg.dmrspusch._cdmGroupsWoData", confDmrsPuschCmd.Flags().Lookup("_cdmGroupsWoData"))
	viper.BindPFlag("nrrg.dmrspusch._numFrontLoadSymbs", confDmrsPuschCmd.Flags().Lookup("_numFrontLoadSymbs"))
	viper.BindPFlag("nrrg.dmrspusch._tdL", confDmrsPuschCmd.Flags().Lookup("_tdL"))
	viper.BindPFlag("nrrg.dmrspusch._fdK", confDmrsPuschCmd.Flags().Lookup("_fdK"))
	confDmrsPuschCmd.Flags().MarkHidden("_dmrsPorts")
	confDmrsPuschCmd.Flags().MarkHidden("_cdmGroupsWoData")
	confDmrsPuschCmd.Flags().MarkHidden("_numFrontLoadSymbs")
	confDmrsPuschCmd.Flags().MarkHidden("_tdL")
	confDmrsPuschCmd.Flags().MarkHidden("_fdK")
}

func initConfPtrsPuschCmd() {
	confPtrsPuschCmd.Flags().BoolVar(&flags.ptrsPusch.puschPtrsEnabled, "puschPtrsEnabled", true, "Enable PTRS of PDSCH[false,true]")
	confPtrsPuschCmd.Flags().IntVar(&flags.ptrsPusch.puschPtrsTimeDensity, "puschPtrsTimeDensity", 1, "The L_PTRS deduced from timeDensity of PTRS-UplinkConfig for CP-OFDM[1,2,4]")
	confPtrsPuschCmd.Flags().IntVar(&flags.ptrsPusch.puschPtrsFreqDensity, "puschPtrsFreqDensity", 2, "The K_PTRS deduced from frequencyDensity of PTRS-UplinkConfig for CP-OFDM[2,4]")
	confPtrsPuschCmd.Flags().StringVar(&flags.ptrsPusch.puschPtrsReOffset, "puschPtrsReOffset", "offset00", "resourceElementOffset of PTRS-UplinkConfig for CP-OFDM[offset00,offset01,offset10,offset11]")
	confPtrsPuschCmd.Flags().StringVar(&flags.ptrsPusch.puschPtrsMaxNumPorts, "puschPtrsMaxNumPorts", "n1", "maxNrofPorts of PTRS-UplinkConfig for CP-OFDM[n1,n2]")
	confPtrsPuschCmd.Flags().IntSliceVar(&flags.ptrsPusch._dmrsPorts, "_dmrsPorts", []int{0}, "Associated DMRS antenna ports for CP-OFDM")
	confPtrsPuschCmd.Flags().IntVar(&flags.ptrsPusch.puschPtrsTimeDensityTp, "puschPtrsTimeDensityTp", 1, "The L_PTRS deduced from timeDensityTransformPrecoding of PTRS-UplinkConfig for DFS-S-OFDM[1,2]")
	confPtrsPuschCmd.Flags().StringVar(&flags.ptrsPusch.puschPtrsGrpPatternTp, "puschPtrsGrpPatternTp", "p0", "The Scheduled-bandwidth column index of 3GPP TS 38.214 Table 6.2.3.2-1, deduced from sampleDensity of PTRS-UplinkConfig for DFS-S-OFDM[p0,p1,p2,p3,p4]")
	confPtrsPuschCmd.Flags().IntVar(&flags.ptrsPusch._numGrpsTp, "_numGrpsTp", 2, "The Number-of-PT-RS-groups of 3GPP TS 38.214 Table 6.2.3.2-1, deduced from sampleDensity of PTRS-UplinkConfig for DFS-S-OFDM")
	confPtrsPuschCmd.Flags().IntVar(&flags.ptrsPusch._samplesPerGrpTp, "_samplesPerGrpTp", 2, "The Number-of-samples-per-PT-RS-group of 3GPP TS 38.214 Table 6.2.3.2-1, deduced from sampleDensity of PTRS-UplinkConfig for DFS-S-OFDM")
	confPtrsPuschCmd.Flags().IntSliceVar(&flags.ptrsPusch._dmrsPortsTp, "_dmrsPortsTp", []int{}, "Associated DMRS antenna ports for DFT-S-OFDM")
	confPtrsPuschCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.ptrspusch.puschPtrsEnabled", confPtrsPuschCmd.Flags().Lookup("puschPtrsEnabled"))
	viper.BindPFlag("nrrg.ptrspusch.puschPtrsTimeDensity", confPtrsPuschCmd.Flags().Lookup("puschPtrsTimeDensity"))
	viper.BindPFlag("nrrg.ptrspusch.puschPtrsFreqDensity", confPtrsPuschCmd.Flags().Lookup("puschPtrsFreqDensity"))
	viper.BindPFlag("nrrg.ptrspusch.puschPtrsReOffset", confPtrsPuschCmd.Flags().Lookup("puschPtrsReOffset"))
	viper.BindPFlag("nrrg.ptrspusch.puschPtrsMaxNumPorts", confPtrsPuschCmd.Flags().Lookup("puschPtrsMaxNumPorts"))
	viper.BindPFlag("nrrg.ptrspusch._dmrsPorts", confPtrsPuschCmd.Flags().Lookup("_dmrsPorts"))
	viper.BindPFlag("nrrg.ptrspusch.puschPtrsTimeDensityTp", confPtrsPuschCmd.Flags().Lookup("puschPtrsTimeDensityTp"))
	viper.BindPFlag("nrrg.ptrspusch.puschPtrsGrpPatternTp", confPtrsPuschCmd.Flags().Lookup("puschPtrsGrpPatternTp"))
	viper.BindPFlag("nrrg.ptrspusch._numGrpsTp", confPtrsPuschCmd.Flags().Lookup("_numGrpsTp"))
	viper.BindPFlag("nrrg.ptrspusch._samplesPerGrpTp", confPtrsPuschCmd.Flags().Lookup("_samplesPerGrpTp"))
	viper.BindPFlag("nrrg.ptrspusch._dmrsPortsTp", confPtrsPuschCmd.Flags().Lookup("_dmrsPortsTp"))
	confPtrsPuschCmd.Flags().MarkHidden("_dmrsPorts")
	confPtrsPuschCmd.Flags().MarkHidden("_numGrpsTp")
	confPtrsPuschCmd.Flags().MarkHidden("_samplesPerGrpTp")
	confPtrsPuschCmd.Flags().MarkHidden("_dmrsPortsTp")
}

func initConfPdschCmd() {
	confPdschCmd.Flags().StringVar(&flags.pdsch.pdschAggFactor, "pdschAggFactor", "n1", "pdsch-AggregationFactor of PDSCH-Config[n1,n2,n4,n8]")
	confPdschCmd.Flags().StringVar(&flags.pdsch.pdschRbgCfg, "pdschRbgCfg", "config1", "rbg-Size of PDSCH-Config[config1,config2]")
	confPdschCmd.Flags().IntVar(&flags.pdsch._rbgSize, "_rbgSize", 16, "RBG size of PDSCH resource allocation type 0")
	confPdschCmd.Flags().StringVar(&flags.pdsch.pdschMcsTable, "pdschMcsTable", "qam256", "mcs-Table of PDSCH-Config[qam64,qam256,qam64LowSE]")
	confPdschCmd.Flags().StringVar(&flags.pdsch.pdschXOh, "pdschXOh", "xOh0", "xOverhead of PDSCH-ServingCellConfig[xOh0,xOh6,xOh12,xOh18]")
	confPdschCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.pdsch.pdschAggFactor", confPdschCmd.Flags().Lookup("pdschAggFactor"))
	viper.BindPFlag("nrrg.pdsch.pdschRbgCfg", confPdschCmd.Flags().Lookup("pdschRbgCfg"))
	viper.BindPFlag("nrrg.pdsch._rbgSize", confPdschCmd.Flags().Lookup("_rbgSize"))
	viper.BindPFlag("nrrg.pdsch.pdschMcsTable", confPdschCmd.Flags().Lookup("pdschMcsTable"))
	viper.BindPFlag("nrrg.pdsch.pdschXOh", confPdschCmd.Flags().Lookup("pdschXOh"))
	confPdschCmd.Flags().MarkHidden("_rbgSize")
}

func initConfPuschCmd() {
	confPuschCmd.Flags().StringVar(&flags.pusch.puschTxCfg, "puschTxCfg", "codebook", "txConfig of PUSCH-Config[codebook,nonCodebook]")
	confPuschCmd.Flags().StringVar(&flags.pusch.puschCbSubset, "puschCbSubset", "fullyAndPartialAndNonCoherent", "codebookSubset of PUSCH-Config[fullyAndPartialAndNonCoherent,partialAndNonCoherent,nonCoherent]")
	confPuschCmd.Flags().IntVar(&flags.pusch.puschCbMaxRankNonCbMaxLayers, "puschCbMaxRankNonCbMaxLayers", 2, "maxRank of PUSCH-Config or maxMIMO-Layers of PUSCH-ServingCellConfig[1..4]")
	confPuschCmd.Flags().IntVar(&flags.pusch.puschFreqHopOffset, "puschFreqHopOffset", 0, "frequencyHoppingOffsetLists of PUSCH-Config[0..274]")
	confPuschCmd.Flags().StringVar(&flags.pusch.puschTp, "puschTp", "disabled", "transformPrecoder of PUSCH-Config[disabled,enabled]")
	confPuschCmd.Flags().StringVar(&flags.pusch.puschAggFactor, "puschAggFactor", "n1", "pusch-AggregationFactor of PUSCH-Config[n1,n2,n4,n8]")
	confPuschCmd.Flags().StringVar(&flags.pusch.puschRbgCfg, "puschRbgCfg", "config1", "rbg-Size of PUSCH-Config[config1,config2]")
	confPuschCmd.Flags().IntVar(&flags.pusch._rbgSize, "_rbgSize", 16, "RBG size of PUSCH resource allocation type 0")
	confPuschCmd.Flags().StringVar(&flags.pusch.puschMcsTable, "puschMcsTable", "qam64", "mcs-Table of PUSCH-Config[qam64,qam256,qam64LowSE]")
	confPuschCmd.Flags().StringVar(&flags.pusch.puschXOh, "puschXOh", "xOh0", "xOverhead of PUSCH-ServingCellConfig[xOh0,xOh6,xOh12,xOh18]")
	confPuschCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.pusch.puschTxCfg", confPuschCmd.Flags().Lookup("puschTxCfg"))
	viper.BindPFlag("nrrg.pusch.puschCbSubset", confPuschCmd.Flags().Lookup("puschCbSubset"))
	viper.BindPFlag("nrrg.pusch.puschCbMaxRankNonCbMaxLayers", confPuschCmd.Flags().Lookup("puschCbMaxRankNonCbMaxLayers"))
	viper.BindPFlag("nrrg.pusch.puschFreqHopOffset", confPuschCmd.Flags().Lookup("puschFreqHopOffset"))
	viper.BindPFlag("nrrg.pusch.puschTp", confPuschCmd.Flags().Lookup("puschTp"))
	viper.BindPFlag("nrrg.pusch.puschAggFactor", confPuschCmd.Flags().Lookup("puschAggFactor"))
	viper.BindPFlag("nrrg.pusch.puschRbgCfg", confPuschCmd.Flags().Lookup("puschRbgCfg"))
	viper.BindPFlag("nrrg.pusch._rbgSize", confPuschCmd.Flags().Lookup("_rbgSize"))
	viper.BindPFlag("nrrg.pusch.puschMcsTable", confPuschCmd.Flags().Lookup("puschMcsTable"))
	viper.BindPFlag("nrrg.pusch.puschXOh", confPuschCmd.Flags().Lookup("puschXOh"))
	confPuschCmd.Flags().MarkHidden("_rbgSize")
}

func initConfNzpCsiRsCmd() {
	confNzpCsiRsCmd.Flags().IntVar(&flags.nzpCsiRs._resSetId, "_resSetId", 0, "nzp-CSI-ResourceSetId of NZP-CSI-RS-ResourceSet")
	confNzpCsiRsCmd.Flags().BoolVar(&flags.nzpCsiRs._trsInfo, "_trsInfo", false, "trs-Info of NZP-CSI-RS-ResourceSet")
	confNzpCsiRsCmd.Flags().IntVar(&flags.nzpCsiRs._resId, "_resId", 0, "nzp-CSI-RS-ResourceId of NZP-CSI-RS-Resource")
	confNzpCsiRsCmd.Flags().StringVar(&flags.nzpCsiRs.nzpCsiRsFreqAllocRow, "nzpCsiRsFreqAllocRow", "row4", "The row of frequencyDomainAllocation of CSI-RS-ResourceMapping[row1,row2,row4,other]")
	confNzpCsiRsCmd.Flags().StringVar(&flags.nzpCsiRs.nzpCsiRsFreqAllocBits, "nzpCsiRsFreqAllocBits", "001", "The bit-string of frequencyDomainAllocation of CSI-RS-ResourceMapping")
	confNzpCsiRsCmd.Flags().StringVar(&flags.nzpCsiRs.nzpCsiRsNumPorts, "nzpCsiRsNumPorts", "p4", "nrofPorts of CSI-RS-ResourceMapping[p1,p2,p4,p8,p12,p16,p24,p32]")
	confNzpCsiRsCmd.Flags().StringVar(&flags.nzpCsiRs.nzpCsiRsCdmType, "nzpCsiRsCdmType", "fd-CDM2", "cdm-Type of CSI-RS-ResourceMapping[noCDM,fd-CDM2,cdm4-FD2-TD2,cdm8-FD2-TD4]")
	confNzpCsiRsCmd.Flags().StringVar(&flags.nzpCsiRs.nzpCsiRsDensity, "nzpCsiRsDensity", "one", "density of CSI-RS-ResourceMapping[evenPRBs,oddPRBs,one,three]")
	confNzpCsiRsCmd.Flags().IntVar(&flags.nzpCsiRs.nzpCsiRsFirstSymb, "nzpCsiRsFirstSymb", 1, "firstOFDMSymbolInTimeDomain of CSI-RS-ResourceMapping[0..13]")
	confNzpCsiRsCmd.Flags().IntVar(&flags.nzpCsiRs.nzpCsiRsFirstSymb2, "nzpCsiRsFirstSymb2", -1, "firstOFDMSymbolInTimeDomain2 of CSI-RS-ResourceMapping[-1 or 0..13]")
	confNzpCsiRsCmd.Flags().IntVar(&flags.nzpCsiRs.nzpCsiRsStartRb, "nzpCsiRsStartRb", 0, "startingRB of CSI-FrequencyOccupation[0..274]")
	confNzpCsiRsCmd.Flags().IntVar(&flags.nzpCsiRs.nzpCsiRsNumRbs, "nzpCsiRsNumRbs", 276, "nrofRBs of CSI-FrequencyOccupation[24..276]")
	confNzpCsiRsCmd.Flags().StringVar(&flags.nzpCsiRs.nzpCsiRsPeriod, "nzpCsiRsPeriod", "slots20", "periodicityAndOffset of NZP-CSI-RS-Resource[slots4,slots5,slots8,slots10,slots16,slots20,slots32,slots40,slots64,slots80,slots160,slots320,slots640]")
	confNzpCsiRsCmd.Flags().IntVar(&flags.nzpCsiRs.nzpCsiRsOffset, "nzpCsiRsOffset", 10, "periodicityAndOffset of NZP-CSI-RS-Resource[0..period-1]")
	confNzpCsiRsCmd.Flags().IntVar(&flags.nzpCsiRs._row, "_row", 4, "The Row of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().StringSliceVar(&flags.nzpCsiRs._kBarLBar, "_kBarLBar", []string{"0_0", "2_0"}, "The constants deduced from (k_bar, l_bar) of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().IntSliceVar(&flags.nzpCsiRs._ki, "_ki", []int{0, 0}, "The index ki deduced from (k_bar, l_bar) of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().IntSliceVar(&flags.nzpCsiRs._li, "_li", []int{0, 0}, "The index li deduced from (k_bar, l_bar) of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().IntSliceVar(&flags.nzpCsiRs._cdmGrpIndj, "_cdmGrpIndj", []int{0, 1}, "The CDM-group-index-j of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().IntSliceVar(&flags.nzpCsiRs._kap, "_kap", []int{0, 1}, "The k_ap of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().IntSliceVar(&flags.nzpCsiRs._lap, "_lap", []int{0}, "The l_ap of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.nzpcsirs._resSetId", confNzpCsiRsCmd.Flags().Lookup("_resSetId"))
	viper.BindPFlag("nrrg.nzpcsirs._trsInfo", confNzpCsiRsCmd.Flags().Lookup("_trsInfo"))
	viper.BindPFlag("nrrg.nzpcsirs._resId", confNzpCsiRsCmd.Flags().Lookup("_resId"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsFreqAllocRow", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsFreqAllocRow"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsFreqAllocBits", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsFreqAllocBits"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsNumPorts", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsNumPorts"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsCdmType", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsCdmType"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsDensity", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsDensity"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsFirstSymb", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsFirstSymb"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsFirstSymb2", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsFirstSymb2"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsStartRb", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsStartRb"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsNumRbs", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsNumRbs"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsPeriod", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsPeriod"))
	viper.BindPFlag("nrrg.nzpcsirs.nzpCsiRsOffset", confNzpCsiRsCmd.Flags().Lookup("nzpCsiRsOffset"))
	viper.BindPFlag("nrrg.nzpcsirs._row", confNzpCsiRsCmd.Flags().Lookup("_row"))
	viper.BindPFlag("nrrg.nzpcsirs._kBarLBar", confNzpCsiRsCmd.Flags().Lookup("_kBarLBar"))
	viper.BindPFlag("nrrg.nzpcsirs._ki", confNzpCsiRsCmd.Flags().Lookup("_ki"))
	viper.BindPFlag("nrrg.nzpcsirs._li", confNzpCsiRsCmd.Flags().Lookup("_li"))
	viper.BindPFlag("nrrg.nzpcsirs._cdmGrpIndj", confNzpCsiRsCmd.Flags().Lookup("_cdmGrpIndj"))
	viper.BindPFlag("nrrg.nzpcsirs._kap", confNzpCsiRsCmd.Flags().Lookup("_kap"))
	viper.BindPFlag("nrrg.nzpcsirs._lap", confNzpCsiRsCmd.Flags().Lookup("_lap"))
	confNzpCsiRsCmd.Flags().MarkHidden("_resSetId")
	confNzpCsiRsCmd.Flags().MarkHidden("_trsInfo")
	confNzpCsiRsCmd.Flags().MarkHidden("_resId")
	confNzpCsiRsCmd.Flags().MarkHidden("_row")
	confNzpCsiRsCmd.Flags().MarkHidden("_kBarLBar")
	confNzpCsiRsCmd.Flags().MarkHidden("_ki")
	confNzpCsiRsCmd.Flags().MarkHidden("_li")
	confNzpCsiRsCmd.Flags().MarkHidden("_cdmGrpIndj")
	confNzpCsiRsCmd.Flags().MarkHidden("_kap")
	confNzpCsiRsCmd.Flags().MarkHidden("_lap")
}

func initConfTrsCmd() {
	confTrsCmd.Flags().IntVar(&flags.trs._resSetId, "_resSetId", 1, "nzp-CSI-ResourceSetId of NZP-CSI-RS-ResourceSet")
	confTrsCmd.Flags().BoolVar(&flags.trs._trsInfo, "_trsInfo", true, "trs-Info of NZP-CSI-RS-ResourceSet")
	confTrsCmd.Flags().IntVar(&flags.trs._firstResId, "_firstResId", 100, "nzp-CSI-RS-ResourceId of NZP-CSI-RS-Resource for the first TRS resource")
	confTrsCmd.Flags().StringVar(&flags.trs._freqAllocRow, "_freqAllocRow", "row1", "The row of frequencyDomainAllocation of CSI-RS-ResourceMapping[row1,row2,row4,other]")
	confTrsCmd.Flags().StringVar(&flags.trs.trsFreqAllocBits, "trsFreqAllocBits", "0001", "The bit-string of frequencyDomainAllocation of CSI-RS-ResourceMapping")
	confTrsCmd.Flags().StringVar(&flags.trs._numPorts, "_numPorts", "p1", "nrofPorts of CSI-RS-ResourceMapping[p1,p2,p4,p8,p12,p16,p24,p32]")
	confTrsCmd.Flags().StringVar(&flags.trs._cdmType, "_cdmType", "noCDM", "cdm-Type of CSI-RS-ResourceMapping[noCDM,fd-CDM2,cdm4-FD2-TD2,cdm8-FD2-TD4]")
	confTrsCmd.Flags().StringVar(&flags.trs._density, "_density", "three", "density of CSI-RS-ResourceMapping[evenPRBs,oddPRBs,one,three]")
	confTrsCmd.Flags().IntSliceVar(&flags.trs.trsFirstSymbs, "trsFirstSymbs", []int{5, 9}, "firstOFDMSymbolInTimeDomain of CSI-RS-ResourceMapping for the two TRS resources in one slot[0..13]")
	confTrsCmd.Flags().IntVar(&flags.trs.trsStartRb, "trsStartRb", 0, "startingRB of CSI-FrequencyOccupation[0..274]")
	confTrsCmd.Flags().IntVar(&flags.trs.trsNumRbs, "trsNumRbs", 276, "nrofRBs of CSI-FrequencyOccupation[24..276]")
	confTrsCmd.Flags().StringVar(&flags.trs.trsPeriod, "trsPeriod", "slots40", "periodicityAndOffset of NZP-CSI-RS-Resource[slots10,slots20,slots40,slots80,slots160,slots320,slots640]")
	confTrsCmd.Flags().IntSliceVar(&flags.trs.trsOffset, "trsOffset", []int{10}, "periodicityAndOffset of NZP-CSI-RS-Resource for at most two consecutive slots[0..period-1]")
	confTrsCmd.Flags().IntVar(&flags.trs._row, "_row", 1, "The Row of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confTrsCmd.Flags().StringSliceVar(&flags.trs._kBarLBar, "_kBarLBar", []string{"0_0", "4_0", "8_0"}, "The constants deduced from (k_bar, l_bar) of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confTrsCmd.Flags().IntSliceVar(&flags.trs._ki, "_ki", []int{0, 0, 0}, "The index ki deduced from (k_bar, l_bar) of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confTrsCmd.Flags().IntSliceVar(&flags.trs._li, "_li", []int{0, 0, 0}, "The index li deduced from (k_bar, l_bar) of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confTrsCmd.Flags().IntSliceVar(&flags.trs._cdmGrpIndj, "_cdmGrpIndj", []int{0, 0, 0}, "The CDM-group-index-j of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confTrsCmd.Flags().IntSliceVar(&flags.trs._kap, "_kap", []int{0}, "The k_ap of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confTrsCmd.Flags().IntSliceVar(&flags.trs._lap, "_lap", []int{0}, "The l_ap of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confTrsCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.trs._resSetId", confTrsCmd.Flags().Lookup("_resSetId"))
	viper.BindPFlag("nrrg.trs._trsInfo", confTrsCmd.Flags().Lookup("_trsInfo"))
	viper.BindPFlag("nrrg.trs._firstResId", confTrsCmd.Flags().Lookup("_firstResId"))
	viper.BindPFlag("nrrg.trs._freqAllocRow", confTrsCmd.Flags().Lookup("_freqAllocRow"))
	viper.BindPFlag("nrrg.trs.trsFreqAllocBits", confTrsCmd.Flags().Lookup("trsFreqAllocBits"))
	viper.BindPFlag("nrrg.trs._numPorts", confTrsCmd.Flags().Lookup("_numPorts"))
	viper.BindPFlag("nrrg.trs._cdmType", confTrsCmd.Flags().Lookup("_cdmType"))
	viper.BindPFlag("nrrg.trs._density", confTrsCmd.Flags().Lookup("_density"))
	viper.BindPFlag("nrrg.trs.trsFirstSymbs", confTrsCmd.Flags().Lookup("trsFirstSymbs"))
	viper.BindPFlag("nrrg.trs.trsStartRb", confTrsCmd.Flags().Lookup("trsStartRb"))
	viper.BindPFlag("nrrg.trs.trsNumRbs", confTrsCmd.Flags().Lookup("trsNumRbs"))
	viper.BindPFlag("nrrg.trs.trsPeriod", confTrsCmd.Flags().Lookup("trsPeriod"))
	viper.BindPFlag("nrrg.trs.trsOffset", confTrsCmd.Flags().Lookup("trsOffset"))
	viper.BindPFlag("nrrg.trs._row", confTrsCmd.Flags().Lookup("_row"))
	viper.BindPFlag("nrrg.trs._kBarLBar", confTrsCmd.Flags().Lookup("_kBarLBar"))
	viper.BindPFlag("nrrg.trs._ki", confTrsCmd.Flags().Lookup("_ki"))
	viper.BindPFlag("nrrg.trs._li", confTrsCmd.Flags().Lookup("_li"))
	viper.BindPFlag("nrrg.trs._cdmGrpIndj", confTrsCmd.Flags().Lookup("_cdmGrpIndj"))
	viper.BindPFlag("nrrg.trs._kap", confTrsCmd.Flags().Lookup("_kap"))
	viper.BindPFlag("nrrg.trs._lap", confTrsCmd.Flags().Lookup("_lap"))
	confTrsCmd.Flags().MarkHidden("_resSetId")
	confTrsCmd.Flags().MarkHidden("_trsInfo")
	confTrsCmd.Flags().MarkHidden("_firstResId")
	confTrsCmd.Flags().MarkHidden("_freqAllocRow")
	confTrsCmd.Flags().MarkHidden("_numPorts")
	confTrsCmd.Flags().MarkHidden("_cdmType")
	confTrsCmd.Flags().MarkHidden("_density")
	confTrsCmd.Flags().MarkHidden("_row")
	confTrsCmd.Flags().MarkHidden("_kBarLBar")
	confTrsCmd.Flags().MarkHidden("_ki")
	confTrsCmd.Flags().MarkHidden("_li")
	confTrsCmd.Flags().MarkHidden("_cdmGrpIndj")
	confTrsCmd.Flags().MarkHidden("_kap")
	confTrsCmd.Flags().MarkHidden("_lap")
}

func initConfCsiImCmd() {
	confCsiImCmd.Flags().IntVar(&flags.csiIm._resSetId, "_resSetId", 0, "csi-IM-ResourceSetId of CSI-IM-ResourceSet")
	confCsiImCmd.Flags().IntVar(&flags.csiIm._resId, "_resId", 0, "csi-IM-ResourceId of CSI-IM-Resource")
	confCsiImCmd.Flags().StringVar(&flags.csiIm.csiImRePattern, "csiImRePattern", "pattern0", "csi-IM-ResourceElementPattern of CSI-IM-Resource[pattern0,pattern1]")
	confCsiImCmd.Flags().StringVar(&flags.csiIm.csiImScLoc, "csiImScLoc", "s8", "subcarrierLocation of csi-IM-ResourceElementPattern of CSI-IM-Resource[s0,s2,s4,s6,s8,s10]")
	confCsiImCmd.Flags().IntVar(&flags.csiIm.csiImSymbLoc, "csiImSymbLoc", 1, "symbolLocation of csi-IM-ResourceElementPattern of CSI-IM-Resource[0..12]")
	confCsiImCmd.Flags().IntVar(&flags.csiIm.csiImStartRb, "csiImStartRb", 0, "startingRB of CSI-FrequencyOccupation[0..274]")
	confCsiImCmd.Flags().IntVar(&flags.csiIm.csiImNumRbs, "csiImNumRbs", 276, "nrofRBs of CSI-FrequencyOccupation[24..276]")
	confCsiImCmd.Flags().StringVar(&flags.csiIm.csiImPeriod, "csiImPeriod", "slots20", "periodicityAndOffset of CSI-IM-Resource[slots4,slots5,slots8,slots10,slots16,slots20,slots32,slots40,slots64,slots80,slots160,slots320,slots640]")
	confCsiImCmd.Flags().IntVar(&flags.csiIm.csiImOffset, "csiImOffset", 10, "periodicityAndOffset of CSI-IM-Resource[0..period-1]")
	confCsiImCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.csiim._resSetId", confCsiImCmd.Flags().Lookup("_resSetId"))
	viper.BindPFlag("nrrg.csiim._resId", confCsiImCmd.Flags().Lookup("_resId"))
	viper.BindPFlag("nrrg.csiim.csiImRePattern", confCsiImCmd.Flags().Lookup("csiImRePattern"))
	viper.BindPFlag("nrrg.csiim.csiImScLoc", confCsiImCmd.Flags().Lookup("csiImScLoc"))
	viper.BindPFlag("nrrg.csiim.csiImSymbLoc", confCsiImCmd.Flags().Lookup("csiImSymbLoc"))
	viper.BindPFlag("nrrg.csiim.csiImStartRb", confCsiImCmd.Flags().Lookup("csiImStartRb"))
	viper.BindPFlag("nrrg.csiim.csiImNumRbs", confCsiImCmd.Flags().Lookup("csiImNumRbs"))
	viper.BindPFlag("nrrg.csiim.csiImPeriod", confCsiImCmd.Flags().Lookup("csiImPeriod"))
	viper.BindPFlag("nrrg.csiim.csiImOffset", confCsiImCmd.Flags().Lookup("csiImOffset"))
	confCsiImCmd.Flags().MarkHidden("_resSetId")
	confCsiImCmd.Flags().MarkHidden("_resId")
}

func initConfCsiReportCmd() {
	confCsiReportCmd.Flags().StringSliceVar(&flags.csiReport._resCfgType, "_resCfgType", []string{"NZP-CSI-RS", "CSI-IM", "TRS"}, "type of CSI-ResourceConfig")
	confCsiReportCmd.Flags().IntSliceVar(&flags.csiReport._resCfgId, "_resCfgId", []int{0, 10, 20}, "csi-ResourceConfigId of CSI-ResourceConfig")
	confCsiReportCmd.Flags().IntSliceVar(&flags.csiReport._resSetId, "_resSetId", []int{0, 0, 1}, "csi-RS-ResourceSetList of CSI-ResourceConfig")
	confCsiReportCmd.Flags().IntSliceVar(&flags.csiReport._resBwpId, "_resBwpId", []int{1, 1, 1}, "bwp-Id of CSI-ResourceConfig")
	confCsiReportCmd.Flags().StringSliceVar(&flags.csiReport._resType, "_resType", []string{"periodic", "periodic", "periodic"}, "resourceType of CSI-ResourceConfig")
	confCsiReportCmd.Flags().IntVar(&flags.csiReport._repCfgId, "_repCfgId", 0, "reportConfigId of CSI-ReportConfig")
	confCsiReportCmd.Flags().IntVar(&flags.csiReport._resCfgIdChnMeas, "_resCfgIdChnMeas", 0, "resourcesForChannelMeasurement of CSI-ReportConfig")
	confCsiReportCmd.Flags().IntVar(&flags.csiReport._resCfgIdCsiImIntf, "_resCfgIdCsiImIntf", 10, "csi-IM-ResourcesForInterference of CSI-ReportConfig")
	confCsiReportCmd.Flags().StringVar(&flags.csiReport._repCfgType, "_repCfgType", "periodic", "reportConfigType of CSI-ReportConfig")
	confCsiReportCmd.Flags().StringVar(&flags.csiReport.csiRepPeriod, "csiRepPeriod", "slots320", "CSI-ReportPeriodicityAndOffset of CSI-ReportConfig[slots4,slots5,slots8,slots10,slots16,slots20,slots40,slots80,slots160,slots320]")
	confCsiReportCmd.Flags().IntVar(&flags.csiReport.csiRepOffset, "csiRepOffset", 7, "CSI-ReportPeriodicityAndOffset of CSI-ReportConfig[0..period-1]")
	confCsiReportCmd.Flags().IntVar(&flags.csiReport._ulBwpId, "_ulBwpId", 1, "uplinkBandwidthPartId of PUCCH-CSI-Resource of CSI-ReportConfig")
	confCsiReportCmd.Flags().IntVar(&flags.csiReport.csiRepPucchRes, "csiRepPucchRes", 2, "pucch-Resource of PUCCH-CSI-Resource of CSI-ReportConfig[2,3,4]")
	confCsiReportCmd.Flags().StringVar(&flags.csiReport._quantity, "_quantity", "cri-RI-PMI-CQI", "reportQuantity of CSI-ReportConfig")
	confCsiReportCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.csireport._resCfgType", confCsiReportCmd.Flags().Lookup("_resCfgType"))
	viper.BindPFlag("nrrg.csireport._resCfgId", confCsiReportCmd.Flags().Lookup("_resCfgId"))
	viper.BindPFlag("nrrg.csireport._resSetId", confCsiReportCmd.Flags().Lookup("_resSetId"))
	viper.BindPFlag("nrrg.csireport._resBwpId", confCsiReportCmd.Flags().Lookup("_resBwpId"))
	viper.BindPFlag("nrrg.csireport._resType", confCsiReportCmd.Flags().Lookup("_resType"))
	viper.BindPFlag("nrrg.csireport._repCfgId", confCsiReportCmd.Flags().Lookup("_repCfgId"))
	viper.BindPFlag("nrrg.csireport._resCfgIdChnMeas", confCsiReportCmd.Flags().Lookup("_resCfgIdChnMeas"))
	viper.BindPFlag("nrrg.csireport._resCfgIdCsiImIntf", confCsiReportCmd.Flags().Lookup("_resCfgIdCsiImIntf"))
	viper.BindPFlag("nrrg.csireport._repCfgType", confCsiReportCmd.Flags().Lookup("_repCfgType"))
	viper.BindPFlag("nrrg.csireport.csiRepPeriod", confCsiReportCmd.Flags().Lookup("csiRepPeriod"))
	viper.BindPFlag("nrrg.csireport.csiRepOffset", confCsiReportCmd.Flags().Lookup("csiRepOffset"))
	viper.BindPFlag("nrrg.csireport._ulBwpId", confCsiReportCmd.Flags().Lookup("_ulBwpId"))
	viper.BindPFlag("nrrg.csireport.csiRepPucchRes", confCsiReportCmd.Flags().Lookup("csiRepPucchRes"))
	viper.BindPFlag("nrrg.csireport._quantity", confCsiReportCmd.Flags().Lookup("_quantity"))
	confCsiReportCmd.Flags().MarkHidden("_resCfgType")
	confCsiReportCmd.Flags().MarkHidden("_resCfgId")
	confCsiReportCmd.Flags().MarkHidden("_resSetId")
	confCsiReportCmd.Flags().MarkHidden("_resBwpId")
	confCsiReportCmd.Flags().MarkHidden("_resType")
	confCsiReportCmd.Flags().MarkHidden("_repCfgId")
	confCsiReportCmd.Flags().MarkHidden("_resCfgIdChnMeas")
	confCsiReportCmd.Flags().MarkHidden("_resCfgIdCsiImIntf")
	confCsiReportCmd.Flags().MarkHidden("_repCfgType")
	confCsiReportCmd.Flags().MarkHidden("_ulBwpId")
	confCsiReportCmd.Flags().MarkHidden("_quantity")
}

func initConfSrsCmd() {
	confSrsCmd.Flags().IntSliceVar(&flags.srs._resId, "_resId", []int{0, 1, 2, 3}, "srs-ResourceId of SRS-Resource")
	confSrsCmd.Flags().StringSliceVar(&flags.srs.srsNumPorts, "srsNumPorts", []string{"ports2", "port1", "port1", "port1"}, "nrofSRS-Ports of SRS-Resource[port1,ports2,ports4]")
	confSrsCmd.Flags().StringSliceVar(&flags.srs.srsNonCbPtrsPort, "srsNonCbPtrsPort", []string{"n0", "n0", "n0", "n0"}, "ptrs-PortIndex of SRS-Resource[n0,n1]")
	confSrsCmd.Flags().StringSliceVar(&flags.srs.srsNumCombs, "srsNumCombs", []string{"n4", "n2", "n2", "n2"}, "transmissionComb of SRS-Resource[n2,n4]")
	confSrsCmd.Flags().IntSliceVar(&flags.srs.srsCombOff, "srsCombOff", []int{0, 0, 0, 0}, "combOffset of SRS-Resource")
	confSrsCmd.Flags().IntSliceVar(&flags.srs.srsCs, "srsCs", []int{11, 0, 0, 0}, "cyclicShift of SRS-Resource")
	confSrsCmd.Flags().IntSliceVar(&flags.srs.srsStartPos, "srsStartPos", []int{3, 0, 0, 0}, "startPosition of SRS-Resource[0..5]")
	confSrsCmd.Flags().StringSliceVar(&flags.srs.srsNumSymbs, "srsNumSymbs", []string{"n4", "n1", "n1", "n1"}, "nrofSymbols of SRS-Resource[n1,n2,n4]")
	confSrsCmd.Flags().StringSliceVar(&flags.srs.srsRepetition, "srsRepetition", []string{"n4", "n1", "n1", "n1"}, "repetitionFactor of SRS-Resource[n1,n2,n4]")
	confSrsCmd.Flags().IntSliceVar(&flags.srs.srsFreqPos, "srsFreqPos", []int{0, 0, 0, 0}, "freqDomainPosition of SRS-Resource[0..67]")
	confSrsCmd.Flags().IntSliceVar(&flags.srs.srsFreqShift, "srsFreqShift", []int{0, 0, 0, 0}, "freqDomainShift of SRS-Resource[0..268]")
	confSrsCmd.Flags().IntSliceVar(&flags.srs.srsCSrs, "srsCSrs", []int{12, 0, 0, 0}, "c-SRS of SRS-Resource[0..63]")
	confSrsCmd.Flags().IntSliceVar(&flags.srs.srsBSrs, "srsBSrs", []int{1, 0, 0, 0}, "b-SRS of SRS-Resource[0..3]")
	confSrsCmd.Flags().IntSliceVar(&flags.srs.srsBHop, "srsBHop", []int{0, 0, 0, 0}, "b-hop of SRS-Resource[0..3]")
	confSrsCmd.Flags().StringSliceVar(&flags.srs._type, "_type", []string{"periodic", "periodic", "periodic", "periodic"}, "resourceType of SRS-Resource")
	confSrsCmd.Flags().StringSliceVar(&flags.srs.srsPeriod, "srsPeriod", []string{"sl10", "sl5", "sl5", "sl5"}, "SRS-PeriodicityAndOffset of SRS-Resource[sl1,sl2,sl4,sl5,sl8,sl10,sl16,sl20,sl32,sl40,sl64,sl80,sl160,sl320,sl640,sl1280,sl2560]")
	confSrsCmd.Flags().IntSliceVar(&flags.srs.srsOffset, "srsOffset", []int{7,0,0,0}, "SRS-PeriodicityAndOffset of SRS-Resource[0..period-1]")
	confSrsCmd.Flags().StringSliceVar(&flags.srs._mSRSb, "_mSRSb", []string{"48_16_8_4", "4_4_4_4", "4_4_4_4", "4_4_4_4"}, "The m_SRS_b with b=B_SRS of 3GPP TS 38.211 Table 6.4.1.4.3-1")
	confSrsCmd.Flags().StringSliceVar(&flags.srs._Nb, "_Nb", []string{"1_3_2_2", "1_1_1_1", "1_1_1_1", "1_1_1_1"}, "The N_b with b=B_SRS of 3GPP TS 38.211 Table 6.4.1.4.3-1")
	confSrsCmd.Flags().IntSliceVar(&flags.srs._resSetId, "_resSetId", []int{0, 1}, "srs-ResourceSetId of SRS-ResourceSet")
	confSrsCmd.Flags().StringSliceVar(&flags.srs.srsSetResIdList, "srsSetResIdList", []string{"0", "0_1_2_3"}, "srs-ResourceIdList of SRS-ResourceSet")
	confSrsCmd.Flags().StringSliceVar(&flags.srs._resType, "_resType", []string{"periodic", "periodic"}, "resourceType of SRS-ResourceSet")
	confSrsCmd.Flags().StringSliceVar(&flags.srs._usage, "_usage", []string{"codebook", "nonCodebook"}, "usage of SRS-ResourceSet")
	confSrsCmd.Flags().StringSliceVar(&flags.srs._dci01NonCbSrsList, "_dci01NonCbSrsList", []string{"-", ""}, "The SRI(s) field of 3GPP TS 38.212 Table 7.3.1.1.2-28 ~ Table 7.3.1.1.2-31")
	confSrsCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.srs._resId", confSrsCmd.Flags().Lookup("_resId"))
	viper.BindPFlag("nrrg.srs.srsNumPorts", confSrsCmd.Flags().Lookup("srsNumPorts"))
	viper.BindPFlag("nrrg.srs.srsNonCbPtrsPort", confSrsCmd.Flags().Lookup("srsNonCbPtrsPort"))
	viper.BindPFlag("nrrg.srs.srsNumCombs", confSrsCmd.Flags().Lookup("srsNumCombs"))
	viper.BindPFlag("nrrg.srs.srsCombOff", confSrsCmd.Flags().Lookup("srsCombOff"))
	viper.BindPFlag("nrrg.srs.srsCs", confSrsCmd.Flags().Lookup("srsCs"))
	viper.BindPFlag("nrrg.srs.srsStartPos", confSrsCmd.Flags().Lookup("srsStartPos"))
	viper.BindPFlag("nrrg.srs.srsNumSymbs", confSrsCmd.Flags().Lookup("srsNumSymbs"))
	viper.BindPFlag("nrrg.srs.srsRepetition", confSrsCmd.Flags().Lookup("srsRepetition"))
	viper.BindPFlag("nrrg.srs.srsFreqPos", confSrsCmd.Flags().Lookup("srsFreqPos"))
	viper.BindPFlag("nrrg.srs.srsFreqShift", confSrsCmd.Flags().Lookup("srsFreqShift"))
	viper.BindPFlag("nrrg.srs.srsCSrs", confSrsCmd.Flags().Lookup("srsCSrs"))
	viper.BindPFlag("nrrg.srs.srsBSrs", confSrsCmd.Flags().Lookup("srsBSrs"))
	viper.BindPFlag("nrrg.srs.srsBHop", confSrsCmd.Flags().Lookup("srsBHop"))
	viper.BindPFlag("nrrg.srs._type", confSrsCmd.Flags().Lookup("_type"))
	viper.BindPFlag("nrrg.srs.srsPeriod", confSrsCmd.Flags().Lookup("srsPeriod"))
	viper.BindPFlag("nrrg.srs.srsOffset", confSrsCmd.Flags().Lookup("srsOffset"))
	viper.BindPFlag("nrrg.srs._mSRSb", confSrsCmd.Flags().Lookup("_mSRSb"))
	viper.BindPFlag("nrrg.srs._Nb", confSrsCmd.Flags().Lookup("_Nb"))
	viper.BindPFlag("nrrg.srs._resSetId", confSrsCmd.Flags().Lookup("_resSetId"))
	viper.BindPFlag("nrrg.srs.srsSetResIdList", confSrsCmd.Flags().Lookup("srsSetResIdList"))
	viper.BindPFlag("nrrg.srs._resType", confSrsCmd.Flags().Lookup("_resType"))
	viper.BindPFlag("nrrg.srs._usage", confSrsCmd.Flags().Lookup("_usage"))
	viper.BindPFlag("nrrg.srs._dci01NonCbSrsList", confSrsCmd.Flags().Lookup("_dci01NonCbSrsList"))
	confSrsCmd.Flags().MarkHidden("_resId")
	confSrsCmd.Flags().MarkHidden("_type")
	confSrsCmd.Flags().MarkHidden("_mSRSb")
	confSrsCmd.Flags().MarkHidden("_Nb")
	confSrsCmd.Flags().MarkHidden("_resSetId")
	confSrsCmd.Flags().MarkHidden("_resType")
	confSrsCmd.Flags().MarkHidden("_usage")
	confSrsCmd.Flags().MarkHidden("_dci01NonCbSrsList")
}

func initConfPucchCmd() {
	confPucchCmd.Flags().StringVar(&flags.pucch.pucchFmtCfgNumSlots, "pucchFmtCfgNumSlots", "n2", "nrofSlots of PUCCH-FormatConfig for PUCCH format 1/3/4[n1,n2,n4,n8]")
	confPucchCmd.Flags().StringVar(&flags.pucch.pucchFmtCfgInterSlotFreqHop, "pucchFmtCfgInterSlotFreqHop", "disabled", "interslotFrequencyHopping of PUCCH-FormatConfig for PUCCH format 1/3/4[disabled,enabled]")
	confPucchCmd.Flags().BoolVar(&flags.pucch.pucchFmtCfgAddDmrs, "pucchFmtCfgAddDmrs", true, "additionalDMRS of PUCCH-FormatConfig for PUCCH format 3/4")
	confPucchCmd.Flags().BoolVar(&flags.pucch.pucchFmtCfgSimAckCsi, "pucchFmtCfgSimAckCsi", true, "simultaneousHARQ-ACK-CSI of PUCCH-FormatConfig for PUCCH format 2/3/4")
	confPucchCmd.Flags().IntSliceVar(&flags.pucch._pucchResId, "_pucchResId", []int{0, 1, 2, 3, 4}, "pucch-ResourceId of PUCCH-Resource")
	confPucchCmd.Flags().StringSliceVar(&flags.pucch._pucchFormat, "_pucchFormat", []string{"format0", "format1", "format2", "format3", "format4"}, "format of PUCCH-Resource")
	confPucchCmd.Flags().IntSliceVar(&flags.pucch._pucchResSetId, "_pucchResSetId", []int{0, 0, 1, 1, 1}, "pucch-ResourceSetId of PUCCH-ResourceSet")
	confPucchCmd.Flags().IntSliceVar(&flags.pucch.pucchStartRb, "pucchStartRb", []int{0, 0, 0, 0, 0}, "startingPRB of PUCCH-ResourceSet[0..274]")
	confPucchCmd.Flags().StringSliceVar(&flags.pucch.pucchIntraSlotFreqHop, "pucchIntraSlotFreqHop", []string{"enabled", "enabled", "disabled", "disabled", "disabled"}, "intraSlotFrequencyHopping of PUCCH-Resource[disabled,enabled]")
	confPucchCmd.Flags().IntSliceVar(&flags.pucch.pucchSecondHopPrb, "pucchSecondHopPrb", []int{272, 272, -1, -1, -1}, "secondHopPRB of PUCCH-Resource[0..274]")
	confPucchCmd.Flags().IntSliceVar(&flags.pucch.pucchNumRbs, "pucchNumRbs", []int{1, 1, 1, 1, 1}, "nrofPRBs of PUCCH-Resource, fixed to 1 for PUCCH format 0/1/4[1..16]")
	confPucchCmd.Flags().IntSliceVar(&flags.pucch.pucchStartSymb, "pucchStartSymb", []int{0, 0, 0, 0, 0}, "startingSymbolIndex of PUCCH-Resource[0..13(format 0/2) or 0..10(format 1/3/4)]")
	confPucchCmd.Flags().IntSliceVar(&flags.pucch.pucchNumSymbs, "pucchNumSymbs", []int{2, 4, 1, 4, 4}, "nrofSymbols of PUCCH-Resource[1..2(format 0/2) or 4..14(format 1/3/4)]")
	confPucchCmd.Flags().IntSliceVar(&flags.pucch._dsrResId, "_dsrResId", []int{0, 1}, "schedulingRequestResourceId of SchedulingRequestResourceConfig")
	confPucchCmd.Flags().IntSliceVar(&flags.pucch._dsrPucchRes, "_dsrPucchRes", []int{0, 1}, "resource of SchedulingRequestResourceConfig")
	confPucchCmd.Flags().StringSliceVar(&flags.pucch.dsrPeriod, "dsrPeriod", []string{"sl10", "sym6or7"}, "periodicityAndOffset of SchedulingRequestResourceConfig[sym2,sym6or7,sl1,sl2,sl4,sl5,sl8,sl10,sl16,sl20,sl40,sl80,sl160,sl320,sl640]")
	confPucchCmd.Flags().IntSliceVar(&flags.pucch.dsrOffset, "dsrOffset", []int{8, 0}, "periodicityAndOffset of SchedulingRequestResourceConfig[0..period-1]")
	confPucchCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.pucch.pucchFmtCfgNumSlots", confPucchCmd.Flags().Lookup("pucchFmtCfgNumSlots"))
	viper.BindPFlag("nrrg.pucch.pucchFmtCfgInterSlotFreqHop", confPucchCmd.Flags().Lookup("pucchFmtCfgInterSlotFreqHop"))
	viper.BindPFlag("nrrg.pucch.pucchFmtCfgAddDmrs", confPucchCmd.Flags().Lookup("pucchFmtCfgAddDmrs"))
	viper.BindPFlag("nrrg.pucch.pucchFmtCfgSimAckCsi", confPucchCmd.Flags().Lookup("pucchFmtCfgSimAckCsi"))
	viper.BindPFlag("nrrg.pucch._pucchResId", confPucchCmd.Flags().Lookup("_pucchResId"))
	viper.BindPFlag("nrrg.pucch._pucchFormat", confPucchCmd.Flags().Lookup("_pucchFormat"))
	viper.BindPFlag("nrrg.pucch._pucchResSetId", confPucchCmd.Flags().Lookup("_pucchResSetId"))
	viper.BindPFlag("nrrg.pucch.pucchStartRb", confPucchCmd.Flags().Lookup("pucchStartRb"))
	viper.BindPFlag("nrrg.pucch.pucchIntraSlotFreqHop", confPucchCmd.Flags().Lookup("pucchIntraSlotFreqHop"))
	viper.BindPFlag("nrrg.pucch.pucchSecondHopPrb", confPucchCmd.Flags().Lookup("pucchSecondHopPrb"))
	viper.BindPFlag("nrrg.pucch.pucchNumRbs", confPucchCmd.Flags().Lookup("pucchNumRbs"))
	viper.BindPFlag("nrrg.pucch.pucchStartSymb", confPucchCmd.Flags().Lookup("pucchStartSymb"))
	viper.BindPFlag("nrrg.pucch.pucchNumSymbs", confPucchCmd.Flags().Lookup("pucchNumSymbs"))
	viper.BindPFlag("nrrg.pucch._dsrResId", confPucchCmd.Flags().Lookup("_dsrResId"))
	viper.BindPFlag("nrrg.pucch._dsrPucchRes", confPucchCmd.Flags().Lookup("_dsrPucchRes"))
	viper.BindPFlag("nrrg.pucch.dsrPeriod", confPucchCmd.Flags().Lookup("dsrPeriod"))
	viper.BindPFlag("nrrg.pucch.dsrOffset", confPucchCmd.Flags().Lookup("dsrOffset"))
	confPucchCmd.Flags().MarkHidden("_pucchResId")
	confPucchCmd.Flags().MarkHidden("_pucchFormat")
	confPucchCmd.Flags().MarkHidden("_pucchResSetId")
	confPucchCmd.Flags().MarkHidden("_dsrResId")
	confPucchCmd.Flags().MarkHidden("_dsrPucchRes")
}

func initConfAdvancedCmd() {
	confAdvancedCmd.Flags().IntVar(&flags.advanced.bestSsb, "bestSsb", 0, "Best SSB index")
	confAdvancedCmd.Flags().IntVar(&flags.advanced.pdcchSlotSib1, "pdcchSlotSib1", -1, "PDCCH slot for SIB1")
	confAdvancedCmd.Flags().IntVar(&flags.advanced.prachOccMsg1, "prachOccMsg1", -1, "PRACH occasion for Msg1")
	confAdvancedCmd.Flags().IntVar(&flags.advanced.pdcchOccMsg2, "pdcchOccMsg2", 4, "PDCCH occasion for Msg2")
	confAdvancedCmd.Flags().IntVar(&flags.advanced.pdcchOccMsg4, "pdcchOccMsg4", 0, "PDCCH occasion for Msg4")
	confAdvancedCmd.Flags().IntVar(&flags.advanced.dsrRes, "dsrRes", 0, "DSR resource index")
	confAdvancedCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.advanced.bestSsb", confAdvancedCmd.Flags().Lookup("bestSsb"))
	viper.BindPFlag("nrrg.advanced.pdcchSlotSib1", confAdvancedCmd.Flags().Lookup("pdcchSlotSib1"))
	viper.BindPFlag("nrrg.advanced.prachOccMsg1", confAdvancedCmd.Flags().Lookup("prachOccMsg1"))
	viper.BindPFlag("nrrg.advanced.pdcchOccMsg2", confAdvancedCmd.Flags().Lookup("pdcchOccMsg2"))
	viper.BindPFlag("nrrg.advanced.pdcchOccMsg4", confAdvancedCmd.Flags().Lookup("pdcchOccMsg4"))
	viper.BindPFlag("nrrg.advanced.dsrRes", confAdvancedCmd.Flags().Lookup("dsrRes"))
}

func loadNrrgFlags() {
	// grid settings
	flags.gridsetting.band = viper.GetString("nrrg.gridsetting.band")
	flags.gridsetting._duplexMode = viper.GetString("nrrg.gridsetting._duplexMode")
	flags.gridsetting._maxDlFreq = viper.GetInt("nrrg.gridsetting._maxDlFreq")
	flags.gridsetting._freqRange = viper.GetString("nrrg.gridsetting._freqRange")

	flags.gridsetting.scs = viper.GetString("nrrg.gridsetting.scs")

	flags.gridsetting._ssbScs = viper.GetString("nrrg.gridsetting._ssbScs")
	flags.gridsetting.gscn = viper.GetInt("nrrg.gridsetting.gscn")
	flags.gridsetting._ssbPattern = viper.GetString("nrrg.gridsetting._ssbPattern")
	flags.gridsetting._kSsb = viper.GetInt("nrrg.gridsetting._kSsb")
	flags.gridsetting._nCrbSsb = viper.GetInt("nrrg.gridsetting._nCrbSsb")
	flags.gridsetting.ssbPeriod = viper.GetString("nrrg.ssbBurst.ssbPeriod")
	flags.gridsetting._maxLBar = viper.GetInt("nrrg.ssbBurst._maxLBar")
	flags.gridsetting._maxL = viper.GetInt("nrrg.ssbBurst._maxL")
	flags.gridsetting.candSsbIndex = viper.GetIntSlice("nrrg.ssbBurst.candSsbIndex")

	flags.gridsetting._carrierScs = viper.GetString("nrrg.gridsetting._carrierScs")
	flags.gridsetting.dlArfcn = viper.GetInt("nrrg.gridsetting.dlArfcn")
	flags.gridsetting.bw = viper.GetString("nrrg.gridsetting.bw")
	flags.gridsetting._carrierNumRbs = viper.GetInt("nrrg.gridsetting._carrierNumRbs")
	flags.gridsetting._offsetToCarrier = viper.GetInt("nrrg.gridsetting._offsetToCarrier")

	flags.gridsetting._mibCommonScs = viper.GetString("nrrg.gridsetting._mibCommonScs")
	flags.gridsetting.rmsiCoreset0 = viper.GetInt("nrrg.gridsetting.rmsiCoreset0")
	flags.gridsetting.rmsiCss0 = viper.GetInt("nrrg.gridsetting.rmsiCss0")
	flags.gridsetting._coreset0MultiplexingPat = viper.GetInt("nrrg.gridsetting._coreset0MultiplexingPat")
	flags.gridsetting._coreset0NumRbs = viper.GetInt("nrrg.gridsetting._coreset0NumRbs")
	flags.gridsetting._coreset0NumSymbs = viper.GetInt("nrrg.gridsetting._coreset0NumSymbs")
	flags.gridsetting._coreset0OffsetList = viper.GetIntSlice("nrrg.gridsetting._coreset0OffsetList")
	flags.gridsetting._coreset0Offset = viper.GetInt("nrrg.gridsetting._coreset0Offset")

	// common settings
	flags.mib.sfn = viper.GetInt("nrrg.mib.sfn")
	flags.mib.hrf = viper.GetInt("nrrg.mib.hrf")
	flags.mib.dmrsTypeAPos = viper.GetString("nrrg.mib.dmrsTypeAPos")

	flags.commonSetting.pci = viper.GetInt("nrrg.commonsetting.pci")
	flags.commonSetting.numUeAp = viper.GetString("nrrg.commonsetting.numUeAp")
	flags.commonSetting._refScs = viper.GetString("nrrg.commonsetting._refScs")
	flags.commonSetting.patPeriod = viper.GetStringSlice("nrrg.commonsetting.patPeriod")
	flags.commonSetting.patNumDlSlots = viper.GetIntSlice("nrrg.commonsetting.patNumDlSlots")
	flags.commonSetting.patNumDlSymbs = viper.GetIntSlice("nrrg.commonsetting.patNumDlSymbs")
	flags.commonSetting.patNumUlSymbs = viper.GetIntSlice("nrrg.commonsetting.patNumUlSymbs")
	flags.commonSetting.patNumUlSlots = viper.GetIntSlice("nrrg.commonsetting.patNumUlSlots")

	flags.css0.css0AggLevel = viper.GetInt("nrrg.css0.css0AggLevel")
	flags.css0.css0NumCandidates = viper.GetString("nrrg.css0.css0NumCandidates")

	flags.coreset1.coreset1FreqRes = viper.GetString("nrrg.coreset1.coreset1FreqRes")
	flags.coreset1.coreset1Duration = viper.GetInt("nrrg.coreset1.coreset1Duration")
	flags.coreset1.coreset1CceRegMap = viper.GetString("nrrg.coreset1.coreset1CceRegMap")
	flags.coreset1.coreset1RegBundleSize = viper.GetString("nrrg.coreset1.coreset1RegBundleSize")
	flags.coreset1.coreset1InterleaverSize = viper.GetString("nrrg.coreset1.coreset1InterleaverSize")
	flags.coreset1.coreset1ShiftInd = viper.GetInt("nrrg.coreset1.coreset1ShiftInd")

	flags.uss.ussPeriod = viper.GetString("nrrg.uss.ussPeriod")
	flags.uss.ussOffset = viper.GetInt("nrrg.uss.ussOffset")
	flags.uss.ussDuration = viper.GetInt("nrrg.uss.ussDuration")
	flags.uss.ussFirstSymbs = viper.GetString("nrrg.uss.ussFirstSymbs")
	flags.uss.ussAggLevel = viper.GetInt("nrrg.uss.ussAggLevel")
	flags.uss.ussNumCandidates = viper.GetString("nrrg.uss.ussNumCandidates")

	flags.dci10._rnti = viper.GetStringSlice("nrrg.dci10._rnti")
	flags.dci10._muPdcch = viper.GetIntSlice("nrrg.dci10._muPdcch")
	flags.dci10._muPdsch = viper.GetIntSlice("nrrg.dci10._muPdsch")
	flags.dci10.dci10TdRa = viper.GetIntSlice("nrrg.dci10.dci10TdRa")
	flags.dci10._tdMappingType = viper.GetStringSlice("nrrg.dci10._tdMappingType")
	flags.dci10._tdK0 = viper.GetIntSlice("nrrg.dci10._tdK0")
	flags.dci10._tdSliv = viper.GetIntSlice("nrrg.dci10._tdSliv")
	flags.dci10._tdStartSymb = viper.GetIntSlice("nrrg.dci10._tdStartSymb")
	flags.dci10._tdNumSymbs = viper.GetIntSlice("nrrg.dci10._tdNumSymbs")
	flags.dci10._fdRaType = viper.GetStringSlice("nrrg.dci10._fdRaType")
	flags.dci10._fdBitwidthRaType1 = viper.GetIntSlice("nrrg.dci10._fdBitwidthRaType1")
	flags.dci10._fdRa = viper.GetStringSlice("nrrg.dci10._fdRa")
	flags.dci10.dci10FdStartRb = viper.GetIntSlice("nrrg.dci10.dci10FdStartRb")
	flags.dci10.dci10FdNumRbs = viper.GetIntSlice("nrrg.dci10.dci10FdNumRbs")
	flags.dci10.dci10FdVrbPrbMappingType = viper.GetStringSlice("nrrg.dci10.dci10FdVrbPrbMappingType")
	flags.dci10._fdBundleSize = viper.GetStringSlice("nrrg.dci10._fdBundleSize")
	flags.dci10.dci10McsCw0 = viper.GetIntSlice("nrrg.dci10.dci10McsCw0")
	flags.dci10._tbs = viper.GetIntSlice("nrrg.dci10._tbs")
	flags.dci10.dci10Msg2TbScaling = viper.GetInt("nrrg.dci10.dci10Msg2TbScaling")
	flags.dci10.dci10Msg4DeltaPri = viper.GetInt("nrrg.dci10.dci10Msg4DeltaPri")
	flags.dci10.dci10Msg4TdK1 = viper.GetInt("nrrg.dci10.dci10Msg4TdK1")

	flags.dci11._rnti = viper.GetString("nrrg.dci11._rnti")
	flags.dci11._muPdcch = viper.GetInt("nrrg.dci11._muPdcch")
	flags.dci11._muPdsch = viper.GetInt("nrrg.dci11._muPdsch")
	flags.dci11._actBwp = viper.GetInt("nrrg.dci11._actBwp")
	flags.dci11._indicatedBwp = viper.GetInt("nrrg.dci11._indicatedBwp")
	flags.dci11.dci11TdRa = viper.GetInt("nrrg.dci11.dci11TdRa")
	flags.dci11.dci11TdMappingType = viper.GetString("nrrg.dci11.dci11TdMappingType")
	flags.dci11.dci11TdK0 = viper.GetInt("nrrg.dci11.dci11TdK0")
	flags.dci11.dci11TdSliv = viper.GetInt("nrrg.dci11.dci11TdSliv")
	flags.dci11.dci11TdStartSymb = viper.GetInt("nrrg.dci11.dci11TdStartSymb")
	flags.dci11.dci11TdNumSymbs = viper.GetInt("nrrg.dci11.dci11TdNumSymbs")
	flags.dci11.dci11FdRaType = viper.GetString("nrrg.dci11.dci11FdRaType")
	flags.dci11._dci11FdBitwidthRaType0 = viper.GetInt("nrrg.dci11._dci11FdBitwidthRaType0")
	flags.dci11._dci11FdBitwidthRaType1 = viper.GetInt("nrrg.dci11._dci11FdBitwidthRaType1")
	flags.dci11.dci11FdRa = viper.GetString("nrrg.dci11.dci11FdRa")
	flags.dci11.dci11FdStartRb = viper.GetInt("nrrg.dci11.dci11FdStartRb")
	flags.dci11.dci11FdNumRbs = viper.GetInt("nrrg.dci11.dci11FdNumRbs")
	flags.dci11.dci11FdVrbPrbMappingType = viper.GetString("nrrg.dci11.dci11FdVrbPrbMappingType")
	flags.dci11.dci11FdBundleSize = viper.GetString("nrrg.dci11.dci11FdBundleSize")
	flags.dci11.dci11McsCw0 = viper.GetInt("nrrg.dci11.dci11McsCw0")
	flags.dci11.dci11McsCw1 = viper.GetInt("nrrg.dci11.dci11McsCw1")
	flags.dci11._tbs = viper.GetIntSlice("nrrg.dci11._tbs")
	flags.dci11.dci11DeltaPri = viper.GetInt("nrrg.dci11.dci11DeltaPri")
	flags.dci11.dci11TdK1 = viper.GetInt("nrrg.dci11.dci11TdK1")
	flags.dci11.dci11AntPorts = viper.GetInt("nrrg.dci11.dci11AntPorts")

	flags.msg3._muPusch = viper.GetInt("nrrg.msg3._muPusch")
	flags.msg3.msg3TdRa = viper.GetInt("nrrg.msg3.msg3TdRa")
	flags.msg3._tdMappingType = viper.GetString("nrrg.msg3._tdMappingType")
	flags.msg3._tdK2 = viper.GetInt("nrrg.msg3._tdK2")
	flags.msg3._tdDelta = viper.GetInt("nrrg.msg3._tdDelta")
	flags.msg3._tdSliv = viper.GetInt("nrrg.msg3._tdSliv")
	flags.msg3._tdStartSymb = viper.GetInt("nrrg.msg3._tdStartSymb")
	flags.msg3._tdNumSymbs = viper.GetInt("nrrg.msg3._tdNumSymbs")
	flags.msg3._fdRaType = viper.GetString("nrrg.msg3._fdRaType")
	flags.msg3.msg3FdFreqHop = viper.GetString("nrrg.msg3.msg3FdFreqHop")
	flags.msg3.msg3FdRa = viper.GetString("nrrg.msg3.msg3FdRa")
	flags.msg3.msg3FdStartRb = viper.GetInt("nrrg.msg3.msg3FdStartRb")
	flags.msg3.msg3FdNumRbs = viper.GetInt("nrrg.msg3.msg3FdNumRbs")
	flags.msg3._fdSecondHopFreqOff = viper.GetInt("nrrg.msg3._fdSecondHopFreqOff")
	flags.msg3.msg3McsCw0 = viper.GetInt("nrrg.msg3.msg3McsCw0")
	flags.msg3._tbs = viper.GetInt("nrrg.msg3._tbs")

	flags.dci01._rnti = viper.GetString("nrrg.dci01._rnti")
	flags.dci01._muPdcch = viper.GetInt("nrrg.dci01._muPdcch")
	flags.dci01._muPusch = viper.GetInt("nrrg.dci01._muPusch")
	flags.dci01._actBwp = viper.GetInt("nrrg.dci01._actBwp")
	flags.dci01._indicatedBwp = viper.GetInt("nrrg.dci01._indicatedBwp")
	flags.dci01.dci01TdRa = viper.GetInt("nrrg.dci01.dci01TdRa")
	flags.dci01.dci01TdMappingType = viper.GetString("nrrg.dci01.dci01TdMappingType")
	flags.dci01.dci01TdK2 = viper.GetInt("nrrg.dci01.dci01TdK2")
	flags.dci01.dci01TdSliv = viper.GetInt("nrrg.dci01.dci01TdSliv")
	flags.dci01.dci01TdStartSymb = viper.GetInt("nrrg.dci01.dci01TdStartSymb")
	flags.dci01.dci01TdNumSymbs = viper.GetInt("nrrg.dci01.dci01TdNumSymbs")
	flags.dci01.dci01FdRaType = viper.GetString("nrrg.dci01.dci01FdRaType")
	flags.dci01.dci01FdFreqHop = viper.GetString("nrrg.dci01.dci01FdFreqHop")
	flags.dci01.dci01FdRa = viper.GetString("nrrg.dci01.dci01FdRa")
	flags.dci01.dci01FdStartRb = viper.GetInt("nrrg.dci01.dci01FdStartRb")
	flags.dci01.dci01FdNumRbs = viper.GetInt("nrrg.dci01.dci01FdNumRbs")
	flags.dci01.dci01McsCw0 = viper.GetInt("nrrg.dci01.dci01McsCw0")
	flags.dci01._tbs = viper.GetInt("nrrg.dci01._tbs")
	flags.dci01.dci01CbTpmiNumLayers = viper.GetInt("nrrg.dci01.dci01CbTpmiNumLayers")
	flags.dci01.dci01Sri = viper.GetString("nrrg.dci01.dci01Sri")
	flags.dci01.dci01AntPorts = viper.GetInt("nrrg.dci01.dci01AntPorts")
	flags.dci01.dci01PtrsDmrsMap = viper.GetInt("nrrg.dci01.dci01PtrsDmrsMap")

	flags.bwp._bwpType = viper.GetStringSlice("nrrg.bwp._bwpType")
	flags.bwp._bwpId = viper.GetIntSlice("nrrg.bwp._bwpId")
	flags.bwp._bwpScs = viper.GetStringSlice("nrrg.bwp._bwpScs")
	flags.bwp._bwpCp = viper.GetStringSlice("nrrg.bwp._bwpCp")
	flags.bwp._bwpLocAndBw = viper.GetIntSlice("nrrg.bwp._bwpLocAndBw")
	flags.bwp._bwpStartRb = viper.GetIntSlice("nrrg.bwp._bwpStartRb")
	flags.bwp._bwpNumRbs = viper.GetIntSlice("nrrg.bwp._bwpNumRbs")

	flags.rach.prachConfId = viper.GetInt("nrrg.rach.prachConfId")
	flags.rach._raFormat = viper.GetString("nrrg.rach._raFormat")
	flags.rach._raX = viper.GetInt("nrrg.rach._raX")
	flags.rach._raY = viper.GetIntSlice("nrrg.rach._raY")
	flags.rach._raSubfNumFr1SlotNumFr2 = viper.GetIntSlice("nrrg.rach._raSubfNumFr1SlotNumFr2")
	flags.rach._raStartingSymb = viper.GetInt("nrrg.rach._raStartingSymb")
	flags.rach._raNumSlotsPerSubfFr1Per60KSlotFr2 = viper.GetInt("nrrg.rach._raNumSlotsPerSubfFr1Per60KSlotFr2")
	flags.rach._raNumOccasionsPerSlot = viper.GetInt("nrrg.rach._raNumOccasionsPerSlot")
	flags.rach._raDuration = viper.GetInt("nrrg.rach._raDuration")
	flags.rach.msg1Scs = viper.GetString("nrrg.rach.msg1Scs")
	flags.rach.msg1Fdm = viper.GetInt("nrrg.rach.msg1Fdm")
	flags.rach.msg1FreqStart = viper.GetInt("nrrg.rach.msg1FreqStart")
	flags.rach.raRespWin = viper.GetString("nrrg.rach.raRespWin")
	flags.rach.totNumPreambs = viper.GetInt("nrrg.rach.totNumPreambs")
	flags.rach.ssbPerRachOccasion = viper.GetString("nrrg.rach.ssbPerRachOccasion")
	flags.rach.cbPreambsPerSsb = viper.GetInt("nrrg.rach.cbPreambsPerSsb")
	flags.rach.contResTimer = viper.GetString("nrrg.rach.contResTimer")
	flags.rach.msg3Tp = viper.GetString("nrrg.rach.msg3Tp")
	flags.rach._raLen = viper.GetInt("nrrg.rach._raLen")
	flags.rach._raNumRbs = viper.GetInt("nrrg.rach._raNumRbs")
	flags.rach._raKBar = viper.GetInt("nrrg.rach._raKBar")

	flags.dmrsCommon._schInfo = viper.GetStringSlice("nrrg.dmrscommon._schInfo")
	flags.dmrsCommon._dmrsType = viper.GetStringSlice("nrrg.dmrscommon._dmrsType")
	flags.dmrsCommon._dmrsAddPos = viper.GetStringSlice("nrrg.dmrscommon._dmrsAddPos")
	flags.dmrsCommon._maxLength = viper.GetStringSlice("nrrg.dmrscommon._maxLength")
	flags.dmrsCommon._dmrsPorts = viper.GetIntSlice("nrrg.dmrscommon._dmrsPorts")
	flags.dmrsCommon._cdmGroupsWoData = viper.GetIntSlice("nrrg.dmrscommon._cdmGroupsWoData")
	flags.dmrsCommon._numFrontLoadSymbs = viper.GetIntSlice("nrrg.dmrscommon._numFrontLoadSymbs")
	flags.dmrsCommon._tdL = viper.GetStringSlice("nrrg.dmrscommon._tdL")
	flags.dmrsCommon._fdK = viper.GetStringSlice("nrrg.dmrscommon._fdK")

	flags.dmrsPdsch.pdschDmrsType = viper.GetString("nrrg.dmrspdsch.pdschDmrsType")
	flags.dmrsPdsch.pdschDmrsAddPos = viper.GetString("nrrg.dmrspdsch.pdschDmrsAddPos")
	flags.dmrsPdsch.pdschMaxLength = viper.GetString("nrrg.dmrspdsch.pdschMaxLength")
	flags.dmrsPdsch._dmrsPorts = viper.GetIntSlice("nrrg.dmrspdsch._dmrsPorts")
	flags.dmrsPdsch._cdmGroupsWoData = viper.GetInt("nrrg.dmrspdsch._cdmGroupsWoData")
	flags.dmrsPdsch._numFrontLoadSymbs = viper.GetInt("nrrg.dmrspdsch._numFrontLoadSymbs")
	flags.dmrsPdsch._tdL = viper.GetString("nrrg.dmrspdsch._tdL")
	flags.dmrsPdsch._fdK = viper.GetString("nrrg.dmrspdsch._fdK")

	flags.ptrsPdsch.pdschPtrsEnabled = viper.GetBool("nrrg.ptrspdsch.pdschPtrsEnabled")
	flags.ptrsPdsch.pdschPtrsTimeDensity = viper.GetInt("nrrg.ptrspdsch.pdschPtrsTimeDensity")
	flags.ptrsPdsch.pdschPtrsFreqDensity = viper.GetInt("nrrg.ptrspdsch.pdschPtrsFreqDensity")
	flags.ptrsPdsch.pdschPtrsReOffset = viper.GetString("nrrg.ptrspdsch.pdschPtrsReOffset")
	flags.ptrsPdsch._dmrsPorts = viper.GetInt("nrrg.ptrspdsch._dmrsPorts")

	flags.dmrsPusch.puschDmrsType = viper.GetString("nrrg.dmrspusch.puschDmrsType")
	flags.dmrsPusch.puschDmrsAddPos = viper.GetString("nrrg.dmrspusch.puschDmrsAddPos")
	flags.dmrsPusch.puschMaxLength = viper.GetString("nrrg.dmrspusch.puschMaxLength")
	flags.dmrsPusch._dmrsPorts = viper.GetIntSlice("nrrg.dmrspusch._dmrsPorts")
	flags.dmrsPusch._cdmGroupsWoData = viper.GetInt("nrrg.dmrspusch._cdmGroupsWoData")
	flags.dmrsPusch._numFrontLoadSymbs = viper.GetInt("nrrg.dmrspusch._numFrontLoadSymbs")
	flags.dmrsPusch._tdL = viper.GetString("nrrg.dmrspusch._tdL")
	flags.dmrsPusch._fdK = viper.GetString("nrrg.dmrspusch._fdK")

	flags.ptrsPusch.puschPtrsEnabled = viper.GetBool("nrrg.ptrspusch.puschPtrsEnabled")
	flags.ptrsPusch.puschPtrsTimeDensity = viper.GetInt("nrrg.ptrspusch.puschPtrsTimeDensity")
	flags.ptrsPusch.puschPtrsFreqDensity = viper.GetInt("nrrg.ptrspusch.puschPtrsFreqDensity")
	flags.ptrsPusch.puschPtrsReOffset = viper.GetString("nrrg.ptrspusch.puschPtrsReOffset")
	flags.ptrsPusch.puschPtrsMaxNumPorts = viper.GetString("nrrg.ptrspusch.puschPtrsMaxNumPorts")
	flags.ptrsPusch._dmrsPorts = viper.GetIntSlice("nrrg.ptrspusch._dmrsPorts")
	flags.ptrsPusch.puschPtrsTimeDensityTp = viper.GetInt("nrrg.ptrspusch.puschPtrsTimeDensityTp")
	flags.ptrsPusch.puschPtrsGrpPatternTp = viper.GetString("nrrg.ptrspusch.puschPtrsGrpPatternTp")
	flags.ptrsPusch._numGrpsTp = viper.GetInt("nrrg.ptrspusch._numGrpsTp")
	flags.ptrsPusch._samplesPerGrpTp = viper.GetInt("nrrg.ptrspusch._samplesPerGrpTp")
	flags.ptrsPusch._dmrsPortsTp = viper.GetIntSlice("nrrg.ptrspusch._dmrsPortsTp")

	flags.pdsch.pdschAggFactor = viper.GetString("nrrg.pdsch.pdschAggFactor")
	flags.pdsch.pdschRbgCfg = viper.GetString("nrrg.pdsch.pdschRbgCfg")
	flags.pdsch._rbgSize = viper.GetInt("nrrg.pdsch._rbgSize")
	flags.pdsch.pdschMcsTable = viper.GetString("nrrg.pdsch.pdschMcsTable")
	flags.pdsch.pdschXOh = viper.GetString("nrrg.pdsch.pdschXOh")

	flags.pusch.puschTxCfg = viper.GetString("nrrg.pusch.puschTxCfg")
	flags.pusch.puschCbSubset = viper.GetString("nrrg.pusch.puschCbSubset")
	flags.pusch.puschCbMaxRankNonCbMaxLayers = viper.GetInt("nrrg.pusch.puschCbMaxRankNonCbMaxLayers")
	flags.pusch.puschFreqHopOffset = viper.GetInt("nrrg.pusch.puschFreqHopOffset")
	flags.pusch.puschTp = viper.GetString("nrrg.pusch.puschTp")
	flags.pusch.puschAggFactor = viper.GetString("nrrg.pusch.puschAggFactor")
	flags.pusch.puschRbgCfg = viper.GetString("nrrg.pusch.puschRbgCfg")
	flags.pusch._rbgSize = viper.GetInt("nrrg.pusch._rbgSize")
	flags.pusch.puschMcsTable = viper.GetString("nrrg.pusch.puschMcsTable")
	flags.pusch.puschXOh = viper.GetString("nrrg.pusch.puschXOh")

	flags.nzpCsiRs._resSetId = viper.GetInt("nrrg.nzpcsirs._resSetId")
	flags.nzpCsiRs._trsInfo = viper.GetBool("nrrg.nzpcsirs._trsInfo")
	flags.nzpCsiRs._resId = viper.GetInt("nrrg.nzpcsirs._resId")
	flags.nzpCsiRs.nzpCsiRsFreqAllocRow = viper.GetString("nrrg.nzpcsirs.nzpCsiRsFreqAllocRow")
	flags.nzpCsiRs.nzpCsiRsFreqAllocBits = viper.GetString("nrrg.nzpcsirs.nzpCsiRsFreqAllocBits")
	flags.nzpCsiRs.nzpCsiRsNumPorts = viper.GetString("nrrg.nzpcsirs.nzpCsiRsNumPorts")
	flags.nzpCsiRs.nzpCsiRsCdmType = viper.GetString("nrrg.nzpcsirs.nzpCsiRsCdmType")
	flags.nzpCsiRs.nzpCsiRsDensity = viper.GetString("nrrg.nzpcsirs.nzpCsiRsDensity")
	flags.nzpCsiRs.nzpCsiRsFirstSymb = viper.GetInt("nrrg.nzpcsirs.nzpCsiRsFirstSymb")
	flags.nzpCsiRs.nzpCsiRsFirstSymb2 = viper.GetInt("nrrg.nzpcsirs.nzpCsiRsFirstSymb2")
	flags.nzpCsiRs.nzpCsiRsStartRb = viper.GetInt("nrrg.nzpcsirs.nzpCsiRsStartRb")
	flags.nzpCsiRs.nzpCsiRsNumRbs = viper.GetInt("nrrg.nzpcsirs.nzpCsiRsNumRbs")
	flags.nzpCsiRs.nzpCsiRsPeriod = viper.GetString("nrrg.nzpcsirs.nzpCsiRsPeriod")
	flags.nzpCsiRs.nzpCsiRsOffset = viper.GetInt("nrrg.nzpcsirs.nzpCsiRsOffset")
	flags.nzpCsiRs._row = viper.GetInt("nrrg.nzpcsirs._row")
	flags.nzpCsiRs._kBarLBar = viper.GetStringSlice("nrrg.nzpcsirs._kBarLBar")
	flags.nzpCsiRs._ki = viper.GetIntSlice("nrrg.nzpcsirs._ki")
	flags.nzpCsiRs._li = viper.GetIntSlice("nrrg.nzpcsirs._li")
	flags.nzpCsiRs._cdmGrpIndj = viper.GetIntSlice("nrrg.nzpcsirs._cdmGrpIndj")
	flags.nzpCsiRs._kap = viper.GetIntSlice("nrrg.nzpcsirs._kap")
	flags.nzpCsiRs._lap = viper.GetIntSlice("nrrg.nzpcsirs._lap")

	flags.trs._resSetId = viper.GetInt("nrrg.trs._resSetId")
	flags.trs._trsInfo = viper.GetBool("nrrg.trs._trsInfo")
	flags.trs._firstResId = viper.GetInt("nrrg.trs._firstResId")
	flags.trs._freqAllocRow = viper.GetString("nrrg.trs._freqAllocRow")
	flags.trs.trsFreqAllocBits = viper.GetString("nrrg.trs.trsFreqAllocBits")
	flags.trs._numPorts = viper.GetString("nrrg.trs._numPorts")
	flags.trs._cdmType = viper.GetString("nrrg.trs._cdmType")
	flags.trs._density = viper.GetString("nrrg.trs._density")
	flags.trs.trsFirstSymbs = viper.GetIntSlice("nrrg.trs.trsFirstSymbs")
	flags.trs.trsStartRb = viper.GetInt("nrrg.trs.trsStartRb")
	flags.trs.trsNumRbs = viper.GetInt("nrrg.trs.trsNumRbs")
	flags.trs.trsPeriod = viper.GetString("nrrg.trs.trsPeriod")
	flags.trs.trsOffset = viper.GetIntSlice("nrrg.trs.trsOffset")
	flags.trs._row = viper.GetInt("nrrg.trs._row")
	flags.trs._kBarLBar = viper.GetStringSlice("nrrg.trs._kBarLBar")
	flags.trs._ki = viper.GetIntSlice("nrrg.trs._ki")
	flags.trs._li = viper.GetIntSlice("nrrg.trs._li")
	flags.trs._cdmGrpIndj = viper.GetIntSlice("nrrg.trs._cdmGrpIndj")
	flags.trs._kap = viper.GetIntSlice("nrrg.trs._kap")
	flags.trs._lap = viper.GetIntSlice("nrrg.trs._lap")

	flags.csiIm._resSetId = viper.GetInt("nrrg.csiim._resSetId")
	flags.csiIm._resId = viper.GetInt("nrrg.csiim._resId")
	flags.csiIm.csiImRePattern = viper.GetString("nrrg.csiim.csiImRePattern")
	flags.csiIm.csiImScLoc = viper.GetString("nrrg.csiim.csiImScLoc")
	flags.csiIm.csiImSymbLoc = viper.GetInt("nrrg.csiim.csiImSymbLoc")
	flags.csiIm.csiImStartRb = viper.GetInt("nrrg.csiim.csiImStartRb")
	flags.csiIm.csiImNumRbs = viper.GetInt("nrrg.csiim.csiImNumRbs")
	flags.csiIm.csiImPeriod = viper.GetString("nrrg.csiim.csiImPeriod")
	flags.csiIm.csiImOffset = viper.GetInt("nrrg.csiim.csiImOffset")

	flags.csiReport._resCfgType = viper.GetStringSlice("nrrg.csireport._resCfgType")
	flags.csiReport._resCfgId = viper.GetIntSlice("nrrg.csireport._resCfgId")
	flags.csiReport._resSetId = viper.GetIntSlice("nrrg.csireport._resSetId")
	flags.csiReport._resBwpId = viper.GetIntSlice("nrrg.csireport._resBwpId")
	flags.csiReport._resType = viper.GetStringSlice("nrrg.csireport._resType")
	flags.csiReport._repCfgId = viper.GetInt("nrrg.csireport._repCfgId")
	flags.csiReport._resCfgIdChnMeas = viper.GetInt("nrrg.csireport._resCfgIdChnMeas")
	flags.csiReport._resCfgIdCsiImIntf = viper.GetInt("nrrg.csireport._resCfgIdCsiImIntf")
	flags.csiReport._repCfgType = viper.GetString("nrrg.csireport._repCfgType")
	flags.csiReport.csiRepPeriod = viper.GetString("nrrg.csireport.csiRepPeriod")
	flags.csiReport.csiRepOffset = viper.GetInt("nrrg.csireport.csiRepOffset")
	flags.csiReport._ulBwpId = viper.GetInt("nrrg.csireport._ulBwpId")
	flags.csiReport.csiRepPucchRes = viper.GetInt("nrrg.csireport.csiRepPucchRes")
	flags.csiReport._quantity = viper.GetString("nrrg.csireport._quantity")

	flags.srs._resId = viper.GetIntSlice("nrrg.srs._resId")
	flags.srs.srsNumPorts = viper.GetStringSlice("nrrg.srs.srsNumPorts")
	flags.srs.srsNonCbPtrsPort = viper.GetStringSlice("nrrg.srs.srsNonCbPtrsPort")
	flags.srs.srsNumCombs = viper.GetStringSlice("nrrg.srs.srsNumCombs")
	flags.srs.srsCombOff = viper.GetIntSlice("nrrg.srs.srsCombOff")
	flags.srs.srsCs = viper.GetIntSlice("nrrg.srs.srsCs")
	flags.srs.srsStartPos = viper.GetIntSlice("nrrg.srs.srsStartPos")
	flags.srs.srsNumSymbs = viper.GetStringSlice("nrrg.srs.srsNumSymbs")
	flags.srs.srsRepetition = viper.GetStringSlice("nrrg.srs.srsRepetition")
	flags.srs.srsFreqPos = viper.GetIntSlice("nrrg.srs.srsFreqPos")
	flags.srs.srsFreqShift = viper.GetIntSlice("nrrg.srs.srsFreqShift")
	flags.srs.srsCSrs = viper.GetIntSlice("nrrg.srs.srsCSrs")
	flags.srs.srsBSrs = viper.GetIntSlice("nrrg.srs.srsBSrs")
	flags.srs.srsBHop = viper.GetIntSlice("nrrg.srs.srsBHop")
	flags.srs._type = viper.GetStringSlice("nrrg.srs._type")
	flags.srs.srsPeriod = viper.GetStringSlice("nrrg.srs.srsPeriod")
	flags.srs.srsOffset = viper.GetIntSlice("nrrg.srs.srsOffset")
	flags.srs._mSRSb = viper.GetStringSlice("nrrg.srs._mSRSb")
	flags.srs._Nb = viper.GetStringSlice("nrrg.srs._Nb")
	flags.srs._resSetId = viper.GetIntSlice("nrrg.srs._resSetId")
	flags.srs.srsSetResIdList = viper.GetStringSlice("nrrg.srs.srsSetResIdList")
	flags.srs._resType = viper.GetStringSlice("nrrg.srs._resType")
	flags.srs._usage = viper.GetStringSlice("nrrg.srs._usage")
	flags.srs._dci01NonCbSrsList = viper.GetStringSlice("nrrg.srs._dci01NonCbSrsList")

	flags.pucch.pucchFmtCfgNumSlots = viper.GetString("nrrg.pucch.pucchFmtCfgNumSlots")
	flags.pucch.pucchFmtCfgInterSlotFreqHop = viper.GetString("nrrg.pucch.pucchFmtCfgInterSlotFreqHop")
	flags.pucch.pucchFmtCfgAddDmrs = viper.GetBool("nrrg.pucch.pucchFmtCfgAddDmrs")
	flags.pucch.pucchFmtCfgSimAckCsi = viper.GetBool("nrrg.pucch.pucchFmtCfgSimAckCsi")
	flags.pucch._pucchResId = viper.GetIntSlice("nrrg.pucch._pucchResId")
	flags.pucch._pucchFormat = viper.GetStringSlice("nrrg.pucch._pucchFormat")
	flags.pucch._pucchResSetId = viper.GetIntSlice("nrrg.pucch._pucchResSetId")
	flags.pucch.pucchStartRb = viper.GetIntSlice("nrrg.pucch.pucchStartRb")
	flags.pucch.pucchIntraSlotFreqHop = viper.GetStringSlice("nrrg.pucch.pucchIntraSlotFreqHop")
	flags.pucch.pucchSecondHopPrb = viper.GetIntSlice("nrrg.pucch.pucchSecondHopPrb")
	flags.pucch.pucchNumRbs = viper.GetIntSlice("nrrg.pucch.pucchNumRbs")
	flags.pucch.pucchStartSymb = viper.GetIntSlice("nrrg.pucch.pucchStartSymb")
	flags.pucch.pucchNumSymbs = viper.GetIntSlice("nrrg.pucch.pucchNumSymbs")
	flags.pucch._dsrResId = viper.GetIntSlice("nrrg.pucch._dsrResId")
	flags.pucch._dsrPucchRes = viper.GetIntSlice("nrrg.pucch._dsrPucchRes")
	flags.pucch.dsrPeriod = viper.GetStringSlice("nrrg.pucch.dsrPeriod")
	flags.pucch.dsrOffset = viper.GetIntSlice("nrrg.pucch.dsrOffset")

	flags.advanced.bestSsb = viper.GetInt("nrrg.advanced.bestSsb")
	flags.advanced.pdcchSlotSib1 = viper.GetInt("nrrg.advanced.pdcchSlotSib1")
	flags.advanced.prachOccMsg1 = viper.GetInt("nrrg.advanced.prachOccMsg1")
	flags.advanced.pdcchOccMsg2 = viper.GetInt("nrrg.advanced.pdcchOccMsg2")
	flags.advanced.pdcchOccMsg4 = viper.GetInt("nrrg.advanced.pdcchOccMsg4")
	flags.advanced.dsrRes = viper.GetInt("nrrg.advanced.dsrRes")
}

var w =[]int{len("Flag"), len("Type"), len("Current Value"), len("Default Value")}
// var w =[]int{len("Flag"), len("Type"), len("Current Value")}

/*
laPrint performs left-aligned printing.
 */
func laPrint(cmd *cobra.Command, args []string) {
	regGreen.Printf("[INFO]: List of parameters\n")
	cmd.Flags().VisitAll(
		func (f *pflag.Flag) {
			if f.Name != "config" && f.Name != "help" {
				if len(f.Name) > w[0] { w[0] = len(f.Name) }
				if len(f.Value.Type()) > w[1] { w[1] = len(f.Value.Type()) }
				if len(f.Value.String()) > w[2] { w[2] = len(f.Value.String()) }
				 if len(f.DefValue) > w[3] { w[3] = len(f.DefValue) }
			}
		})

	for i := 0; i < len(w); i++ {
		w[i] += 4
	}

	 fmt.Printf("%-*v%-*v%-*v%-*v%v\n", w[0], "Flag", w[1], "Type", w[2], "Current Value", w[3], "Default Value", "Modifiable")
	// fmt.Printf("%-*v%-*v%-*v%v\n", w[0], "Flag", w[1], "Type", w[2], "Current Value", "Modifiable")
	cmd.Flags().VisitAll(
		func (f *pflag.Flag) {
			if f.Name != "config" && f.Name != "help" {
				if f.Hidden {
					fmt.Printf("%-*v%-*v%-*v%-*v%v\n", w[0], f.Name, w[1], f.Value.Type(), w[2], f.Value, w[3], f.DefValue, !f.Hidden)
				} else {
					regBlue.Printf("%-*v%-*v%-*v%-*v%v\n", w[0], f.Name, w[1], f.Value.Type(), w[2], f.Value, w[3], f.DefValue, !f.Hidden)
				}
				// fmt.Printf("%-*v%-*v%-*v%v\n", w[0], f.Name, w[1], f.Value.Type(), w[2], f.Value, !f.Hidden)
			}
		})
}