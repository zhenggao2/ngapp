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
	mapset "github.com/deckarep/golang-set"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/zhenggao2/ngapp/nrgrid"
	"github.com/zhenggao2/ngapp/utils"
	"math"
	"sort"
	"strconv"
	"strings"
)

var (
	flags   NrrgFlags
	rgd     NrrgData
	minChBw int

	//boldRed    = color.New(color.FgHiRed).Add(color.Bold).SprintFunc()
	regRed = color.New(color.FgHiRed)
	//boldGreen  = color.New(color.FgHiGreen).Add(color.Bold).SprintFunc()
	regGreen = color.New(color.FgHiGreen)
	//boldBlue   = color.New(color.FgHiBlue).Add(color.Bold).SprintFunc()
	regBlue = color.New(color.FgHiBlue)
	//boldYellow = color.New(color.FgHiYellow).Add(color.Bold).SprintFunc()
	regYellow = color.New(color.FgHiYellow)
	//boldMagenta = color.New(color.FgHiMagenta).Add(color.Bold).SprintFunc()
	regMagenta = color.New(color.FgHiMagenta)
	//boldCyan = color.New(color.FgHiCyan).Add(color.Bold).SprintFunc()
	regCyan = color.New(color.FgHiCyan)
)

type NrrgFlags struct {
	gridsetting GridSettingFlags
	tdduldl     TddUlDlFlags
	searchspace SearchSpaceFlags
	dldci       DlDciFlags
	uldci       UlDciFlags
	bwp         BwpFlags
	rach        RachFlags
	dmrsCommon  DmrsCommonFlags
	pdsch       PdschFlags
	pusch       PuschFlags
	pucch       PucchFlags
	csi         CsiFlags
	srs         SrsFlags
	advanced    AdvancedFlags
}

// grid setting flags
type GridSettingFlags struct {
	scs string // Unified subcarrier spacing of SSB, RMSI, carrier and BWP

	band        string // NR frequency band, n1~n256 for FR1, n257~n512 for FR2-1 and FR2-2
	_duplexMode string // Duplex mode, which can be FDD/TDD/SUL/SDL
	_maxDlFreq  int    // Maximum DL frequency in MHz
	_freqRange  string // Type of frequency range, which can be FR1, FR2-1 or FR2-2
	_unlicensed bool   // Whether the frequency band can be used for NR-U

	_ssbScs      string  // Subcarrier spacing of SSB
	gscn         int     // GSCN of SSB
	_ssbPattern  string  // SSB pattern, which can be Case A, Case B, Case C, Case D, Case E, Case F, and Case G
	_kSsbScs     float64 // Subcarrier spacing of k_SSB
	_kSsb        int     // The k_SSB which is determined by the ssb-SubcarrierOffset of MIB and PBCH payload
	_nCrbSsbScs  float64 // Subcarrier spacing of N_CRB_SSB
	_nCrbSsb     int     // The N_CRB_SSB which can be obtained from offsetToPointA
	ssbPeriod    string  // Periodicity of SSB
	_maxLBar     int     // Maximum number of SSBs in a half frame
	_maxL        int     // Maximum number of transmitted SSBs within a half frame in a cell
	candSsbIndex []int   // List of candidate SSB index

	_carrierScs      string // The subcarrierSpacing of SCS-SpecificCarrier
	bw               string // Channel bandwidth of the carrier in MHz
	dlArfcn          int    // DL ARFCN of the carrier
	_carrierNumRbs   int    // The carrierBandwidth of SCS-SpecificCarrier
	_offsetToCarrier int    // The offsetToCarrier of SCS-SpecificCarrier

	pci int // Physical cell identity, which can be 0~1007

	_mibCommonScs            string // The subCarrierSpacingCommon of MIB
	rmsiCoreset0             int    // The pdcch-ConfigSIB1 of MIB which determines CORESET0
	_coreset0MultiplexingPat int    // The multiplexing pattern of CORESET0, which can be multiplexing pattern 1/2/3
	_coreset0NumRbs          int    // Number of RBs of CORESET0
	_coreset0NumSymbs        int    // Number of OFDM symbols of CORESET0
	_coreset0OffsetList      []int  // List of offset of CORESET0
	_coreset0Offset          int    // Actual offset of CORESET0
	rmsiCss0                 int    // The pdcch-ConfigSIB1 of MIB which determines CSS0
	_css0AggLevel            int    // CCE aggregation level of the PDCCH candidates of CSS0
	_css0NumCandidates       string // Number of PDCCH candidates of CSS0
	dmrsTypeAPos             string // The dmrs-TypeA-Position of MIB
	_sfn                     int    // The systemFrameNumber of MIB
	_hrf                     int    // The half-frame indicator for SSB transmission
}

// common setting flags
type TddUlDlFlags struct {
	_refScs       string   // The referenceSubcarrierSpacing of TDD-UL-DL-ConfigCommon
	patPeriod     []string // The dl-UL-TransmissionPeriodicityv of TDD-UL-DL-Pattern, and max length is 2
	patNumDlSlots []int    // The nrofDownlinkSlots of TDD-UL-DL-Pattern, and max length is 2
	patNumDlSymbs []int    // The nrofDownlinkSymbols of TDD-UL-DL-Pattern, and max length is 2
	patNumUlSymbs []int    // The nrofUplinkSymbols of TDD-UL-DL-Pattern, and max length is 2
	patNumUlSlots []int    // The nrofUplinkSlots of TDD-UL-DL-Pattern, and max length is 2
}

// initial/dedicated UL/DL BWP
type BwpFlags struct {
	_bwpType     []string
	_bwpId       []int
	_bwpScs      []string
	_bwpCp       []string
	_bwpLocAndBw []int
	_bwpStartRb  []int
	_bwpNumRbs   []int
}

const (
	// BWP tags, not the BWP-Id

	INI_DL_BWP int = 0
	DED_DL_BWP int = 1
	INI_UL_BWP int = 2
	DED_UL_BWP int = 3

	// DL DCI tags

	DCI_10_SIB1  int = 0 // rnti = SI-RNTI
	DCI_10_MSG2  int = 1 // rnti = RA-RNTI
	DCI_10_MSG4  int = 2 // rnti = TC-RNTI
	DCI_11_PDSCH int = 3 // rnti = C-RNTI
	DCI_10_MSGB  int = 4 // rnti = MSGB-RNTI (two-steps CBRA)

	// UL DCI tags

	RAR_UL_MSG3   int = 0 // rnti = RA-RNTI (for RAR UL grant)
	DCI_01_PUSCH  int = 1 // rnti = C-RNTI
	RA_UL_MSGA    int = 2 // rnti = RA-RNTI, together with RAPID (two-steps CBRA)
	FBRAR_UL_MSG3 int = 3 // rnti = TC-RNTI (two-steps CBRA)

	// Common DMRS tags
	DMRS_DCI_10_SIB1 int = 0
	DMRS_DCI_10_MSG2 int = 1
	DMRS_DCI_10_MSG4 int = 2
	DMRS_RAR_UL_MSG3 int = 3
)

// Search space
type SearchSpaceFlags struct {
	_coreset1FdRes            string // the frequencyDomainResources of ControlResourceSet, the length of which is 45 bits
	coreset1StartCrb          int    // the starting CRB, which must fulfill coreset1StartCrb % 6 == 0
	coreset1NumRbs            int    // number of RBs, which must fulfill coreset1NumRbs % 6 == 0
	_coreset1Duration         int    // the duration of ControlResourceSet, which can be 1/2/3; fixed to 1 for simplicity
	coreset1CceRegMappingType string // the cce-REG-MappingType of ControlResourceSet, which can be interleaved or nonInterleaved
	coreset1RegBundleSize     string // the reg-BundleSize of ControlResourceSet when cce-REG-MappingType is interleaved, which can be n2/n3/n6
	coreset1InterleaverSize   string // the interleaverSize of ControlResourceSet when cce-REG-MappingType is interleaved, which can be n2/n3/n6
	_coreset1ShiftIndex       int    // the shiftIndex of ControlResourceSet when coreset1CceRegMappingType is interleaved
	// coreset1PrecoderGranularity string

	_ssId                         []int    // the searchSpaceId of SearchSpace
	_ssType                       []string // the type of search space, which can be type0a/type1/type2/type3/uss
	_ssCoresetId                  []int    // the controlResourceSetId of SearchSpace
	_ssDuration                   []int    // the duration of SearchSpace
	_ssMonitoringSymbolWithinSlot []string // the monitoringSymbolsWithinSlot of SearchSpace, which can be 100/110/111 corresponding to the first 3 symbols
	ssAggregationLevel            []string // the aggregationLevel of SearchSpace, which can be AL1/AL2/AL4/AL8/AL16
	ssNumOfPdcchCandidates        []string // the nrofCandidates of SearchSpace, which can be n0/n1/n2/n3/n4/n5/n6/n8
	_ssPeriodicity                []string // the monitoringSlotPeriodicityAndOffset of SearchSpace
	_ssSlotOffset                 []int    // the monitoringSlotPeriodicityAndOffset of SearchSpace
}

// DL DCI 1_0/1_1 for PDSCH scheduling
type DlDciFlags struct {
	_tag                []string // tag of DL DCI, such as DCI_10_SIB1, DCI_10_MSG2, DCI_10_MSG4, DCI_11_PDSCH
	_rnti               []string // the n_RNTI used for PDSCH scrambling (38.211 7.3.1.1 Scrambling)
	_muPdcch            []int    // the u_PDCCH used for Ks calculation (38.214 5.1.2.1	Resource allocation in time domain)
	_muPdsch            []int    // the u_PDSCH used for Ks calculation (38.214 5.1.2.1	Resource allocation in time domain)
	_indicatedBwp       []int    // the "Bandwidth part indicator" field of DCI 1_1, which should be 0(initial DL BWP) for DCI 1_0
	tdra                []int    // the "Time domain resource assignment" field of DCI 1_0/1_1
	_tdMappingType      []string // the PDSCH mapping type, which can be typeA or typeB (38.214 Table 5.1.2.1-1: Valid S and L combinations)
	_tdK0               []int    // the K0 of PDSCH TDRA (38.214 5.1.2.1.1	Determination of the resource allocation table to be used for PDSCH)
	_tdSliv             []int    // the SLIV of PDSCH TDRA (38.214 5.1.2.1	Resource allocation in time domain)
	_tdStartSymb        []int    // the starting symbol S (38.214 5.1.2.1	Resource allocation in time domain)
	_tdNumSymbs         []int    // the number of consecutive symbols L (38.214 5.1.2.1	Resource allocation in time domain)
	_fdRaType           []string // the PDSCH resource allocation type, which can be raType0 or raType1 (38.214 5.1.2.2	Resource allocation in frequency domain)
	_fdBitsRaType0      int      // the number of bits of the "Frequency domain resource assignment" field when raType0 is configured
	_fdBitsRaType1      []int    // the number of bits of the "Frequency domain resource assignment" field when raType1 is configured
	_fdRa               []string // the "Frequency domain resource assignment" field of DCI 1_0/1_1
	fdStartRb           []int    // the starting VRB RB_start (38.214 5.1.2.2	Resource allocation in frequency domain)
	fdNumRbs            []int    // the number of contiguously allocated resource blocks L_RBs (38.214 5.1.2.2	Resource allocation in frequency domain)
	fdVrbPrbMappingType []string // the "VRB-to-PRB mapping" field of DCI 1_0/1_1
	fdBundleSize        []string // the vrb-ToPRB-Interleaver of PDSCH-Config, which can be n2 or n4
	mcsCw0              []int    // the "Modulation and coding scheme" field for transport block 1 of DCI 1_0/1_1
	_tbsCw0             []int    // the calculated TBS of transport block 1
	mcsCw1              int      // the "Modulation and coding scheme" field for transport block 2 of DCI 1_1
	_tbsCw1             int      // the calculated TBS of transport block 2
	tbScalingFactor     float64  // the TB scaling factor S (38.214 Table 5.1.3.2-2: Scaling factor of Ninfo for P-RNTI, RA-RNTI and MSGB-RNTI)
	deltaPri            int      // the "PUCCH resource indicator" field of DCI 1_0/1_1
	tdK1                int      // the "PDSCH-to-HARQ_feedback timing indicator" field of DCI 1_0/1_1
	antennaPorts        int      // the "Antenna port(s)" field of DCI 1_1
}

// UL DCI 0_1 or RAR UL grant for PUSCH scheduling
type UlDciFlags struct {
	_tag                   []string // tag of UL DCI, such as RAR_UL_MSG3, DCI_01_PUSCH, RA_UL_MSGA, FBRAR_UL_MSG3
	_rnti                  []string // the n_RNTI used for PUSCH scrambling (38.211 6.3.1.1	Scrambling)
	_muPdcch               []int    // the u_PDCCH used for Ks calculation (38.214 6.1.2.1	Resource allocation in time domain)
	_muPusch               []int    // the u_PUSCH used for Ks calculation (38.214 6.1.2.1	Resource allocation in time domain)
	_indicatedBwp          []int    // the "Bandwidth part indicator" field of DCI 0_1, which should be 0(initial UL BWP) for Msg3 scheduled by RAR UL grant
	tdra                   []int    // the "Time domain resource assignment" field of DCI 0_1 or RAR UL grant
	_tdMappingType         []string // the PUSCH mapping type, which can be typeA or typeB (38.214 Table 6.1.2.1-1: Valid S and L combinations)
	_tdK2                  []int    // the K2 of PUSCH TDRA (38.214 6.1.2.1.1	Determination of the resource allocation table to be used for PUSCH)
	_tdDelta               int      // the delta for Msg3 PUSCH scheduled by RAR UL grant (38.213 8.3	PUSCH scheduled by RAR UL grant)
	_tdSliv                []int    // the SLIV of PUSCH TDRA (38.214 6.1.2.1	Resource allocation in time domain)
	_tdStartSymb           []int    // the starting symbol S (38.214 6.1.2.1	Resource allocation in time domain)
	_tdNumSymbs            []int    // the number of consecutive symbols L (38.214 6.1.2.1	Resource allocation in time domain)
	_fdRaType              []string // the PUSCH resource allocation type, which can be raType0 or raType1 (38.214 6.1.2.2	Resource allocation in frequency domain)
	fdFreqHop              []string // the "Frequency hopping flag" field of DCI 0_1 or RAR UL grant (38.214 6.3	UE PUSCH frequency hopping procedure)
	_fdFreqHopOffset       []int    // the frequency offset of 2nd hop (38.214 6.3	UE PUSCH frequency hopping procedure)
	_fdBitsRaType0         int      // the number of bits of the "Frequency domain resource assignment" field when raType0 is configured
	_fdBitsRaType1         []int    // the number of bits of the "Frequency domain resource assignment" field when raType1 is configured
	_fdRa                  []string // the "Frequency domain resource assignment" field of DCI 0_1 or RAR UL grant
	fdStartRb              []int    // the starting VRB RB_start (38.214 5.1.2.2	Resource allocation in frequency domain)
	fdNumRbs               []int    // the number of contiguously allocated resource blocks L_RBs (38.214 6.1.2.2	Resource allocation in frequency domain)
	mcsCw0                 []int    // the "Modulation and coding scheme" field for transport block 1 of DCI 0_1 or RAR UL grant
	_tbs                   []int    // the calculated TBS of transport block 1
	precodingInfoNumLayers int      // the "Precoding information and number of layers" field of DCI 0_1
	srsResIndicator        int      // the "SRS resource indicator" field of DCI 0_1
	antennaPorts           int      // the "Antenna ports" field of DCI 0_1
	ptrsDmrsAssociation    int      // the "PTRS-DMRS association" field of DCI 0_1
}

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
	_msg1Scs                           string // the msg1-SubcarrierSpacing of RACH-ConfigCommon, which can be 15/30KHz for FR1, 60/120KHz for FR2-1 and 120/480/960KHz for FR2-2
	msg1Fdm                            int
	msg1FreqStart                      int
	totNumPreambs                      int
	ssbPerRachOccasion                 string
	cbPreambsPerSsb                    int
	raRespWin                          string
	msg3Tp                             string
	contResTimer                       string
	_raLen                             int
	_raNumRbs                          int
	_raKBar                            int
}

// SIB1/Msg2/Msg4/Msg3 DMRS flags
type DmrsCommonFlags struct {
	_tag               []string // tag of DMRS, such as DCI_10_SIB1, DCI_10_MSG2, DCI_10_MSG4, RAR_UL_MSG3, FBRAR_UL_MSG3
	_dmrsType          []string // the dmrs-Type, which can be type1 or type2
	_dmrsAddPos        []string // the dmrs-AdditionalPosition, which can be pos0, pos1, pos2 or pos3
	_maxLength         []string // the maxLength, which can be len1 or len2
	_dmrsPorts         []int    // DMRS antenna port(s)
	_cdmGroupsWoData   []int    // number of CDM groups without data
	_numFrontLoadSymbs []int    // number of front-load OFDM symbol(s) for DMRS
	_tdL               [][]int  // TD pattern of DMRS in a single slot
	_tdL2              []int    // TD pattern of DMRS in a single slot, only valid for the 2nd hop of Msg3 DMRS when intra-slot frequency hopping is enabled
	_fdK               [][]int  // FD pattern of DMRS in a single PRB
}

// PDSCH-config and PDSCH-ServingCellConfig
type PdschFlags struct {
	_pdschAggFactor string // the pdsch-AggregationFactor of PDSCH-Config, which can be n1, n2, n4 or n8
	pdschRbgCfg     string // the rbg-Size of PDSCH-Config, which can be config1 or config2
	_rbgSize        int    // the Nominal RBG size P of FDRA type 0
	pdschMcsTable   string // the mcs-Table or mcs-Table-r17 of PDSCH-Config, which can be qam64, qam256, qam1024, qam64LowSE
	pdschXOh        string // the xOverhead of PDSCH-ServingCellConfig, which can be xoh0, xoh6, xoh12, xoh18
	pdschMaxLayers  int    // the maxMIMO-Layers of PDSCH-ServingCellConfig

	pdschDmrsType      string // the dmrs-Type of DMRS-DownlinkConfig, which can be type1 or type2
	pdschDmrsAddPos    string // the dmrs-AdditionalPosition of DMRS-DownlinkConfig, which can be pos0, pos1, pos2 or pos3
	pdschMaxLength     string // the maxLength of DMRS-DownlinkConfig, which can be len1 or len2
	_dmrsPorts         []int  // DMRS antenna port(s)
	_cdmGroupsWoData   int    // number of CDM groups without data
	_numFrontLoadSymbs int    // number of front-load OFDM symbol(s) for DMRS
	_tdL               []int  // TD pattern of DMRS in a single slot
	_fdK               []int  // FD pattern of DMRS in a single PRB

	pdschPtrsEnabled     bool
	pdschPtrsTimeDensity int
	pdschPtrsFreqDensity int
	pdschPtrsReOffset    string
	_ptrsDmrsPorts       int
}

// PUSCH-config and PUSCH-ServingCellConfig
type PuschFlags struct {
	puschTxCfg                   string // the txConfig of PUSCH-Config, which can be codebook or nonCodebook
	puschCbSubset                string // the codebookSubset of PUSCH-Config, which can be fullyAndPartialAndNonCoherent, partialAndNonCoherent or nonCoherent
	puschCbMaxRankNonCbMaxLayers int    // the maxRank of PUSCH-Config in case of CB PUSCH or the maxNumberMIMO-LayersNonCB-PUSCH in case of nonCB PUSCH
	puschTp                      string // the transformPrecoder of PUSCH-Config
	_puschAggFactor              string // the pusch-AggregationFactor of PUSCH-Config, which can be n1, n2, n4 or n8
	puschRbgCfg                  string // the rbg-Size of PUSCH-Config, which can be config1 or config2
	_rbgSize                     int    // the Nominal RBG size P of FDRA type 0
	puschMcsTable                string // the mcs-Table or mcs-TableTransformPrecoder of PUSCH-Config, which can be qam64, qam256 or qam64LowSE
	puschXOh                     string // the xOverhead of PUSCH-ServingCellConfig, which can be xoh0, xoh6, xoh12, xoh18
	_puschRepType                string // pusch-RepTypeIndicatorDCI-0-1-r16 or pusch-RepTypeIndicatorDCI-0-2-r16 of PUSCH-Config, which can be typeA or typeB

	puschDmrsType         string // the dmrs-Type of DMRS-UplinkConfig, which can be type1 or type2
	puschDmrsAddPos       string // the dmrs-AdditionalPosition of DMRS-UplinkConfig, which can be pos0, pos1, pos2 or pos3
	puschMaxLength        string // the maxLength of DMRS-UplinkConfig, which can be len1 or len2
	_dmrsPorts            []int  // DMRS antenna port(s)
	_cdmGroupsWoData      int    // number of CDM groups without data
	_numFrontLoadSymbs    int    // number of front-load OFDM symbol(s) for DMRS
	_tdL                  []int  // TD pattern of DMRS in a single slot
	_tdL2                 []int  // TD pattern of DMRS in a single slot, only valid for the 2nd hop of DMRS for PUSCH when intra-slot frequency hopping is enabled
	_fdK                  []int  // FD pattern of DMRS in a single PRB
	_nonCbSri             []int  // list of SRI(s) in case of nonCodebook transmission
	_dmrsPosLBar          []int  // PUSCH DMRS positions l_bar when intra-slot FH is disabled, or l_bar of 1st hop when intra-slot FH is enabled
	_dmrsPosLBarSecondHop []int  // PUSCH DMRS positions l_bar of 2nd hop when intra-slot is enabled

	puschPtrsEnabled       bool
	puschPtrsTimeDensity   int
	puschPtrsFreqDensity   int
	puschPtrsReOffset      string
	puschPtrsMaxNumPorts   string
	puschPtrsTimeDensityTp int
	puschPtrsGrpPatternTp  string
	_numGrpsTp             int
	_samplesPerGrpTp       int
	_ptrsDmrsPorts         []int
}

// PUCCH-Config
type PucchFlags struct {
	_numSlots         string // the nrofSlots of PUCCH-FormatConfig, which can be n1/n2/n4/n8
	_interSlotFreqHop string // the interslotFrequencyHopping of PUSCH-FormatConfig, which can be enabled or disabled
	_addDmrs          bool   // the additionalDMRS of PUSCH-FormatConfig
	_simHarqAckCsi    bool   // the simultaneousHARQ-ACK-CSI of PUSCH-FormatConfig

	_pucchResId  []int    // the pucch-ResourceId of PUCCH-Resource
	_pucchFormat []string // the format of PUCCH-Resource, which can be format1/format3
	//_pucchResSetId        []int
	_pucchStartRb          []int    // the startingPRB of PUCCH-Resource
	_pucchIntraSlotFreqHop []string // the intraSlotFrequencyHopping of PUCCH-Resource, which can be enabled or disabled
	_pucchSecondHopPrb     []int    // the secondHopPRB of PUCCH-Resource
	_pucchNumRbs           []int    // the nrofPRBs of PUCCH-format1 and PUCCH-format3
	_pucchStartSymb        []int    // the startingSymbolIndex of PUCCH-format1 and PUCCH-format3
	_pucchNumSymbs         []int    // the nrofSymbols of PUCCH-format1 and PUCCH-format3

	//_dsrResId    []int
	dsrPeriod    string // the periodicityAndOffset of SchedulingRequestResourceConfig
	dsrOffset    int    // the periodicityAndOffset of SchedulingRequestResourceConfig
	_dsrPucchRes int    // the resource of SchedulingRequestResourceConfig
}

// NZP-CSI-RS resources for channel measurement, TRS and CSI-IM resource (ZP-CSI-RS)
type CsiFlags struct {
	_resSetId     []int
	_trsInfo      []string
	_resId        []int
	freqAllocRow  []string
	freqAllocBits []string
	_numPorts     []string
	_cdmType      []string
	_density      []string
	_firstSymb    []int
	//_firstSymb2   []int
	_startRb []int
	_numRbs  []int
	period   []string
	offset   []int

	_tdLoc [2]nrgrid.CsiRsLocInfo

	_csiImRePattern string
	_csiImScLoc     string
	_csiImSymbLoc   int
	_csiImStartRb   int
	_csiImNumRbs    int
	_csiImPeriod    string
	_csiImOffset    int

	_resType        string
	_repCfgType     string
	csiRepPeriod    string
	csiRepOffset    int
	_csiRepPucchRes int
	_quantity       string
}

// SRS resource
type SrsFlags struct {
	_resId            []int
	srsNumPorts       []string
	_srsNonCbPtrsPort []string
	srsNumCombs       []string
	srsCombOff        []int
	srsCs             []int
	srsStartPos       []int
	srsNumSymbs       []string
	srsRepetition     []string
	srsFreqPos        []int
	srsFreqShift      []int
	srsCSrs           []int
	srsBSrs           []int
	srsBHop           []int
	_resType          []string
	srsPeriod         []string
	srsOffset         []int
	_mSRSb            []string
	_Nb               []string
	// SRS resource set
	_resSetId       []int
	srsSetResIdList []string
	_resSetType     []string
	_usage          []string
}

// Advanced settings
type AdvancedFlags struct {
	bestSsb       int
	pdcchSlotSib1 int
	prachOccMsg1  int
	pdcchOccMsg2  int
	pdcchOccMsg4  int
	//dsrRes        int
}

type DataPerSlot struct {
	res  []int      // REs in a slot, ordering: subcarrier per symbol, then symbol per slot
	tags mapset.Set // Physical signals/channels mapped in a slot
}

//
type NrrgData struct {
	subfPerRf   int
	slotPerSubf int
	slotPerRf   int
	symbPerSlot int
	symbPerSubf int
	symbPerRf   int
	scPerRb     int
	scPerSlot   int
	scPerSubf   int
	scPerRf     int

	gridTdd   map[string][]DataPerSlot
	gridFddUl map[string][]DataPerSlot
	gridFddDl map[string][]DataPerSlot

	ssbFirstSymbs  []int
	ssbSc0Rb0      int
	coreset0Sc0Rb0 int

	coreset0NumCces    int
	coreset0RegBundles []int
	coreset0Cces       []int
}

// nrrgCmd represents the "nrrg" command
var nrrgCmd = &cobra.Command{
	Use:   "nrrg",
	Short: "NR resource grid tool",
	Long:  `CMD "nrrg" generates NR resource grid according to configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		viper.WriteConfig()
	},
}

// gridSettingCmd represents the "nrrg gridsetting" command
var gridSettingCmd = &cobra.Command{
	Use:   "gridsetting",
	Short: "",
	Long:  `CMD "nrrg gridsetting" can be used to get/set resource grid settings.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()

		// initialization
		if flags.dmrsCommon._tdL == nil {
			flags.dmrsCommon._tdL = make([][]int, 4)
		}
		if flags.dmrsCommon._fdK == nil {
			flags.dmrsCommon._fdK = make([][]int, 4)
		}
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
			flags.tdduldl._refScs = scs
			// update u_PDCCH/u_PDSCH/u_PUSCH in DCI 0_1/1_1 and Msg3 PUSCH
			u := nrgrid.Scs2Mu[flags.gridsetting._carrierScs]
			flags.uldci._muPdcch = []int{u, u}
			flags.uldci._muPusch = []int{u, u}
			flags.uldci._tdDelta = nrgrid.PuschTimeAllocMsg3K2Delta[flags.gridsetting._carrierScs]
			// update SCS of initial UL BWP and dedicated UL/DL BWP
			flags.bwp._bwpScs[DED_DL_BWP] = flags.gridsetting._carrierScs
			flags.bwp._bwpScs[INI_UL_BWP] = flags.gridsetting._carrierScs
			flags.bwp._bwpScs[DED_UL_BWP] = flags.gridsetting._carrierScs
			// get SR periodicity and offset(38.331 vh30 periodicityAndOffset and periodicityAndOffset-r17 of SchedulingRequestResourceConfig)
			fmt.Printf("Available SR periodicity: %v\n", nrgrid.SrPeriodSet[flags.gridsetting._carrierScs])
			// update TRS periodicity (2023/2/20: For simplicity, TRS is not supported!)
			fmt.Printf("Available TRS periodicity: %v\n", []string{"slots10", "slots20", "slots40", "slots80", "slots160", "slots320", "slots640"}[u:u+4])

			// update u_PDCCH/u_PDSCH for SIB1/Msg2/Msg4
			u = nrgrid.Scs2Mu[flags.gridsetting._mibCommonScs]
			flags.dldci._muPdcch = []int{u, u, u, u}
			flags.dldci._muPdsch = []int{u, u, u, u}
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
				regRed.Printf("[ERR]: Invalid carrier bandwidth for %v: carrierBw=%v\n", fr, bw)
				return
			}
			flags.gridsetting._carrierNumRbs = nrgrid.NrbFr1[carrierScsVal][idx]

			// update RB_Start and L_RB for initial UL BWP and dedicated UL/DL BWP
			flags.bwp._bwpStartRb[DED_DL_BWP] = 0
			flags.bwp._bwpNumRbs[DED_DL_BWP] = flags.gridsetting._carrierNumRbs
			flags.bwp._bwpLocAndBw[DED_DL_BWP], _ = makeRiv(flags.gridsetting._carrierNumRbs, 0, 275)
			flags.bwp._bwpStartRb[INI_UL_BWP] = 0
			flags.bwp._bwpNumRbs[INI_UL_BWP] = flags.gridsetting._carrierNumRbs
			flags.bwp._bwpLocAndBw[INI_UL_BWP], _ = makeRiv(flags.gridsetting._carrierNumRbs, 0, 275)
			flags.bwp._bwpStartRb[DED_UL_BWP] = 0
			flags.bwp._bwpNumRbs[DED_UL_BWP] = flags.gridsetting._carrierNumRbs
			flags.bwp._bwpLocAndBw[DED_UL_BWP], _ = makeRiv(flags.gridsetting._carrierNumRbs, 0, 275)

			// update number of bits of FDRA field
			bitsRaType0Dl := utils.CeilInt(float64(flags.gridsetting._carrierNumRbs) / float64(flags.pdsch._rbgSize))
			bitsRaType0Ul := utils.CeilInt(float64(flags.gridsetting._carrierNumRbs) / float64(flags.pusch._rbgSize))
			bitsRaType1Bwp0 := utils.CeilInt(math.Log2(float64(flags.gridsetting._coreset0NumRbs) * (float64(flags.gridsetting._coreset0NumRbs) + 1) / 2))
			bitsRaType1Bwp1 := utils.CeilInt(math.Log2(float64(flags.gridsetting._carrierNumRbs) * (float64(flags.gridsetting._carrierNumRbs) + 1) / 2))
			flags.dldci._fdBitsRaType0 = bitsRaType0Dl
			flags.dldci._fdBitsRaType1 = []int{}
			for i, _ := range flags.dldci._rnti {
				if i == DCI_10_SIB1 || i == DCI_10_MSG2 || i == DCI_10_MSG4 {
					flags.dldci._fdBitsRaType1 = append(flags.dldci._fdBitsRaType1, bitsRaType1Bwp0)
				} else if i == DCI_11_PDSCH {
					flags.dldci._fdBitsRaType1 = append(flags.dldci._fdBitsRaType1, bitsRaType1Bwp1)
				} else if i == DCI_10_MSGB {
					flags.dldci._fdBitsRaType1 = append(flags.dldci._fdBitsRaType1, bitsRaType1Bwp0)
				}
			}
			flags.uldci._fdBitsRaType0 = bitsRaType0Ul
			flags.uldci._fdBitsRaType1 = []int{}
			for i, _ := range flags.uldci._rnti {
				if i == RAR_UL_MSG3 || i == FBRAR_UL_MSG3 {
					flags.uldci._fdBitsRaType1 = append(flags.uldci._fdBitsRaType1, 14)
				} else if i == DCI_01_PUSCH {
					flags.uldci._fdBitsRaType1 = append(flags.uldci._fdBitsRaType1, bitsRaType1Bwp1)
				} else if i == RA_UL_MSGA {
					// For MsgA, FDRA is signalling via msgA-PUSCH-Config-r16: frequencyStartMsgA-PUSCH-r16 and nrofPRBs-PerMsgA-PO-r16
					flags.uldci._fdBitsRaType1 = append(flags.uldci._fdBitsRaType1, -1)
				}
			}

			// update frequency offset for 2nd hop when intra-slot frequency hopping is enabled for PUSCH
			// refer to 38.213 vh40
			// Table 8.3-1: Frequency offset for second hop of PUSCH transmission with frequency hopping scheduled by RAR UL grant or of Msg3 PUSCH retransmission
			// refer to 38.214 vh40
			// 6.3.1	Frequency hopping for PUSCH repetition Type A and for TB processing over multiple slots
			// For simplicity, assume that only codepoint 0 or 00 is used!
			flags.uldci._fdFreqHopOffset = []int{}
			for range flags.uldci._rnti {
				flags.uldci._fdFreqHopOffset = append(flags.uldci._fdFreqHopOffset, utils.FloorInt(float64(flags.gridsetting._carrierNumRbs)/2))
			}
		}

		// process gridsetting.dmrsTypeAPos
		if cmd.Flags().Lookup("dmrsTypeAPos").Changed {
			regGreen.Printf("[INFO]: Processing gridSetting.dmrsTypeAPos...\n")

			dmrsTypeAPos := flags.gridsetting.dmrsTypeAPos

			// validate CORESET duration
			// refer to 3GPP TS 38.211 vf80: 7.3.2.2	Control-resource set (CORESET)
			// N_CORESET_symb = 3 is supported only if the higher-layer parameter dmrs-TypeA-Position equals 3;
			if flags.gridsetting._coreset0NumSymbs == 3 && dmrsTypeAPos != "pos3" {
				fmt.Printf("coreset0NumSymbs can be 3 only if dmrs-TypeA-Position is pos3! (corest0NumSymbs=%v,dmrsTypeAPos=%v)\n", flags.gridsetting._coreset0NumSymbs, flags.gridsetting.dmrsTypeAPos)
				return
			}
			if flags.searchspace._coreset1Duration == 3 && dmrsTypeAPos != "pos3" {
				fmt.Printf("coreset1Duration can be 3 only if dmrs-TypeA-Position is pos3! (coreset1Duration=%v,dmrsTypeAPos=%v)\n", flags.searchspace._coreset1Duration, flags.gridsetting.dmrsTypeAPos)
				return
			}

			// validate TDRA of DCI 1_0/1_1
			err := validatePdsch()
			if err != nil {
				regRed.Printf("[ERR]: " + err.Error())
				return
			}

			// validate TDRA of Msg3 PUSCH scheduled by RAR Msg2
			err = validatePusch()
			if err != nil {
				regRed.Printf("[ERR]: " + err.Error())
				return
			}
		}

		// process gridsetting.pci
		if cmd.Flags().Lookup("pci").Changed {
			flags.searchspace._coreset1ShiftIndex = flags.gridsetting.pci
		}

		regGreen.Printf("[INFO]: Post-processing...\n")
		// update rach info
		err := updateRach()
		if err != nil {
			regRed.Printf("[ERR]: %v\n", err.Error())
			return
		}

		// update n_CRB_SSB/k_SSB
		updateKSsbAndNCrbSsb()

		// validate CORESET0
		err = validateCoreset0()
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

		// validate search space
		err = validateSearchSpace()
		if err != nil {
			regRed.Printf("[ERR]: %s\n", err.Error())
			return
		}

		// validate PUCCH
		err = validatePucch()
		if err != nil {
			regRed.Printf("[ERR]: %s\n", err.Error())
			return
		}

		// validate CSI-RS
		err = validateCsi()
		if err != nil {
			regRed.Printf("[ERR]: %s\n", err.Error())
			return
		}

		laPrint(cmd, args)
		viper.WriteConfig()

		// trigger NRRG simulation
		regGreen.Printf("[INFO]: Init NRRG data...\n")
		err = initNrrgData()
		if err != nil {
			regRed.Printf("[ERR]: %s\n", err.Error())
			return
		}

		// trigger NRRG simulation
		regGreen.Printf("[INFO]: Start 5GNR simulation...\n")

		hsfn := 0
		sfn := flags.gridsetting._sfn
		slot := 0

		// DL always-on transmission
		regYellow.Printf("[5GNR SIM]Init always-on-transmission(SSB/PDCCH/SIB1) @ [HSFN=%d, SFN=%d, Slot=%d]\n", hsfn, sfn, slot)
		err = alwaysOnTr(hsfn, sfn, slot)
		if err != nil {
			regRed.Printf("[ERR]: %s\n", err.Error())
			return
		}

		/*
			// receiving SIB1
			regYellow.Printf("[5GNR SIM]UE recv SSB/SIB1 @ [HSFN=%d, SFN=%d]\n", hsfn, sfn)
			hsfn, sfn, slot, err = recvSib1(hsfn, sfn)
			if err != nil {
				regRed.Printf("[ERR]: %s\n", err.Error())
				return
			}

			// sending Msg1(PRACH)
			regYellow.Printf("[5GNR SIM]UE send PRACH(Msg1) @ [HSFN=%d, SFN=%d, Slot=%d]\n", hsfn, sfn, slot)
			hsfn, sfn, slot, err = sendMsg1(hsfn, sfn, slot)
			if err != nil {
				regRed.Printf("[ERR]: %s\n", err.Error())
				return
			}

			// monitoring PDCCH for Msg2(RAR)
			regYellow.Printf("[5GNR SIM]UE recv PDCCH(DCI 1_0, RA-RNTI) @ [HSFN=%d, SFN=%d, Slot=%d]\n", hsfn, sfn, slot)
			hsfn, sfn, slot, err = monitorPdcch(hsfn, sfn, slot, "dci10", "RA-RNTI")
			if err != nil {
				regRed.Printf("[ERR]: %s\n", err.Error())
				return
			}

			// receiving Msg2(RAR)
			regYellow.Printf("[5GNR SIM]UE recv RAR(Msg2) @ [HSFN=%d, SFN=%d, Slot=%d]\n", hsfn, sfn, slot)
			hsfn, sfn, slot, err = recvMsg2(hsfn, sfn, slot)
			if err != nil {
				regRed.Printf("[ERR]: %s\n", err.Error())
				return
			}

			// sending Msg3 PUSCH
			regYellow.Printf("[5GNR SIM]UE send Msg3 @ [HSFN=%d, SFN=%d, Slot=%d]\n", hsfn, sfn, slot)
			hsfn, sfn, slot, err = sendMsg3(hsfn, sfn, slot)
			if err != nil {
				regRed.Printf("[ERR]: %s\n", err.Error())
				return
			}

			// monitoring PDCCH for Msg4
			regYellow.Printf("[5GNR SIM]UE recv PDCCH(DCI 1_0, TC-RNTI) @ [HSFN=%d, SFN=%d, Slot=%d]\n", hsfn, sfn, slot)
			hsfn, sfn, slot, err = monitorPdcch(hsfn, sfn, slot, "dci10", "TC-RNTI")
			if err != nil {
				regRed.Printf("[ERR]: %s\n", err.Error())
				return
			}

			// receiving Msg4
			regYellow.Printf("[5GNR SIM]UE recv Msg4 @ [HSFN=%d, SFN=%d, Slot=%d]\n", hsfn, sfn, slot)
			hsfn, sfn, slot, err = recvMsg4(hsfn, sfn, slot)
			if err != nil {
				regRed.Printf("[ERR]: %s\n", err.Error())
				return
			}

			// sending HARQ-AN of Msg4(PUCCH)
			regYellow.Printf("[5GNR SIM]UE send PUCCH(Msg4 HARQ) @ [HSFN=%d, SFN=%d, Slot=%d]\n", hsfn, sfn, slot)
			hsfn, sfn, slot, err = sendPucch(hsfn, sfn, slot, true, false, false, "common") //harq=True, sr=False, csi=False, pucchResSet='common'
			if err != nil {
				regRed.Printf("[ERR]: %s\n", err.Error())
				return
			}
		*/

		// UL always-on transmission (pCSI/SRS)
		regYellow.Printf("[5GNR SIM]Init always-on-transmission(periodic CSI-RS/SRS) @ [HSFN=%d, SFN=%d, Slot=%d]\n", hsfn, sfn, slot)
		err = alwaysOnTr(hsfn, sfn, slot)
		if err != nil {
			regRed.Printf("[ERR]: %s\n", err.Error())
			return
		}

		/*
			self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]Start 5GNR simulation</b></font>')
			nrGrid = NgNrGrid(self.ngwin, self.args)
			hsfn = 0
			sfn = int(self.args['mib']['sfn'])

			self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]Init always-on-transmission(SSB/PDCCH/SIB1) @ [HSFN=%d, SFN=%d, Slot=%d]</b></font>' % (hsfn, sfn, 0))
			nrGrid.alwaysOnTr(hsfn, sfn, 0)

			# receiving SIB1
			self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]UE recv SSB/SIB1 @ [HSFN=%d, SFN=%d]</b></font>' % (hsfn, sfn))
			hsfn, sfn, slot = nrGrid.recvSib1(hsfn, sfn)

			# sending Msg1(PRACH)
			if hsfn is not None and sfn is not None and slot is not None:
				self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]UE send PRACH(Msg1) @ [HSFN=%d, SFN=%d, Slot=%d]</b></font>' % (hsfn, sfn, slot))
				hsfn, sfn, slot = nrGrid.sendMsg1(hsfn, sfn, slot)

			# monitoring PDCCH for Msg2
			if hsfn is not None and sfn is not None and slot is not None:
				self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]UE recv PDCCH(DCI 1_0, RA-RNTI) @ [HSFN=%d, SFN=%d, Slot=%d]</b></font>' % (hsfn, sfn, slot))
				hsfn, sfn, slot = nrGrid.monitorPdcch(hsfn, sfn, slot, dci='dci10', rnti='ra-rnti')

			# receiving Msg2(RAR)
			if hsfn is not None and sfn is not None and slot is not None:
				self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]UE recv RAR(Msg2) @ [HSFN=%d, SFN=%d, Slot=%d]</b></font>' % (hsfn, sfn, slot))
				hsfn, sfn, slot = nrGrid.recvMsg2(hsfn, sfn, slot)

			# sending Msg3(PUSCH)
			if hsfn is not None and sfn is not None and slot is not None:
				self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]UE send Msg3 @ [HSFN=%d, SFN=%d, Slot=%d]</b></font>' % (hsfn, sfn, slot))
				hsfn, sfn, slot = nrGrid.sendMsg3(hsfn, sfn, slot)

			# monitoring PDCCH for Msg4
			if hsfn is not None and sfn is not None and slot is not None:
				self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]UE recv PDCCH(DCI 1_0, TC-RNTI) @ [HSFN=%d, SFN=%d, Slot=%d]</b></font>' % (hsfn, sfn, slot))
				hsfn, sfn, slot = nrGrid.monitorPdcch(hsfn, sfn, slot, dci='dci10', rnti='tc-rnti')

			# receiving Msg4
			if hsfn is not None and sfn is not None and slot is not None:
				self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]UE recv Msg4 @ [HSFN=%d, SFN=%d, Slot=%d]</b></font>' % (hsfn, sfn, slot))
				hsfn, sfn, slot = nrGrid.recvMsg4(hsfn, sfn, slot)

			# sending Msg4 HARQ feedback(PUCCH)
			if hsfn is not None and sfn is not None and slot is not None:
				self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]UE send PUCCH(Msg4 HARQ) @ [HSFN=%d, SFN=%d, Slot=%d]</b></font>' % (hsfn, sfn, slot))
				hsfn, sfn, slot = nrGrid.sendPucch(hsfn, sfn, slot, harq=True, sr=False, csi=False, pucchResSet='common')

			# triggering always-on-transmission of periodic CSI-RS/SRS
			if hsfn is not None and sfn is not None and slot is not None:
				self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]Init always-on-transmission(periodic CSI-RS/SRS) @ [HSFN=%d, SFN=%d, Slot=%d]</b></font>' % (hsfn, sfn, slot))
				nrGrid.alwaysOnTr(hsfn, sfn, slot)

			# sending PUCCH for DSR
			self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]UE send DSR @ [HSFN=%d, SFN=%d, Slot=%d]</b></font>' % (hsfn, sfn, slot))
			retry = 0  # retry 16 frames at most, which is 160ms while max DSR period is 80slots
			while True:
				_hsfn, _sfn, _slot = nrGrid.sendPucch(hsfn, sfn, slot, harq=False, sr=True, csi=False, pucchResSet='dedicated')
				retry += 1
				if _slot is None and retry < 16:
					hsfn, sfn = nrGrid.incSfn(hsfn, sfn, 1)
					slot = 0
					nrGrid.alwaysOnTr(hsfn, sfn, 0)
					if nrGrid.error:
						break
				else:
					hsfn, sfn, slot = _hsfn, _sfn, _slot
					break

			# monitoring PDCCH for PUSCH(Msg5)
			# start pdcch monitoring at next slot since last slot of DSR is 'U' or 'S'
			slotsPerRf = 10 * 2 ** {'15KHz':0, '30KHz':1, '60KHz':2, '120KHz':3, '240KHz':4}[self.args['dedUlBwp']['scs']]
			hsfn, sfn, slot = nrGrid.incSlot(hsfn, sfn, slot, slotsPerRf, 1)
			hsfn, sfn, slot = nrGrid.monitorPdcch(hsfn, sfn, slot, dci='dci01', rnti='c-rnti')

			# sending PUSCH(Msg5)
			# hsfn, sfn = nrGrid.sendPusch(hsfn, sfn)

			# monitoring PDCCH for normal PDSCH
			# if hsfn is not None and sfn is not None and slot is not None:
			#     hsfn, sfn, slot = nrGrid.monitorPdcch(hsfn, sfn, dci='dci11', rnti='c-rnti')

			# receiving PDSCH
			# hsfn, sfn = nrGrid.recvPdsch(hsfn, sfn)

			# sending PDSCH HARQ feedback(PUCCH), together with CSI
			# hsfn, sfn = nrGrid.sendPucch(hsfn, sfn)

			# export grid to excel
			if not nrGrid.error:
				if self.ngwin.enableDebug:
					# Don't waste time for waiting exportToExcel to finish!
					self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]Exporting to excel skipped</b></font>')
					pass
				else:
					self.ngwin.logEdit.append('<font color=green><b>[5GNR SIM]Exporting to excel, please wait</b></font>')
					nrGrid.exportToExcel()
		*/
	},
}

func initNrrgData() error {
	rgd.subfPerRf = 10
	rgd.slotPerSubf = int(math.Exp2(float64(nrgrid.Scs2Mu[flags.gridsetting.scs])))
	rgd.slotPerRf = rgd.slotPerSubf * rgd.subfPerRf
	rgd.symbPerSlot = 14
	rgd.symbPerSubf = rgd.symbPerSlot * rgd.slotPerSubf
	rgd.symbPerRf = rgd.symbPerSlot * rgd.slotPerRf
	rgd.scPerRb = 12
	rgd.scPerSlot = rgd.scPerRb * rgd.symbPerSlot
	rgd.scPerSubf = rgd.scPerSlot * rgd.slotPerSubf
	rgd.scPerRf = rgd.scPerSlot * rgd.slotPerRf

	if flags.gridsetting._duplexMode == "TDD" {
		rgd.gridTdd = make(map[string][]DataPerSlot)
	} else {
		rgd.gridFddUl = make(map[string][]DataPerSlot)
		rgd.gridFddDl = make(map[string][]DataPerSlot)
	}

	var s1, s3 []int
	var s2 int
	if flags.gridsetting._ssbPattern == "Case A" && flags.gridsetting._ssbScs == "15KHz" {
		s1 = []int{2, 8}
		s2 = 14
		if !flags.gridsetting._unlicensed {
			if flags.gridsetting._maxDlFreq <= 3000 {
				s3 = []int{0, 1}
			} else {
				s3 = []int{0, 1, 2, 3}
			}
		} else {
			s3 = []int{0, 1, 2, 3, 4}
		}
	} else if flags.gridsetting._ssbPattern == "Case B" && flags.gridsetting._ssbScs == "30KHz" {
		s1 = []int{4, 8, 16, 20}
		s2 = 28
		if flags.gridsetting._maxDlFreq <= 3000 {
			s3 = []int{0}
		} else {
			s3 = []int{0, 1}
		}
	} else if flags.gridsetting._ssbPattern == "Case C" && flags.gridsetting._ssbScs == "30KHz" {
		s1 = []int{2, 8}
		s2 = 14
		if !flags.gridsetting._unlicensed {
			if flags.gridsetting._duplexMode == "FDD" {
				if flags.gridsetting._maxDlFreq <= 3000 {
					s3 = []int{0, 1}
				} else {
					s3 = []int{0, 1, 2, 3}
				}
			} else if flags.gridsetting._duplexMode == "TDD" {
				if flags.gridsetting._maxDlFreq < 1880 {
					s3 = []int{0, 1}
				} else {
					s3 = []int{0, 1, 2, 3}
				}
			}
		} else {
			s3 = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		}
	} else if flags.gridsetting._ssbPattern == "Case D" && flags.gridsetting._ssbScs == "120KHz" {
		s1 = []int{4, 8, 16, 20}
		s2 = 28
		s3 = []int{0, 1, 2, 3, 5, 6, 7, 8, 10, 11, 12, 13, 15, 16, 17, 18}
	} else if flags.gridsetting._ssbPattern == "Case E" && flags.gridsetting._ssbScs == "240KHz" {
		s1 = []int{8, 12, 16, 20, 32, 36, 40, 44}
		s2 = 56
		s3 = []int{0, 1, 2, 3, 5, 6, 7, 8}
	} else if ((flags.gridsetting._ssbPattern == "Case F" && flags.gridsetting._ssbScs == "480KHz") || (flags.gridsetting._ssbPattern == "Case G" && flags.gridsetting._ssbScs == "960KHz")) && flags.gridsetting._freqRange == "FR2-2" {
		s1 = []int{2, 9}
		s2 = 14
		s3 = utils.PyRange(0, 32, 1)
	}

	for _, i := range s1 {
		for _, j := range s3 {
			rgd.ssbFirstSymbs = append(rgd.ssbFirstSymbs, i+s2*j)
		}
	}
	sort.Ints(rgd.ssbFirstSymbs)

	rmsiScs, _ := strconv.Atoi(flags.gridsetting._mibCommonScs[:len(flags.gridsetting._mibCommonScs)-3])
	rgd.ssbSc0Rb0 = (flags.gridsetting._nCrbSsb*12*int(flags.gridsetting._nCrbSsbScs)+flags.gridsetting._kSsb*int(flags.gridsetting._kSsbScs))/rmsiScs - flags.gridsetting._offsetToCarrier*12
	rgd.coreset0Sc0Rb0 = rgd.ssbSc0Rb0 - flags.gridsetting._coreset0Offset*12
	fmt.Printf("offsetToCarrier=%v, nCrbSsb=%v(SCS=%.0fKHz), kSsb=%v(SCS=%.0fKHz) -> ssbSc0Rb0=%v\n", flags.gridsetting._offsetToCarrier, flags.gridsetting._nCrbSsb, flags.gridsetting._nCrbSsbScs, flags.gridsetting._kSsb, flags.gridsetting._kSsbScs, rgd.ssbSc0Rb0)
	fmt.Printf("coreset0Offset=%v -> coreset0Sc0Rb0=%v\n", flags.gridsetting._coreset0Offset, rgd.coreset0Sc0Rb0)

	rgd.coreset0NumCces = flags.gridsetting._coreset0NumSymbs * flags.gridsetting._coreset0NumRbs / 6
	if flags.gridsetting._css0AggLevel > rgd.coreset0NumCces {
		return errors.New(fmt.Sprintf("Invalid configurations of CSS0/CORESET0: aggregation level=%v while total number of CCEs=%v!", flags.gridsetting._css0AggLevel, rgd.coreset0NumCces))
	}
	//TODO

	return nil
}

func alwaysOnTr(hsfn, sfn, slot int) error {

	return nil
}

func updateRach() error {
	regYellow.Printf("-->calling updateRach\n")

	var p *nrgrid.RachInfo
	var exist bool
	if flags.gridsetting._freqRange == "FR1" {
		if flags.gridsetting._duplexMode == "FDD" {
			p, exist = nrgrid.RaCfgFr1FddSUl[flags.rach.prachConfId]
		} else {
			p, exist = nrgrid.RaCfgFr1Tdd[flags.rach.prachConfId]
		}
	} else {
		p, exist = nrgrid.RaCfgFr2Tdd[flags.rach.prachConfId]
	}

	if !exist {
		return errors.New(fmt.Sprintf("Invalid configurations for PRACH: %v,%v with prach-ConfigurationIndex=%v\n",
			flags.gridsetting._freqRange, flags.gridsetting._duplexMode, flags.rach.prachConfId))
	}

	fmt.Printf("RACH Info: %v\n", *p)

	flags.rach._raFormat = p.Format
	flags.rach._raX = p.X
	flags.rach._raY = p.Y
	flags.rach._raSubfNumFr1SlotNumFr2 = p.SubfNumFr1SlotNumFr2
	flags.rach._raStartingSymb = p.StartingSymb
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

	if utils.ContainsStr([]string{"0", "1", "2", "3"}, flags.rach._raFormat) {
		flags.rach._raLen = 839
		if flags.rach._raFormat == "3" {
			flags.rach._msg1Scs = "5KHz"
		} else {
			flags.rach._msg1Scs = "1.25KHz"
		}
	} else {
		// refer to 38.211 vh40 Table 6.3.3.1-2
		// L_RA=1151/571 for NR-U are not supported!
		flags.rach._raLen = 139
		flags.rach._msg1Scs = flags.gridsetting.scs
	}

	key := fmt.Sprintf("%v_%v_%v", flags.rach._raLen, flags.rach._msg1Scs[:len(flags.rach._msg1Scs)-3], flags.gridsetting.scs[:len(flags.gridsetting.scs)-3])
	p2, exist2 := nrgrid.NumRbsRaAndKBar[key]
	if !exist2 {
		return errors.New(fmt.Sprintf("Invalid key(=%v) when referring NumRbsRaAndKBar!\n", key))
	}
	flags.rach._raNumRbs = p2[0]
	flags.rach._raKBar = p2[1]

	// refer to 38.211 vh40
	// 5.3.2	OFDM baseband signal generation for PRACH
	// - n_RA_slot is given by
	//   - if deltaf_RA is {30, 120}kHz and either of "Number of PRACH slots within a subframe" in Tables 6.3.3.2-2 to 6.3.3.2-3 or "Number of PRACH slots within a 60 kHz slot" in Table 6.3.3.2-4 is equal to 1, then n_RA_slot = 0, otherwise n_RA_slot = {0,1}
	if flags.rach._raNumSlotsPerSubfFr1Per60KSlotFr2 == 2 && flags.rach._msg1Scs != "30KHz" && flags.rach._msg1Scs != "120KHz" {
		return errors.New(fmt.Sprintf("Msg1 SCS must be 30KHz for FR1 or 120KHz for FR2-1 when NumSlotsPerSubfFr1Per60KSlotFr2 = 2!\n"))
	}

	// refer to 38.331 vh30 totalNumberOfRA-Preambles
	// Total number of preambles used for contention based and contention free 4-step or 2-step random access in the RACH resources defined in RACH-ConfigCommon, excluding preambles used for other purposes (e.g. for SI request). If the field is absent, all 64 preambles are available for RA.
	// The setting should be consistent with the setting of ssb-perRACH-OccasionAndCB-PreamblesPerSSB, i.e. it should be a multiple of the number of SSBs per RACH occasion.
	if nrgrid.SsbPerRachOccasion2Float[flags.rach.ssbPerRachOccasion] > 1 && flags.rach.totNumPreambs%int(nrgrid.SsbPerRachOccasion2Float[flags.rach.ssbPerRachOccasion]) != 0 {
		return errors.New(fmt.Sprintf("The totalNumberOfRA-Preambles should be a multiple of the number of SSBs per RACH occasion. (totNumPreambs=%v, ssbPerRachoccasion=%v)\n", flags.rach.totNumPreambs, flags.rach.ssbPerRachOccasion))
	}

	return nil
}

// convert ARFCN to F_REF(MHz) (refer to 38.104 vh80)
//  Table 5.4.2.1-1: NR-ARFCN parameters for the global frequency raster
func arfcn2Fref(arfcn int, maxFreq int) float64 {
	if maxFreq < 3000 {
		return float64(arfcn) * 0.005
	} else if maxFreq < 24250 {
		return 3000 + 0.015*(float64(arfcn)-600000)
	} else {
		return 24250.08 + 0.06*(float64(arfcn)-2016667)
	}
}

// convert GSCN to SS_REF(MHz) (refer to 38.104 vh80)
//  Table 5.4.3.1-1: GSCN parameters for the global frequency raster
func gscn2Ssref(gscn int, maxFreq int) float64 {
	if maxFreq < 3000 {
		N := math.Floor((float64(gscn) + 1.5) / 3)
		M := math.Mod(float64(gscn)+1.5, 3) * 2
		ssRef := 1.2*N + 0.05*M

		fmt.Printf("GSCN=%v, N=%v, M=%v, SS_REF=%vMHz\n", gscn, N, M, ssRef)
		return ssRef
	} else if maxFreq < 24250 {
		N := gscn - 7499
		return 3000 + 1.44*float64(N)
	} else {
		N := gscn - 22256
		return 24250.08 + 17.28*float64(N)
	}
}

// calculate N_CRB_SSB and k_SSB given GSCN and DL ARFCN
func updateKSsbAndNCrbSsb() error {
	regYellow.Printf("-->calling updateKSsbAndNCrbSsb\n")

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
	ssFreqSc0Rb0 := ssFreq - 120*ssbScs/1000

	dlFreq := arfcn2Fref(flags.gridsetting.dlArfcn, flags.gridsetting._maxDlFreq)
	var dlFreqPointA float64
	if flags.gridsetting._carrierNumRbs%2 == 0 {
		dlFreqPointA = dlFreq - 12*float64(flags.gridsetting._carrierNumRbs/2)*carrierScs/1000
	} else {
		dlFreqPointA = dlFreq - 12*(math.Floor(float64(flags.gridsetting._carrierNumRbs)/2)+6)*carrierScs/1000
	}

	nCrbSsb := math.Floor((ssFreqSc0Rb0 - dlFreqPointA) / (12 * flags.gridsetting._nCrbSsbScs / 1000))
	kSsb := (ssFreqSc0Rb0 - dlFreqPointA - 12*flags.gridsetting._nCrbSsbScs/1000*nCrbSsb) / (flags.gridsetting._kSsbScs / 1000)

	fmt.Printf("%v: nCrbSsb SCS=%.0fKHz, kSsb SCS=%.0fKHz\n", flags.gridsetting._freqRange, flags.gridsetting._nCrbSsbScs, flags.gridsetting._kSsbScs)
	fmt.Printf("ssFreq=%vMHz, ssFreqSc0Rb0=%vMHz, dlFreq=%vMHz, dlFreqPointA=%vMHz, nCrbSsb=%v, kSsb=%v\n",
		ssFreq, ssFreqSc0Rb0, dlFreq, dlFreqPointA, nCrbSsb, kSsb)

	flags.gridsetting._nCrbSsb = int(nCrbSsb)
	flags.gridsetting._kSsb = int(math.Ceil(kSsb))

	return nil
}

func validateCoreset0() error {
	regYellow.Printf("-->calling validateCoreset0\n")

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
		minBw = utils.MaxInt([]int{flags.gridsetting._coreset0NumRbs, flags.gridsetting._coreset0Offset + 20*ssbScsVal/rmsiScsVal})
	} else {
		minBw = flags.gridsetting._coreset0NumRbs - flags.gridsetting._coreset0Offset
	}
	if numRbsRmsiScs < minBw {
		return errors.New(fmt.Sprintf("Invalid configurations for CORESET0: numRbsRmsiScs=%v, minBw=%v(coreset0NumRbs=%v,offset=%v)\n", numRbsRmsiScs, minBw, flags.gridsetting._coreset0NumRbs, flags.gridsetting._coreset0Offset))
	}

	// validate coreste0NumSymbs against dmrs-pointA-Position
	// refer to 3GPP TS 38.211 vf80: 7.3.2.2	Control-resource set (CORESET)
	// N_CORESET_symb = 3 is supported only if the higher-layer parameter dmrs-TypeA-Position equals 3;
	if flags.gridsetting._coreset0NumSymbs == 3 && flags.gridsetting.dmrsTypeAPos != "pos3" {
		return errors.New(fmt.Sprintf("coreset0NumSymbs can be 3 only if dmrs-TypeA-Position is pos3! (corest0NumSymbs=%v,dmrsTypeAPos=%v)\n", flags.gridsetting._coreset0NumSymbs, flags.gridsetting.dmrsTypeAPos))
	}

	// update info of initial dl bwp
	if flags.gridsetting._coreset0Offset >= 0 {
		upper := utils.MinInt([]int{numRbsRmsiScs - flags.gridsetting._coreset0NumRbs, numRbsRmsiScs - (flags.gridsetting._coreset0NumRbs + 20*ssbScsVal/rmsiScsVal)})
		fmt.Printf("Available RB_Start for Initial DL BWP: [%v..%v]\n", 0, upper)
	} else {
		upper := utils.MinInt([]int{numRbsRmsiScs - flags.gridsetting._coreset0NumRbs, numRbsRmsiScs - (flags.gridsetting._coreset0NumRbs + 20*ssbScsVal/rmsiScsVal)})
		fmt.Printf("Available RB_Start for Initial DL BWP: [%v..%v]\n", -flags.gridsetting._coreset0Offset, upper)
	}
	fmt.Printf("Available L_RBs for Initial DL BWP: [%v]\n", flags.gridsetting._coreset0NumRbs)

	return nil
}

func validateCss0() error {
	regYellow.Printf("-->calling validateCss0\n")

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

func validateSearchSpace() error {
	regYellow.Printf("-->calling validateSearchSpace\n")

	// validate CORESET1
	crbStart := flags.searchspace.coreset1StartCrb
	numRbs := flags.searchspace.coreset1NumRbs
	cceRegMapping := flags.searchspace.coreset1CceRegMappingType
	duration := flags.searchspace._coreset1Duration
	L, _ := strconv.Atoi(flags.searchspace.coreset1RegBundleSize[1:])

	if crbStart%6 == 0 && numRbs%6 == 0 {
		fdres := []byte(fmt.Sprintf("%045b", 0))
		// refer to 38.213 vh40
		// 10.1	UE procedure for determining physical downlink control channel assignment
		// ...if a CORESET is not associated with any search space set configured with freqMonitorLocations, the bits of the bitmap have a one-to-one mapping with non-overlapping groups of 6 consecutive PRBs, in ascending order of the PRB index in the DL BWP bandwidth of N_BWP_RB PRBs with starting common RB position N_start_BWP, where the first common RB of the first group of 6 PRBs has common RB index 6*ceil(N_start_BWP/6) if rb-Offset is not provided...
		rb0grp0 := utils.CeilInt(6 * float64(flags.bwp._bwpStartRb[DED_DL_BWP]) / 6)
		bit0 := (crbStart - rb0grp0) / 6
		nbits := numRbs / 6
		for i := 0; i < nbits; i++ {
			fdres[bit0+i] = '1'
		}
		flags.searchspace._coreset1FdRes = string(fdres)
	} else {
		return errors.New(fmt.Sprintf("Both coreset1StartCrb and coreset1NumRbs must be multiples of 6!"))
	}

	if cceRegMapping == "nonInterleaved" {
		flags.searchspace.coreset1RegBundleSize = "n6"
	} else {
		if (duration == 1 && !utils.ContainsInt([]int{2, 6}, L)) || (utils.ContainsInt([]int{2, 3}, duration) && !utils.ContainsInt([]int{duration, 6}, L)) {
			return errors.New(fmt.Sprintf("For interleaved CCE-to-REG mapping, L = 2/6 for coreset1Duration=1 and L = duration/6 for coreset1Duration=2/3.(cceRegMapping=%v, duration=%v, L=%v)\n", cceRegMapping, duration, L))
		}
	}

	// TODO: validate type3 CSS/USS MonitoringSymbolWithinSlot
	// refer to 38.213 vh40
	// 10.1	UE procedure for determining physical downlink control channel assignment
	// If the monitoringSymbolsWithinSlot indicates to a UE to monitor PDCCH in a subset of up to three consecutive symbols that are same in every slot where the UE monitors PDCCH for all search space sets, the UE does not expect to be configured with a PDCCH SCS other than 15 kHz if the subset includes at least one symbol after the third symbol.
	// A UE does not expect to be provided a first symbol and a number of consecutive symbols for a CORESET that results to a PDCCH candidate mapping to symbols of different slots.
	// A UE does not expect any two PDCCH monitoring occasions on an active DL BWP, for a same search space set or for different search space sets, in a same CORESET to be separated by a non-zero number of symbols that is smaller than the CORESET duration.

	return nil
}

func validatePucch() error {
	regYellow.Printf("-->calling validatePucch\n")

	if flags.pucch._interSlotFreqHop == "enabled" {
		for _, v := range flags.pucch._pucchIntraSlotFreqHop {
			if v == "enabled" {
				return errors.New(fmt.Sprintf("For long PUCCH over multiple slots, the intra and inter slot frequency hopping cannot be enabled at the same time for a UE."))
			}
		}
	}

	flags.csi._csiRepPucchRes = 1

	return nil
}

func validateCsi() error {
	regYellow.Printf("-->calling validateCsi\n")

	if len(flags.csi._resId) != 2 {
		return errors.New(fmt.Sprintf("Only two NZP-CSI-RS resources can be configured, one for CSI report, and the other for TRS."))
	}

	// update nzpCsiRsInfo
	for i := range flags.csi._resId {
		row := flags.csi.freqAllocRow[i]
		irow, _ := strconv.Atoi(row[3:])
		bits := flags.csi.freqAllocBits[i]
		if !((row == "row1" && len(bits) == 4) || (row == "row2" && len(bits) == 12) || (row == "row4" && len(bits) == 3) || (!utils.ContainsStr([]string{"row1", "row2", "row4"}, row) && len(bits) == 6)) {
			return errors.New(fmt.Sprintf("[CSI-RS resourceId=%v] Invalid length of freqAllocBits(=%v) for freqAllocRow=%v!", flags.csi._resId[i], bits, row))
		}

		ports, _ := strconv.Atoi(flags.csi._numPorts[i][1:])
		density := map[string]string{"evenPRBs": "0.5", "oddPRBs": "0.5", "one": "1", "three": "3"}[flags.csi._density[i]]
		key := fmt.Sprintf("%v_%v_%v", ports, density, flags.csi._cdmType[i])
		p, exist := nrgrid.CsiRsLoc[key]
		if !exist {
			return errors.New(fmt.Sprintf("Invalid key(=%v) when referring CsiRsLoc!", key))
		} else {
			for _, v := range p {
				if v.Row == irow {
					fmt.Printf("NZP-CSI-RS Info(resourceId=%v): %v\n", flags.csi._resId[i], v)
					flags.csi._tdLoc[i] = v
				}
			}
		}
	}

	// validate TRS
	for i, v := range flags.csi._trsInfo {
		if v == "true" {
			if flags.gridsetting._freqRange == "FR1" {
				if !utils.ContainsInt([]int{4, 5, 6}, flags.csi._firstSymb[i]) {
					return errors.New(fmt.Sprintf("Only time-domain locations (4,8)/(5,9)/(6,10) are supported for TRS with FR1!"))
				}
			} else {
				if !utils.ContainsInt(utils.PyRange(0, 10, 1), flags.csi._firstSymb[i]) {
					return errors.New(fmt.Sprintf("Only time-domain locations (0,4)~(9,13) are supported for TRS with FR2!"))
				}
			}
		}
	}

	// update CSI-IM resource
	flags.csi._csiImPeriod = flags.csi.period[0]
	flags.csi._csiImOffset = flags.csi.offset[0]
	flags.csi._csiImStartRb = flags.csi._startRb[0]
	flags.csi._csiImNumRbs = flags.csi._numRbs[0]
	if flags.csi._numPorts[0] == "p4" {
		if flags.csi.freqAllocRow[0] == "row4" {
			flags.csi._csiImRePattern = "pattern1"
			flags.csi._csiImScLoc = fmt.Sprintf("s%v", (flags.csi._tdLoc[0].Ki[0]+4)%12)
			flags.csi._csiImSymbLoc = flags.csi._tdLoc[0].Li[0]
		} else {
			flags.csi._csiImRePattern = "pattern0"
			flags.csi._csiImScLoc = fmt.Sprintf("s%v", (flags.csi._tdLoc[0].Ki[0]+2)%12)
			flags.csi._csiImSymbLoc = flags.csi._tdLoc[0].Li[0]
		}
	}

	return nil
}

// calculate RIV (refer to 38.214 vh40)
//  5.1.2.2.2	Downlink resource allocation type 1
func makeRiv(L_RBs, RB_start, N_BWP_size int) (int, error) {
	if L_RBs < 1 || L_RBs > (N_BWP_size-RB_start) {
		return -1, errors.New(fmt.Sprintf("Invalid combination of L_RBs=%d, RB_start=%d, N_BWP_size=%d.\n", L_RBs, RB_start, N_BWP_size))
	}

	var riv int
	if (L_RBs - 1) <= int(math.Floor(float64(N_BWP_size)/2)) {
		riv = N_BWP_size*(L_RBs-1) + RB_start
	} else {
		riv = N_BWP_size*(N_BWP_size-L_RBs+1) + (N_BWP_size - 1 - RB_start)
	}

	return riv, nil
}

// parse RIV (refer to 38.214 vh40)
//  5.1.2.2.2	Downlink resource allocation type 1
func parseRiv(riv, N_BWP_size int) ([]int, error) {
	div := riv / N_BWP_size
	rem := riv % N_BWP_size

	L_RBs := []int{div + 1, N_BWP_size + 1 - div}
	RB_start := []int{rem, N_BWP_size - 1 - rem}
	if L_RBs[0] >= 1 && L_RBs[0] <= (N_BWP_size-RB_start[0]) && L_RBs[0] <= utils.FloorInt(float64(N_BWP_size)/2) {
		return []int{L_RBs[0], RB_start[0]}, nil
	} else if L_RBs[1] >= 1 && L_RBs[1] <= (N_BWP_size-RB_start[1]) && L_RBs[1] > utils.FloorInt(float64(N_BWP_size)/2) {
		return []int{L_RBs[1], RB_start[1]}, nil
	} else {
		regRed.Printf("[ERR]: Fail to parse RIV, where RIV=%v, N_BWP_size=%v.\n", riv, N_BWP_size)
		return []int{-1, -1}, errors.New(fmt.Sprintf("Fail to parse RIV, where RIV=%v, N_BWP_size=%v.\n", riv, N_BWP_size))
	}
}

// validatePdsch validates the "Time domain resource assignment" field of DCI 1_0/1_1, updates associated DMRS, and calculate TBS.
func validatePdsch() error {
	regYellow.Printf("-->calling validatePdsch\n")

	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1-1: Valid S and L combinations
	// Note 1:	S = 3 is applicable only if dmrs-TypeA-Position = 3
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-1: Applicable PDSCH time domain resource allocation for DCI formats 1_0, 1_1, 4_0, 4_1 and 4_2
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-1A: Applicable PDSCH time domain resource allocation for DCI format 1_2
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-2: Default PDSCH time domain resource allocation A for normal CP
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-3: Default PDSCH time domain resource allocation A for extended CP
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-4: Default PDSCH time domain resource allocation B
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-5: Default PDSCH time domain resource allocation C
	for i, _ := range flags.dldci._rnti {
		if i == DCI_10_SIB1 || i == DCI_10_MSG2 || i == DCI_10_MSG4 {
			// validate TDRA
			err := validateDci10PdschTdRa(i)
			if err != nil {
				return err
			}

			// update TBS info
			err = updateDci10PdschTbs(i)
			if err != nil {
				return err
			}

		} else if i == DCI_11_PDSCH {
			// validate TDRA
			err := validateDci11PdschTdRa()
			if err != nil {
				return err
			}

			// validate 'Antenna port(s)' and update TBS
			err = validateDci11PdschAntPorts()
			if err != nil {
				return err
			}
		} else if i == DCI_10_MSGB {
			// TODO: MSGB transmission scheduled by DCI 1_0 with MSGB-RNTI for two-steps CBRA
		}
	}

	return nil
}

// validateDci10PdschTdRa validates the "Time domain resource assignment" field of DCI 1_0, updates associated DMRS, and calculate TBS.
func validateDci10PdschTdRa(i int) error {
	//regYellow.Printf("-->calling validatePdsch\n")

	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1-1: Valid S and L combinations
	// Note 1:	S = 3 is applicable only if dmrs-TypeA-Position = 3
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-1: Applicable PDSCH time domain resource allocation for DCI formats 1_0, 1_1, 4_0, 4_1 and 4_2
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-1A: Applicable PDSCH time domain resource allocation for DCI format 1_2
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-2: Default PDSCH time domain resource allocation A for normal CP
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-3: Default PDSCH time domain resource allocation A for extended CP
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-4: Default PDSCH time domain resource allocation B
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-5: Default PDSCH time domain resource allocation C
	dmrsTypeAPos := flags.gridsetting.dmrsTypeAPos
	rnti := flags.dldci._rnti[i]

	// processing TDRA of DCI 1_0
	row := flags.dldci.tdra[i] + 1
	key := fmt.Sprintf("%v_%v", row, dmrsTypeAPos[3:])
	var p *nrgrid.TimeAllocInfo
	var exist bool

	switch rnti {
	case "SI-RNTI":
		switch flags.gridsetting._coreset0MultiplexingPat {
		case 1:
			p, exist = nrgrid.PdschTimeAllocDefANormCp[key]
		case 2:
			// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-4: Default PDSCH time domain resource allocation B
			// Note 1: If the PDSCH was scheduled with SI-RNTI in PDCCH Type0 common search space, the UE may assume that this PDSCH resource allocation is not applied.
			if utils.ContainsInt(nrgrid.PdschTimeAllocDefBNote1Set, flags.dldci.tdra[i]+1) {
				return errors.New(fmt.Sprintf("Row %v is invalid for SIB1 (refer to 'Note 1' of Table 5.1.2.1.1-4 of TS 38.214).", flags.dldci.tdra[i]+1))
			}
			p, exist = nrgrid.PdschTimeAllocDefB[key]
		case 3:
			// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-5: Default PDSCH time domain resource allocation C
			// Note 1: The UE may assume that this PDSCH resource allocation is not used, if the PDSCH was scheduled with SI-RNTI in PDCCH Type0 common search space.
			// Note 2:	This applies for Case F and Case G candidate SS/PBCH block pattern described in clause 4 of [6, TS 38.213]
			if utils.ContainsInt(nrgrid.PdschTimeAllocDefCNote1Set, flags.dldci.tdra[i]+1) {
				return errors.New(fmt.Sprintf("Row %v is invalid for SIB1 (refer to 'Note 1' of Table 5.1.2.1.1-5 of TS 38.214).", flags.dldci.tdra[i]+1))
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
		return errors.New(fmt.Sprintf("Invalid PDSCH time domain allocation: tdra=%v, dmrsTypeAPos=%v\n", flags.dldci.tdra[i], flags.gridsetting.dmrsTypeAPos))
	} else {
		// update DCI 1_0 info
		fmt.Printf("TimeAllocInfo(tag=%v, rnti=%v, coreset0MultiplexingPat=%v): %v\n", flags.dldci._tag[i], rnti, flags.gridsetting._coreset0MultiplexingPat, *p)
		flags.dldci._tdMappingType[i] = p.MappingType
		flags.dldci._tdK0[i] = p.K0K2
		flags.dldci._tdStartSymb[i] = p.S
		flags.dldci._tdNumSymbs[i] = p.L
		sliv, _ := nrgrid.ToSliv(p.S, p.L, "PDSCH", p.MappingType, "normal", "")
		flags.dldci._tdSliv[i] = sliv

		// update DMRS info
		// refer to 3GPP TS 38.214 vh40: 5.1.6.2	DM-RS reception procedure
		// When receiving PDSCH scheduled by DCI format 1_0, 4_0, or 4_1, the UE shall assume the number of DM-RS CDM groups without data is 1 which corresponds to CDM group 0 for the case of PDSCH with allocation duration of 2 symbols, and the UE shall assume that the number of DM-RS CDM groups without data is 2 which corresponds to CDM group {0,1} for all other cases.
		// When receiving PDSCH scheduled by DCI format 1_0, 4_0, or 4_1, ...the UE shall assume that the PDSCH is not present in any symbol carrying DM-RS except for PDSCH with allocation duration of 2 symbols with PDSCH mapping type B (described in clause 7.4.1.1.2 of [4, TS 38.211]), and a single symbol front-loaded DM-RS of configuration type 1 on DM-RS port 1000 is transmitted, and that all the remaining orthogonal antenna ports are not associated with transmission of PDSCH to another UE and in addition:
		// 	-For PDSCH with mapping type A and type B, the UE shall assume dmrs-AdditionalPosition='pos2' and up to two additional single-symbol DM-RS present in a slot according to the PDSCH duration indicated in the DCI as defined in Clause 7.4.1.1 of [4, TS 38.211], and
		//	-For PDSCH with allocation duration of 2 symbols with mapping type B, the UE shall assume that the PDSCH is present in the symbol carrying DM-RS.
		if p.L == 2 {
			flags.dmrsCommon._cdmGroupsWoData[i] = 1
		} else {
			flags.dmrsCommon._cdmGroupsWoData[i] = 2
		}
		flags.dmrsCommon._dmrsType[i] = "type1"
		flags.dmrsCommon._dmrsPorts[i] = 1000
		flags.dmrsCommon._maxLength[i] = "len1"
		flags.dmrsCommon._numFrontLoadSymbs[i] = 1
		flags.dmrsCommon._dmrsAddPos[i] = "pos2"

		// update TD/FD pattern of DMRS
		flags.dmrsCommon._tdL[i], flags.dmrsCommon._fdK[i] = getDmrsPdschTdFdPattern("type1", p.MappingType, p.S, p.L, 1, "pos2", flags.dmrsCommon._cdmGroupsWoData[i])
		fmt.Printf("TD pattern within a slot of DMRS for %v: %v\n", flags.dmrsCommon._tag[i], flags.dmrsCommon._tdL[i])
		fmt.Printf("FD pattern within a PRB of DMRS for %v: %v\n", flags.dmrsCommon._tag[i], flags.dmrsCommon._fdK[i])
	}

	return nil
}

// updateDci10PdschTbs updates the TBS field of DCI 1_0 scheduling Sib1/Msg2/Msg4.
//  i: index of the flags.dci10 slices[0-SIB1, 1-Msg2, 2-Msg4]
func updateDci10PdschTbs(i int) error {
	// regYellow.Printf("-->calling updateDci10PdschTbs\n")

	td := flags.dldci._tdNumSymbs[i]
	ld := 0
	fd := flags.dldci.fdNumRbs[i]
	mcs := flags.dldci.mcsCw0[i]

	// update FDRA
	riv, err := makeRiv(flags.dldci.fdNumRbs[i], flags.dldci.fdStartRb[i], flags.bwp._bwpNumRbs[INI_DL_BWP])
	if err != nil {
		return err
	}
	flags.dldci._fdRa[i] = fmt.Sprintf("%0*b", flags.dldci._fdBitsRaType1[i], riv)
	fmt.Printf("PDSCH(tag=%v): RIV=%v, FDRA bits=%v\n", flags.dldci._tag[i], riv, flags.dldci._fdRa[i])

	// refer to 3GPP TS 38.211 vh40: 7.4.1.1.2	Mapping to physical resources (DMRS for PDSCH)
	// -for PDSCH mapping type A, ld is the duration between the first OFDM symbol of the slot and the last OFDM symbol of the scheduled PDSCH resources in the slot
	// -for PDSCH mapping type B, ld is the duration of the scheduled PDSCH resources
	if flags.dldci._tdMappingType[i] == "typeA" {
		ld = flags.dldci._tdStartSymb[i] + flags.dldci._tdNumSymbs[i]
	} else {
		ld = td
	}

	key2 := fmt.Sprintf("%v_%v_%v", ld, flags.dldci._tdMappingType[i], flags.dmrsCommon._dmrsAddPos[i])
	// refer to 3GPP TS 38.214 vh40:
	// When receiving PDSCH scheduled by DCI format 1_0, 4_0, or 4_1, the UE shall assume that ..., and a single symbol front-loaded DM-RS of configuration type 1 on DM-RS port 1000 is transmitted, and ...
	dmrs, exist := nrgrid.DmrsPdschPosOneSymb[key2]
	if !exist || dmrs == nil {
		return errors.New(fmt.Sprintf("Invalid DMRS for PDSCH settings: rnti=%v, numFrontLoadSymbs=%v, key=%v\n", flags.dldci._rnti[i], 1, key2))
	}

	// refer to 3GPP TS 38.211 vh40: 7.4.1.1.2	Mapping to physical resources (DMRS for PDSCH)
	// For PDSCH mapping type A,
	// 	- the case dmrs-AdditionalPosition equals to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2'.
	//	- l_d = 3 and l_d = 4 symbols in Tables 7.4.1.1.2-3 and 7.4.1.1.2-4 respectively is only applicable when dmrs-TypeA-Position is equal to 'pos2'.
	//	- single-symbol DM-RS, l1=11 except if all of the following conditions are fulfilled in which case l1=12: (2023/2/22: DSS is not supported!)
	// For PDSCH mapping type B,
	// 	- if ... and the front-loaded DM-RS of the PDSCH allocation collides with resources reserved for a search space set associated with a CORESET...(2023/2/23: Assume no collision between PDSCH DMRS and CORESET!)
	//  - if the PDSCH duration ld is less than or equal to 4 OFDM symbols, only single-symbol DM-RS is supported.
	//	- if the higher-layer parameter lte-CRS-ToMatchAround, lte-CRS-PatternList1, or lte-CRS-PatternList2 is configured,...(2023/2/23: DSS is not supported!)
	dmrsTypeAPos := flags.gridsetting.dmrsTypeAPos
	if flags.dldci._tdMappingType[i] == "typeA" {
		if (ld == 3 || ld == 4) && dmrsTypeAPos != "pos2" {
			return errors.New(fmt.Sprintf("For PDSCH mapping type A, ld = 3 and ld = 4 symbols in Tables 7.4.1.1.2-3 and 7.4.1.1.2-4 respectively is only applicable when dmrs-TypeA-Position is equal to 'pos2'.\nld=%v, dmrsTypeAPos=%v\n", ld, dmrsTypeAPos))
		}
	}

	dmrsOh := (2 * flags.dmrsCommon._cdmGroupsWoData[i]) * len(dmrs)
	fmt.Printf("PDSCH(tag=%v) DMRS overhead: cdmGroupsWoData=%v, key=%v, dmrs=%v\n", flags.dldci._tag[i], flags.dmrsCommon._cdmGroupsWoData[i], key2, dmrs)

	tbs, err := getTbs("PDSCH", false, flags.dldci._rnti[i], "qam64", td, fd, mcs, 1, dmrsOh, 0, 1)
	if err != nil {
		return err
	} else {
		fmt.Printf("PDSCH(tag=%v) CW0 TBS=%v bits\n", flags.dldci._tag[i], tbs)
		flags.dldci._tbsCw0[i] = tbs
	}

	fmt.Println()

	return nil
}

// validateDci11PdschTdRa validates the "Time domain resource assignment" field of DCI 1_1 scheduling PDSCH.
func validateDci11PdschTdRa() error {
	//regYellow.Printf("-->calling validateDci11PdschTdRa\n")

	dmrsTypeAPos := flags.gridsetting.dmrsTypeAPos

	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-1: Applicable PDSCH time domain resource allocation for DCI formats 1_0, 1_1, 4_0, 4_1 and 4_2
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-2: Default PDSCH time domain resource allocation A for normal CP
	// refer to 3GPP TS 38.214 vh40: Table 5.1.2.1.1-3: Default PDSCH time domain resource allocation A for extended CP
	// 2023/2/23: For simplicity, pdsch-AllocationList configured PDSCH TDRA is not supported!
	row := flags.dldci.tdra[DCI_11_PDSCH] + 1
	key := fmt.Sprintf("%v_%v", row, dmrsTypeAPos[3:])
	var p *nrgrid.TimeAllocInfo
	var exist bool

	if flags.bwp._bwpCp[DED_DL_BWP] == "normal" {
		p, exist = nrgrid.PdschTimeAllocDefANormCp[key]
	} else {
		p, exist = nrgrid.PdschTimeAllocDefAExtCp[key]
	}

	if !exist {
		return errors.New(fmt.Sprintf("Invalid PDSCH time domain allocation: dci11TdRa=%v, dmrsTypeAPos=%v\n", flags.dldci.tdra[DCI_11_PDSCH], flags.gridsetting.dmrsTypeAPos))
	} else {
		fmt.Printf("TimeAllocInfo(tag=%v, rnti=C-RNTI): %v\n", flags.dldci._tag[DCI_11_PDSCH], *p)
		flags.dldci._tdMappingType[DCI_11_PDSCH] = p.MappingType
		flags.dldci._tdK0[DCI_11_PDSCH] = p.K0K2
		flags.dldci._tdStartSymb[DCI_11_PDSCH] = p.S
		flags.dldci._tdNumSymbs[DCI_11_PDSCH] = p.L
		sliv, _ := nrgrid.ToSliv(p.S, p.L, "PDSCH", p.MappingType, "normal", "")
		flags.dldci._tdSliv[DCI_11_PDSCH] = sliv
	}

	return nil
}

// validateDci11PdschAntPorts validates PDSCH configurations, updates DMRS/PTRS for PDSCH and updates PDSCH TBS.
func validateDci11PdschAntPorts() error {
	//regYellow.Printf("-->calling validateDci11PdschAntPorts\n")

	dmrsType := flags.pdsch.pdschDmrsType
	maxLength := flags.pdsch.pdschMaxLength
	var mcsSet []int
	if flags.dldci.mcsCw0[DCI_11_PDSCH] >= 0 {
		mcsSet = append(mcsSet, flags.dldci.mcsCw0[DCI_11_PDSCH])
	}
	if flags.dldci.mcsCw1 >= 0 {
		mcsSet = append(mcsSet, flags.dldci.mcsCw1)
	}

	ap := flags.dldci.antennaPorts

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
		return errors.New(fmt.Sprintf("Invalid settings for DCI 1_1 'Antenna port(s)'.\ndmrsType=%v, maxLength=%v, mcsSet=%v\n", dmrsType, maxLength, mcsSet))
	}
	fmt.Printf("Available 'Antenna port(s)' field of DCI 1_1(dmrsType=%v,maxLen=%v,mcsSet=%v,ap=%v): %v\n", dmrsType, maxLength, mcsSet, ap, tokens)

	if !exist || p == nil {
		return errors.New(fmt.Sprintf("Invalid settings for DCI 1_1 'Antenna port(s)'.\ndmrsType=%v, maxLength=%v, mcsSet=%v, antPorts=%v\n", dmrsType, maxLength, mcsSet, ap))
	}

	if len(p.DmrsPorts) > flags.pdsch.pdschMaxLayers {
		return errors.New(fmt.Sprintf("Invalid settings for DCI 1_1 'Antenna port(s)'.\nantPorts=%v, dmrsPorts=%v, while pdschMaxLayers=%v!\n", ap, p.DmrsPorts, flags.pdsch.pdschMaxLayers))
	}

	for i := range p.DmrsPorts {
		p.DmrsPorts[i] += 1000
	}
	fmt.Printf("AntPortsInfo(PDSCH and its DMRS): %v\n", *p)

	flags.pdsch._cdmGroupsWoData = p.CdmGroups
	flags.pdsch._dmrsPorts = p.DmrsPorts
	flags.pdsch._numFrontLoadSymbs = p.NumDmrsSymbs

	// determine TD/FD pattern of DMRS for PDSCH
	flags.pdsch._tdL, flags.pdsch._fdK = getDmrsPdschTdFdPattern(dmrsType, flags.dldci._tdMappingType[DCI_11_PDSCH], flags.dldci._tdStartSymb[DCI_11_PDSCH], flags.dldci._tdNumSymbs[DCI_11_PDSCH], p.NumDmrsSymbs, flags.pdsch.pdschDmrsAddPos, p.CdmGroups)
	fmt.Printf("TD pattern within a slot of DMRS for PDSCH(DCI 1_1): %v\n", flags.pdsch._tdL)
	fmt.Printf("FD pattern within a PRB of DMRS for PDSCH(DCI 1_1): %v\n", flags.pdsch._fdK)

	// update PTRS for PDSCH
	maxDmrsPorts := utils.MaxInt(flags.pdsch._dmrsPorts)
	noPtrs := false
	// refer to 3GPP TS 38.214 vh40: 5.1.6.2	DM-RS reception procedure
	// If a UE receiving PDSCH scheduled by DCI format 1_2 is configured with the higher layer parameter phaseTrackingRS in dmrs-DownlinkForPDSCH-MappingTypeA-DCI-1-2  or dmrs-DownlinkForPDSCH-MappingTypeB-DCI-1-2
	// or a UE receiving PDSCH scheduled by DCI format 1_0 or DCI format 1_1 is configured with the higher layer parameter phaseTrackingRS in dmrs-DownlinkForPDSCH-MappingTypeA or dmrs-DownlinkForPDSCH-MappingTypeB,
	// the UE may assume that the following configurations are not occurring simultaneously for the received PDSCH:
	// - any DM-RS ports among 1004-1007 or 1006-1011 for DM-RS configurations type 1 and type 2, respectively are scheduled for the UE and the other UE(s) sharing the DM-RS REs on the same CDM group(s), and
	// - PT-RS is transmitted to the UE.
	if (dmrsType == "type1" && maxDmrsPorts >= 1004) || (dmrsType == "type2" && maxDmrsPorts >= 1006) {
		noPtrs = true
	}
	fmt.Printf("PDSCH noPtrs=%v\n", noPtrs)

	if noPtrs {
		flags.pdsch.pdschPtrsEnabled = false
	} else {
		// refer to 3GPP TS 38.214 vh40: 5.1.6.3	PT-RS reception procedure
		// If a UE is scheduled with one codeword, the PT-RS antenna port is associated with the lowest indexed DM-RS antenna port among the DM-RS antenna ports assigned for the PDSCH.
		// If a UE is scheduled with two codewords, the PT-RS antenna port is associated with the lowest indexed DM-RS antenna port among the DM-RS antenna ports assigned for the codeword with the higher MCS. If the MCS indices of the two codewords are the same, the PT-RS antenna port is associated with the lowest indexed DM-RS antenna port assigned for codeword 0.
		if len(mcsSet) == 1 {
			flags.pdsch._ptrsDmrsPorts = flags.pdsch._dmrsPorts[0]
		} else {
			// refer to 3GPP TS 38.211 vh40: Table 7.3.1.3-1: Codeword-to-layer mapping for spatial multiplexing.
			// refer to 3GPP TS 38.211 vh40: 7.3.1.4	Antenna port mapping
			numLayersCw0 := utils.FloorInt(float64(len(flags.pdsch._dmrsPorts)) / 2)
			if mcsSet[0] >= mcsSet[1] {
				flags.pdsch._ptrsDmrsPorts = flags.pdsch._dmrsPorts[0]
			} else {
				flags.pdsch._ptrsDmrsPorts = flags.pdsch._dmrsPorts[numLayersCw0]
			}
		}
	}

	// update PDSCH TBS
	fdRaType := flags.dldci._fdRaType[DCI_11_PDSCH]
	fdRa := flags.dldci._fdRa[DCI_11_PDSCH]
	if fdRaType == "raType0" && len(fdRa) != flags.dldci._fdBitsRaType0 {
		return errors.New(fmt.Sprintf("Invalid 'Frequency domain resource assignment' field of DCI 1_1: fdRaType=%v, fdRa=%v, len(fdRa)=%v, bitsRaType0=%v\n", fdRaType, fdRa, len(fdRa), flags.dldci._fdBitsRaType0))
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
		fd = flags.dldci.fdNumRbs[DCI_11_PDSCH]
		// update FDRA
		riv, err := makeRiv(flags.dldci.fdNumRbs[DCI_11_PDSCH], flags.dldci.fdStartRb[DCI_11_PDSCH], flags.bwp._bwpNumRbs[DED_DL_BWP])
		if err != nil {
			return err
		}
		flags.dldci._fdRa[DCI_11_PDSCH] = fmt.Sprintf("%0*b", flags.dldci._fdBitsRaType1[DCI_11_PDSCH], riv)
		fmt.Printf("PDSCH(tag=%v): RIV=%v, FDRA bits=%v\n", flags.dldci._tag[DCI_11_PDSCH], riv, flags.dldci._fdRa[DCI_11_PDSCH])
	}

	// calculate DMRS overhead
	td := flags.dldci._tdNumSymbs[DCI_11_PDSCH]
	ld := 0
	tdMappingType := flags.dldci._tdMappingType[DCI_11_PDSCH]
	dmrsAddPos := flags.pdsch.pdschDmrsAddPos

	// refer to 3GPP TS 38.211 vh40: 7.4.1.1.2	Mapping to physical resources (DMRS for PDSCH)
	// -for PDSCH mapping type A, ld is the duration between the first OFDM symbol of the slot and the last OFDM symbol of the scheduled PDSCH resources in the slot
	// -for PDSCH mapping type B, ld is the duration of the scheduled PDSCH resources
	if tdMappingType == "typeA" {
		ld = flags.dldci._tdStartSymb[DCI_11_PDSCH] + flags.dldci._tdNumSymbs[DCI_11_PDSCH]
	} else {
		ld = td
	}

	key := fmt.Sprintf("%v_%v_%v", ld, tdMappingType, dmrsAddPos)
	var dmrs []int
	if flags.pdsch._numFrontLoadSymbs == 1 {
		dmrs, exist = nrgrid.DmrsPdschPosOneSymb[key]
	} else {
		dmrs, exist = nrgrid.DmrsPdschPosTwoSymbs[key]
	}

	if !exist || dmrs == nil {
		return errors.New(fmt.Sprintf("Invalid DMRS for PDSCH settings: rnti=%v, numFrontLoadSymbs=%v, key=%v\n", flags.dldci._rnti[DCI_11_PDSCH], flags.pdsch._numFrontLoadSymbs, key))
	}

	// refer to 3GPP TS 38.211 vh40: 7.4.1.1.2	Mapping to physical resources (DMRS for PDSCH)
	// For PDSCH mapping type A,
	// 	- the case dmrs-AdditionalPosition equals to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2'.
	//	- l_d = 3 and l_d = 4 symbols in Tables 7.4.1.1.2-3 and 7.4.1.1.2-4 respectively is only applicable when dmrs-TypeA-Position is equal to 'pos2'.
	//	- single-symbol DM-RS, l1=11 except if all of the following conditions are fulfilled in which case l1=12: (2023/2/22: DSS is not supported!)
	// For PDSCH mapping type B,
	// 	- if ... and the front-loaded DM-RS of the PDSCH allocation collides with resources reserved for a search space set associated with a CORESET...(2023/2/23: Assume no collision between PDSCH DMRS and CORESET!)
	//  - if the PDSCH duration ld is less than or equal to 4 OFDM symbols, only single-symbol DM-RS is supported.
	//	- if the higher-layer parameter lte-CRS-ToMatchAround, lte-CRS-PatternList1, or lte-CRS-PatternList2 is configured,...(2023/2/23: DSS is not supported!)
	dmrsTypeAPos := flags.gridsetting.dmrsTypeAPos
	if tdMappingType == "typeA" && dmrsAddPos == "pos3" && dmrsTypeAPos != "pos2" {
		return errors.New(fmt.Sprintf("For PDSCH mapping type A, the case dmrs-AdditionalPosition equals to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2'.\npdschDmrsAddPos=%v,dmrsTypeAPos=%v\n", flags.pdsch.pdschDmrsAddPos, dmrsTypeAPos))
	}
	if tdMappingType == "typeA" {
		if (ld == 3 || ld == 4) && dmrsTypeAPos != "pos2" {
			return errors.New(fmt.Sprintf("For PDSCH mapping type A, ld = 3 and ld = 4 symbols in Tables 7.4.1.1.2-3 and 7.4.1.1.2-4 respectively is only applicable when dmrs-TypeA-Position is equal to 'pos2'.\nld=%v, dmrsTypeAPos=%v\n", ld, dmrsTypeAPos))
		}
	}
	if tdMappingType == "typeB" && (ld <= 4) && flags.pdsch._numFrontLoadSymbs != 1 {
		return errors.New(fmt.Sprintf("For PDSCH mapping type B, if the PDSCH duration ld is less than or equal to 4 OFDM symbols, only single-symbol DM-RS is supported.\n tdMappingType=%v, ld=%v, numFrontLoadSymbs=%v\n", tdMappingType, ld, flags.pdsch._numFrontLoadSymbs))
	}

	dmrsOh := (2 * flags.pdsch._cdmGroupsWoData) * len(dmrs)
	fmt.Printf("PDSCH(tag=%v) DMRS overhead: cdmGroupsWoData=%v, key=%v, dmrs=%v\n", flags.dldci._tag[DCI_11_PDSCH], flags.pdsch._cdmGroupsWoData, key, dmrs)

	xoh, _ := strconv.Atoi(flags.pdsch.pdschXOh[3:])
	if flags.dldci.mcsCw0[DCI_11_PDSCH] >= 0 {
		tbs, err := getTbs("PDSCH", false, "C-RNTI", flags.pdsch.pdschMcsTable, td, fd, flags.dldci.mcsCw0[DCI_11_PDSCH], len(flags.pdsch._dmrsPorts), dmrsOh, xoh, 1)
		if err != nil {
			return err
		} else {
			fmt.Printf("PDSCH(tag=%v) CW0 TBS=%v bits\n", flags.dldci._tag[DCI_11_PDSCH], tbs)
			flags.dldci._tbsCw0[DCI_11_PDSCH] = tbs
		}
	}

	if flags.dldci.mcsCw1 >= 0 {
		tbs, err := getTbs("PDSCH", false, "C-RNTI", flags.pdsch.pdschMcsTable, td, fd, flags.dldci.mcsCw1, len(flags.pdsch._dmrsPorts), dmrsOh, xoh, 1)
		if err != nil {
			return err
		} else {
			fmt.Printf("PDSCH(tag=%v) CW1 TBS=%v bits\n", flags.dldci._tag[DCI_11_PDSCH], tbs)
			flags.dldci._tbsCw1 = tbs
		}
	}

	fmt.Println()

	return nil
}

func validatePusch() error {
	regYellow.Printf("-->calling validatePusch\n")

	for i, _ := range flags.uldci._rnti {
		if i == RAR_UL_MSG3 {
			// validate TDRA
			err := validateRarUlMsg3Tdra()
			if err != nil {
				return err
			}

			// update TBS
			err = updateRarUlMsg3PuschTbs()
			if err != nil {
				return err
			}
		} else if i == DCI_01_PUSCH {
			// validate TDRA
			err := validateDci01PuschTdRa()
			if err != nil {
				return err
			}
			// validate 'Antenna port(s)' field and update TBS
			err = validateDci01PuschAntPorts()
			if err != nil {
				return err
			}
		} else if i == RA_UL_MSGA {
			// TODO: MSGA transmission for two-steps CBRA
		} else if i == FBRAR_UL_MSG3 {
			// TODO: MSG3 transmission scheduled by fallbackRAR UL grant for two-setps CBRA
		}
	}

	return nil
}

func validateRarUlMsg3Tdra() error {
	// refer to 3GPP TS 38.214 vh40:
	//  - Table 6.1.2.1.1-1: Applicable PUSCH time domain resource allocation for common search space and DCI format 0_0 in UE specific search space
	//  - Table 6.1.2.1.1-1A: Applicable PUSCH time domain resource allocation for DCI format 0_1 in UE specific search space scrambled with C-RNTI, MCS-C-RNTI, CS-RNTI or SP-CSI-RNTI
	//  - Table 6.1.2.1.1-1B: Applicable PUSCH time domain resource allocation for DCI format 0_2 in UE specific search space scrambled with C-RNTI, MCS-C-RNTI, CS-RNTI or SP-CSI-RNTI
	//  - Table 6.1.2.1.1-2: Default PUSCH time domain resource allocation A for normal CP
	//  - Table 6.1.2.1.1-3: Default PUSCH time domain resource allocation A for extended CP
	row := flags.uldci.tdra[RAR_UL_MSG3] + 1
	var p *nrgrid.TimeAllocInfo
	var exist bool

	if flags.bwp._bwpCp[INI_UL_BWP] == "normal" {
		p, exist = nrgrid.PuschTimeAllocDefANormCp[row]
	} else {
		p, exist = nrgrid.PuschTimeAllocDefAExtCp[row]
	}

	if !exist {
		return errors.New(fmt.Sprintf("Invalid PUSCH time domain allocation: tdra=%v, dmrsTypeAPos=%v\n", flags.uldci.tdra[RAR_UL_MSG3], flags.gridsetting.dmrsTypeAPos))
	} else {
		// update Msg3 info
		fmt.Printf("TimeAllocInfo(tag=%v, rnti=%v): %v\n", flags.uldci._tag[RAR_UL_MSG3], flags.uldci._rnti[RAR_UL_MSG3], *p)
		flags.uldci._tdMappingType[RAR_UL_MSG3] = p.MappingType
		flags.uldci._tdK2[RAR_UL_MSG3] = p.K0K2 + nrgrid.PuschTimeAllocK2j[flags.gridsetting.scs]
		flags.uldci._tdDelta = nrgrid.PuschTimeAllocMsg3K2Delta[flags.gridsetting.scs]
		flags.uldci._tdStartSymb[RAR_UL_MSG3] = p.S
		flags.uldci._tdNumSymbs[RAR_UL_MSG3] = p.L
		sliv, _ := nrgrid.ToSliv(p.S, p.L, "PUSCH", p.MappingType, "normal", "typeA")
		flags.uldci._tdSliv[RAR_UL_MSG3] = sliv
	}

	return nil
}

// updateRarUlMsg3PuschTbs updates the TBS field of Msg3 PUSCH scheduled by RAR Msg2.
func updateRarUlMsg3PuschTbs() error {
	//regYellow.Printf("-->calling updateRarUlMsg3PuschTbs\n")

	td := flags.uldci._tdNumSymbs[RAR_UL_MSG3]
	fd := flags.uldci.fdNumRbs[RAR_UL_MSG3]

	// validate L_RBs when transform precoding is enabled
	if flags.rach.msg3Tp == "enabled" {
		// valid PUSCH PRB allocations when transforming precoding is enabled
		lrbsPuschTp := initLrbsPuschTp(flags.bwp._bwpNumRbs[INI_UL_BWP])

		if !utils.ContainsInt(lrbsPuschTp, fd) {
			lt, gt := utils.NearestInt(lrbsPuschTp, fd)
			return errors.New(fmt.Sprintf("L_RBs(=%v) must be 2^x*3^y*5^z, where x/y/z>=0, when transforming precoding is enabled, and the nearest values are: [%v, %v].\n", fd, lt, gt))
		}
	}

	// update FDRA
	riv, err := makeRiv(flags.uldci.fdNumRbs[RAR_UL_MSG3], flags.uldci.fdStartRb[RAR_UL_MSG3], flags.bwp._bwpNumRbs[INI_UL_BWP])
	if err != nil {
		return err
	}
	flags.uldci._fdRa[RAR_UL_MSG3] = fmt.Sprintf("%0*b", flags.uldci._fdBitsRaType1[RAR_UL_MSG3], riv)
	fmt.Printf("PUSCH(tag=%v): RIV=%v, FDRA bits=%v\n", flags.uldci._tag[RAR_UL_MSG3], riv, flags.uldci._fdRa[RAR_UL_MSG3])
	if flags.uldci.fdFreqHop[RAR_UL_MSG3] != "disabled" {
		var ulHopBits int
		if flags.bwp._bwpNumRbs[INI_UL_BWP] >= 50 {
			ulHopBits = 2
		} else {
			ulHopBits = 1
		}

		v, _ := strconv.Atoi(flags.uldci._fdRa[RAR_UL_MSG3][:ulHopBits])
		if v != 0 {
			return errors.New(fmt.Sprintf("The first %v bits of RIV must be all zeros when frequency hopping is enabled!", ulHopBits))
		}
	}

	// update DMRS for Msg3 PUSCH
	// 2023-3-10: assume Msg3 follows the same rules as DCI 0_0 with TC-RNTI which is used for Msg3 retransmission
	if td <= 2 && flags.rach.msg3Tp == "disabled" {
		flags.dmrsCommon._cdmGroupsWoData[DMRS_RAR_UL_MSG3] = 1
	} else {
		flags.dmrsCommon._cdmGroupsWoData[DMRS_RAR_UL_MSG3] = 2
	}
	flags.dmrsCommon._dmrsType[DMRS_RAR_UL_MSG3] = "type1"
	flags.dmrsCommon._dmrsPorts[DMRS_RAR_UL_MSG3] = 0
	flags.dmrsCommon._maxLength[DMRS_RAR_UL_MSG3] = "len1"
	flags.dmrsCommon._numFrontLoadSymbs[DMRS_RAR_UL_MSG3] = 1

	// refer to 3GPP TS 38.214 vh40: 6.2.2	UE DM-RS transmission procedure
	// When transmitted PUSCH is neither scheduled by DCI format 0_1/0_2 with CRC scrambled by C-RNTI, CS-RNTI, SP-CSI-RNTI or MCS-C-RNTI, nor corresponding to a configured grant, nor being a PUSCH for Type-2 random access procedure, the UE shall use single symbol front-loaded DM-RS of configuration type 1 on DM-RS port 0 and the remaining REs not used for DM-RS in the symbols are not used for any PUSCH transmission except for PUSCH with allocation duration of 2 or less OFDM symbols with transform precoding disabled, additional DM-RS can be transmitted according to the scheduling type and the PUSCH duration as specified in Table 6.4.1.1.3-3 of [4, TS38.211] for frequency hopping disabled and as specified in Table 6.4.1.1.3-6 of [4, TS38.211] for frequency hopping enabled, and
	// If frequency hopping is disabled:
	// -	The UE shall assume dmrs-AdditionalPosition equals to 'pos2' and up to two additional DM-RS can be transmitted according to PUSCH duration, or
	// If frequency hopping is enabled:
	// -	The UE shall assume dmrs-AdditionalPosition equals to 'pos1' and up to one additional DM-RS can be transmitted according to PUSCH duration.
	if flags.uldci.fdFreqHop[RAR_UL_MSG3] == "intra-slot" {
		flags.dmrsCommon._dmrsAddPos[DMRS_RAR_UL_MSG3] = "pos1"
	} else {
		flags.dmrsCommon._dmrsAddPos[DMRS_RAR_UL_MSG3] = "pos2"
	}

	// determine TD/FD pattern of DMRS for Msg3 PUSCH
	flags.dmrsCommon._tdL[DMRS_RAR_UL_MSG3], flags.dmrsCommon._tdL2, flags.dmrsCommon._fdK[DMRS_RAR_UL_MSG3] = getDmrsPuschTdFdPattern("type1", flags.uldci._tdMappingType[RAR_UL_MSG3], flags.uldci._tdStartSymb[RAR_UL_MSG3], flags.uldci._tdNumSymbs[RAR_UL_MSG3], 1, flags.dmrsCommon._dmrsAddPos[DMRS_RAR_UL_MSG3], flags.dmrsCommon._cdmGroupsWoData[DMRS_RAR_UL_MSG3], flags.uldci.fdFreqHop[RAR_UL_MSG3])
	if flags.uldci.fdFreqHop[RAR_UL_MSG3] != "intra-slot" {
		fmt.Printf("TD pattern within a slot of DMRS for %v: %v\n", flags.dmrsCommon._tag[DMRS_RAR_UL_MSG3], flags.dmrsCommon._tdL[DMRS_RAR_UL_MSG3])
	} else {
		fmt.Printf("TD pattern within a slot of DMRS for %v (1st hop): %v\n", flags.dmrsCommon._tag[DMRS_RAR_UL_MSG3], flags.dmrsCommon._tdL[DMRS_RAR_UL_MSG3])
		fmt.Printf("TD pattern within a slot of DMRS for %v (2nd hop): %v\n", flags.dmrsCommon._tag[DMRS_RAR_UL_MSG3], flags.dmrsCommon._tdL2)
	}
	fmt.Printf("FD pattern within a PRB of DMRS for %v: %v\n", flags.dmrsCommon._tag[DMRS_RAR_UL_MSG3], flags.dmrsCommon._fdK[DMRS_RAR_UL_MSG3])

	// calculate DMRS overhead
	tdMappingType := flags.uldci._tdMappingType[RAR_UL_MSG3]
	dmrsAddPos := flags.dmrsCommon._dmrsAddPos[DMRS_RAR_UL_MSG3]
	freqHop := flags.uldci.fdFreqHop[RAR_UL_MSG3]
	var v1, v2, v []int
	var e1, e2, e bool
	var key1, key2, key string
	if freqHop == "intra-slot" {
		// refer to 3GPP 38.211 vh40 6.4.1.1.3
		// ld is the duration per hop according to Table 6.4.1.1.3-6 if intra-slot frequency hopping is used
		// refer to 3GPP 38.214 vf30 6.3
		// In case of intra-slot frequency hopping, ... The number of symbols in the first hop is given by floor(N_PUSCH_symb/2) , the number of symbols in the second hop is given by N_PUSCH_symb - floor(N_PUSCH_symb/2) , where N_PUSCH_symb is the length of the PUSCH transmission in OFDM symbols in one slot.
		ld1 := utils.FloorInt(float64(td) / 2)
		ld2 := td - ld1

		// refer to 3GPP 38.211 vh40 6.4.1.1.3
		// ...and the position l0 of the first DM-RS symbol depends on the mapping type:
		//  -	for PUSCH mapping type A:
		//    -	l0 is given by the higher-layer parameter dmrs-TypeA-Position
		//  -	for PUSCH mapping type B:
		//    -	l0 = 0
		var l0 int
		if tdMappingType == "typeA" {
			l0, _ = strconv.Atoi(flags.gridsetting.dmrsTypeAPos[3:])
		} else {
			l0 = 0
		}

		// refer to 3GPP 38.211 vh40 6.4.1.1.3
		// if the higher-layer parameter dmrs-AdditionalPosition is not set to 'pos0' and intra-slot frequency hopping is enabled according to clause 7.3.1.1.2 in [4, TS 38.212] and by higher layer, Tables 6.4.1.1.3-6 shall be used assuming dmrs-AdditionalPosition is equal to 'pos1' for each hop.
		if dmrsAddPos != "pos0" {
			dmrsAddPos = "pos1"
		}

		key1 = fmt.Sprintf("%v_%v_%v_%v_1st", ld1, tdMappingType, l0, dmrsAddPos)
		key2 = fmt.Sprintf("%v_%v_%v_%v_2nd", ld2, tdMappingType, l0, dmrsAddPos)
		v1, e1 = nrgrid.DmrsPuschPosOneSymbWithIntraSlotFh[key1]
		v2, e2 = nrgrid.DmrsPuschPosOneSymbWithIntraSlotFh[key2]
		if !e1 || v1 == nil {
			return errors.New(fmt.Sprintf("Invalid key(=%v) when referring DmrsPuschPosOneSymbWithIntraSlotFh!", key1))
		}
		if !e2 || v2 == nil {
			return errors.New(fmt.Sprintf("Invalid key(=%v) when referring DmrsPuschPosOneSymbWithIntraSlotFh!", key2))
		}
	} else {
		// refer to 38.211 vh40 6.4.1.1.3
		// ld is the duration between the first OFDM symbol of the slot and the last OFDM symbol of the scheduled PUSCH resources in the slot for PUSCH mapping type A according to Tables 6.4.1.1.3-3 and 6.4.1.1.3-4 if intra-slot frequency hopping is not used, or
		// ld is the duration of scheduled PUSCH resources for PUSCH mapping type B according to Tables 6.4.1.1.3-3 and 6.4.1.1.3-4 if intra-slot frequency hopping is not used
		var ld int
		if tdMappingType == "typeA" {
			ld = flags.uldci._tdStartSymb[RAR_UL_MSG3] + td
		} else {
			ld = td
		}
		key = fmt.Sprintf("%v_%v_%v", ld, tdMappingType, dmrsAddPos)
		v, e = nrgrid.DmrsPuschPosOneSymbWoIntraSlotFh[key]
		if !e || v == nil {
			return errors.New(fmt.Sprintf("Invalid key(=%v) when referring DmrsPuschPosOneSymbWoIntraSlotFh!", key))
		}
	}

	// refer to 38.211 vh40 6.4.1.1.3
	// For PUSCH mapping type A,
	//  - the case dmrs-AdditionalPosition is equal to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2';
	//  - ld=4 symbols in Table 6.4.1.1.3-4 is only applicable when dmrs-TypeA-Position is equal to 'pos2'.
	// Table 6.4.1.1.3-4: PUSCH DM-RS positions l_bar within a slot for double-symbol DM-RS and intra-slot frequency hopping disabled.
	// Rationale: dmrsAddPos of Msg3 is either 'pos1' or 'pos2', so restriction #1 is not relevant. Table 6.4.1.1.3-4 is for double symbols front-loaded DMRS, so restriction #2 is not relevant.

	var dmrsOh int
	if freqHop == "intra-slot" {
		dmrsOh = (2 * flags.dmrsCommon._cdmGroupsWoData[DMRS_RAR_UL_MSG3]) * (len(v1) + len(v2))
		fmt.Printf("PUSCH(tag=%v) DMRS overhead: cdmGroupsWoData=%v, key1=%v, val1=%v, key2=%v, val2=%v\n", flags.uldci._tag[RAR_UL_MSG3], flags.dmrsCommon._cdmGroupsWoData[DMRS_RAR_UL_MSG3], key1, v1, key2, v2)
	} else {
		dmrsOh = (2 * flags.dmrsCommon._cdmGroupsWoData[DMRS_RAR_UL_MSG3]) * len(v)
		fmt.Printf("PUSCH(tag=%v) DMRS overhead: cdmGroupsWoData=%v, key=%v, val=%v\n", flags.uldci._tag[RAR_UL_MSG3], flags.dmrsCommon._cdmGroupsWoData[DMRS_RAR_UL_MSG3], key, v)
	}

	// 38.214 vh40 6.1.4.2	Transport block size determination
	// For Msg3 or MsgA PUSCH transmission the N_PRB_oh is always set to 0.
	// 38.214 vh40 6.1.1	Transmission schemes
	// If PUSCH is scheduled by DCI format 0_0, the PUSCH transmission is based on a single antenna port.
	tbs, err := getTbs("PUSCH", flags.rach.msg3Tp == "enabled", "MSG3", "qam64", td, fd, flags.uldci.mcsCw0[RAR_UL_MSG3], 1, dmrsOh, 0, 1)
	if err != nil {
		return err
	} else {
		fmt.Printf("PUSCH(tag=%v) CW0 TBS=%v bits\n", flags.uldci._tag[RAR_UL_MSG3], tbs)
		flags.uldci._tbs[RAR_UL_MSG3] = tbs
	}
	fmt.Println()

	return nil
}

// validateDci01PuschTdRa validates the "Time domain resource assignment" field of DCI 0_1 scheduling PDSCH.
func validateDci01PuschTdRa() error {
	//regYellow.Printf("-->calling validateDci01PuschTdRa\n")

	// refer to 3GPP TS 38.214 vh40:
	//  - Table 6.1.2.1.1-1: Applicable PUSCH time domain resource allocation for common search space and DCI format 0_0 in UE specific search space
	//  - Table 6.1.2.1.1-1A: Applicable PUSCH time domain resource allocation for DCI format 0_1 in UE specific search space scrambled with C-RNTI, MCS-C-RNTI, CS-RNTI or SP-CSI-RNTI
	//  - Table 6.1.2.1.1-1B: Applicable PUSCH time domain resource allocation for DCI format 0_2 in UE specific search space scrambled with C-RNTI, MCS-C-RNTI, CS-RNTI or SP-CSI-RNTI
	//  - Table 6.1.2.1.1-2: Default PUSCH time domain resource allocation A for normal CP
	//  - Table 6.1.2.1.1-3: Default PUSCH time domain resource allocation A for extended CP
	// 2023/3/5: For simplicity, pusch-AllocationList configured PUSCH TDRA is not supported!
	row := flags.uldci.tdra[DCI_01_PUSCH] + 1
	var p *nrgrid.TimeAllocInfo
	var exist bool

	if flags.bwp._bwpCp[DED_UL_BWP] == "normal" {
		p, exist = nrgrid.PuschTimeAllocDefANormCp[row]
	} else {
		p, exist = nrgrid.PuschTimeAllocDefAExtCp[row]
	}

	if !exist {
		return errors.New(fmt.Sprintf("Invalid PUSCH time domain allocation: tdra=%v, dmrsTypeAPos=%v\n", flags.uldci.tdra, flags.gridsetting.dmrsTypeAPos))
	} else {
		// update uldci info
		fmt.Printf("TimeAllocInfo(tag=%v, rnti=%v): %v\n", flags.uldci._tag[DCI_01_PUSCH], flags.uldci._rnti[DCI_01_PUSCH], *p)
		flags.uldci._tdMappingType[DCI_01_PUSCH] = p.MappingType
		flags.uldci._tdK2[DCI_01_PUSCH] = p.K0K2 + nrgrid.PuschTimeAllocK2j[flags.gridsetting.scs]
		flags.uldci._tdStartSymb[DCI_01_PUSCH] = p.S
		flags.uldci._tdNumSymbs[DCI_01_PUSCH] = p.L
		sliv, _ := nrgrid.ToSliv(p.S, p.L, "PUSCH", p.MappingType, "normal", flags.pusch._puschRepType)
		flags.uldci._tdSliv[DCI_01_PUSCH] = sliv
	}

	return nil
}

// validateDci01PuschAntPorts validates PUSCH configurations, updates DMRS/PTRS for PUSCH and updates PUSCH TBS.
func validateDci01PuschAntPorts() error {
	//regYellow.Printf("-->calling validateDci01PuschAntPorts\n")

	// determine rank
	var rank, tpmi int
	var coherence string
	if flags.pusch.puschTxCfg == "codebook" {
		// 3GPP 38.212 vh40
		// Table 7.3.1.1.2-32: SRI indication or Second SRI indication, for codebook based PUSCH transmission, if ul-FullPowerTransmission is not configured, or ul-FullPowerTransmission = fullpowerMode1, or ul-FullPowerTransmission = fullpowerMode2, or ul-FullPowerTransmission = fullpower and
		// 2023/2/23: For CB PUSCH, if N_SRS=1, SRI=0; if N_SRS=2, SRI=0/1; N_SRS=3/4 is not supported! For simplicity, use SRI as index to the SRS resource.
		// 3GPP 38.214 vh40: 6.1.1.1	Codebook based UL transmission
		// The UE shall transmit PUSCH using the same antenna port(s) as the SRS port(s) in the SRS resource indicated by the DCI format 0_1 or 0_2 or by configuredGrantConfig according to clause 6.1.2.3.
		var apCbPusch []int
		switch flags.srs.srsNumPorts[flags.uldci.srsResIndicator] {
		case "port1":
			apCbPusch = []int{1000}
		case "ports2":
			apCbPusch = []int{1000, 1001}
		case "ports4":
			apCbPusch = []int{1000, 1001, 1002, 1003}
		}
		fmt.Printf("CB PUSCH using antenna port(s): %v - %v\n", flags.srs.srsNumPorts[flags.uldci.srsResIndicator], apCbPusch)

		numAp := len(apCbPusch)
		tp := flags.pusch.puschTp
		// maxRank IE of PUSCH-Config, 38.331 vh30
		//  Subset of PMIs addressed by TRIs from 1 to ULmaxRank (see TS 38.214 [19], clause 6.1.1.1). The field maxRank applies to DCI format 0_1 and the field maxRankDCI-0-2 applies to DCI format 0_2 (see TS 38.214 [19], clause 6.1.1.1).
		maxRank := flags.pusch.puschCbMaxRankNonCbMaxLayers
		cbSubset := flags.pusch.puschCbSubset
		precoding := flags.uldci.precodingInfoNumLayers

		key := fmt.Sprintf("%v_%v", map[string]int{"fullyAndPartialAndNonCoherent": 0, "partialAndNonCoherent": 1, "nonCoherent": 2}[cbSubset], precoding)
		if numAp == 4 && tp == "disabled" && utils.ContainsInt([]int{2, 3, 4}, maxRank) {
			p, exist := nrgrid.Dci01TpmiAp4Tp0MaxRank234[key]
			if !exist || p == nil {
				return errors.New(fmt.Sprintf("Invalid key(=%v) when referring Dci01TpmiAp4Tp0MaxRank234!", key))
			}
			rank, tpmi = p[0], p[1]
			if tpmi >= 0 && tpmi <= 11 {
				coherence = "nonCoherent"
			} else if tpmi >= 12 && tpmi <= 31 {
				coherence = "partialCoherent"
			} else {
				coherence = "fullyCoherent"
			}
		} else if numAp == 4 && (tp == "enabled" || (tp == "disabled" && maxRank == 1)) {
			p, exist := nrgrid.Dci01TpmiAp4Tp1OrTp0MaxRank1[key]
			if !exist || p == nil {
				return errors.New(fmt.Sprintf("Invalid key(=%v) when referring Dci01TpmiAp4Tp1OrTp0MaxRank1!", key))
			}
			rank, tpmi = p[0], p[1]
			if tpmi >= 0 && tpmi <= 3 {
				coherence = "nonCoherent"
			} else if tpmi >= 4 && tpmi <= 11 {
				coherence = "partialCoherent"
			} else {
				coherence = "fullyCoherent"
			}
		} else if numAp == 2 && tp == "disabled" && maxRank == 2 {
			p, exist := nrgrid.Dci01TpmiAp2Tp0MaxRank2[key]
			if !exist || p == nil {
				return errors.New(fmt.Sprintf("Invalid key(=%v) when referring Dci01TpmiAp2Tp0MaxRank2!", key))
			}
			rank, tpmi = p[0], p[1]
			if tpmi >= 0 && tpmi <= 2 {
				coherence = "nonCoherent"
			} else {
				coherence = "fullyCoherent"
			}
		} else if numAp == 2 && (tp == "enabled" || (tp == "disabled" && maxRank == 1)) {
			p, exist := nrgrid.Dci01TpmiAp2Tp1OrTp0MaxRank1[key]
			if !exist || p == nil {
				return errors.New(fmt.Sprintf("Invalid key(=%v) when referring Dci01TpmiAp2Tp1OrTp0MaxRank1!", key))
			}
			rank, tpmi = p[0], p[1]
			if tpmi >= 0 && tpmi <= 1 {
				coherence = "nonCoherent"
			} else {
				coherence = "fullyCoherent"
			}
		} else if numAp == 1 {
			rank = 1
			tpmi = -1
		} else {
			return errors.New(fmt.Sprintf("Invalid PUSCH configurations(numAp=%v, tp=%v, maxRank=%v)!", numAp, tp, maxRank))
		}

		if numAp > 1 {
			fmt.Printf("CB PUSCH Rank=%v, TPMI=%v, Coherence=%v (with PUSCH CB Subset=%v, DCI 0_1 Precoding Field=%v)\n", rank, tpmi, coherence, cbSubset, precoding)
		} else {
			fmt.Printf("CB PUSCH Rank=%v (with PUSCH CB Subset=%v)\n", rank, cbSubset)
		}
	} else {
		tokens := strings.Split(flags.srs.srsSetResIdList[1], "_")
		Nsrs := len(tokens)

		// maxNumberMIMO-LayersNonCB-PUSCH, 38.306 vh30
		//  Defines supported maximum number of MIMO layers at the UE for PUSCH transmission using non-codebook precoding.
		// or maxMIMO-Layers IE of PUSCH-ServingCellConfig, 38.331 vh30
		//  Indicates the maximum MIMO layer to be used for PUSCH in all BWPs of the corresponding UL of this serving cell (see TS 38.212 [17], clause 5.4.2.1).
		Lmax := flags.pusch.puschCbMaxRankNonCbMaxLayers

		var apNonCbPusch []int
		if Nsrs == 1 {
			rank = 1
			apNonCbPusch = []int{1000}
		} else {
			key := fmt.Sprintf("%v_%v_%v", Lmax, Nsrs, flags.uldci.srsResIndicator)
			p, exist := nrgrid.Dci01NonCbSri[key]
			if !exist || p == nil {
				return errors.New(fmt.Sprintf("Invalid key(=%v) when referring Dci01NonCbSri!", key))
			}
			rank = len(p)

			// 3GPP 38.214 vh40
			// 6.1.1.2	Non-Codebook based UL transmission
			// The UE shall transmit PUSCH using the same antenna ports as the SRS port(s) in the SRS resource(s) indicated by SRI(s) given by DCI format 0_1 or 0_2 or by configuredGrantConfig according to clause 6.1.2.3, where the SRS port in (i+1)-th SRS resource in the SRS resource set is indexed as pi=1000+i.
			for _, sri := range p {
				apNonCbPusch = append(apNonCbPusch, 1000+sri)
			}

			flags.pusch._nonCbSri = p
		}

		fmt.Printf("NonCB PUSCH using antenna port(s): %v\n", apNonCbPusch)
	}

	// update DMRS for PUSCH
	tp := flags.pusch.puschTp
	dmrsType := flags.pusch.puschDmrsType
	maxLen := flags.pusch.puschMaxLength
	key := fmt.Sprintf("%v_%v_%v_%v_%v", map[string]int{"disabled": 0, "enabled": 1}[tp], dmrsType[len(dmrsType)-1:], maxLen[len(maxLen)-1:], rank, flags.uldci.antennaPorts)
	p, exist := nrgrid.Dci01AntPorts[key]
	if !exist || p == nil {
		return errors.New(fmt.Sprintf("Invalid key(=%v) when referring Dci01AntPorts!", key))
	}
	fmt.Printf("AntPortsInfo(DMRS for PUSCH): %v\n", *p)
	flags.pusch._cdmGroupsWoData = p.CdmGroups
	flags.pusch._dmrsPorts = p.DmrsPorts
	flags.pusch._numFrontLoadSymbs = p.NumDmrsSymbs

	// 3GPP 38.214 vh40
	// 6.1.1.1	Codebook based UL transmission
	// 6.1.1.2	Non-Codebook based UL transmission
	// The DM-RS antenna ports[p~{0}...p~{v-1}]in Clause 6.4.1.1.3 of [4, TS38.211] are determined according to the ordering of DM-RS port(s) given by Tables 7.3.1.1.2-6 to 7.3.1.1.2-23 in Clause 7.3.1.1.2 of [5, TS 38.212].
	if rank != len(flags.pusch._dmrsPorts) {
		return errors.New(fmt.Sprintf("Inconsistent PUSCH rank and DMRS ports!(rank=%v, dmrsPorts=%v)", rank, flags.pusch._dmrsPorts))
	}

	// 3GPP 38.212 vh40
	// 7.3.1.1.2	Format 0_1
	// -	Frequency hopping flag – 0 or 1 bit:
	// 	-	1 bit according to Table 7.3.1.1.1-3 otherwise, only applicable to resource allocation type 1, as defined in Clause 6.3 of [6, TS 38.214].
	if flags.uldci.fdFreqHop[DCI_01_PUSCH] != "disabled" && flags.uldci._fdRaType[DCI_01_PUSCH] != "raType1" {
		return errors.New(fmt.Sprintf("Frequency hopping is only applicable to PUSCH resource allocation type 1!"))
	}

	// determine TD/FD pattern of DMRS for PUSCH
	flags.pusch._tdL, flags.pusch._tdL2, flags.pusch._fdK = getDmrsPuschTdFdPattern(dmrsType, flags.uldci._tdMappingType[DCI_01_PUSCH], flags.uldci._tdStartSymb[DCI_01_PUSCH], flags.uldci._tdNumSymbs[DCI_01_PUSCH], p.NumDmrsSymbs, flags.pusch.puschDmrsAddPos, p.CdmGroups, flags.uldci.fdFreqHop[DCI_01_PUSCH])
	if flags.uldci.fdFreqHop[DCI_01_PUSCH] != "intra-slot" {
		fmt.Printf("TD pattern within a slot of DMRS for PUSCH (DCI 0_1): %v\n", flags.pusch._tdL)
	} else {
		fmt.Printf("TD pattern within a slot of DMRS for PUSCH (DCI 0_1) (1st hop): %v\n", flags.pusch._tdL)
		fmt.Printf("TD pattern within a slot of DMRS for PUSCH (DCI 0_1) (2nd hop): %v\n", flags.pusch._tdL2)
	}
	fmt.Printf("FD pattern within a PRB of DMRS for PUSCH (DCI 0_1): %v\n", flags.pusch._fdK)

	// update PTRS for PUSCH
	// 3GPP 38.214 vh40
	// 6.2.2	UE DM-RS transmission procedure
	// If a UE transmitting PUSCH scheduled by DCI format 0_2 is configured with the higher layer parameter phaseTrackingRS in dmrs-UplinkForPUSCH-MappingTypeA-DCI-0-2 or dmrs-UplinkForPUSCH-MappingTypeB-DCI-0-2, or a UE transmitting PUSCH scheduled by DCI format 0_0 or DCI format 0_1 is configured with the higher layer parameter phaseTrackingRS in dmrs-UplinkForPUSCH-MappingTypeA or dmrs-UplinkForPUSCH-MappingTypeB, the UE may assume that the following configurations are not occurring simultaneously for the transmitted PUSCH
	//   - any DM-RS ports among 4-7 or 6-11 for DM-RS configurations type 1 and type 2, respectively are scheduled for the UE and PT-RS is transmitted from the UE.
	var dmrsApSetNoPtrs []int
	if flags.pusch.puschDmrsType == "type1" {
		dmrsApSetNoPtrs = utils.PyRange(4, 8, 1)
	} else {
		dmrsApSetNoPtrs = utils.PyRange(6, 12, 1)
	}
	noPtrs := false
	for _, ap := range flags.pusch._dmrsPorts {
		if utils.ContainsInt(dmrsApSetNoPtrs, ap) {
			noPtrs = true
			break
		}
	}
	fmt.Printf("PUSCH noPTRS=%v\n", noPtrs)

	if !noPtrs && flags.pusch.puschTp == "enabled" {
		flags.pusch._ptrsDmrsPorts = flags.pusch._dmrsPorts
	} else if !noPtrs {
		flags.pusch._ptrsDmrsPorts = []int{}
		if flags.pusch.puschTxCfg == "codebook" {
			if flags.pusch.puschPtrsMaxNumPorts == "n1" {
				// refer to 38.214 vh40
				// 6.2.3.1	UE PT-RS transmission procedure when transform precoding is not enabled
				// If a UE has reported the capability of supporting full-coherent UL transmission, the UE shall expect the number of UL PT-RS ports to be configured as one if UL-PTRS is configured

				// refer to 38.212 vh40
				// Table 7.3.1.1.2-25: PTRS-DMRS association or Second PTRS-DMRS association for UL PTRS port 0
				if flags.uldci.ptrsDmrsAssociation < len(flags.pusch._dmrsPorts) {
					flags.pusch._ptrsDmrsPorts = []int{flags.pusch._dmrsPorts[flags.uldci.ptrsDmrsAssociation]}
				} else {
					return errors.New(fmt.Sprintf("Invalid ptrsDmrsAssociation(=%v) where len(dmrsPorts)=%v!", flags.uldci.ptrsDmrsAssociation, len(flags.pusch._dmrsPorts)))
				}
			} else if coherence != "fullyCoherent" {
				key = fmt.Sprintf("%v_%v_%v", flags.srs.srsNumPorts[flags.uldci.srsResIndicator], rank, tpmi)
				p, exist := nrgrid.CbPuschTpmiDmrsAssociation[key]
				if !exist || p == nil {
					return errors.New(fmt.Sprintf("Invalid key(=%v) when referring CbPuschTpmiDmrsAssociation!", key))
				}
				fmt.Printf("CbPuschTpmiDmrsAssociation: %v\n", p)

				ptrsDmrsMapping := make([][]int, 2)
				for i := 0; i < len(p); i++ {
					for _, ap := range strings.Split(p[i], ",") {
						if ap == "-" {
							break
						} else {
							apv, _ := strconv.Atoi(ap)

							if (i == 0 || i == 2) && !utils.ContainsInt(ptrsDmrsMapping[0], apv) {
								ptrsDmrsMapping[0] = append(ptrsDmrsMapping[0], apv)
							} else if (i == 1 || i == 3) && !utils.ContainsInt(ptrsDmrsMapping[1], apv) {
								ptrsDmrsMapping[1] = append(ptrsDmrsMapping[1], apv)
							}
						}
					}
				}

				fmt.Printf("CB PUSCH ptrsDmrsMapping=%v\n", ptrsDmrsMapping)

				// refer to 38.212 vh40
				// Table 7.3.1.1.2-26: PTRS-DMRS association or Second PTRS-DMRS association for UL PTRS ports 0 and 1
				var msb, lsb int
				switch flags.uldci.ptrsDmrsAssociation {
				case 0:
					msb = 0
					lsb = 0
				case 1:
					msb = 0
					lsb = 1
				case 2:
					msb = 1
					lsb = 0
				case 3:
					msb = 1
					lsb = 1
				}

				if (msb == 0 && len(ptrsDmrsMapping[0]) > 0) || (msb == 1 && len(ptrsDmrsMapping[0]) == 2) {
					flags.pusch._ptrsDmrsPorts = append(flags.pusch._ptrsDmrsPorts, ptrsDmrsMapping[0][msb])
				}

				if (lsb == 0 && len(ptrsDmrsMapping[0]) > 0) || (lsb == 1 && len(ptrsDmrsMapping[1]) == 2) {
					flags.pusch._ptrsDmrsPorts = append(flags.pusch._ptrsDmrsPorts, ptrsDmrsMapping[1][lsb])
				}
			}
		} else {
			var nonCbSrsResIds []int
			for _, v := range strings.Split(flags.srs.srsSetResIdList[1], "_") {
				vi, _ := strconv.Atoi(v)
				nonCbSrsResIds = append(nonCbSrsResIds, vi)
			}

			ptrsDmrsMapping := map[int][]int{0: {}, 1: {}}
			for i, sri := range flags.pusch._nonCbSri {
				if flags.srs._srsNonCbPtrsPort[nonCbSrsResIds[sri]] == "n0" {
					ptrsDmrsMapping[0] = append(ptrsDmrsMapping[0], flags.pusch._dmrsPorts[i])
				} else {
					ptrsDmrsMapping[1] = append(ptrsDmrsMapping[1], flags.pusch._dmrsPorts[i])
				}
			}

			// determine number of PTRS antenna port(s)
			var numPtrsAp int
			if len(ptrsDmrsMapping[0]) > 0 && len(ptrsDmrsMapping[1]) == 0 {
				numPtrsAp = 1
			} else if utils.ContainsInt([]int{1, 2}, len(ptrsDmrsMapping[0])) && utils.ContainsInt([]int{1, 2}, len(ptrsDmrsMapping[1])) {
				numPtrsAp = 2
			} else {
				return errors.New(fmt.Sprintf("Invalid SRS setting for nonCodebook PUSCH! (ptrsDmrsMapping=%v)", ptrsDmrsMapping))
			}

			fmt.Printf("non-CB PUSCH ptrsDmrsMapping=%v\n", ptrsDmrsMapping)

			// determine associated DMRS port per PTRS port
			if numPtrsAp == 1 {
				// refer to 38.212 vh40
				// Table 7.3.1.1.2-25: PTRS-DMRS association or Second PTRS-DMRS association for UL PTRS port 0
				if flags.uldci.ptrsDmrsAssociation < len(flags.pusch._dmrsPorts) {
					flags.pusch._ptrsDmrsPorts = []int{flags.pusch._dmrsPorts[flags.uldci.ptrsDmrsAssociation]}
				} else {
					return errors.New(fmt.Sprintf("Invalid ptrsDmrsAssociation(=%v) where len(dmrsPorts)=%v!", flags.uldci.ptrsDmrsAssociation, len(flags.pusch._dmrsPorts)))
				}
			} else {
				// refer to 38.212 vh40
				// Table 7.3.1.1.2-26: PTRS-DMRS association or Second PTRS-DMRS association for UL PTRS ports 0 and 1
				var msb, lsb int
				switch flags.uldci.ptrsDmrsAssociation {
				case 0:
					msb = 0
					lsb = 0
				case 1:
					msb = 0
					lsb = 1
				case 2:
					msb = 1
					lsb = 0
				case 3:
					msb = 1
					lsb = 1
				}

				if msb == 0 || (msb == 1 && len(ptrsDmrsMapping[0]) == 2) {
					flags.pusch._ptrsDmrsPorts = append(flags.pusch._ptrsDmrsPorts, ptrsDmrsMapping[0][msb])
				}

				if lsb == 0 || (lsb == 1 && len(ptrsDmrsMapping[1]) == 2) {
					flags.pusch._ptrsDmrsPorts = append(flags.pusch._ptrsDmrsPorts, ptrsDmrsMapping[1][lsb])
				}
			}
		}

		fmt.Printf("DMRS port(s) with associated PTRS for PUSCH: %v\n", flags.pusch._ptrsDmrsPorts)
	}

	// update PUSCH TBS
	fdRaType := flags.uldci._fdRaType[DCI_01_PUSCH]
	fdRa := flags.uldci._fdRa[DCI_01_PUSCH]
	// 38.214 vh40 6.1.2.2	Resource allocation in frequency domain
	// Uplink resource allocation scheme type 0 is supported for PUSCH only when transform precoding is disabled.
	// Uplink resource allocation scheme type 1 and type 2 are supported for PUSCH for both cases when transform precoding is enabled or disabled.
	if tp == "enabled" && fdRaType == "raType0" {
		return errors.New(fmt.Sprintf("Uplink resource allocation scheme type 0 is supported for PUSCH only when transform precoding is disabled.\n"))
	}
	if fdRaType == "raType0" && len(fdRa) != flags.uldci._fdBitsRaType0 {
		return errors.New(fmt.Sprintf("Invalid 'Frequency domain resource assignment' field of DCI 0_1: fdRaType=%v, fdRa=%v, len(fdRa)=%v, bitsRaType0=%v\n", fdRaType, fdRa, len(fdRa), flags.uldci._fdBitsRaType0))
	}

	fd := 0
	if fdRaType == "raType0" {
		rbgs := getRaType0Rbgs(flags.bwp._bwpStartRb[DED_UL_BWP], flags.bwp._bwpNumRbs[DED_UL_BWP], flags.pdsch._rbgSize)
		for i, c := range fdRa {
			if c == '1' {
				fd += rbgs[i]
			}
		}
	} else {
		fd = flags.uldci.fdNumRbs[DCI_01_PUSCH]

		// validate L_RBs when transform precoding is enabled
		if tp == "enabled" {
			// valid PUSCH PRB allocations when transforming precoding is enabled
			lrbsPuschTp := initLrbsPuschTp(flags.bwp._bwpNumRbs[DED_UL_BWP])

			if !utils.ContainsInt(lrbsPuschTp, fd) {
				lt, gt := utils.NearestInt(lrbsPuschTp, fd)
				return errors.New(fmt.Sprintf("L_RBs(=%v) must be 2^x*3^y*5^z, where x/y/z>=0, when transforming precoding is enabled, and the nearest values are: [%v, %v].\n", fd, lt, gt))
			}
		}

		// update FDRA
		riv, err := makeRiv(flags.uldci.fdNumRbs[DCI_01_PUSCH], flags.uldci.fdStartRb[DCI_01_PUSCH], flags.bwp._bwpNumRbs[DED_UL_BWP])
		if err != nil {
			return err
		}
		flags.uldci._fdRa[DCI_01_PUSCH] = fmt.Sprintf("%0*b", flags.uldci._fdBitsRaType1[DCI_01_PUSCH], riv)
		fmt.Printf("PUSCH(tag=%v): RIV=%v, FDRA bits=%v\n", flags.uldci._tag[DCI_01_PUSCH], riv, flags.uldci._fdRa[DCI_01_PUSCH])
		if flags.uldci.fdFreqHop[DCI_01_PUSCH] != "disabled" {
			var ulHopBits int
			if flags.bwp._bwpNumRbs[DED_UL_BWP] >= 50 {
				ulHopBits = 2
			} else {
				ulHopBits = 1
			}

			v, _ := strconv.Atoi(flags.uldci._fdRa[DCI_01_PUSCH][:ulHopBits])
			if v != 0 {
				return errors.New(fmt.Sprintf("The first %v bits of RIV must be all zeros when frequency hopping is enabled!", ulHopBits))
			}
		}
	}

	// calculate DMRS overhead
	td := flags.uldci._tdNumSymbs[DCI_01_PUSCH]
	ld := 0
	tdMappingType := flags.uldci._tdMappingType[DCI_01_PUSCH]
	dmrsAddPos := flags.pusch.puschDmrsAddPos
	freqHop := flags.uldci.fdFreqHop[DCI_01_PUSCH]
	if freqHop == "intra-slot" {
		// refer to 3GPP 38.211 vh40 6.4.1.1.3
		// ld is the duration per hop according to Table 6.4.1.1.3-6 if intra-slot frequency hopping is used
		// refer to 3GPP 38.214 vf30 6.3
		// In case of intra-slot frequency hopping, ... The number of symbols in the first hop is given by floor(N_PUSCH_symb/2) , the number of symbols in the second hop is given by N_PUSCH_symb - floor(N_PUSCH_symb/2) , where N_PUSCH_symb is the length of the PUSCH transmission in OFDM symbols in one slot.
		ld1 := utils.FloorInt(float64(td) / 2)
		ld2 := td - ld1

		// refer to 3GPP 38.211 vh40 6.4.1.1.3
		// ...and the position l0 of the first DM-RS symbol depends on the mapping type:
		//  -	for PUSCH mapping type A:
		//    -	l0 is given by the higher-layer parameter dmrs-TypeA-Position
		//  -	for PUSCH mapping type B:
		//    -	l0 = 0
		var l0 int
		if tdMappingType == "typeA" {
			l0, _ = strconv.Atoi(flags.gridsetting.dmrsTypeAPos[3:])
		} else {
			l0 = 0
		}

		// refer to 3GPP 38.211 vh40 6.4.1.1.3
		// if the higher-layer parameter dmrs-AdditionalPosition is not set to 'pos0' and intra-slot frequency hopping is enabled according to clause 7.3.1.1.2 in [4, TS 38.212] and by higher layer, Tables 6.4.1.1.3-6 shall be used assuming dmrs-AdditionalPosition is equal to 'pos1' for each hop.
		if dmrsAddPos != "pos0" {
			dmrsAddPos = "pos1"
		}

		key1 := fmt.Sprintf("%v_%v_%v_%v_1st", ld1, tdMappingType, l0, dmrsAddPos)
		key2 := fmt.Sprintf("%v_%v_%v_%v_2nd", ld2, tdMappingType, l0, dmrsAddPos)
		if flags.pusch._numFrontLoadSymbs == 1 {
			p1, e1 := nrgrid.DmrsPuschPosOneSymbWithIntraSlotFh[key1]
			p2, e2 := nrgrid.DmrsPuschPosOneSymbWithIntraSlotFh[key2]
			if !e1 || p1 == nil {
				return errors.New(fmt.Sprintf("Invalid key(=%v) when referring DmrsPuschPosOneSymbWithIntraSlotFh!", key1))
			}
			if !e2 || p2 == nil {
				return errors.New(fmt.Sprintf("Invalid key(=%v) when referring DmrsPuschPosOneSymbWithIntraSlotFh!", key2))
			}
			flags.pusch._dmrsPosLBar = p1
			flags.pusch._dmrsPosLBarSecondHop = p2
		} else {
			return errors.New(fmt.Sprintf("Only single-symbol front-load DMRS for PUSCH is supported when intra-slot frequency hopping is enabled! (number of front-load DMRS = %v)", flags.pusch._numFrontLoadSymbs))
		}
	} else {
		// refer to 38.211 vh40 6.4.1.1.3
		// ld is the duration between the first OFDM symbol of the slot and the last OFDM symbol of the scheduled PUSCH resources in the slot for PUSCH mapping type A according to Tables 6.4.1.1.3-3 and 6.4.1.1.3-4 if intra-slot frequency hopping is not used, or
		// ld is the duration of scheduled PUSCH resources for PUSCH mapping type B according to Tables 6.4.1.1.3-3 and 6.4.1.1.3-4 if intra-slot frequency hopping is not used
		if tdMappingType == "typeA" {
			ld = flags.uldci._tdStartSymb[DCI_01_PUSCH] + td
		} else {
			ld = td
		}

		key = fmt.Sprintf("%v_%v_%v", ld, tdMappingType, dmrsAddPos)
		if flags.pusch._numFrontLoadSymbs == 1 {
			p, e := nrgrid.DmrsPuschPosOneSymbWoIntraSlotFh[key]
			if !e || p == nil {
				return errors.New(fmt.Sprintf("Invalid key(=%v) when referring DmrsPuschPosOneSymbWoIntraSlotFh!", key))
			}
			flags.pusch._dmrsPosLBar = p
		} else {
			p, e := nrgrid.DmrsPuschPosTwoSymbsWoIntraSlotFh[key]
			if !e || p == nil {
				return errors.New(fmt.Sprintf("Invalid key(=%v) when referring DmrsPuschPosTwoSymbsWoIntraSlotFh!", key))
			}
			flags.pusch._dmrsPosLBar = p
		}
	}

	// refer to 38.211 vh40 6.4.1.1.3
	// For PUSCH mapping type A,
	//  - the case dmrs-AdditionalPosition is equal to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2';
	//  - ld=4 symbols in Table 6.4.1.1.3-4 is only applicable when dmrs-TypeA-Position is equal to 'pos2'.
	// Table 6.4.1.1.3-4: PUSCH DM-RS positions l_bar within a slot for double-symbol DM-RS and intra-slot frequency hopping disabled.
	if tdMappingType == "typeA" && dmrsAddPos == "pos3" && flags.gridsetting.dmrsTypeAPos != "pos2" {
		return errors.New(fmt.Sprintf("For PUSCH mapping type A, the case dmrs-AdditionalPosition is equal to 'pos3' is only supported when dmrs-TypeA-Position is equal to 'pos2'!(dmrsTypeAPos=%v)", flags.gridsetting.dmrsTypeAPos))
	}
	if tdMappingType == "typeA" && ld == 4 && flags.pusch._numFrontLoadSymbs == 2 && flags.gridsetting.dmrsTypeAPos != "pos2" {
		return errors.New(fmt.Sprintf("For PUSCH mapping type A, ld=4 symbols in Table 6.4.1.1.3-4 is only applicable when dmrs-TypeA-Position is equal to 'pos2'!(dmrsTypeAPos=%v)", flags.gridsetting.dmrsTypeAPos))
	}

	var dmrsOh int
	if freqHop == "intra-slot" {
		dmrsOh = (2 * flags.pusch._cdmGroupsWoData) * (len(flags.pusch._dmrsPosLBar) + len(flags.pusch._dmrsPosLBarSecondHop))
		fmt.Printf("PUSCH(tag=%v) DMRS overhead: cdmGroupsWoData=%v, l_bar of 1st hop=%v, l_bar of 2nd hop=%v\n", flags.uldci._tag[DCI_01_PUSCH], flags.pusch._cdmGroupsWoData, flags.pusch._dmrsPosLBar, flags.pusch._dmrsPosLBarSecondHop)
	} else {
		dmrsOh = (2 * flags.pusch._cdmGroupsWoData) * len(flags.pusch._dmrsPosLBar)
		fmt.Printf("PUSCH(tag=%v) DMRS overhead: cdmGroupsWoData=%v, l_bar=%v\n", flags.uldci._tag[DCI_01_PUSCH], flags.pusch._cdmGroupsWoData, flags.pusch._dmrsPosLBar)
	}

	xoh, _ := strconv.Atoi(flags.pusch.puschXOh[3:])
	tbs, err := getTbs("PUSCH", flags.pusch.puschTp == "enabled", "C-RNTI", flags.pusch.puschMcsTable, td, fd, flags.uldci.mcsCw0[DCI_01_PUSCH], len(flags.pusch._dmrsPorts), dmrsOh, xoh, 1)
	if err != nil {
		return err
	} else {
		fmt.Printf("PUSCH(tag=%v) CW0 TBS=%v bits\n", flags.uldci._tag[DCI_01_PUSCH], tbs)
		flags.uldci._tbs[DCI_01_PUSCH] = tbs
	}
	fmt.Println()

	return nil
}

// getRaType0Rbgs return RBGs for PDSCH/PUSCH resource allocation Type 0.
func getRaType0Rbgs(bwpStart, bwpSize, P int) []int {
	//regYellow.Printf("-->calling getRaType0Rbgs\n")

	bitwidth := utils.CeilInt((float64(bwpSize) + float64(bwpStart%P)) / float64(P))
	rbgs := make([]int, bitwidth)
	for i := 0; i < bitwidth; i++ {
		if i == 0 {
			rbgs[i] = P - bwpStart%P
		} else if i == bitwidth-1 {
			if (bwpStart+bwpSize)%P > 0 {
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

//getTbs calculates TBS for PUSCH/PDSCH.
//	sch: PUSCH or PDSCH
//	tp: PUSCH transform percoding flag
//	rnti: C-RNTI, SI-RNTI, RA-RNTI, TC-RNTI or MSG3
//	mcsTab: qam64, qam64LowSE, qam256 or qam1024
//	td: number of symbols
//	fd: number of PRBs
//	mcs: MCS
//	layer: number of spatial multiplexing layers
//	dmrs: overhead of DMRS
//	xoh: the xOverhead
//	scale: TB scaling for Msg2
func getTbs(sch string, tp bool, rnti string, mcsTab string, td int, fd int, mcs int, layer int, dmrs int, xoh int, scale float64) (int, error) {
	// regYellow.Printf("-->calling getTbs\n")

	rntiSet := []string{"C-RNTI", "SI-RNTI", "RA-RNTI", "TC-RNTI", "MSG3"}
	mcsTabSet := []string{"qam1024", "qam256", "qam64", "qam64LowSE"}

	if !utils.ContainsStr(rntiSet, rnti) || !utils.ContainsStr(mcsTabSet, mcsTab) {
		return 0, errors.New(fmt.Sprintf("Invalid RNTI or MCS table!\n"))
	}

	// refer to 3GPP TS 38.214 vh40
	// 5.1.3	Modulation order, target code rate, redundancy version and transport block size determination
	// 6.1.4	Modulation order, redundancy version and transport block size determination
	// 1st step: get Qm and R(x1024)
	var p *nrgrid.McsInfo
	if sch == "PDSCH" || (sch == "PUSCH" && !tp) {
		if sch == "PDSCH" && rnti == "C-RNTI" && mcsTab == "qam1024" {
			p = nrgrid.PdschMcsTabQam1024[mcs]
		} else if rnti == "C-RNTI" && mcsTab == "qam256" {
			p = nrgrid.PdschMcsTabQam256[mcs]
		} else if rnti == "C-RNTI" && mcsTab == "qam64LowSE" {
			p = nrgrid.PdschMcsTabQam64LowSE[mcs]
		} else {
			p = nrgrid.PdschMcsTabQam64[mcs]
		}
	} else if sch == "PUSCH" && tp {
		if rnti == "C-RNTI" && mcsTab == "qam256" {
			p = nrgrid.PdschMcsTabQam256[mcs]
		} else if rnti == "C-RNTI" && mcsTab == "qam64LowSE" {
			p = nrgrid.PuschTpMcsTabQam64LowSE[mcs]
		} else {
			p = nrgrid.PuschTpMcsTabQam64[mcs]
		}
	}

	if p == nil {
		return 0, errors.New(fmt.Sprintf("Invalid MCS: sch=%v, tp=%v, rnti=%v, mcsTab=%v, mcs=%v\n", sch, tp, rnti, mcsTab, mcs))
	}
	Qm, R := p.ModOrder, p.CodeRate

	// The UE is not expected to decode a PDSCH scheduled with P-RNTI, RA-RNTI, SI-RNTI and Qm > 2.
	// FIXME: assume PDSCH scheduled with TC-RNTI has the same restraint.
	if (rnti == "RA-RNTI" || rnti == "SI-RNTI" || rnti == "TC-RNTI") && Qm > 2 {
		return 0, errors.New(fmt.Sprintf("The UE is not expected to decode a PDSCH scheduled with P-RNTI, RA-RNTI, SI-RNTI and Qm > 2.\nMcsInfo=%v\n", *p))
	}

	// 2nd step: get N_RE
	N_RE_ap := 12*td - dmrs - xoh
	min := utils.MinInt([]int{156, N_RE_ap})
	N_RE := min * fd

	// 3rd step: get N_info
	N_info := scale * float64(N_RE) * (R / 1024) * float64(Qm) * float64(layer)

	// 4th step: get TBS
	var tbs int
	if N_info <= 3824 {
		n := utils.MaxInt([]int{3, utils.FloorInt(math.Log2(N_info)) - 6})
		n2 := 1 << n
		N_info_ap := utils.MaxInt([]int{24, n2 * utils.FloorInt(N_info/float64(n2))})
		for _, v := range nrgrid.TbsTabLessThan3824 {
			if v >= N_info_ap {
				tbs = v
				break
			}
		}
	} else {
		n := utils.FloorInt(math.Log2(N_info-24)) - 5
		n2 := 1 << n
		N_info_ap := utils.MaxInt([]int{3840, n2 * utils.RoundInt((N_info-24)/float64(n2))})
		if R <= 256 {
			C := utils.CeilInt(float64(N_info_ap+24) / 3816)
			tbs = 8*C*utils.CeilInt(float64(N_info_ap+24)/float64(8*C)) - 24
		} else {
			if N_info_ap > 8424 {
				C := utils.CeilInt(float64(N_info_ap+24) / 8424)
				tbs = 8*C*utils.CeilInt(float64(N_info_ap+24)/float64(8*C)) - 24
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

// getDmrsPdschTdFdPattern determines the DMRS for PDSCH TD/FD pattern
//  dmrsType: DMRS configuration type, which can be type1 or type2
//	tdMappingType: PDSCH mapping type, which can be typeA or typeB
//	slivS: the S of PDSCH SLIV
//	slivL: the L of PDSCH SLIV
//	numFrontLoadSymbs: number of front-load OFDM symbol(s) of DMRS for PDSCH
//	dmrsAddPos: the dmrs-AdditionalPosition
//	cdmGroupsWoData: number of CDM group(s) without data
func getDmrsPdschTdFdPattern(dmrsType string, tdMappingType string, slivS int, slivL int, numFrontLoadSymbs int, dmrsAddPos string, cdmGroupsWoData int) ([]int, []int) {
	// determine TD pattern
	var tdL0, tdLd int
	if tdMappingType == "typeA" {
		tdL0, _ = strconv.Atoi(flags.gridsetting.dmrsTypeAPos[3:])
		tdLd = slivS + slivL
	} else {
		tdL0 = 0
		tdLd = slivL
	}

	var tdLbar, tdLap []int
	if numFrontLoadSymbs == 1 {
		tdLbar = nrgrid.DmrsPdschPosOneSymb[fmt.Sprintf("%v_%v_%v", tdLd, tdMappingType, dmrsAddPos)]
		tdLap = []int{0}
	} else {
		tdLbar = nrgrid.DmrsPdschPosTwoSymbs[fmt.Sprintf("%v_%v_%v", tdLd, tdMappingType, dmrsAddPos)]
		tdLap = []int{0, 1}
	}

	// replace tdLbar[0] with l0
	tdLbar[0] = tdL0

	var tdL []int
	for _, i := range tdLbar {
		for _, j := range tdLap {
			tdL = append(tdL, i+j)
		}
	}

	// determine FD pattern within a single PRB
	var fdN, fdKap, fdDelta []int
	if dmrsType == "type1" {
		fdN = []int{0, 1, 2}
		fdKap = []int{0, 1}
		fdDelta = []int{0, 1}[:cdmGroupsWoData]
	} else {
		fdN = []int{0, 1}
		fdKap = []int{0, 1}
		fdDelta = []int{0, 2, 4}[:cdmGroupsWoData]
	}

	fdK := make([]int, 12)
	for _, i := range fdN {
		for _, j := range fdKap {
			for _, k := range fdDelta {
				if dmrsType == "type1" {
					fdK[4*i+2*j+k] = 1
				} else {
					fdK[6*i+j+k] = 1
				}
			}
		}
	}

	return tdL, fdK
}

// getDmrsPuschTdFdPattern determines the DMRS for PUSCH TD/FD pattern
//  dmrsType: DMRS configuration type, which can be type1 or type2
//	tdMappingType: PUSCH mapping type, which can be typeA or typeB
//	slivS: the S of PUSCH SLIV
//	slivL: the L of PUSCH SLIV
//	numFrontLoadSymbs: number of front-load OFDM symbol(s) of DMRS for PUSCH
//	dmrsAddPos: the dmrs-AdditionalPosition
//	cdmGroupsWoData: number of CDM group(s) without data
//	freqHop: indicate whether intra-slot frequency hopping is enabled
func getDmrsPuschTdFdPattern(dmrsType string, tdMappingType string, slivS int, slivL int, numFrontLoadSymbs int, dmrsAddPos string, cdmGroupsWoData int, freqHop string) ([]int, []int, []int) {
	// determine TD pattern
	var tdL0, tdLd int
	var tdLbar, tdLap []int
	var tdL, tdL2 []int
	if freqHop != "intra-slot" {
		if tdMappingType == "typeA" {
			tdL0, _ = strconv.Atoi(flags.gridsetting.dmrsTypeAPos[3:])
			tdLd = slivS + slivL
		} else {
			tdL0 = 0
			tdLd = slivL
		}

		if numFrontLoadSymbs == 1 {
			tdLbar = nrgrid.DmrsPuschPosOneSymbWoIntraSlotFh[fmt.Sprintf("%v_%v_%v", tdLd, tdMappingType, dmrsAddPos)]
			tdLap = []int{0}
		} else {
			tdLbar = nrgrid.DmrsPuschPosTwoSymbsWoIntraSlotFh[fmt.Sprintf("%v_%v_%v", tdLd, tdMappingType, dmrsAddPos)]
			tdLap = []int{0, 1}
		}
		// replace tdLbar[0] with l0
		tdLbar[0] = tdL0

		for _, i := range tdLbar {
			for _, j := range tdLap {
				tdL = append(tdL, i+j)
			}
		}
	} else {
		tdLd1 := utils.FloorInt(float64(slivL) / 2)
		tdLd2 := slivL - tdLd1
		if tdMappingType == "typeA" {
			tdL0, _ = strconv.Atoi(flags.gridsetting.dmrsTypeAPos[3:])
		} else {
			tdL0 = 0
		}

		// refer to 3GPP 38.211 vh40 6.4.1.1.3
		// if the higher-layer parameter dmrs-AdditionalPosition is not set to 'pos0' and intra-slot frequency hopping is enabled according to clause 7.3.1.1.2 in [4, TS 38.212] and by higher layer, Tables 6.4.1.1.3-6 shall be used assuming dmrs-AdditionalPosition is equal to 'pos1' for each hop.
		if dmrsAddPos != "pos0" {
			dmrsAddPos = "pos1"
		}

		if numFrontLoadSymbs == 1 {
			// 1st hop
			tdLbar = nrgrid.DmrsPuschPosOneSymbWithIntraSlotFh[fmt.Sprintf("%v_%v_%v_%v_1st", tdLd1, tdMappingType, tdL0, dmrsAddPos)]
			tdLap = []int{0}
			for _, i := range tdLbar {
				for _, j := range tdLap {
					tdL = append(tdL, i+j)
				}
			}

			// 2nd hop
			tdLbar = nrgrid.DmrsPuschPosOneSymbWithIntraSlotFh[fmt.Sprintf("%v_%v_%v_%v_2nd", tdLd2, tdMappingType, tdL0, dmrsAddPos)]
			for _, i := range tdLbar {
				for _, j := range tdLap {
					tdL2 = append(tdL2, i+j)
				}
			}
		} else {
			return nil, nil, nil
		}
	}

	// determine FD pattern within a single PRB
	var fdN, fdKap, fdDelta []int
	if dmrsType == "type1" {
		fdN = []int{0, 1, 2}
		fdKap = []int{0, 1}
		fdDelta = []int{0, 1}[:cdmGroupsWoData]
	} else {
		fdN = []int{0, 1}
		fdKap = []int{0, 1}
		fdDelta = []int{0, 2, 4}[:cdmGroupsWoData]
	}

	fdK := make([]int, 12)
	for _, i := range fdN {
		for _, j := range fdKap {
			for _, k := range fdDelta {
				if dmrsType == "type1" {
					fdK[4*i+2*j+k] = 1
				} else {
					fdK[6*i+j+k] = 1
				}
			}
		}
	}

	return tdL, tdL2, fdK
}

// tddUlDlCmd represents the "nrrg tdduldl" command
var tddUlDlCmd = &cobra.Command{
	Use:   "tdduldl",
	Short: "",
	Long:  `CMD "nrrg tdduldl" can be used to get/set TDD-UL-DL-Config related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// searchSpaceCmd represents the "nrrg searchspace" command
var searchSpaceCmd = &cobra.Command{
	Use:   "searchspace",
	Short: "",
	Long:  `CMD "nrrg searchspace" can be used to get/set search space related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// dlDciCmd represents the "nrrg dldci" command
var dlDciCmd = &cobra.Command{
	Use:   "dldci",
	Short: "",
	Long:  `CMD "nrrg dldci" can be used to get/set DCI 1_0/1_1 related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// ulDciCmd represents the "nrrg uldci" command
var ulDciCmd = &cobra.Command{
	Use:   "uldci",
	Short: "",
	Long:  `CMD "nrrg uldci" can be used to get/set DCI 0_1 or RAR UL grant related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// bwpCmd represents the "nrrg bwp" command
var bwpCmd = &cobra.Command{
	Use:   "bwp",
	Short: "",
	Long:  `CMD "nrrg bwp" can be used to get/set generic BWP related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// rachCmd represents the "nrrg rach" command
var rachCmd = &cobra.Command{
	Use:   "rach",
	Short: "",
	Long:  `CMD "nrrg rach" can be used to get/set random access related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// dmrsCommonCmd represents the "nrrg dmrscommon" command
var dmrsCommonCmd = &cobra.Command{
	Use:   "dmrscommon",
	Short: "",
	Long:  `CMD "nrrg dmrscommon" can be used to get/set DMRS of SIB1/Msg2/Msg4/Msg3 related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// pdschCmd represents the "nrrg pdsch" command
var pdschCmd = &cobra.Command{
	Use:   "pdsch",
	Short: "",
	Long:  `CMD "nrrg pdsch" can be used to get/set PDSCH-config or PDSCH-ServingCellConfig related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()

		// process pdsch.pdschRbgCfg
		if cmd.Flags().Lookup("pdschRbgCfg").Changed {
			flags.pdsch._rbgSize = getNominalRbgSize(flags.bwp._bwpNumRbs[DED_DL_BWP], flags.pdsch.pdschRbgCfg)
		}

		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

func getNominalRbgSize(bwpSize int, config string) int {
	// refer to 38.214 vh40
	// Table 5.1.2.2.1-1:Nominal RBG size P
	// Table 6.1.2.2.1-1:Nominal RBG size P
	var cat int
	if bwpSize >= 1 && bwpSize <= 36 {
		cat = 0
	} else if bwpSize >= 37 && bwpSize <= 72 {
		cat = 1
	} else if bwpSize >= 73 && bwpSize <= 144 {
		cat = 2
	} else if bwpSize >= 145 && bwpSize <= 275 {
		cat = 3
	}

	if config == "config1" {
		return []int{2, 4, 8, 16}[cat]
	} else {
		return []int{4, 8, 16, 16}[cat]
	}
}

// puschCmd represents the "nrrg pusch" command
var puschCmd = &cobra.Command{
	Use:   "pusch",
	Short: "",
	Long:  `CMD "nrrg pusch" can be used to get/set PUSCH-config or PUSCH-ServingCellConfig related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()

		// process pusch.puschRbgCfg
		if cmd.Flags().Lookup("puschRbgCfg").Changed {
			flags.pusch._rbgSize = getNominalRbgSize(flags.bwp._bwpNumRbs[DED_UL_BWP], flags.pusch.puschRbgCfg)
		}

		// process pusch.puschPtrsGrpPatternTp
		if cmd.Flags().Lookup("puschPtrsGrpPatternTp").Changed {
			// 38.214 vh40 Table 6.2.3.2-1: PT-RS group pattern as a function of scheduled bandwidth
			// There are 5 categories of 'Scheduled bandwidth', which are simplified as pat0~pat4.
			v := map[string][]int{"pat0": {2, 2}, "pat1": {2, 4}, "pat2": {4, 2}, "pat3": {4, 4}, "pat4": {8, 4}}[flags.pusch.puschPtrsGrpPatternTp]
			flags.pusch._numGrpsTp = v[0]
			flags.pusch._samplesPerGrpTp = v[1]
		}

		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// csiCmd represents the "nrrg csi" command
var csiCmd = &cobra.Command{
	Use:   "csi",
	Short: "",
	Long:  `CMD "nrrg csi" can be used to get/set CSI-RS related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// srsCmd represents the "nrrg srs" command
var srsCmd = &cobra.Command{
	Use:   "srs",
	Short: "",
	Long:  `CMD "nrrg srs" can be used to get/set SRS-Resource and SRS-ResourceSet related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// pucchCmd represents the "nrrg pucch" command
var pucchCmd = &cobra.Command{
	Use:   "pucch",
	Short: "",
	Long:  `CMD "nrrg pucch" can be used to get/set PUCCH-Config related network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// advancedCmd represents the "nrrg advanced" command
var advancedCmd = &cobra.Command{
	Use:   "advanced",
	Short: "",
	Long:  `CMD "nrrg advanced" can be used to get/set advanced network configurations.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadNrrgFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()
		laPrint(cmd, args)
		viper.WriteConfig()
	},
}

// TODO: add more subcmd here!!!

// simCmd represents the "nrrg sim" command
var simCmd = &cobra.Command{
	Use:   "sim",
	Short: "",
	Long:  `CMD "nrrg sim" can be used to perform static NR-Uu simulation.`,
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: func(cmd *cobra.Command, args []string) {
		viper.WatchConfig()
		fmt.Println("nrrg sim called")
		viper.WriteConfig()

		/*
			// Examples of 'expression evaluation'
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
		*/
	},
}

func init() {
	nrrgCmd.AddCommand(gridSettingCmd)
	nrrgCmd.AddCommand(tddUlDlCmd)
	nrrgCmd.AddCommand(searchSpaceCmd)
	nrrgCmd.AddCommand(dlDciCmd)
	nrrgCmd.AddCommand(ulDciCmd)
	nrrgCmd.AddCommand(bwpCmd)
	nrrgCmd.AddCommand(rachCmd)
	nrrgCmd.AddCommand(dmrsCommonCmd)
	nrrgCmd.AddCommand(pdschCmd)
	nrrgCmd.AddCommand(puschCmd)
	nrrgCmd.AddCommand(pucchCmd)
	nrrgCmd.AddCommand(csiCmd)
	nrrgCmd.AddCommand(srsCmd)
	nrrgCmd.AddCommand(advancedCmd)

	if cmdFlags&CMD_FLAG_NRRG != 0 {
		rootCmd.AddCommand(nrrgCmd)
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nrrgCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	initGridSettingCmd()
	initTddUlDlCmd()
	initSearchSpaceCmd()
	initDlDciCmd()
	initUlDciCmd()
	initBwpCmd()
	initRachCmd()
	initDmrsCommonCmd()
	initPdschCmd()
	initPuschCmd()
	initCsiCmd()
	initSrsCmd()
	initPucchCmd()
	initAdvancedCmd()
}

func initGridSettingCmd() {
	// freqBand part
	gridSettingCmd.Flags().StringVar(&flags.gridsetting.band, "band", "n28", "Operating band")
	gridSettingCmd.Flags().StringVar(&flags.gridsetting._duplexMode, "_duplexMode", "FDD", "Duplex mode")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._maxDlFreq, "_maxDlFreq", 803, "Maximum DL frequency(MHz)")
	gridSettingCmd.Flags().StringVar(&flags.gridsetting._freqRange, "_freqRange", "FR1", "Frequency range(FR1/FR2)")
	gridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting.band", gridSettingCmd.Flags().Lookup("band"))
	viper.BindPFlag("nrrg.gridsetting._duplexMode", gridSettingCmd.Flags().Lookup("_duplexMode"))
	viper.BindPFlag("nrrg.gridsetting._maxDlFreq", gridSettingCmd.Flags().Lookup("_maxDlFreq"))
	viper.BindPFlag("nrrg.gridsetting._freqRange", gridSettingCmd.Flags().Lookup("_freqRange"))
	gridSettingCmd.Flags().MarkHidden("_duplexMode")
	gridSettingCmd.Flags().MarkHidden("_maxDlFreq")
	gridSettingCmd.Flags().MarkHidden("_freqRange")

	// SCS
	gridSettingCmd.Flags().StringVar(&flags.gridsetting.scs, "scs", "15KHz", "Subcarrier spacing for SSB/RMSI/Carrier/BWP etc.")
	gridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting.scs", gridSettingCmd.Flags().Lookup("scs"))

	// ssbGrid part and ssbBurst part
	gridSettingCmd.Flags().StringVar(&flags.gridsetting._ssbScs, "_ssbScs", "15KHz", "SSB subcarrier spacing")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting.gscn, "gscn", 1931, "SSB GSCN")
	gridSettingCmd.Flags().StringVar(&flags.gridsetting._ssbPattern, "_ssbPattern", "Case A", "SSB pattern")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._kSsb, "_kSsb", 2, "k_SSB[0..23]")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._nCrbSsb, "_nCrbSsb", 69, "n_CRB_SSB")
	gridSettingCmd.Flags().StringVar(&flags.gridsetting.ssbPeriod, "ssbPeriod", "20ms", "ssb-PeriodicityServingCell[5ms,10ms,20ms,40ms,80ms,160ms]")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._maxLBar, "_maxLBar", 4, "L_max_bar as specified in 38.213")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._maxL, "_maxL", 4, "L_max as specified in 38.213")
	gridSettingCmd.Flags().IntSliceVar(&flags.gridsetting.candSsbIndex, "candSsbIndex", []int{0, 1, 2, 3}, "List of candidate SSB index")
	gridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting._ssbScs", gridSettingCmd.Flags().Lookup("_ssbScs"))
	viper.BindPFlag("nrrg.gridsetting.gscn", gridSettingCmd.Flags().Lookup("gscn"))
	viper.BindPFlag("nrrg.gridsetting._ssbPattern", gridSettingCmd.Flags().Lookup("_ssbPattern"))
	viper.BindPFlag("nrrg.gridsetting._kSsb", gridSettingCmd.Flags().Lookup("_kSsb"))
	viper.BindPFlag("nrrg.gridsetting._nCrbSsb", gridSettingCmd.Flags().Lookup("_nCrbSsb"))
	viper.BindPFlag("nrrg.gridsetting.ssbPeriod", gridSettingCmd.Flags().Lookup("ssbPeriod"))
	viper.BindPFlag("nrrg.gridsetting._maxLBar", gridSettingCmd.Flags().Lookup("_maxLBar"))
	viper.BindPFlag("nrrg.gridsetting._maxL", gridSettingCmd.Flags().Lookup("_maxL"))
	viper.BindPFlag("nrrg.gridsetting.candSsbIndex", gridSettingCmd.Flags().Lookup("candSsbIndex"))
	gridSettingCmd.Flags().MarkHidden("_ssbScs")
	gridSettingCmd.Flags().MarkHidden("_ssbPattern")
	gridSettingCmd.Flags().MarkHidden("_kSsb")
	gridSettingCmd.Flags().MarkHidden("_nCrbSsb")
	gridSettingCmd.Flags().MarkHidden("_maxLBar")
	gridSettingCmd.Flags().MarkHidden("_maxL")

	// carrierGrid part and MIB-subCarrierSpacingCommon
	gridSettingCmd.Flags().StringVar(&flags.gridsetting._carrierScs, "_carrierScs", "15KHz", "subcarrierSpacing of SCS-SpecificCarrier")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting.dlArfcn, "dlArfcn", 154600, "DL ARFCN")
	gridSettingCmd.Flags().StringVar(&flags.gridsetting.bw, "bw", "30MHz", "Transmission bandwidth(MHz)")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._carrierNumRbs, "_carrierNumRbs", 160, "carrierBandwidth(N_RB) of SCS-SpecificCarrier")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._offsetToCarrier, "_offsetToCarrier", 0, "_offsetToCarrier of SCS-SpecificCarrier")
	gridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting._carrierScs", gridSettingCmd.Flags().Lookup("_carrierScs"))
	viper.BindPFlag("nrrg.gridsetting.dlArfcn", gridSettingCmd.Flags().Lookup("dlArfcn"))
	viper.BindPFlag("nrrg.gridsetting.bw", gridSettingCmd.Flags().Lookup("bw"))
	viper.BindPFlag("nrrg.gridsetting._carrierNumRbs", gridSettingCmd.Flags().Lookup("_carrierNumRbs"))
	viper.BindPFlag("nrrg.gridsetting._offsetToCarrier", gridSettingCmd.Flags().Lookup("_offsetToCarrier"))
	gridSettingCmd.Flags().MarkHidden("_carrierScs")
	gridSettingCmd.Flags().MarkHidden("_carrierNumRbs")
	gridSettingCmd.Flags().MarkHidden("_offsetToCarrier")

	// PCI
	gridSettingCmd.Flags().IntVar(&flags.gridsetting.pci, "pci", 0, "Physical cell identity[0..1007]")
	gridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting.pci", gridSettingCmd.Flags().Lookup("pci"))

	// MIB part
	gridSettingCmd.Flags().StringVar(&flags.gridsetting._mibCommonScs, "_mibCommonScs", "15KHz", "subCarrierSpacingCommon of MIB")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting.rmsiCoreset0, "rmsiCoreset0", 7, "coresetZero of PDCCH-ConfigSIB1[0..15]")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._coreset0MultiplexingPat, "_coreset0MultiplexingPat", 1, "Multiplexing pattern of CORESET0")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._coreset0NumRbs, "_coreset0NumRbs", 48, "Number of PRBs of CORESET0")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._coreset0NumSymbs, "_coreset0NumSymbs", 1, "Number of OFDM symbols of CORESET0")
	gridSettingCmd.Flags().IntSliceVar(&flags.gridsetting._coreset0OffsetList, "_coreset0OffsetList", []int{16}, "List of offset of CORESET0")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._coreset0Offset, "_coreset0Offset", 16, "Offset of CORESET0")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting.rmsiCss0, "rmsiCss0", 4, "searchSpaceZero of PDCCH-ConfigSIB1[0..15]")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._css0AggLevel, "_css0AggLevel", 4, "CCE aggregation level of CSS0[4,8,16]")
	gridSettingCmd.Flags().StringVar(&flags.gridsetting._css0NumCandidates, "_css0NumCandidates", "n4", "Number of PDCCH candidates of CSS0[n1,n2,n4]")
	gridSettingCmd.Flags().StringVar(&flags.gridsetting.dmrsTypeAPos, "dmrsTypeAPos", "pos2", "dmrs-TypeA-Position[pos2,pos3]")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._sfn, "_sfn", 0, "System frame number(SFN)[0..1023]")
	gridSettingCmd.Flags().IntVar(&flags.gridsetting._hrf, "_hrf", 0, "Half frame bit[0,1]")
	gridSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.gridsetting._mibCommonScs", gridSettingCmd.Flags().Lookup("_mibCommonScs"))
	viper.BindPFlag("nrrg.gridsetting.rmsiCoreset0", gridSettingCmd.Flags().Lookup("rmsiCoreset0"))
	viper.BindPFlag("nrrg.gridsetting._coreset0MultiplexingPat", gridSettingCmd.Flags().Lookup("_coreset0MultiplexingPat"))
	viper.BindPFlag("nrrg.gridsetting._coreset0NumRbs", gridSettingCmd.Flags().Lookup("_coreset0NumRbs"))
	viper.BindPFlag("nrrg.gridsetting._coreset0NumSymbs", gridSettingCmd.Flags().Lookup("_coreset0NumSymbs"))
	viper.BindPFlag("nrrg.gridsetting._coreset0OffsetList", gridSettingCmd.Flags().Lookup("_coreset0OffsetList"))
	viper.BindPFlag("nrrg.gridsetting._coreset0Offset", gridSettingCmd.Flags().Lookup("_coreset0Offset"))
	viper.BindPFlag("nrrg.gridsetting.rmsiCss0", gridSettingCmd.Flags().Lookup("rmsiCss0"))
	viper.BindPFlag("nrrg.gridsetting._css0AggLevel", gridSettingCmd.Flags().Lookup("_css0AggLevel"))
	viper.BindPFlag("nrrg.gridsetting._css0NumCandidates", gridSettingCmd.Flags().Lookup("_css0NumCandidates"))
	viper.BindPFlag("nrrg.gridsetting.dmrsTypeAPos", gridSettingCmd.Flags().Lookup("dmrsTypeAPos"))
	viper.BindPFlag("nrrg.gridsetting._sfn", gridSettingCmd.Flags().Lookup("_sfn"))
	viper.BindPFlag("nrrg.gridsetting._hrf", gridSettingCmd.Flags().Lookup("_hrf"))
	gridSettingCmd.Flags().MarkHidden("_mibCommonScs")
	gridSettingCmd.Flags().MarkHidden("_coreset0MultiplexingPat")
	gridSettingCmd.Flags().MarkHidden("_coreset0NumRbs")
	gridSettingCmd.Flags().MarkHidden("_coreset0NumSymbs")
	gridSettingCmd.Flags().MarkHidden("_coreset0OffsetList")
	gridSettingCmd.Flags().MarkHidden("_coreset0Offset")
	gridSettingCmd.Flags().MarkHidden("_css0AggLevel")
	gridSettingCmd.Flags().MarkHidden("_css0NumCandidates")
	gridSettingCmd.Flags().MarkHidden("_sfn")
	gridSettingCmd.Flags().MarkHidden("_hrf")
}

func initTddUlDlCmd() {
	tddUlDlCmd.Flags().StringVar(&flags.tdduldl._refScs, "_refScs", "30KHz", "referenceSubcarrierSpacing of TDD-UL-DL-ConfigCommon")
	tddUlDlCmd.Flags().StringSliceVar(&flags.tdduldl.patPeriod, "patPeriod", []string{"5ms"}, "dl-UL-TransmissionPeriodicity of TDD-UL-DL-ConfigCommon[0.5ms,0.625ms,1ms,1.25ms,2ms,2.5ms,3ms,4ms,5ms,10ms]")
	tddUlDlCmd.Flags().IntSliceVar(&flags.tdduldl.patNumDlSlots, "patNumDlSlots", []int{7}, "nrofDownlinkSlot of TDD-UL-DL-ConfigCommon[0..80]")
	tddUlDlCmd.Flags().IntSliceVar(&flags.tdduldl.patNumDlSymbs, "patNumDlSymbs", []int{6}, "nrofDownlinkSymbols of TDD-UL-DL-ConfigCommon[0..13]")
	tddUlDlCmd.Flags().IntSliceVar(&flags.tdduldl.patNumUlSymbs, "patNumUlSymbs", []int{4}, "nrofUplinkSymbols of TDD-UL-DL-ConfigCommon[0..13]")
	tddUlDlCmd.Flags().IntSliceVar(&flags.tdduldl.patNumUlSlots, "patNumUlSlots", []int{2}, "nrofUplinkSlots of TDD-UL-DL-ConfigCommon[0..80]")
	tddUlDlCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.tdduldl._refScs", tddUlDlCmd.Flags().Lookup("_refScs"))
	viper.BindPFlag("nrrg.tdduldl.patPeriod", tddUlDlCmd.Flags().Lookup("patPeriod"))
	viper.BindPFlag("nrrg.tdduldl.patNumDlSlots", tddUlDlCmd.Flags().Lookup("patNumDlSlots"))
	viper.BindPFlag("nrrg.tdduldl.patNumDlSymbs", tddUlDlCmd.Flags().Lookup("patNumDlSymbs"))
	viper.BindPFlag("nrrg.tdduldl.patNumUlSymbs", tddUlDlCmd.Flags().Lookup("patNumUlSymbs"))
	viper.BindPFlag("nrrg.tdduldl.patNumUlSlots", tddUlDlCmd.Flags().Lookup("patNumUlSlots"))
	tddUlDlCmd.Flags().MarkHidden("_refScs")
}

func initSearchSpaceCmd() {
	searchSpaceCmd.Flags().StringVar(&flags.searchspace._coreset1FdRes, "_coreset1FdRes", "111111111111111111111111111111111111111111111", "frequencyDomainResources of ControlResourceSet")
	searchSpaceCmd.Flags().IntVar(&flags.searchspace.coreset1StartCrb, "coreset1StartCrb", 0, "starting CRB of CORESET1")
	searchSpaceCmd.Flags().IntVar(&flags.searchspace.coreset1NumRbs, "coreset1NumRbs", 120, "number of RBs of CORESET1")
	searchSpaceCmd.Flags().IntVar(&flags.searchspace._coreset1Duration, "_coreset1Duration", 1, "duration of ControlResourceSet[1..3]")
	searchSpaceCmd.Flags().StringVar(&flags.searchspace.coreset1CceRegMappingType, "coreset1CceRegMappingType", "interleaved", "cce-REG-MappingType of ControlResourceSet[1..3]")
	searchSpaceCmd.Flags().StringVar(&flags.searchspace.coreset1RegBundleSize, "coreset1RegBundleSize", "n2", "reg-BundleSize of ControlResourceSet[n2,n3,n6]")
	searchSpaceCmd.Flags().StringVar(&flags.searchspace.coreset1InterleaverSize, "coreset1InterleaverSize", "n3", "interleaverSize of ControlResourceSet[n2,n3,n6]")
	searchSpaceCmd.Flags().IntVar(&flags.searchspace._coreset1ShiftIndex, "_coreset1ShiftIndex", 0, "shiftIndex of ControlResourceSet[0..274]")
	searchSpaceCmd.Flags().IntSliceVar(&flags.searchspace._ssId, "_ssId", []int{1, 2, 3, 4, 5}, "search space id")
	searchSpaceCmd.Flags().StringSliceVar(&flags.searchspace._ssType, "_ssType", []string{"type0a", "type1", "type2", "type3", "uss"}, "search space type")
	searchSpaceCmd.Flags().IntSliceVar(&flags.searchspace._ssCoresetId, "_ssCoresetId", []int{0, 0, 0, 1, 1}, "associated CORESET id")
	searchSpaceCmd.Flags().IntSliceVar(&flags.searchspace._ssDuration, "_ssDuration", []int{1, 1, 1, 1, 1}, "search space duration in slot(s)")
	searchSpaceCmd.Flags().StringSliceVar(&flags.searchspace._ssMonitoringSymbolWithinSlot, "_ssMonitoringSymbolWithinSlot", []string{"100", "110", "100", "110", "110"}, "the MonitoringSymbolWithinSlot")
	searchSpaceCmd.Flags().StringSliceVar(&flags.searchspace.ssAggregationLevel, "ssAggregationLevel", []string{"AL4", "AL4", "AL4", "AL4", "AL4"}, "aggregation level")
	searchSpaceCmd.Flags().StringSliceVar(&flags.searchspace.ssNumOfPdcchCandidates, "ssNumOfPdcchCandidates", []string{"n2", "n2", "n2", "n5", "n5"}, "number of PDCCH candidates")
	searchSpaceCmd.Flags().StringSliceVar(&flags.searchspace._ssPeriodicity, "_ssPeriodicity", []string{"sl1", "sl1", "sl1", "sl1", "sl1"}, "search space periodicity")
	searchSpaceCmd.Flags().IntSliceVar(&flags.searchspace._ssSlotOffset, "_ssSlotOffset", []int{0, 0, 0, 0, 0}, "search space slot offset")
	searchSpaceCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.searchspace._coreset1FdRes", searchSpaceCmd.Flags().Lookup("_coreset1FdRes"))
	viper.BindPFlag("nrrg.searchspace.coreset1StartCrb", searchSpaceCmd.Flags().Lookup("coreset1StartCrb"))
	viper.BindPFlag("nrrg.searchspace.coreset1NumRbs", searchSpaceCmd.Flags().Lookup("coreset1NumRbs"))
	viper.BindPFlag("nrrg.searchspace._coreset1Duration", searchSpaceCmd.Flags().Lookup("_coreset1Duration"))
	viper.BindPFlag("nrrg.searchspace.coreset1CceRegMappingType", searchSpaceCmd.Flags().Lookup("coreset1CceRegMappingType"))
	viper.BindPFlag("nrrg.searchspace.coreset1RegBundleSize", searchSpaceCmd.Flags().Lookup("coreset1RegBundleSize"))
	viper.BindPFlag("nrrg.searchspace.coreset1InterleaverSize", searchSpaceCmd.Flags().Lookup("coreset1InterleaverSize"))
	viper.BindPFlag("nrrg.searchspace._coreset1ShiftIndex", searchSpaceCmd.Flags().Lookup("_coreset1ShiftIndex"))
	viper.BindPFlag("nrrg.searchspace._ssId", searchSpaceCmd.Flags().Lookup("_ssId"))
	viper.BindPFlag("nrrg.searchspace._ssType", searchSpaceCmd.Flags().Lookup("_ssType"))
	viper.BindPFlag("nrrg.searchspace._ssCoresetId", searchSpaceCmd.Flags().Lookup("_ssCoresetId"))
	viper.BindPFlag("nrrg.searchspace._ssDuration", searchSpaceCmd.Flags().Lookup("_ssDuration"))
	viper.BindPFlag("nrrg.searchspace._ssMonitoringSymbolWithinSlot", searchSpaceCmd.Flags().Lookup("_ssMonitoringSymbolWithinSlot"))
	viper.BindPFlag("nrrg.searchspace.ssAggregationLevel", searchSpaceCmd.Flags().Lookup("ssAggregationLevel"))
	viper.BindPFlag("nrrg.searchspace.ssNumOfPdcchCandidates", searchSpaceCmd.Flags().Lookup("ssNumOfPdcchCandidates"))
	viper.BindPFlag("nrrg.searchspace._ssPeriodicity", searchSpaceCmd.Flags().Lookup("_ssPeriodicity"))
	viper.BindPFlag("nrrg.searchspace._ssSlotOffset", searchSpaceCmd.Flags().Lookup("_ssSlotOffset"))
	searchSpaceCmd.Flags().MarkHidden("_coreset1FdRes")
	searchSpaceCmd.Flags().MarkHidden("_coreset1Duration")
	searchSpaceCmd.Flags().MarkHidden("_coreset1ShiftIndex")
	searchSpaceCmd.Flags().MarkHidden("_ssId")
	searchSpaceCmd.Flags().MarkHidden("_ssType")
	searchSpaceCmd.Flags().MarkHidden("_ssCoresetId")
	searchSpaceCmd.Flags().MarkHidden("_ssDuration")
	searchSpaceCmd.Flags().MarkHidden("_ssMonitoringSymbolWithinSlot")
	searchSpaceCmd.Flags().MarkHidden("_ssPeriodicity")
	searchSpaceCmd.Flags().MarkHidden("_ssSlotOffset")
}

func initDlDciCmd() {
	dlDciCmd.Flags().StringSliceVar(&flags.dldci._tag, "_tag", []string{"DCI_10_SIB1", "DCI_10_MSG2", "DCI_10_MSG4", "DCI_11_PDSCH"}, "DCI tag")
	dlDciCmd.Flags().StringSliceVar(&flags.dldci._rnti, "_rnti", []string{"SI-RNTI", "RA-RNTI", "TC-RNTI", "C-RNTI"}, "RNTI for DCI 1_0/1_1")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci._muPdcch, "_muPdcch", []int{1, 1, 1, 1}, "Subcarrier spacing of PDCCH")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci._muPdsch, "_muPdsch", []int{1, 1, 1, 1}, "Subcarrier spacing of PDSCH")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci._indicatedBwp, "_indicatedBwp", []int{0, 0, 0, 1}, "Bandwidth part indicator field of DCI 1_1")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci.tdra, "tdra", []int{10, 10, 10, 10}, "Time-domain-resource-assignment field of DCI 1_0")
	dlDciCmd.Flags().StringSliceVar(&flags.dldci._tdMappingType, "_tdMappingType", []string{"typeA", "typeA", "typeA", "typeA"}, "Mapping type for PDSCH time-domain allocation")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci._tdK0, "_tdK0", []int{0, 0, 0, 0}, "Slot offset K0 for PDSCH time-domain allocation")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci._tdSliv, "_tdSliv", []int{26, 26, 26, 26}, "SLIV for PDSCH time-domain allocation")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci._tdStartSymb, "_tdStartSymb", []int{12, 12, 12, 12}, "Starting symbol S for PDSCH time-domain allocation")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci._tdNumSymbs, "_tdNumSymbs", []int{2, 2, 2, 2}, "Number of OFDM symbols L for PDSCH time-domain allocation")
	dlDciCmd.Flags().StringSliceVar(&flags.dldci._fdRaType, "_fdRaType", []string{"raType1", "raType1", "raType1", "raType1"}, "resourceAllocation for PDSCH frequency-domain allocation")
	dlDciCmd.Flags().IntVar(&flags.dldci._fdBitsRaType0, "_fdBitsRaType0", 11, "Bitwidth of PDSCH frequency-domain allocation for RA Type 1")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci._fdBitsRaType1, "_fdBitsRaType1", []int{11, 11, 11, 11}, "Bitwidth of PDSCH frequency-domain allocation for RA Type 1")
	dlDciCmd.Flags().StringSliceVar(&flags.dldci._fdRa, "_fdRa", []string{"00001011111", "00001011111", "00001011111", ""}, "Frequency-domain-resource-assignment field of DCI 1_0/1_1")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci.fdStartRb, "fdStartRb", []int{0, 0, 0, 0}, "RB_start of RIV for PDSCH frequency-domain allocation")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci.fdNumRbs, "fdNumRbs", []int{48, 48, 48, 160}, "L_RBs of RIV for PDSCH frequency-domain allocation")
	dlDciCmd.Flags().StringSliceVar(&flags.dldci.fdVrbPrbMappingType, "fdVrbPrbMappingType", []string{"interleaved", "interleaved", "interleaved", "interleaved"}, "VRB-to-PRB-mapping field of DCI 1_0/1_1")
	dlDciCmd.Flags().StringSliceVar(&flags.dldci.fdBundleSize, "fdBundleSize", []string{"n2", "n2", "n2", "n2"}, "L(vrb-ToPRB-Interleaver) for PDSCH frequency-domain allocation")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci.mcsCw0, "mcsCw0", []int{2, 2, 2, 27}, "Modulation-and-coding-scheme field of DCI 1_0/1_1 for the 1st TB")
	dlDciCmd.Flags().IntSliceVar(&flags.dldci._tbsCw0, "_tbsCw0", []int{408, 408, 408, 408}, "Transport block size(bits) for PDSCH CW0")
	dlDciCmd.Flags().IntVar(&flags.dldci.mcsCw1, "mcsCw1", -1, "Modulation-and-coding-scheme field of DCI 1_1 for the 2nd TB (-1 to disable the 2nd TB)")
	dlDciCmd.Flags().IntVar(&flags.dldci._tbsCw1, "_tbsCw1", -1, "Transport block size(bits) for PDSCH CW1")
	dlDciCmd.Flags().Float64Var(&flags.dldci.tbScalingFactor, "tbScalingFactor", 1, "TB scaling factor[0,0.5,0.25]")
	dlDciCmd.Flags().IntVar(&flags.dldci.deltaPri, "deltaPri", 1, "PUCCH-resource-indicator field of DCI 1_0/1_1[0..7]")
	dlDciCmd.Flags().IntVar(&flags.dldci.tdK1, "tdK1", 2, "PDSCH-to-HARQ_feedback-timing-indicator(K1) field of DCI 1_0/1_1[0..7]")
	dlDciCmd.Flags().IntVar(&flags.dldci.antennaPorts, "antennaPorts", 10, "Antenna port(s) field of DCI 1_1")
	dlDciCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dldci._tag", dlDciCmd.Flags().Lookup("_tag"))
	viper.BindPFlag("nrrg.dldci._rnti", dlDciCmd.Flags().Lookup("_rnti"))
	viper.BindPFlag("nrrg.dldci._muPdcch", dlDciCmd.Flags().Lookup("_muPdcch"))
	viper.BindPFlag("nrrg.dldci._muPdsch", dlDciCmd.Flags().Lookup("_muPdsch"))
	viper.BindPFlag("nrrg.dldci._indicatedBwp", dlDciCmd.Flags().Lookup("_indicatedBwp"))
	viper.BindPFlag("nrrg.dldci.tdra", dlDciCmd.Flags().Lookup("tdra"))
	viper.BindPFlag("nrrg.dldci._tdMappingType", dlDciCmd.Flags().Lookup("_tdMappingType"))
	viper.BindPFlag("nrrg.dldci._tdK0", dlDciCmd.Flags().Lookup("_tdK0"))
	viper.BindPFlag("nrrg.dldci._tdSliv", dlDciCmd.Flags().Lookup("_tdSliv"))
	viper.BindPFlag("nrrg.dldci._tdStartSymb", dlDciCmd.Flags().Lookup("_tdStartSymb"))
	viper.BindPFlag("nrrg.dldci._tdNumSymbs", dlDciCmd.Flags().Lookup("_tdNumSymbs"))
	viper.BindPFlag("nrrg.dldci._fdRaType", dlDciCmd.Flags().Lookup("_fdRaType"))
	viper.BindPFlag("nrrg.dldci._fdBitsRaType0", dlDciCmd.Flags().Lookup("_fdBitsRaType0"))
	viper.BindPFlag("nrrg.dldci._fdBitsRaType1", dlDciCmd.Flags().Lookup("_fdBitsRaType1"))
	viper.BindPFlag("nrrg.dldci._fdRa", dlDciCmd.Flags().Lookup("_fdRa"))
	viper.BindPFlag("nrrg.dldci.fdStartRb", dlDciCmd.Flags().Lookup("fdStartRb"))
	viper.BindPFlag("nrrg.dldci.fdNumRbs", dlDciCmd.Flags().Lookup("fdNumRbs"))
	viper.BindPFlag("nrrg.dldci.fdVrbPrbMappingType", dlDciCmd.Flags().Lookup("fdVrbPrbMappingType"))
	viper.BindPFlag("nrrg.dldci.fdBundleSize", dlDciCmd.Flags().Lookup("fdBundleSize"))
	viper.BindPFlag("nrrg.dldci.mcsCw0", dlDciCmd.Flags().Lookup("mcsCw0"))
	viper.BindPFlag("nrrg.dldci._tbsCw0", dlDciCmd.Flags().Lookup("_tbsCw0"))
	viper.BindPFlag("nrrg.dldci.mcsCw1", dlDciCmd.Flags().Lookup("mcsCw1"))
	viper.BindPFlag("nrrg.dldci._tbsCw1", dlDciCmd.Flags().Lookup("_tbsCw1"))
	viper.BindPFlag("nrrg.dldci.tbScalingFactor", dlDciCmd.Flags().Lookup("tbScalingFactor"))
	viper.BindPFlag("nrrg.dldci.deltaPri", dlDciCmd.Flags().Lookup("deltaPri"))
	viper.BindPFlag("nrrg.dldci.tdK1", dlDciCmd.Flags().Lookup("tdK1"))
	viper.BindPFlag("nrrg.dldci.antennaPorts", dlDciCmd.Flags().Lookup("antennaPorts"))
	dlDciCmd.Flags().MarkHidden("_tag")
	dlDciCmd.Flags().MarkHidden("_rnti")
	dlDciCmd.Flags().MarkHidden("_muPdcch")
	dlDciCmd.Flags().MarkHidden("_muPdsch")
	dlDciCmd.Flags().MarkHidden("_indicatedBwp")
	dlDciCmd.Flags().MarkHidden("_tdMappingType")
	dlDciCmd.Flags().MarkHidden("_tdK0")
	dlDciCmd.Flags().MarkHidden("_tdSliv")
	dlDciCmd.Flags().MarkHidden("_tdStartSymb")
	dlDciCmd.Flags().MarkHidden("_tdNumSymbs")
	dlDciCmd.Flags().MarkHidden("_fdRaType")
	dlDciCmd.Flags().MarkHidden("_fdBitsRaType0")
	dlDciCmd.Flags().MarkHidden("_fdBitsRaType1")
	dlDciCmd.Flags().MarkHidden("_fdRa")
	dlDciCmd.Flags().MarkHidden("_tbsCw0")
	dlDciCmd.Flags().MarkHidden("_tbsCw1")
	dlDciCmd.Flags().MarkHidden("tbScalingFactor")
}

func initUlDciCmd() {
	ulDciCmd.Flags().StringSliceVar(&flags.uldci._tag, "_tag", []string{"RAR_MSG3", "DCI_01_PUSCH"}, "RNTI for DCI 0_1")
	ulDciCmd.Flags().StringSliceVar(&flags.uldci._rnti, "_rnti", []string{"RA-RNTI", "C-RNTI"}, "RNTI for DCI 0_1")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci._muPdcch, "_muPdcch", []int{1, 1}, "Subcarrier spacing of PDCCH[0..3]")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci._muPusch, "_muPusch", []int{1, 1}, "Subcarrier spacing of PUSCH[0..3]")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci._indicatedBwp, "_indicatedBwp", []int{0, 1}, "Bandwidth-part-indicator field of DCI 0_1[0..1]")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci.tdra, "tdra", []int{7, 7}, "Time-domain-resource-assignment field of DCI 0_1[0..15]")
	ulDciCmd.Flags().StringSliceVar(&flags.uldci._tdMappingType, "_tdMappingType", []string{"typeA", "typeA"}, "Mapping type for PUSCH time-domain allocation[typeA,typeB]")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci._tdK2, "_tdK2", []int{2, 2}, "Slot offset K2 for PUSCH time-domain allocation[0..32]")
	ulDciCmd.Flags().IntVar(&flags.uldci._tdDelta, "_tdDelta", 2, "The delta for Msg3 PUSCH scheduled by RAR UL grant(38.214 8.3)")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci._tdSliv, "_tdSliv", []int{27, 27}, "SLIV for PUSCH time-domain allocation")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci._tdStartSymb, "_tdStartSymb", []int{0, 0}, "Starting symbol S for PUSCH time-domain allocation")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci._tdNumSymbs, "_tdNumSymbs", []int{14, 14}, "Number of OFDM symbols L for PUSCH time-domain allocation")
	ulDciCmd.Flags().StringSliceVar(&flags.uldci._fdRaType, "_fdRaType", []string{"raType1", "raType1"}, "resourceAllocation for PUSCH frequency-domain allocation[raType0,raType1]")
	ulDciCmd.Flags().StringSliceVar(&flags.uldci.fdFreqHop, "fdFreqHop", []string{"intra-slot", "disabled"}, "Frequency-hopping-flag field for DCI 0_1[disabled,intraSlot,interSlot]")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci._fdFreqHopOffset, "_fdFreqHopOffset", []int{0, 0}, "frequencyHoppingOffsetLists of PUSCH-Config[0..274]")
	ulDciCmd.Flags().IntVar(&flags.uldci._fdBitsRaType0, "_fdBitsRaType0", 11, "Bitwidth of PUSCH frequency-domain allocation for RA Type 1")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci._fdBitsRaType1, "_fdBitsRaType1", []int{14, 11}, "Bitwidth of PUSCH frequency-domain allocation for RA Type 1")
	ulDciCmd.Flags().StringSliceVar(&flags.uldci._fdRa, "_fdRa", []string{"", "0000001000100001"}, "Frequency-domain-resource-assignment field of DCI 0_1")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci.fdStartRb, "fdStartRb", []int{0, 0}, "RB_start of RIV for PUSCH frequency-domain allocation")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci.fdNumRbs, "fdNumRbs", []int{3, 160}, "L_RBs of RIV for PUSCH frequency-domain allocation")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci.mcsCw0, "mcsCw0", []int{2, 28}, "Modulation-and-coding-scheme-cw0 field of DCI 0_1[0..28]")
	ulDciCmd.Flags().IntSliceVar(&flags.uldci._tbs, "_tbs", []int{3624, 475584}, "Transport block size(bits) for PUSCH")
	ulDciCmd.Flags().IntVar(&flags.uldci.precodingInfoNumLayers, "precodingInfoNumLayers", 2, "Precoding-information-and-number-of-layers field of DCI 0_1[0..63]")
	ulDciCmd.Flags().IntVar(&flags.uldci.srsResIndicator, "srsResIndicator", 0, "SRS-resource-indicator field of DCI 0_1")
	ulDciCmd.Flags().IntVar(&flags.uldci.antennaPorts, "antennaPorts", 0, "Antenna_port(s) field of DCI 0_1[0..7]")
	ulDciCmd.Flags().IntVar(&flags.uldci.ptrsDmrsAssociation, "ptrsDmrsAssociation", 0, "PTRS-DMRS-association field of DCI 0_1[0..3]")
	ulDciCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.uldci._tag", ulDciCmd.Flags().Lookup("_tag"))
	viper.BindPFlag("nrrg.uldci._rnti", ulDciCmd.Flags().Lookup("_rnti"))
	viper.BindPFlag("nrrg.uldci._muPdcch", ulDciCmd.Flags().Lookup("_muPdcch"))
	viper.BindPFlag("nrrg.uldci._muPusch", ulDciCmd.Flags().Lookup("_muPusch"))
	viper.BindPFlag("nrrg.uldci._indicatedBwp", ulDciCmd.Flags().Lookup("_indicatedBwp"))
	viper.BindPFlag("nrrg.uldci.tdra", ulDciCmd.Flags().Lookup("tdra"))
	viper.BindPFlag("nrrg.uldci._tdMappingType", ulDciCmd.Flags().Lookup("_tdMappingType"))
	viper.BindPFlag("nrrg.uldci._tdK2", ulDciCmd.Flags().Lookup("_tdK2"))
	viper.BindPFlag("nrrg.uldci._tdDelta", ulDciCmd.Flags().Lookup("_tdDelta"))
	viper.BindPFlag("nrrg.uldci._tdSliv", ulDciCmd.Flags().Lookup("_tdSliv"))
	viper.BindPFlag("nrrg.uldci._tdStartSymb", ulDciCmd.Flags().Lookup("_tdStartSymb"))
	viper.BindPFlag("nrrg.uldci._tdNumSymbs", ulDciCmd.Flags().Lookup("_tdNumSymbs"))
	viper.BindPFlag("nrrg.uldci._fdRaType", ulDciCmd.Flags().Lookup("_fdRaType"))
	viper.BindPFlag("nrrg.uldci.fdFreqHop", ulDciCmd.Flags().Lookup("fdFreqHop"))
	viper.BindPFlag("nrrg.uldci._fdFreqHopOffset", ulDciCmd.Flags().Lookup("_fdFreqHopOffset"))
	viper.BindPFlag("nrrg.uldci._fdBitsRaType0", ulDciCmd.Flags().Lookup("_fdBitsRaType0"))
	viper.BindPFlag("nrrg.uldci._fdBitsRaType1", ulDciCmd.Flags().Lookup("_fdBitsRaType1"))
	viper.BindPFlag("nrrg.uldci._fdRa", ulDciCmd.Flags().Lookup("_fdRa"))
	viper.BindPFlag("nrrg.uldci.fdStartRb", ulDciCmd.Flags().Lookup("fdStartRb"))
	viper.BindPFlag("nrrg.uldci.fdNumRbs", ulDciCmd.Flags().Lookup("fdNumRbs"))
	viper.BindPFlag("nrrg.uldci.mcsCw0", ulDciCmd.Flags().Lookup("mcsCw0"))
	viper.BindPFlag("nrrg.uldci._tbs", ulDciCmd.Flags().Lookup("_tbs"))
	viper.BindPFlag("nrrg.uldci.precodingInfoNumLayers", ulDciCmd.Flags().Lookup("precodingInfoNumLayers"))
	viper.BindPFlag("nrrg.uldci.srsResIndicator", ulDciCmd.Flags().Lookup("srsResIndicator"))
	viper.BindPFlag("nrrg.uldci.antennaPorts", ulDciCmd.Flags().Lookup("antennaPorts"))
	viper.BindPFlag("nrrg.uldci.ptrsDmrsAssociation", ulDciCmd.Flags().Lookup("ptrsDmrsAssociation"))
	ulDciCmd.Flags().MarkHidden("_tag")
	ulDciCmd.Flags().MarkHidden("_rnti")
	ulDciCmd.Flags().MarkHidden("_muPdcch")
	ulDciCmd.Flags().MarkHidden("_muPusch")
	ulDciCmd.Flags().MarkHidden("_indicatedBwp")
	ulDciCmd.Flags().MarkHidden("_tdMappingType")
	ulDciCmd.Flags().MarkHidden("_tdK2")
	ulDciCmd.Flags().MarkHidden("_tdDelta")
	ulDciCmd.Flags().MarkHidden("_tdSliv")
	ulDciCmd.Flags().MarkHidden("_tdStartSymb")
	ulDciCmd.Flags().MarkHidden("_tdNumSymbs")
	ulDciCmd.Flags().MarkHidden("_fdRaType")
	ulDciCmd.Flags().MarkHidden("_fdFreqHopOffset")
	ulDciCmd.Flags().MarkHidden("_fdBitsRaType0")
	ulDciCmd.Flags().MarkHidden("_fdBitsRaType1")
	ulDciCmd.Flags().MarkHidden("_fdRa")
	ulDciCmd.Flags().MarkHidden("_tbs")
}

func initBwpCmd() {
	bwpCmd.Flags().StringSliceVar(&flags.bwp._bwpType, "_bwpType", []string{"iniDlBwp", "dedDlBwp", "iniUlBwp", "dedUlBwp"}, "BWP type")
	bwpCmd.Flags().IntSliceVar(&flags.bwp._bwpId, "_bwpId", []int{0, 1, 0, 1}, "bwp-Id of BWP-Uplink or BWP-Downlink")
	bwpCmd.Flags().StringSliceVar(&flags.bwp._bwpScs, "_bwpScs", []string{"30KHz", "30KHz", "30KHz", "30KHz"}, "subcarrierSpacing of BWP")
	bwpCmd.Flags().StringSliceVar(&flags.bwp._bwpCp, "_bwpCp", []string{"normal", "normal", "normal", "normal"}, "cyclicPrefix of BWP")
	bwpCmd.Flags().IntSliceVar(&flags.bwp._bwpLocAndBw, "_bwpLocAndBw", []int{12925, 1099, 1099, 1099}, "locationAndBandwidth of BWP")
	bwpCmd.Flags().IntSliceVar(&flags.bwp._bwpStartRb, "_bwpStartRb", []int{0, 0, 0, 0}, "RB_start of BWP")
	bwpCmd.Flags().IntSliceVar(&flags.bwp._bwpNumRbs, "_bwpNumRbs", []int{48, 273, 273, 273}, "L_RBs of BWP")
	bwpCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.bwp._bwpType", bwpCmd.Flags().Lookup("_bwpType"))
	viper.BindPFlag("nrrg.bwp._bwpId", bwpCmd.Flags().Lookup("_bwpId"))
	viper.BindPFlag("nrrg.bwp._bwpScs", bwpCmd.Flags().Lookup("_bwpScs"))
	viper.BindPFlag("nrrg.bwp._bwpCp", bwpCmd.Flags().Lookup("_bwpCp"))
	viper.BindPFlag("nrrg.bwp._bwpLocAndBw", bwpCmd.Flags().Lookup("_bwpLocAndBw"))
	viper.BindPFlag("nrrg.bwp._bwpStartRb", bwpCmd.Flags().Lookup("_bwpStartRb"))
	viper.BindPFlag("nrrg.bwp._bwpNumRbs", bwpCmd.Flags().Lookup("_bwpNumRbs"))
	bwpCmd.Flags().MarkHidden("_bwpType")
	bwpCmd.Flags().MarkHidden("_bwpId")
	bwpCmd.Flags().MarkHidden("_bwpScs")
	bwpCmd.Flags().MarkHidden("_bwpCp")
	bwpCmd.Flags().MarkHidden("_bwpLocAndBw")
	bwpCmd.Flags().MarkHidden("_bwpStartRb")
	bwpCmd.Flags().MarkHidden("_bwpNumRbs")
}

func initRachCmd() {
	rachCmd.Flags().IntVar(&flags.rach.prachConfId, "prachConfId", 148, "prach-ConfigurationIndex of RACH-ConfigGeneric[0..255]")
	rachCmd.Flags().StringVar(&flags.rach._raFormat, "_raFormat", "B4", "Preamble format")
	rachCmd.Flags().IntVar(&flags.rach._raX, "_raX", 2, "The x in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	rachCmd.Flags().IntSliceVar(&flags.rach._raY, "_raY", []int{1}, "The y in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	rachCmd.Flags().IntSliceVar(&flags.rach._raSubfNumFr1SlotNumFr2, "_raSubfNumFr1SlotNumFr2", []int{9}, "The Subframe-number in 3GPP TS 38.211 Table 6.3.3.2-2 and Table 6.3.3.2-3, or the Slot-number in Table 6.3.3.2-4")
	rachCmd.Flags().IntVar(&flags.rach._raStartingSymb, "_raStartingSymb", 0, "The Starting-symbol in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	rachCmd.Flags().IntVar(&flags.rach._raNumSlotsPerSubfFr1Per60KSlotFr2, "_raNumSlotsPerSubfFr1Per60KSlotFr2", 1, "The Number-of-PRACH-slots-within-a-subframe in 3GPP TS 38.211 Table 6.3.3.2-2 and Table 6.3.3.2-3, or the Number-of-PRACH-slots-within-a-60-kHz-slot in Table 6.3.3.2-4")
	rachCmd.Flags().IntVar(&flags.rach._raNumOccasionsPerSlot, "_raNumOccasionsPerSlot", 1, "The Number-of-time-domain-PRACH-occasions-within-a-PRACH-slot in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	rachCmd.Flags().IntVar(&flags.rach._raDuration, "_raDuration", 12, "The PRACH duration in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	rachCmd.Flags().StringVar(&flags.rach._msg1Scs, "_msg1Scs", "30KHz", "msg1-SubcarrierSpacing of RACH-ConfigCommon")
	rachCmd.Flags().IntVar(&flags.rach.msg1Fdm, "msg1Fdm", 1, "msg1-FDM of RACH-ConfigGeneric[1,2,4,8]")
	rachCmd.Flags().IntVar(&flags.rach.msg1FreqStart, "msg1FreqStart", 0, "msg1-FrequencyStart of RACH-ConfigGeneric[0..274]")
	rachCmd.Flags().IntVar(&flags.rach.totNumPreambs, "totNumPreambs", 64, "totalNumberOfRA-Preambles of RACH-ConfigCommon[1..64]")
	rachCmd.Flags().StringVar(&flags.rach.ssbPerRachOccasion, "ssbPerRachOccasion", "one", "ssb-perRACH-Occasion of RACH-ConfigGeneric[oneEighth,oneFourth,oneHalf,one,two,four,eight,sixteen]")
	rachCmd.Flags().IntVar(&flags.rach.cbPreambsPerSsb, "cbPreambsPerSsb", 64, "cb-PreamblesPerSSB of RACH-ConfigCommon[depends on ssbPerRachOccasion]")
	rachCmd.Flags().StringVar(&flags.rach.raRespWin, "raRespWin", "sl20", "ra-ResponseWindow of RACH-ConfigGeneric[sl1,sl2,sl4,sl8,sl10,sl20,sl40,sl80]")
	rachCmd.Flags().StringVar(&flags.rach.msg3Tp, "msg3Tp", "disabled", "msg3-transformPrecoder of RACH-ConfigGeneric[disabled,enabled]")
	rachCmd.Flags().StringVar(&flags.rach.contResTimer, "contResTimer", "sf64", "ra-ContentionResolutionTimer of RACH-ConfigGeneric[sf8,sf16,sf24,sf32,sf40,sf48,sf56,sf64]")
	rachCmd.Flags().IntVar(&flags.rach._raLen, "_raLen", 139, "L_RA of 3GPP TS 38.211 Table 6.3.3.1-1 and Table 6.3.3.1-2")
	rachCmd.Flags().IntVar(&flags.rach._raNumRbs, "_raNumRbs", 12, "Allocation-expressed-in-number-of-RBs-for-PUSCH of 3GPP TS 38.211 Table 6.3.3.2-1")
	rachCmd.Flags().IntVar(&flags.rach._raKBar, "_raKBar", 2, "k_bar of 3GPP TS 38.211 Table 6.3.3.2-1")
	rachCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.rach.prachConfId", rachCmd.Flags().Lookup("prachConfId"))
	viper.BindPFlag("nrrg.rach._raFormat", rachCmd.Flags().Lookup("_raFormat"))
	viper.BindPFlag("nrrg.rach._raX", rachCmd.Flags().Lookup("_raX"))
	viper.BindPFlag("nrrg.rach._raY", rachCmd.Flags().Lookup("_raY"))
	viper.BindPFlag("nrrg.rach._raSubfNumFr1SlotNumFr2", rachCmd.Flags().Lookup("_raSubfNumFr1SlotNumFr2"))
	viper.BindPFlag("nrrg.rach._raStartingSymb", rachCmd.Flags().Lookup("_raStartingSymb"))
	viper.BindPFlag("nrrg.rach._raNumSlotsPerSubfFr1Per60KSlotFr2", rachCmd.Flags().Lookup("_raNumSlotsPerSubfFr1Per60KSlotFr2"))
	viper.BindPFlag("nrrg.rach._raNumOccasionsPerSlot", rachCmd.Flags().Lookup("_raNumOccasionsPerSlot"))
	viper.BindPFlag("nrrg.rach._raDuration", rachCmd.Flags().Lookup("_raDuration"))
	viper.BindPFlag("nrrg.rach._msg1Scs", rachCmd.Flags().Lookup("_msg1Scs"))
	viper.BindPFlag("nrrg.rach.msg1Fdm", rachCmd.Flags().Lookup("msg1Fdm"))
	viper.BindPFlag("nrrg.rach.msg1FreqStart", rachCmd.Flags().Lookup("msg1FreqStart"))
	viper.BindPFlag("nrrg.rach.totNumPreambs", rachCmd.Flags().Lookup("totNumPreambs"))
	viper.BindPFlag("nrrg.rach.ssbPerRachOccasion", rachCmd.Flags().Lookup("ssbPerRachOccasion"))
	viper.BindPFlag("nrrg.rach.cbPreambsPerSsb", rachCmd.Flags().Lookup("cbPreambsPerSsb"))
	viper.BindPFlag("nrrg.rach.raRespWin", rachCmd.Flags().Lookup("raRespWin"))
	viper.BindPFlag("nrrg.rach.msg3Tp", rachCmd.Flags().Lookup("msg3Tp"))
	viper.BindPFlag("nrrg.rach.contResTimer", rachCmd.Flags().Lookup("contResTimer"))
	viper.BindPFlag("nrrg.rach._raLen", rachCmd.Flags().Lookup("_raLen"))
	viper.BindPFlag("nrrg.rach._raNumRbs", rachCmd.Flags().Lookup("_raNumRbs"))
	viper.BindPFlag("nrrg.rach._raKBar", rachCmd.Flags().Lookup("_raKBar"))
	rachCmd.Flags().MarkHidden("_raFormat")
	rachCmd.Flags().MarkHidden("_raX")
	rachCmd.Flags().MarkHidden("_raY")
	rachCmd.Flags().MarkHidden("_raSubfNumFr1SlotNumFr2")
	rachCmd.Flags().MarkHidden("_raStartingSymb")
	rachCmd.Flags().MarkHidden("_raNumSlotsPerSubfFr1Per60KSlotFr2")
	rachCmd.Flags().MarkHidden("_raNumOccasionsPerSlot")
	rachCmd.Flags().MarkHidden("_raDuration")
	rachCmd.Flags().MarkHidden("_msg1Scs")
	rachCmd.Flags().MarkHidden("_raLen")
	rachCmd.Flags().MarkHidden("_raNumRbs")
	rachCmd.Flags().MarkHidden("_raKBar")
}

func initDmrsCommonCmd() {
	dmrsCommonCmd.Flags().StringSliceVar(&flags.dmrsCommon._tag, "_tag", []string{"SIB1", "Msg2", "Msg4", "Msg3"}, "Information of UL/DL-SCH")
	dmrsCommonCmd.Flags().StringSliceVar(&flags.dmrsCommon._dmrsType, "_dmrsType", []string{"type1", "type1", "type1", "type1"}, "dmrs-Type as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	dmrsCommonCmd.Flags().StringSliceVar(&flags.dmrsCommon._dmrsAddPos, "_dmrsAddPos", []string{"pos0", "pos0", "pos0", "pos1"}, "dmrs-AdditionalPosition as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	dmrsCommonCmd.Flags().StringSliceVar(&flags.dmrsCommon._maxLength, "_maxLength", []string{"len1", "len1", "len1", "len1"}, "maxLength as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	dmrsCommonCmd.Flags().IntSliceVar(&flags.dmrsCommon._dmrsPorts, "_dmrsPorts", []int{1000, 1000, 1000, 0}, "DMRS antenna ports")
	dmrsCommonCmd.Flags().IntSliceVar(&flags.dmrsCommon._cdmGroupsWoData, "_cdmGroupsWoData", []int{1, 1, 1, 2}, "CDM group(s) without data")
	dmrsCommonCmd.Flags().IntSliceVar(&flags.dmrsCommon._numFrontLoadSymbs, "_numFrontLoadSymbs", []int{1, 1, 1, 1}, "Number of front-load DMRS symbols")
	dmrsCommonCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dmrscommon._tag", dmrsCommonCmd.Flags().Lookup("_tag"))
	viper.BindPFlag("nrrg.dmrscommon._dmrsType", dmrsCommonCmd.Flags().Lookup("_dmrsType"))
	viper.BindPFlag("nrrg.dmrscommon._dmrsAddPos", dmrsCommonCmd.Flags().Lookup("_dmrsAddPos"))
	viper.BindPFlag("nrrg.dmrscommon._maxLength", dmrsCommonCmd.Flags().Lookup("_maxLength"))
	viper.BindPFlag("nrrg.dmrscommon._dmrsPorts", dmrsCommonCmd.Flags().Lookup("_dmrsPorts"))
	viper.BindPFlag("nrrg.dmrscommon._cdmGroupsWoData", dmrsCommonCmd.Flags().Lookup("_cdmGroupsWoData"))
	viper.BindPFlag("nrrg.dmrscommon._numFrontLoadSymbs", dmrsCommonCmd.Flags().Lookup("_numFrontLoadSymbs"))
	dmrsCommonCmd.Flags().MarkHidden("_tag")
	dmrsCommonCmd.Flags().MarkHidden("_dmrsType")
	dmrsCommonCmd.Flags().MarkHidden("_dmrsAddPos")
	dmrsCommonCmd.Flags().MarkHidden("_maxLength")
	dmrsCommonCmd.Flags().MarkHidden("_dmrsPorts")
	dmrsCommonCmd.Flags().MarkHidden("_cdmGroupsWoData")
	dmrsCommonCmd.Flags().MarkHidden("_numFrontLoadSymbs")
}

func initPdschCmd() {
	pdschCmd.Flags().StringVar(&flags.pdsch._pdschAggFactor, "_pdschAggFactor", "n1", "pdsch-AggregationFactor of PDSCH-Config[n1,n2,n4,n8]")
	pdschCmd.Flags().StringVar(&flags.pdsch.pdschRbgCfg, "pdschRbgCfg", "config1", "rbg-Size of PDSCH-Config[config1,config2]")
	pdschCmd.Flags().IntVar(&flags.pdsch._rbgSize, "_rbgSize", 16, "RBG size of PDSCH resource allocation type 0")
	pdschCmd.Flags().StringVar(&flags.pdsch.pdschMcsTable, "pdschMcsTable", "qam256", "mcs-Table of PDSCH-Config[qam64,qam256,qam64LowSE]")
	pdschCmd.Flags().StringVar(&flags.pdsch.pdschXOh, "pdschXOh", "xOh0", "xOverhead of PDSCH-ServingCellConfig[xOh0,xOh6,xOh12,xOh18]")
	pdschCmd.Flags().IntVar(&flags.pdsch.pdschMaxLayers, "pdschMaxLayers", 2, "maxMIMO-Layers of PDSCH-ServingCellConfig[1..8]")
	pdschCmd.Flags().StringVar(&flags.pdsch.pdschDmrsType, "pdschDmrsType", "type1", "dmrs-Type as in DMRS-DownlinkConfig[type1,type2]")
	pdschCmd.Flags().StringVar(&flags.pdsch.pdschDmrsAddPos, "pdschDmrsAddPos", "pos0", "dmrs-additionalPosition as in DMRS-DownlinkConfig[pos0,pos1,pos2,pos3]")
	pdschCmd.Flags().StringVar(&flags.pdsch.pdschMaxLength, "pdschMaxLength", "len1", "maxLength as in DMRS-DownlinkConfig[len1,len2]")
	pdschCmd.Flags().IntSliceVar(&flags.pdsch._dmrsPorts, "_dmrsPorts", []int{1000, 1001, 1002, 1003}, "DMRS antenna ports")
	pdschCmd.Flags().IntVar(&flags.pdsch._cdmGroupsWoData, "_cdmGroupsWoData", 2, "CDM group(s) without data")
	pdschCmd.Flags().IntVar(&flags.pdsch._numFrontLoadSymbs, "_numFrontLoadSymbs", 1, "Number of front-load DMRS symbols")
	pdschCmd.Flags().BoolVar(&flags.pdsch.pdschPtrsEnabled, "pdschPtrsEnabled", true, "Enable PTRS of PDSCH[false,true]")
	pdschCmd.Flags().IntVar(&flags.pdsch.pdschPtrsTimeDensity, "pdschPtrsTimeDensity", 1, "The L_PTRS deduced from timeDensity of PTRS-DownlinkConfig[1,2,4]")
	pdschCmd.Flags().IntVar(&flags.pdsch.pdschPtrsFreqDensity, "pdschPtrsFreqDensity", 2, "The K_PTRS deduced from frequencyDensity of PTRS-DownlinkConfig[2,4]")
	pdschCmd.Flags().StringVar(&flags.pdsch.pdschPtrsReOffset, "pdschPtrsReOffset", "offset00", "resourceElementOffset of PTRS-DownlinkConfig[offset00,offset01,offset10,offset11]")
	pdschCmd.Flags().IntVar(&flags.pdsch._ptrsDmrsPorts, "_ptrsDmrsPorts", 1000, "Associated DMRS antenna port")
	pdschCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.pdsch._pdschAggFactor", pdschCmd.Flags().Lookup("_pdschAggFactor"))
	viper.BindPFlag("nrrg.pdsch.pdschRbgCfg", pdschCmd.Flags().Lookup("pdschRbgCfg"))
	viper.BindPFlag("nrrg.pdsch._rbgSize", pdschCmd.Flags().Lookup("_rbgSize"))
	viper.BindPFlag("nrrg.pdsch.pdschMcsTable", pdschCmd.Flags().Lookup("pdschMcsTable"))
	viper.BindPFlag("nrrg.pdsch.pdschXOh", pdschCmd.Flags().Lookup("pdschXOh"))
	viper.BindPFlag("nrrg.pdsch.pdschMaxLayers", pdschCmd.Flags().Lookup("pdschMaxLayers"))
	viper.BindPFlag("nrrg.pdsch.pdschDmrsType", pdschCmd.Flags().Lookup("pdschDmrsType"))
	viper.BindPFlag("nrrg.pdsch.pdschDmrsAddPos", pdschCmd.Flags().Lookup("pdschDmrsAddPos"))
	viper.BindPFlag("nrrg.pdsch.pdschMaxLength", pdschCmd.Flags().Lookup("pdschMaxLength"))
	viper.BindPFlag("nrrg.pdsch._dmrsPorts", pdschCmd.Flags().Lookup("_dmrsPorts"))
	viper.BindPFlag("nrrg.pdsch._cdmGroupsWoData", pdschCmd.Flags().Lookup("_cdmGroupsWoData"))
	viper.BindPFlag("nrrg.pdsch._numFrontLoadSymbs", pdschCmd.Flags().Lookup("_numFrontLoadSymbs"))
	viper.BindPFlag("nrrg.pdsch.pdschPtrsEnabled", pdschCmd.Flags().Lookup("pdschPtrsEnabled"))
	viper.BindPFlag("nrrg.pdsch.pdschPtrsTimeDensity", pdschCmd.Flags().Lookup("pdschPtrsTimeDensity"))
	viper.BindPFlag("nrrg.pdsch.pdschPtrsFreqDensity", pdschCmd.Flags().Lookup("pdschPtrsFreqDensity"))
	viper.BindPFlag("nrrg.pdsch.pdschPtrsReOffset", pdschCmd.Flags().Lookup("pdschPtrsReOffset"))
	viper.BindPFlag("nrrg.pdsch._ptrsDmrsPorts", pdschCmd.Flags().Lookup("_ptrsDmrsPorts"))
	pdschCmd.Flags().MarkHidden("_pdschAggFactor")
	pdschCmd.Flags().MarkHidden("_rbgSize")
	pdschCmd.Flags().MarkHidden("_dmrsPorts")
	pdschCmd.Flags().MarkHidden("_cdmGroupsWoData")
	pdschCmd.Flags().MarkHidden("_numFrontLoadSymbs")
	pdschCmd.Flags().MarkHidden("_ptrsDmrsPorts")
}

func initPuschCmd() {
	puschCmd.Flags().StringVar(&flags.pusch.puschTxCfg, "puschTxCfg", "codebook", "txConfig of PUSCH-Config[codebook,nonCodebook]")
	puschCmd.Flags().StringVar(&flags.pusch.puschCbSubset, "puschCbSubset", "fullyAndPartialAndNonCoherent", "codebookSubset of PUSCH-Config[fullyAndPartialAndNonCoherent,partialAndNonCoherent,nonCoherent]")
	puschCmd.Flags().IntVar(&flags.pusch.puschCbMaxRankNonCbMaxLayers, "puschCbMaxRankNonCbMaxLayers", 2, "maxRank of PUSCH-Config or maxMIMO-Layers of PUSCH-ServingCellConfig[1..4]")
	puschCmd.Flags().StringVar(&flags.pusch.puschTp, "puschTp", "disabled", "transformPrecoder of PUSCH-Config[disabled,enabled]")
	puschCmd.Flags().StringVar(&flags.pusch._puschAggFactor, "_puschAggFactor", "n1", "pusch-AggregationFactor of PUSCH-Config[n1,n2,n4,n8]")
	puschCmd.Flags().StringVar(&flags.pusch.puschRbgCfg, "puschRbgCfg", "config1", "rbg-Size of PUSCH-Config[config1,config2]")
	puschCmd.Flags().IntVar(&flags.pusch._rbgSize, "_rbgSize", 16, "RBG size of PUSCH resource allocation type 0")
	puschCmd.Flags().StringVar(&flags.pusch.puschMcsTable, "puschMcsTable", "qam64", "mcs-Table of PUSCH-Config[qam64,qam256,qam64LowSE]")
	puschCmd.Flags().StringVar(&flags.pusch.puschXOh, "puschXOh", "xOh0", "xOverhead of PUSCH-ServingCellConfig[xOh0,xOh6,xOh12,xOh18]")
	puschCmd.Flags().StringVar(&flags.pusch._puschRepType, "_puschRepType", "typeA", "pusch-RepTypeIndicator of PUSCH-Config[typeA,typeB]")
	puschCmd.Flags().StringVar(&flags.pusch.puschDmrsType, "puschDmrsType", "type1", "dmrs-Type as in DMRS-UplinkConfig[type1,type2]")
	puschCmd.Flags().StringVar(&flags.pusch.puschDmrsAddPos, "puschDmrsAddPos", "pos0", "dmrs-additionalPosition as in DMRS-UplinkConfig[pos0,pos1,pos2,pos3]")
	puschCmd.Flags().StringVar(&flags.pusch.puschMaxLength, "puschMaxLength", "len1", "maxLength as in DMRS-UplinkConfig[len1,len2]")
	puschCmd.Flags().IntSliceVar(&flags.pusch._dmrsPorts, "_dmrsPorts", []int{0, 1}, "DMRS antenna ports")
	puschCmd.Flags().IntVar(&flags.pusch._cdmGroupsWoData, "_cdmGroupsWoData", 1, "CDM group(s) without data")
	puschCmd.Flags().IntVar(&flags.pusch._numFrontLoadSymbs, "_numFrontLoadSymbs", 1, "Number of front-load DMRS symbols")
	puschCmd.Flags().BoolVar(&flags.pusch.puschPtrsEnabled, "puschPtrsEnabled", true, "Enable PTRS of PDSCH[false,true]")
	puschCmd.Flags().IntVar(&flags.pusch.puschPtrsTimeDensity, "puschPtrsTimeDensity", 1, "The L_PTRS deduced from timeDensity of PTRS-UplinkConfig for CP-OFDM[1,2,4]")
	puschCmd.Flags().IntVar(&flags.pusch.puschPtrsFreqDensity, "puschPtrsFreqDensity", 2, "The K_PTRS deduced from frequencyDensity of PTRS-UplinkConfig for CP-OFDM[2,4]")
	puschCmd.Flags().StringVar(&flags.pusch.puschPtrsReOffset, "puschPtrsReOffset", "offset00", "resourceElementOffset of PTRS-UplinkConfig for CP-OFDM[offset00,offset01,offset10,offset11]")
	puschCmd.Flags().StringVar(&flags.pusch.puschPtrsMaxNumPorts, "puschPtrsMaxNumPorts", "n1", "maxNrofPorts of PTRS-UplinkConfig for CP-OFDM[n1,n2]")
	puschCmd.Flags().IntVar(&flags.pusch.puschPtrsTimeDensityTp, "puschPtrsTimeDensityTp", 1, "The L_PTRS deduced from timeDensityTransformPrecoding of PTRS-UplinkConfig for DFS-S-OFDM[1,2]")
	puschCmd.Flags().StringVar(&flags.pusch.puschPtrsGrpPatternTp, "puschPtrsGrpPatternTp", "pat0", "The Scheduled-bandwidth column index of 3GPP TS 38.214 Table 6.2.3.2-1, deduced from sampleDensity of PTRS-UplinkConfig for DFS-S-OFDM[pat0,pat1,pat2,pat3,pat4]")
	puschCmd.Flags().IntVar(&flags.pusch._numGrpsTp, "_numGrpsTp", 2, "The Number-of-PT-RS-groups of 3GPP TS 38.214 Table 6.2.3.2-1, deduced from sampleDensity of PTRS-UplinkConfig for DFS-S-OFDM")
	puschCmd.Flags().IntVar(&flags.pusch._samplesPerGrpTp, "_samplesPerGrpTp", 2, "The Number-of-samples-per-PT-RS-group of 3GPP TS 38.214 Table 6.2.3.2-1, deduced from sampleDensity of PTRS-UplinkConfig for DFS-S-OFDM")
	puschCmd.Flags().IntSliceVar(&flags.pusch._ptrsDmrsPorts, "_ptrsDmrsPorts", []int{0}, "Associated DMRS antenna ports for CP-OFDM")
	//puschCmd.Flags().IntVar(&flags.pusch._ptrsDmrsPortsTp, "_ptrsDmrsPortsTp", -1, "Associated DMRS antenna ports for DFT-S-OFDM")
	puschCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.pusch.puschTxCfg", puschCmd.Flags().Lookup("puschTxCfg"))
	viper.BindPFlag("nrrg.pusch.puschCbSubset", puschCmd.Flags().Lookup("puschCbSubset"))
	viper.BindPFlag("nrrg.pusch.puschCbMaxRankNonCbMaxLayers", puschCmd.Flags().Lookup("puschCbMaxRankNonCbMaxLayers"))
	viper.BindPFlag("nrrg.pusch.puschTp", puschCmd.Flags().Lookup("puschTp"))
	viper.BindPFlag("nrrg.pusch._puschAggFactor", puschCmd.Flags().Lookup("_puschAggFactor"))
	viper.BindPFlag("nrrg.pusch.puschRbgCfg", puschCmd.Flags().Lookup("puschRbgCfg"))
	viper.BindPFlag("nrrg.pusch._rbgSize", puschCmd.Flags().Lookup("_rbgSize"))
	viper.BindPFlag("nrrg.pusch.puschMcsTable", puschCmd.Flags().Lookup("puschMcsTable"))
	viper.BindPFlag("nrrg.pusch.puschXOh", puschCmd.Flags().Lookup("puschXOh"))
	viper.BindPFlag("nrrg.pusch._puschRepType", puschCmd.Flags().Lookup("_puschRepType"))
	viper.BindPFlag("nrrg.pusch.puschDmrsType", puschCmd.Flags().Lookup("puschDmrsType"))
	viper.BindPFlag("nrrg.pusch.puschDmrsAddPos", puschCmd.Flags().Lookup("puschDmrsAddPos"))
	viper.BindPFlag("nrrg.pusch.puschMaxLength", puschCmd.Flags().Lookup("puschMaxLength"))
	viper.BindPFlag("nrrg.pusch._dmrsPorts", puschCmd.Flags().Lookup("_dmrsPorts"))
	viper.BindPFlag("nrrg.pusch._cdmGroupsWoData", puschCmd.Flags().Lookup("_cdmGroupsWoData"))
	viper.BindPFlag("nrrg.pusch._numFrontLoadSymbs", puschCmd.Flags().Lookup("_numFrontLoadSymbs"))
	viper.BindPFlag("nrrg.pusch.puschPtrsEnabled", puschCmd.Flags().Lookup("puschPtrsEnabled"))
	viper.BindPFlag("nrrg.pusch.puschPtrsTimeDensity", puschCmd.Flags().Lookup("puschPtrsTimeDensity"))
	viper.BindPFlag("nrrg.pusch.puschPtrsFreqDensity", puschCmd.Flags().Lookup("puschPtrsFreqDensity"))
	viper.BindPFlag("nrrg.pusch.puschPtrsReOffset", puschCmd.Flags().Lookup("puschPtrsReOffset"))
	viper.BindPFlag("nrrg.pusch.puschPtrsMaxNumPorts", puschCmd.Flags().Lookup("puschPtrsMaxNumPorts"))
	viper.BindPFlag("nrrg.pusch.puschPtrsTimeDensityTp", puschCmd.Flags().Lookup("puschPtrsTimeDensityTp"))
	viper.BindPFlag("nrrg.pusch.puschPtrsGrpPatternTp", puschCmd.Flags().Lookup("puschPtrsGrpPatternTp"))
	viper.BindPFlag("nrrg.pusch._numGrpsTp", puschCmd.Flags().Lookup("_numGrpsTp"))
	viper.BindPFlag("nrrg.pusch._samplesPerGrpTp", puschCmd.Flags().Lookup("_samplesPerGrpTp"))
	//viper.BindPFlag("nrrg.pusch._ptrsDmrsPortsTp", puschCmd.Flags().Lookup("_ptrsDmrsPortsTp"))
	viper.BindPFlag("nrrg.pusch._ptrsDmrsPorts", puschCmd.Flags().Lookup("_ptrsDmrsPorts"))
	puschCmd.Flags().MarkHidden("_puschAggFactor")
	puschCmd.Flags().MarkHidden("_rbgSize")
	puschCmd.Flags().MarkHidden("_puschRepType")
	puschCmd.Flags().MarkHidden("_dmrsPorts")
	puschCmd.Flags().MarkHidden("_cdmGroupsWoData")
	puschCmd.Flags().MarkHidden("_numFrontLoadSymbs")
	puschCmd.Flags().MarkHidden("_numGrpsTp")
	puschCmd.Flags().MarkHidden("_samplesPerGrpTp")
	puschCmd.Flags().MarkHidden("_ptrsDmrsPortsTp")
	puschCmd.Flags().MarkHidden("_ptrsDmrsPorts")
}

func initCsiCmd() {
	csiCmd.Flags().IntSliceVar(&flags.csi._resSetId, "_resSetId", []int{0, 1}, "nzp-CSI-ResourceSetId of NZP-CSI-RS-ResourceSet")
	csiCmd.Flags().StringSliceVar(&flags.csi._trsInfo, "_trsInfo", []string{"false", "true"}, "trs-Info of NZP-CSI-RS-ResourceSet")
	csiCmd.Flags().IntSliceVar(&flags.csi._resId, "_resId", []int{0, 1}, "nzp-CSI-RS-ResourceId of NZP-CSI-RS-Resource")
	csiCmd.Flags().StringSliceVar(&flags.csi.freqAllocRow, "freqAllocRow", []string{"row4", "row1"}, "The row of frequencyDomainAllocation of CSI-RS-ResourceMapping[row1,row2,row4,other]")
	csiCmd.Flags().StringSliceVar(&flags.csi.freqAllocBits, "freqAllocBits", []string{"001", "0001"}, "The bit-string of frequencyDomainAllocation of CSI-RS-ResourceMapping")
	csiCmd.Flags().StringSliceVar(&flags.csi._numPorts, "_numPorts", []string{"p4", "p1"}, "nrofPorts of CSI-RS-ResourceMapping[p1,p2,p4,p8,p12,p16,p24,p32]")
	csiCmd.Flags().StringSliceVar(&flags.csi._cdmType, "_cdmType", []string{"fd-CDM2", "noCDM"}, "cdm-Type of CSI-RS-ResourceMapping[noCDM,fd-CDM2,cdm4-FD2-TD2,cdm8-FD2-TD4]")
	csiCmd.Flags().StringSliceVar(&flags.csi._density, "_density", []string{"one", "three"}, "density of CSI-RS-ResourceMapping[evenPRBs,oddPRBs,one,three]")
	csiCmd.Flags().IntSliceVar(&flags.csi._firstSymb, "_firstSymb", []int{13, 6}, "firstOFDMSymbolInTimeDomain of CSI-RS-ResourceMapping[0..13]")
	//csiCmd.Flags().IntSliceVar(&flags.csi._firstSymb2, "_firstSymb2", []int{-1, -1}, "firstOFDMSymbolInTimeDomain2 of CSI-RS-ResourceMapping[-1 or 0..13]")
	csiCmd.Flags().IntSliceVar(&flags.csi._startRb, "_startRb", []int{0, 0}, "startingRB of CSI-FrequencyOccupation[0..274]")
	csiCmd.Flags().IntSliceVar(&flags.csi._numRbs, "_numRbs", []int{160, 160}, "nrofRBs of CSI-FrequencyOccupation[24..276]")
	csiCmd.Flags().StringSliceVar(&flags.csi.period, "period", []string{"slots20", "slots10"}, "periodicityAndOffset of NZP-CSI-RS-Resource[slots4,slots5,slots8,slots10,slots16,slots20,slots32,slots40,slots64,slots80,slots160,slots320,slots640]")
	csiCmd.Flags().IntSliceVar(&flags.csi.offset, "offset", []int{6, 0}, "periodicityAndOffset of NZP-CSI-RS-Resource[0..period-1]")
	csiCmd.Flags().StringVar(&flags.csi._csiImRePattern, "_csiImRePattern", "pattern1", "csi-IM-ResourceElementPattern of CSI-IM-Resource[pattern0,pattern1]")
	csiCmd.Flags().StringVar(&flags.csi._csiImScLoc, "_csiImScLoc", "s4", "subcarrierLocation of csi-IM-ResourceElementPattern of CSI-IM-Resource[s0,s2,s4,s6,s8,s10]")
	csiCmd.Flags().IntVar(&flags.csi._csiImSymbLoc, "_csiImSymbLoc", 13, "symbolLocation of csi-IM-ResourceElementPattern of CSI-IM-Resource[0..12]")
	csiCmd.Flags().IntVar(&flags.csi._csiImStartRb, "_csiImStartRb", 0, "startingRB of CSI-FrequencyOccupation[0..274]")
	csiCmd.Flags().IntVar(&flags.csi._csiImNumRbs, "_csiImNumRbs", 160, "nrofRBs of CSI-FrequencyOccupation[24..276]")
	csiCmd.Flags().StringVar(&flags.csi._csiImPeriod, "_csiImPeriod", "slots20", "periodicityAndOffset of CSI-IM-Resource[slots4,slots5,slots8,slots10,slots16,slots20,slots32,slots40,slots64,slots80,slots160,slots320,slots640]")
	csiCmd.Flags().IntVar(&flags.csi._csiImOffset, "_csiImOffset", 6, "periodicityAndOffset of CSI-IM-Resource[0..period-1]")
	csiCmd.Flags().StringVar(&flags.csi._resType, "_resType", "periodic", "resourceType of CSI-ResourceConfig")
	csiCmd.Flags().StringVar(&flags.csi._repCfgType, "_repCfgType", "periodic", "reportConfigType of CSI-ReportConfig")
	csiCmd.Flags().StringVar(&flags.csi.csiRepPeriod, "csiRepPeriod", "slots40", "CSI-ReportPeriodicityAndOffset of CSI-ReportConfig[slots4,slots5,slots8,slots10,slots16,slots20,slots40,slots80,slots160,slots320]")
	csiCmd.Flags().IntVar(&flags.csi.csiRepOffset, "csiRepOffset", 8, "CSI-ReportPeriodicityAndOffset of CSI-ReportConfig[0..period-1]")
	csiCmd.Flags().IntVar(&flags.csi._csiRepPucchRes, "_csiRepPucchRes", 1, "pucch-Resource of PUCCH-CSI-Resource of CSI-ReportConfig")
	csiCmd.Flags().StringVar(&flags.csi._quantity, "_quantity", "cri-RI-PMI-CQI", "reportQuantity of CSI-ReportConfig")
	csiCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.csi._resSetId", csiCmd.Flags().Lookup("_resSetId"))
	viper.BindPFlag("nrrg.csi._trsInfo", csiCmd.Flags().Lookup("_trsInfo"))
	viper.BindPFlag("nrrg.csi._resId", csiCmd.Flags().Lookup("_resId"))
	viper.BindPFlag("nrrg.csi.freqAllocRow", csiCmd.Flags().Lookup("freqAllocRow"))
	viper.BindPFlag("nrrg.csi.freqAllocBits", csiCmd.Flags().Lookup("freqAllocBits"))
	viper.BindPFlag("nrrg.csi._numPorts", csiCmd.Flags().Lookup("_numPorts"))
	viper.BindPFlag("nrrg.csi._cdmType", csiCmd.Flags().Lookup("_cdmType"))
	viper.BindPFlag("nrrg.csi._density", csiCmd.Flags().Lookup("_density"))
	viper.BindPFlag("nrrg.csi._firstSymb", csiCmd.Flags().Lookup("_firstSymb"))
	//viper.BindPFlag("nrrg.csi._firstSymb2", csiCmd.Flags().Lookup("_firstSymb2"))
	viper.BindPFlag("nrrg.csi._startRb", csiCmd.Flags().Lookup("_startRb"))
	viper.BindPFlag("nrrg.csi._numRbs", csiCmd.Flags().Lookup("_numRbs"))
	viper.BindPFlag("nrrg.csi.period", csiCmd.Flags().Lookup("period"))
	viper.BindPFlag("nrrg.csi.offset", csiCmd.Flags().Lookup("offset"))
	viper.BindPFlag("nrrg.csi._csiImRePattern", csiCmd.Flags().Lookup("_csiImRePattern"))
	viper.BindPFlag("nrrg.csi._csiImScLoc", csiCmd.Flags().Lookup("_csiImScLoc"))
	viper.BindPFlag("nrrg.csi._csiImSymbLoc", csiCmd.Flags().Lookup("_csiImSymbLoc"))
	viper.BindPFlag("nrrg.csi._csiImStartRb", csiCmd.Flags().Lookup("_csiImStartRb"))
	viper.BindPFlag("nrrg.csi._csiImNumRbs", csiCmd.Flags().Lookup("_csiImNumRbs"))
	viper.BindPFlag("nrrg.csi._csiImPeriod", csiCmd.Flags().Lookup("_csiImPeriod"))
	viper.BindPFlag("nrrg.csi._csiImOffset", csiCmd.Flags().Lookup("_csiImOffset"))
	viper.BindPFlag("nrrg.csi._resType", csiCmd.Flags().Lookup("_resType"))
	viper.BindPFlag("nrrg.csi._repCfgType", csiCmd.Flags().Lookup("_repCfgType"))
	viper.BindPFlag("nrrg.csi.csiRepPeriod", csiCmd.Flags().Lookup("csiRepPeriod"))
	viper.BindPFlag("nrrg.csi.csiRepOffset", csiCmd.Flags().Lookup("csiRepOffset"))
	viper.BindPFlag("nrrg.csi._csiRepPucchRes", csiCmd.Flags().Lookup("_csiRepPucchRes"))
	viper.BindPFlag("nrrg.csi._quantity", csiCmd.Flags().Lookup("_quantity"))
	csiCmd.Flags().MarkHidden("_resSetId")
	csiCmd.Flags().MarkHidden("_trsInfo")
	csiCmd.Flags().MarkHidden("_resId")
	csiCmd.Flags().MarkHidden("_numPorts")
	csiCmd.Flags().MarkHidden("_cdmType")
	csiCmd.Flags().MarkHidden("_density")
	csiCmd.Flags().MarkHidden("_firstSymb")
	//csiCmd.Flags().MarkHidden("_firstSymb2")
	csiCmd.Flags().MarkHidden("_startRb")
	csiCmd.Flags().MarkHidden("_numRbs")
	csiCmd.Flags().MarkHidden("_csiImRePattern")
	csiCmd.Flags().MarkHidden("_csiImScLoc")
	csiCmd.Flags().MarkHidden("_csiImSymbLoc")
	csiCmd.Flags().MarkHidden("_csiImStartRb")
	csiCmd.Flags().MarkHidden("_csiImNumRbs")
	csiCmd.Flags().MarkHidden("_csiImPeriod")
	csiCmd.Flags().MarkHidden("_csiImOffset")
	csiCmd.Flags().MarkHidden("_resType")
	csiCmd.Flags().MarkHidden("_repCfgType")
	csiCmd.Flags().MarkHidden("_csiRepPucchRes")
	csiCmd.Flags().MarkHidden("_quantity")
}

func initSrsCmd() {
	srsCmd.Flags().IntSliceVar(&flags.srs._resId, "_resId", []int{0, 1, 2, 3, 4}, "srs-ResourceId of SRS-Resource")
	srsCmd.Flags().StringSliceVar(&flags.srs.srsNumPorts, "srsNumPorts", []string{"ports2", "port1", "port1", "port1", "port1"}, "nrofSRS-Ports of SRS-Resource[port1,ports2,ports4]")
	srsCmd.Flags().StringSliceVar(&flags.srs._srsNonCbPtrsPort, "_srsNonCbPtrsPort", []string{"-", "n0", "n0", "n1", "n1"}, "ptrs-PortIndex of SRS-Resource[n0,n1]")
	srsCmd.Flags().StringSliceVar(&flags.srs.srsNumCombs, "srsNumCombs", []string{"n4", "n2", "n2", "n2", "n2"}, "transmissionComb of SRS-Resource[n2,n4]")
	srsCmd.Flags().IntSliceVar(&flags.srs.srsCombOff, "srsCombOff", []int{0, 0, 0, 0, 0}, "combOffset of SRS-Resource")
	srsCmd.Flags().IntSliceVar(&flags.srs.srsCs, "srsCs", []int{11, 0, 0, 0, 0}, "cyclicShift of SRS-Resource")
	srsCmd.Flags().IntSliceVar(&flags.srs.srsStartPos, "srsStartPos", []int{3, 0, 0, 0, 0}, "startPosition of SRS-Resource[0..5]")
	srsCmd.Flags().StringSliceVar(&flags.srs.srsNumSymbs, "srsNumSymbs", []string{"n4", "n1", "n1", "n1", "n1"}, "nrofSymbols of SRS-Resource[n1,n2,n4]")
	srsCmd.Flags().StringSliceVar(&flags.srs.srsRepetition, "srsRepetition", []string{"n4", "n1", "n1", "n1", "n1"}, "repetitionFactor of SRS-Resource[n1,n2,n4]")
	srsCmd.Flags().IntSliceVar(&flags.srs.srsFreqPos, "srsFreqPos", []int{0, 0, 0, 0, 0}, "freqDomainPosition of SRS-Resource[0..67]")
	srsCmd.Flags().IntSliceVar(&flags.srs.srsFreqShift, "srsFreqShift", []int{0, 0, 0, 0, 0}, "freqDomainShift of SRS-Resource[0..268]")
	srsCmd.Flags().IntSliceVar(&flags.srs.srsCSrs, "srsCSrs", []int{12, 0, 0, 0, 0}, "c-SRS of SRS-Resource[0..63]")
	srsCmd.Flags().IntSliceVar(&flags.srs.srsBSrs, "srsBSrs", []int{1, 0, 0, 0, 0}, "b-SRS of SRS-Resource[0..3]")
	srsCmd.Flags().IntSliceVar(&flags.srs.srsBHop, "srsBHop", []int{0, 0, 0, 0, 0}, "b-hop of SRS-Resource[0..3]")
	srsCmd.Flags().StringSliceVar(&flags.srs._resType, "_resType", []string{"periodic", "periodic", "periodic", "periodic", "periodic"}, "resourceType of SRS-Resource")
	srsCmd.Flags().StringSliceVar(&flags.srs.srsPeriod, "srsPeriod", []string{"sl10", "sl5", "sl5", "sl5", "sl5"}, "SRS-PeriodicityAndOffset of SRS-Resource[sl1,sl2,sl4,sl5,sl8,sl10,sl16,sl20,sl32,sl40,sl64,sl80,sl160,sl320,sl640,sl1280,sl2560]")
	srsCmd.Flags().IntSliceVar(&flags.srs.srsOffset, "srsOffset", []int{7, 0, 0, 0, 0}, "SRS-PeriodicityAndOffset of SRS-Resource[0..period-1]")
	srsCmd.Flags().StringSliceVar(&flags.srs._mSRSb, "_mSRSb", []string{"48_16_8_4", "4_4_4_4", "4_4_4_4", "4_4_4_4", "4_4_4_4"}, "The m_SRS_b with b=B_SRS of 3GPP TS 38.211 Table 6.4.1.4.3-1")
	srsCmd.Flags().StringSliceVar(&flags.srs._Nb, "_Nb", []string{"1_3_2_2", "1_1_1_1", "1_1_1_1", "1_1_1_1", "1_1_1_1"}, "The N_b with b=B_SRS of 3GPP TS 38.211 Table 6.4.1.4.3-1")
	srsCmd.Flags().IntSliceVar(&flags.srs._resSetId, "_resSetId", []int{0, 1, 2}, "srs-ResourceSetId of SRS-ResourceSet")
	srsCmd.Flags().StringSliceVar(&flags.srs.srsSetResIdList, "srsSetResIdList", []string{"0", "1_2_3_4", "1_2"}, "srs-ResourceIdList of SRS-ResourceSet")
	srsCmd.Flags().StringSliceVar(&flags.srs._resSetType, "_resSetType", []string{"periodic", "periodic", "periodic"}, "resourceType of SRS-ResourceSet")
	srsCmd.Flags().StringSliceVar(&flags.srs._usage, "_usage", []string{"codebook", "nonCodebook", "antennaSwitching"}, "usage of SRS-ResourceSet")
	srsCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.srs._resId", srsCmd.Flags().Lookup("_resId"))
	viper.BindPFlag("nrrg.srs.srsNumPorts", srsCmd.Flags().Lookup("srsNumPorts"))
	viper.BindPFlag("nrrg.srs._srsNonCbPtrsPort", srsCmd.Flags().Lookup("_srsNonCbPtrsPort"))
	viper.BindPFlag("nrrg.srs.srsNumCombs", srsCmd.Flags().Lookup("srsNumCombs"))
	viper.BindPFlag("nrrg.srs.srsCombOff", srsCmd.Flags().Lookup("srsCombOff"))
	viper.BindPFlag("nrrg.srs.srsCs", srsCmd.Flags().Lookup("srsCs"))
	viper.BindPFlag("nrrg.srs.srsStartPos", srsCmd.Flags().Lookup("srsStartPos"))
	viper.BindPFlag("nrrg.srs.srsNumSymbs", srsCmd.Flags().Lookup("srsNumSymbs"))
	viper.BindPFlag("nrrg.srs.srsRepetition", srsCmd.Flags().Lookup("srsRepetition"))
	viper.BindPFlag("nrrg.srs.srsFreqPos", srsCmd.Flags().Lookup("srsFreqPos"))
	viper.BindPFlag("nrrg.srs.srsFreqShift", srsCmd.Flags().Lookup("srsFreqShift"))
	viper.BindPFlag("nrrg.srs.srsCSrs", srsCmd.Flags().Lookup("srsCSrs"))
	viper.BindPFlag("nrrg.srs.srsBSrs", srsCmd.Flags().Lookup("srsBSrs"))
	viper.BindPFlag("nrrg.srs.srsBHop", srsCmd.Flags().Lookup("srsBHop"))
	viper.BindPFlag("nrrg.srs._resType", srsCmd.Flags().Lookup("_resType"))
	viper.BindPFlag("nrrg.srs.srsPeriod", srsCmd.Flags().Lookup("srsPeriod"))
	viper.BindPFlag("nrrg.srs.srsOffset", srsCmd.Flags().Lookup("srsOffset"))
	viper.BindPFlag("nrrg.srs._mSRSb", srsCmd.Flags().Lookup("_mSRSb"))
	viper.BindPFlag("nrrg.srs._Nb", srsCmd.Flags().Lookup("_Nb"))
	viper.BindPFlag("nrrg.srs._resSetId", srsCmd.Flags().Lookup("_resSetId"))
	viper.BindPFlag("nrrg.srs.srsSetResIdList", srsCmd.Flags().Lookup("srsSetResIdList"))
	viper.BindPFlag("nrrg.srs._resSetType", srsCmd.Flags().Lookup("_resSetType"))
	viper.BindPFlag("nrrg.srs._usage", srsCmd.Flags().Lookup("_usage"))
	srsCmd.Flags().MarkHidden("_resId")
	srsCmd.Flags().MarkHidden("_srsNonCbPtrsPort")
	srsCmd.Flags().MarkHidden("_resType")
	srsCmd.Flags().MarkHidden("_mSRSb")
	srsCmd.Flags().MarkHidden("_Nb")
	srsCmd.Flags().MarkHidden("_resSetId")
	srsCmd.Flags().MarkHidden("_resSetType")
	srsCmd.Flags().MarkHidden("_usage")
}

func initPucchCmd() {
	pucchCmd.Flags().StringVar(&flags.pucch._numSlots, "_numSlots", "n1", "nrofSlots of PUCCH-FormatConfig for PUCCH format 1/3/4[n1,n2,n4,n8]")
	pucchCmd.Flags().StringVar(&flags.pucch._interSlotFreqHop, "_interSlotFreqHop", "disabled", "interslotFrequencyHopping of PUCCH-FormatConfig for PUCCH format 1/3/4[disabled,enabled]")
	pucchCmd.Flags().BoolVar(&flags.pucch._addDmrs, "_addDmrs", true, "additionalDMRS of PUCCH-FormatConfig for PUCCH format 3/4")
	pucchCmd.Flags().BoolVar(&flags.pucch._simHarqAckCsi, "_simHarqAckCsi", true, "simultaneousHARQ-ACK-CSI of PUCCH-FormatConfig for PUCCH format 2/3/4")
	pucchCmd.Flags().IntSliceVar(&flags.pucch._pucchResId, "_pucchResId", []int{0, 1, 2}, "pucch-ResourceId of PUCCH-Resource")
	pucchCmd.Flags().StringSliceVar(&flags.pucch._pucchFormat, "_pucchFormat", []string{"format1", "format3", "format1"}, "format of PUCCH-Resource")
	//pucchCmd.Flags().IntSliceVar(&flags.pucch._pucchResSetId, "_pucchResSetId", []int{0, 0, 1, 1, 1}, "pucch-ResourceSetId of PUCCH-ResourceSet")
	pucchCmd.Flags().IntSliceVar(&flags.pucch._pucchStartRb, "_pucchStartRb", []int{1, 2, 0}, "startingPRB of PUCCH-ResourceSet[0..274]")
	pucchCmd.Flags().StringSliceVar(&flags.pucch._pucchIntraSlotFreqHop, "_pucchIntraSlotFreqHop", []string{"enabled", "enabled", "enabled"}, "intraSlotFrequencyHopping of PUCCH-Resource[disabled,enabled]")
	pucchCmd.Flags().IntSliceVar(&flags.pucch._pucchSecondHopPrb, "_pucchSecondHopPrb", []int{158, 157, 159}, "secondHopPRB of PUCCH-Resource[0..274]")
	pucchCmd.Flags().IntSliceVar(&flags.pucch._pucchNumRbs, "_pucchNumRbs", []int{1, 1, 1}, "nrofPRBs of PUCCH-Resource, fixed to 1 for PUCCH format 0/1/4[1..16]")
	pucchCmd.Flags().IntSliceVar(&flags.pucch._pucchStartSymb, "_pucchStartSymb", []int{0, 0, 0}, "startingSymbolIndex of PUCCH-Resource[0..13(format 0/2) or 0..10(format 1/3/4)]")
	pucchCmd.Flags().IntSliceVar(&flags.pucch._pucchNumSymbs, "_pucchNumSymbs", []int{14, 14, 14}, "nrofSymbols of PUCCH-Resource[1..2(format 0/2) or 4..14(format 1/3/4)]")
	//pucchCmd.Flags().IntSliceVar(&flags.pucch._dsrResId, "_dsrResId", []int{0, 1}, "schedulingRequestResourceId of SchedulingRequestResourceConfig")
	pucchCmd.Flags().StringVar(&flags.pucch.dsrPeriod, "dsrPeriod", "sl20", "periodicityAndOffset of SchedulingRequestResourceConfig[sym2,sym6or7,sl1,sl2,sl4,sl5,sl8,sl10,sl16,sl20,sl40,sl80,sl160,sl320,sl640]")
	pucchCmd.Flags().IntVar(&flags.pucch.dsrOffset, "dsrOffset", 2, "periodicityAndOffset of SchedulingRequestResourceConfig[0..period-1]")
	pucchCmd.Flags().IntVar(&flags.pucch._dsrPucchRes, "_dsrPucchRes", 2, "resource of SchedulingRequestResourceConfig")
	pucchCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.pucch._numSlots", pucchCmd.Flags().Lookup("_numSlots"))
	viper.BindPFlag("nrrg.pucch._interSlotFreqHop", pucchCmd.Flags().Lookup("_interSlotFreqHop"))
	viper.BindPFlag("nrrg.pucch._addDmrs", pucchCmd.Flags().Lookup("_addDmrs"))
	viper.BindPFlag("nrrg.pucch._simHarqAckCsi", pucchCmd.Flags().Lookup("_simHarqAckCsi"))
	viper.BindPFlag("nrrg.pucch._pucchResId", pucchCmd.Flags().Lookup("_pucchResId"))
	viper.BindPFlag("nrrg.pucch._pucchFormat", pucchCmd.Flags().Lookup("_pucchFormat"))
	//viper.BindPFlag("nrrg.pucch._pucchResSetId", pucchCmd.Flags().Lookup("_pucchResSetId"))
	viper.BindPFlag("nrrg.pucch._pucchStartRb", pucchCmd.Flags().Lookup("_pucchStartRb"))
	viper.BindPFlag("nrrg.pucch._pucchIntraSlotFreqHop", pucchCmd.Flags().Lookup("_pucchIntraSlotFreqHop"))
	viper.BindPFlag("nrrg.pucch._pucchSecondHopPrb", pucchCmd.Flags().Lookup("_pucchSecondHopPrb"))
	viper.BindPFlag("nrrg.pucch._pucchNumRbs", pucchCmd.Flags().Lookup("_pucchNumRbs"))
	viper.BindPFlag("nrrg.pucch._pucchStartSymb", pucchCmd.Flags().Lookup("_pucchStartSymb"))
	viper.BindPFlag("nrrg.pucch._pucchNumSymbs", pucchCmd.Flags().Lookup("_pucchNumSymbs"))
	//viper.BindPFlag("nrrg.pucch._dsrResId", pucchCmd.Flags().Lookup("_dsrResId"))
	viper.BindPFlag("nrrg.pucch.dsrPeriod", pucchCmd.Flags().Lookup("dsrPeriod"))
	viper.BindPFlag("nrrg.pucch.dsrOffset", pucchCmd.Flags().Lookup("dsrOffset"))
	viper.BindPFlag("nrrg.pucch._dsrPucchRes", pucchCmd.Flags().Lookup("_dsrPucchRes"))
	pucchCmd.Flags().MarkHidden("_numSlots")
	pucchCmd.Flags().MarkHidden("_interSlotFreqHop")
	pucchCmd.Flags().MarkHidden("_addDmrs")
	pucchCmd.Flags().MarkHidden("_simHarqAckCsi")
	pucchCmd.Flags().MarkHidden("_pucchResId")
	pucchCmd.Flags().MarkHidden("_pucchFormat")
	//pucchCmd.Flags().MarkHidden("_pucchResSetId")
	pucchCmd.Flags().MarkHidden("_pucchStartRb")
	pucchCmd.Flags().MarkHidden("_pucchIntraSlotFreqHop")
	pucchCmd.Flags().MarkHidden("_pucchSecondHopPrb")
	pucchCmd.Flags().MarkHidden("_pucchNumRbs")
	pucchCmd.Flags().MarkHidden("_pucchStartSymb")
	pucchCmd.Flags().MarkHidden("_pucchNumSymbs")
	//pucchCmd.Flags().MarkHidden("_dsrResId")
	pucchCmd.Flags().MarkHidden("_dsrPucchRes")
}

func initAdvancedCmd() {
	advancedCmd.Flags().IntVar(&flags.advanced.bestSsb, "bestSsb", 0, "Best SSB index")
	advancedCmd.Flags().IntVar(&flags.advanced.pdcchSlotSib1, "pdcchSlotSib1", -1, "PDCCH slot for SIB1")
	advancedCmd.Flags().IntVar(&flags.advanced.prachOccMsg1, "prachOccMsg1", -1, "PRACH occasion for Msg1")
	advancedCmd.Flags().IntVar(&flags.advanced.pdcchOccMsg2, "pdcchOccMsg2", 4, "PDCCH occasion for Msg2")
	advancedCmd.Flags().IntVar(&flags.advanced.pdcchOccMsg4, "pdcchOccMsg4", 0, "PDCCH occasion for Msg4")
	//advancedCmd.Flags().IntVar(&flags.advanced.dsrRes, "dsrRes", 0, "DSR resource index")
	advancedCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.advanced.bestSsb", advancedCmd.Flags().Lookup("bestSsb"))
	viper.BindPFlag("nrrg.advanced.pdcchSlotSib1", advancedCmd.Flags().Lookup("pdcchSlotSib1"))
	viper.BindPFlag("nrrg.advanced.prachOccMsg1", advancedCmd.Flags().Lookup("prachOccMsg1"))
	viper.BindPFlag("nrrg.advanced.pdcchOccMsg2", advancedCmd.Flags().Lookup("pdcchOccMsg2"))
	viper.BindPFlag("nrrg.advanced.pdcchOccMsg4", advancedCmd.Flags().Lookup("pdcchOccMsg4"))
	//viper.BindPFlag("nrrg.advanced.dsrRes", advancedCmd.Flags().Lookup("dsrRes"))
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
	flags.gridsetting.ssbPeriod = viper.GetString("nrrg.gridsetting.ssbPeriod")
	flags.gridsetting._maxLBar = viper.GetInt("nrrg.gridsetting._maxLBar")
	flags.gridsetting._maxL = viper.GetInt("nrrg.gridsetting._maxL")
	flags.gridsetting.candSsbIndex = viper.GetIntSlice("nrrg.gridsetting.candSsbIndex")

	flags.gridsetting._carrierScs = viper.GetString("nrrg.gridsetting._carrierScs")
	flags.gridsetting.dlArfcn = viper.GetInt("nrrg.gridsetting.dlArfcn")
	flags.gridsetting.bw = viper.GetString("nrrg.gridsetting.bw")
	flags.gridsetting._carrierNumRbs = viper.GetInt("nrrg.gridsetting._carrierNumRbs")
	flags.gridsetting._offsetToCarrier = viper.GetInt("nrrg.gridsetting._offsetToCarrier")

	flags.gridsetting.pci = viper.GetInt("nrrg.gridsetting.pci")

	flags.gridsetting._mibCommonScs = viper.GetString("nrrg.gridsetting._mibCommonScs")
	flags.gridsetting.rmsiCoreset0 = viper.GetInt("nrrg.gridsetting.rmsiCoreset0")
	flags.gridsetting._coreset0MultiplexingPat = viper.GetInt("nrrg.gridsetting._coreset0MultiplexingPat")
	flags.gridsetting._coreset0NumRbs = viper.GetInt("nrrg.gridsetting._coreset0NumRbs")
	flags.gridsetting._coreset0NumSymbs = viper.GetInt("nrrg.gridsetting._coreset0NumSymbs")
	flags.gridsetting._coreset0OffsetList = viper.GetIntSlice("nrrg.gridsetting._coreset0OffsetList")
	flags.gridsetting._coreset0Offset = viper.GetInt("nrrg.gridsetting._coreset0Offset")
	flags.gridsetting.rmsiCss0 = viper.GetInt("nrrg.gridsetting.rmsiCss0")
	flags.gridsetting._css0AggLevel = viper.GetInt("nrrg.gridsetting._css0AggLevel")
	flags.gridsetting._css0NumCandidates = viper.GetString("nrrg.gridsetting._css0NumCandidates")
	flags.gridsetting.dmrsTypeAPos = viper.GetString("nrrg.gridsetting.dmrsTypeAPos")
	flags.gridsetting._sfn = viper.GetInt("nrrg.gridsetting._sfn")
	flags.gridsetting._hrf = viper.GetInt("nrrg.gridsetting._hrf")

	// common settings
	flags.tdduldl._refScs = viper.GetString("nrrg.tdduldl._refScs")
	flags.tdduldl.patPeriod = viper.GetStringSlice("nrrg.tdduldl.patPeriod")
	flags.tdduldl.patNumDlSlots = viper.GetIntSlice("nrrg.tdduldl.patNumDlSlots")
	flags.tdduldl.patNumDlSymbs = viper.GetIntSlice("nrrg.tdduldl.patNumDlSymbs")
	flags.tdduldl.patNumUlSymbs = viper.GetIntSlice("nrrg.tdduldl.patNumUlSymbs")
	flags.tdduldl.patNumUlSlots = viper.GetIntSlice("nrrg.tdduldl.patNumUlSlots")

	flags.searchspace._coreset1FdRes = viper.GetString("nrrg.searchspace._coreset1FdRes")
	flags.searchspace.coreset1StartCrb = viper.GetInt("nrrg.searchspace.coreset1StartCrb")
	flags.searchspace.coreset1NumRbs = viper.GetInt("nrrg.searchspace.coreset1NumRbs")
	flags.searchspace._coreset1Duration = viper.GetInt("nrrg.searchspace._coreset1Duration")
	flags.searchspace.coreset1CceRegMappingType = viper.GetString("nrrg.searchspace.coreset1CceRegMappingType")
	flags.searchspace.coreset1RegBundleSize = viper.GetString("nrrg.searchspace.coreset1RegBundleSize")
	flags.searchspace.coreset1InterleaverSize = viper.GetString("nrrg.searchspace.coreset1InterleaverSize")
	flags.searchspace._coreset1ShiftIndex = viper.GetInt("nrrg.searchspace._coreset1ShiftIndex")
	flags.searchspace._ssId = viper.GetIntSlice("nrrg.searchspace._ssId")
	flags.searchspace._ssType = viper.GetStringSlice("nrrg.searchspace._ssType")
	flags.searchspace._ssCoresetId = viper.GetIntSlice("nrrg.searchspace._ssCoresetId")
	flags.searchspace._ssDuration = viper.GetIntSlice("nrrg.searchspace._ssDuration")
	flags.searchspace._ssMonitoringSymbolWithinSlot = viper.GetStringSlice("nrrg.searchspace._ssMonitoringSymbolWithinSlot")
	flags.searchspace.ssAggregationLevel = viper.GetStringSlice("nrrg.searchspace.ssAggregationLevel")
	flags.searchspace.ssNumOfPdcchCandidates = viper.GetStringSlice("nrrg.searchspace.ssNumOfPdcchCandidates")
	flags.searchspace._ssPeriodicity = viper.GetStringSlice("nrrg.searchspace._ssPeriodicity")
	flags.searchspace._ssSlotOffset = viper.GetIntSlice("nrrg.searchspace._ssSlotOffset")

	flags.dldci._tag = viper.GetStringSlice("nrrg.dldci._tag")
	flags.dldci._rnti = viper.GetStringSlice("nrrg.dldci._rnti")
	flags.dldci._muPdcch = viper.GetIntSlice("nrrg.dldci._muPdcch")
	flags.dldci._muPdsch = viper.GetIntSlice("nrrg.dldci._muPdsch")
	flags.dldci._indicatedBwp = viper.GetIntSlice("nrrg.dldci._indicatedBwp")
	flags.dldci.tdra = viper.GetIntSlice("nrrg.dldci.tdra")
	flags.dldci._tdMappingType = viper.GetStringSlice("nrrg.dldci._tdMappingType")
	flags.dldci._tdK0 = viper.GetIntSlice("nrrg.dldci._tdK0")
	flags.dldci._tdSliv = viper.GetIntSlice("nrrg.dldci._tdSliv")
	flags.dldci._tdStartSymb = viper.GetIntSlice("nrrg.dldci._tdStartSymb")
	flags.dldci._tdNumSymbs = viper.GetIntSlice("nrrg.dldci._tdNumSymbs")
	flags.dldci._fdRaType = viper.GetStringSlice("nrrg.dldci._fdRaType")
	flags.dldci._fdBitsRaType0 = viper.GetInt("nrrg.dldci._fdBitsRaType0")
	flags.dldci._fdBitsRaType1 = viper.GetIntSlice("nrrg.dldci._fdBitsRaType1")
	flags.dldci._fdRa = viper.GetStringSlice("nrrg.dldci._fdRa")
	flags.dldci.fdStartRb = viper.GetIntSlice("nrrg.dldci.fdStartRb")
	flags.dldci.fdNumRbs = viper.GetIntSlice("nrrg.dldci.fdNumRbs")
	flags.dldci.fdVrbPrbMappingType = viper.GetStringSlice("nrrg.dldci.fdVrbPrbMappingType")
	flags.dldci.fdBundleSize = viper.GetStringSlice("nrrg.dldci.fdBundleSize")
	flags.dldci.mcsCw0 = viper.GetIntSlice("nrrg.dldci.mcsCw0")
	flags.dldci._tbsCw0 = viper.GetIntSlice("nrrg.dldci._tbsCw0")
	flags.dldci.mcsCw1 = viper.GetInt("nrrg.dldci.mcsCw1")
	flags.dldci._tbsCw1 = viper.GetInt("nrrg.dldci._tbsCw1")
	flags.dldci.tbScalingFactor = viper.GetFloat64("nrrg.dldci.tbScalingFactor")
	flags.dldci.deltaPri = viper.GetInt("nrrg.dldci.deltaPri")
	flags.dldci.tdK1 = viper.GetInt("nrrg.dldci.tdK1")
	flags.dldci.antennaPorts = viper.GetInt("nrrg.dldci.antennaPorts")

	flags.uldci._tag = viper.GetStringSlice("nrrg.uldci._tag")
	flags.uldci._rnti = viper.GetStringSlice("nrrg.uldci._rnti")
	flags.uldci._muPdcch = viper.GetIntSlice("nrrg.uldci._muPdcch")
	flags.uldci._muPusch = viper.GetIntSlice("nrrg.uldci._muPusch")
	flags.uldci._indicatedBwp = viper.GetIntSlice("nrrg.uldci._indicatedBwp")
	flags.uldci.tdra = viper.GetIntSlice("nrrg.uldci.tdra")
	flags.uldci._tdMappingType = viper.GetStringSlice("nrrg.uldci._tdMappingType")
	flags.uldci._tdK2 = viper.GetIntSlice("nrrg.uldci._tdK2")
	flags.uldci._tdSliv = viper.GetIntSlice("nrrg.uldci._tdSliv")
	flags.uldci._tdStartSymb = viper.GetIntSlice("nrrg.uldci._tdStartSymb")
	flags.uldci._tdNumSymbs = viper.GetIntSlice("nrrg.uldci._tdNumSymbs")
	flags.uldci._fdRaType = viper.GetStringSlice("nrrg.uldci._fdRaType")
	flags.uldci.fdFreqHop = viper.GetStringSlice("nrrg.uldci.fdFreqHop")
	flags.uldci._fdFreqHopOffset = viper.GetIntSlice("nrrg.uldci._fdFreqHopOffset")
	flags.uldci._fdBitsRaType0 = viper.GetInt("nrrg.uldci._fdBitsRaType0")
	flags.uldci._fdBitsRaType1 = viper.GetIntSlice("nrrg.uldci._fdBitsRaType1")
	flags.uldci._fdRa = viper.GetStringSlice("nrrg.uldci._fdRa")
	flags.uldci.fdStartRb = viper.GetIntSlice("nrrg.uldci.fdStartRb")
	flags.uldci.fdNumRbs = viper.GetIntSlice("nrrg.uldci.fdNumRbs")
	flags.uldci.mcsCw0 = viper.GetIntSlice("nrrg.uldci.mcsCw0")
	flags.uldci._tbs = viper.GetIntSlice("nrrg.uldci._tbs")
	flags.uldci.precodingInfoNumLayers = viper.GetInt("nrrg.uldci.precodingInfoNumLayers")
	flags.uldci.srsResIndicator = viper.GetInt("nrrg.uldci.srsResIndicator")
	flags.uldci.antennaPorts = viper.GetInt("nrrg.uldci.antennaPorts")
	flags.uldci.ptrsDmrsAssociation = viper.GetInt("nrrg.uldci.ptrsDmrsAssociation")

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
	flags.rach._msg1Scs = viper.GetString("nrrg.rach._msg1Scs")
	flags.rach.msg1Fdm = viper.GetInt("nrrg.rach.msg1Fdm")
	flags.rach.msg1FreqStart = viper.GetInt("nrrg.rach.msg1FreqStart")
	flags.rach.totNumPreambs = viper.GetInt("nrrg.rach.totNumPreambs")
	flags.rach.ssbPerRachOccasion = viper.GetString("nrrg.rach.ssbPerRachOccasion")
	flags.rach.cbPreambsPerSsb = viper.GetInt("nrrg.rach.cbPreambsPerSsb")
	flags.rach.raRespWin = viper.GetString("nrrg.rach.raRespWin")
	flags.rach.msg3Tp = viper.GetString("nrrg.rach.msg3Tp")
	flags.rach.contResTimer = viper.GetString("nrrg.rach.contResTimer")
	flags.rach._raLen = viper.GetInt("nrrg.rach._raLen")
	flags.rach._raNumRbs = viper.GetInt("nrrg.rach._raNumRbs")
	flags.rach._raKBar = viper.GetInt("nrrg.rach._raKBar")

	flags.dmrsCommon._tag = viper.GetStringSlice("nrrg.dmrscommon._tag")
	flags.dmrsCommon._dmrsType = viper.GetStringSlice("nrrg.dmrscommon._dmrsType")
	flags.dmrsCommon._dmrsAddPos = viper.GetStringSlice("nrrg.dmrscommon._dmrsAddPos")
	flags.dmrsCommon._maxLength = viper.GetStringSlice("nrrg.dmrscommon._maxLength")
	flags.dmrsCommon._dmrsPorts = viper.GetIntSlice("nrrg.dmrscommon._dmrsPorts")
	flags.dmrsCommon._cdmGroupsWoData = viper.GetIntSlice("nrrg.dmrscommon._cdmGroupsWoData")
	flags.dmrsCommon._numFrontLoadSymbs = viper.GetIntSlice("nrrg.dmrscommon._numFrontLoadSymbs")

	flags.pdsch._pdschAggFactor = viper.GetString("nrrg.pdsch._pdschAggFactor")
	flags.pdsch.pdschRbgCfg = viper.GetString("nrrg.pdsch.pdschRbgCfg")
	flags.pdsch._rbgSize = viper.GetInt("nrrg.pdsch._rbgSize")
	flags.pdsch.pdschMcsTable = viper.GetString("nrrg.pdsch.pdschMcsTable")
	flags.pdsch.pdschXOh = viper.GetString("nrrg.pdsch.pdschXOh")
	flags.pdsch.pdschMaxLayers = viper.GetInt("nrrg.pdsch.pdschMaxLayers")

	flags.pdsch.pdschDmrsType = viper.GetString("nrrg.pdsch.pdschDmrsType")
	flags.pdsch.pdschDmrsAddPos = viper.GetString("nrrg.pdsch.pdschDmrsAddPos")
	flags.pdsch.pdschMaxLength = viper.GetString("nrrg.pdsch.pdschMaxLength")
	flags.pdsch._dmrsPorts = viper.GetIntSlice("nrrg.pdsch._dmrsPorts")
	flags.pdsch._cdmGroupsWoData = viper.GetInt("nrrg.pdsch._cdmGroupsWoData")
	flags.pdsch._numFrontLoadSymbs = viper.GetInt("nrrg.pdsch._numFrontLoadSymbs")

	flags.pdsch.pdschPtrsEnabled = viper.GetBool("nrrg.pdsch.pdschPtrsEnabled")
	flags.pdsch.pdschPtrsTimeDensity = viper.GetInt("nrrg.pdsch.pdschPtrsTimeDensity")
	flags.pdsch.pdschPtrsFreqDensity = viper.GetInt("nrrg.pdsch.pdschPtrsFreqDensity")
	flags.pdsch.pdschPtrsReOffset = viper.GetString("nrrg.pdsch.pdschPtrsReOffset")
	flags.pdsch._ptrsDmrsPorts = viper.GetInt("nrrg.pdsch._ptrsDmrsPorts")

	flags.pusch.puschDmrsType = viper.GetString("nrrg.pusch.puschDmrsType")
	flags.pusch.puschDmrsAddPos = viper.GetString("nrrg.pusch.puschDmrsAddPos")
	flags.pusch.puschMaxLength = viper.GetString("nrrg.pusch.puschMaxLength")
	flags.pusch._dmrsPorts = viper.GetIntSlice("nrrg.pusch._dmrsPorts")
	flags.pusch._cdmGroupsWoData = viper.GetInt("nrrg.pusch._cdmGroupsWoData")
	flags.pusch._numFrontLoadSymbs = viper.GetInt("nrrg.pusch._numFrontLoadSymbs")

	flags.pusch.puschPtrsEnabled = viper.GetBool("nrrg.pusch.puschPtrsEnabled")
	flags.pusch.puschPtrsTimeDensity = viper.GetInt("nrrg.pusch.puschPtrsTimeDensity")
	flags.pusch.puschPtrsFreqDensity = viper.GetInt("nrrg.pusch.puschPtrsFreqDensity")
	flags.pusch.puschPtrsReOffset = viper.GetString("nrrg.pusch.puschPtrsReOffset")
	flags.pusch.puschPtrsMaxNumPorts = viper.GetString("nrrg.pusch.puschPtrsMaxNumPorts")
	flags.pusch.puschPtrsTimeDensityTp = viper.GetInt("nrrg.pusch.puschPtrsTimeDensityTp")
	flags.pusch.puschPtrsGrpPatternTp = viper.GetString("nrrg.pusch.puschPtrsGrpPatternTp")
	flags.pusch._numGrpsTp = viper.GetInt("nrrg.pusch._numGrpsTp")
	flags.pusch._samplesPerGrpTp = viper.GetInt("nrrg.pusch._samplesPerGrpTp")
	//flags.pusch._ptrsDmrsPortsTp = viper.GetInt("nrrg.pusch._ptrsDmrsPortsTp")
	flags.pusch._ptrsDmrsPorts = viper.GetIntSlice("nrrg.pusch._ptrsDmrsPorts")

	flags.pusch.puschTxCfg = viper.GetString("nrrg.pusch.puschTxCfg")
	flags.pusch.puschCbSubset = viper.GetString("nrrg.pusch.puschCbSubset")
	flags.pusch.puschCbMaxRankNonCbMaxLayers = viper.GetInt("nrrg.pusch.puschCbMaxRankNonCbMaxLayers")
	flags.pusch.puschTp = viper.GetString("nrrg.pusch.puschTp")
	flags.pusch._puschAggFactor = viper.GetString("nrrg.pusch._puschAggFactor")
	flags.pusch.puschRbgCfg = viper.GetString("nrrg.pusch.puschRbgCfg")
	flags.pusch._rbgSize = viper.GetInt("nrrg.pusch._rbgSize")
	flags.pusch.puschMcsTable = viper.GetString("nrrg.pusch.puschMcsTable")
	flags.pusch.puschXOh = viper.GetString("nrrg.pusch.puschXOh")
	flags.pusch._puschRepType = viper.GetString("nrrg.pusch._puschRepType")

	flags.csi._resSetId = viper.GetIntSlice("nrrg.csi._resSetId")
	flags.csi._trsInfo = viper.GetStringSlice("nrrg.csi._trsInfo")
	flags.csi._resId = viper.GetIntSlice("nrrg.csi._resId")
	flags.csi.freqAllocRow = viper.GetStringSlice("nrrg.csi.freqAllocRow")
	flags.csi.freqAllocBits = viper.GetStringSlice("nrrg.csi.freqAllocBits")
	flags.csi._numPorts = viper.GetStringSlice("nrrg.csi._numPorts")
	flags.csi._cdmType = viper.GetStringSlice("nrrg.csi._cdmType")
	flags.csi._density = viper.GetStringSlice("nrrg.csi._density")
	flags.csi._firstSymb = viper.GetIntSlice("nrrg.csi._firstSymb")
	//flags.csi._firstSymb2 = viper.GetInt("nrrg.csi._firstSymb2")
	flags.csi._startRb = viper.GetIntSlice("nrrg.csi._startRb")
	flags.csi._numRbs = viper.GetIntSlice("nrrg.csi._numRbs")
	flags.csi.period = viper.GetStringSlice("nrrg.csi.period")
	flags.csi.offset = viper.GetIntSlice("nrrg.csi.offset")

	flags.csi._csiImRePattern = viper.GetString("nrrg.csi._csiImRePattern")
	flags.csi._csiImScLoc = viper.GetString("nrrg.csi._csiImScLoc")
	flags.csi._csiImSymbLoc = viper.GetInt("nrrg.csi._csiImSymbLoc")
	flags.csi._csiImStartRb = viper.GetInt("nrrg.csi._csiImStartRb")
	flags.csi._csiImNumRbs = viper.GetInt("nrrg.csi._csiImNumRbs")
	flags.csi._csiImPeriod = viper.GetString("nrrg.csi._csiImPeriod")
	flags.csi._csiImOffset = viper.GetInt("nrrg.csi._csiImOffset")

	flags.csi._resType = viper.GetString("nrrg.csi._resType")
	flags.csi._repCfgType = viper.GetString("nrrg.csi._repCfgType")
	flags.csi.csiRepPeriod = viper.GetString("nrrg.csi.csiRepPeriod")
	flags.csi.csiRepOffset = viper.GetInt("nrrg.csi.csiRepOffset")
	flags.csi._csiRepPucchRes = viper.GetInt("nrrg.csi._csiRepPucchRes")
	flags.csi._quantity = viper.GetString("nrrg.csi._quantity")

	flags.srs._resId = viper.GetIntSlice("nrrg.srs._resId")
	flags.srs.srsNumPorts = viper.GetStringSlice("nrrg.srs.srsNumPorts")
	flags.srs._srsNonCbPtrsPort = viper.GetStringSlice("nrrg.srs._srsNonCbPtrsPort")
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
	flags.srs._resType = viper.GetStringSlice("nrrg.srs._resType")
	flags.srs.srsPeriod = viper.GetStringSlice("nrrg.srs.srsPeriod")
	flags.srs.srsOffset = viper.GetIntSlice("nrrg.srs.srsOffset")
	flags.srs._mSRSb = viper.GetStringSlice("nrrg.srs._mSRSb")
	flags.srs._Nb = viper.GetStringSlice("nrrg.srs._Nb")
	flags.srs._resSetId = viper.GetIntSlice("nrrg.srs._resSetId")
	flags.srs.srsSetResIdList = viper.GetStringSlice("nrrg.srs.srsSetResIdList")
	flags.srs._resSetType = viper.GetStringSlice("nrrg.srs._resSetType")
	flags.srs._usage = viper.GetStringSlice("nrrg.srs._usage")

	flags.pucch._numSlots = viper.GetString("nrrg.pucch._numSlots")
	flags.pucch._interSlotFreqHop = viper.GetString("nrrg.pucch._interSlotFreqHop")
	flags.pucch._addDmrs = viper.GetBool("nrrg.pucch._addDmrs")
	flags.pucch._simHarqAckCsi = viper.GetBool("nrrg.pucch._simHarqAckCsi")
	flags.pucch._pucchResId = viper.GetIntSlice("nrrg.pucch._pucchResId")
	flags.pucch._pucchFormat = viper.GetStringSlice("nrrg.pucch._pucchFormat")
	//flags.pucch._pucchResSetId = viper.GetIntSlice("nrrg.pucch._pucchResSetId")
	flags.pucch._pucchStartRb = viper.GetIntSlice("nrrg.pucch._pucchStartRb")
	flags.pucch._pucchIntraSlotFreqHop = viper.GetStringSlice("nrrg.pucch._pucchIntraSlotFreqHop")
	flags.pucch._pucchSecondHopPrb = viper.GetIntSlice("nrrg.pucch._pucchSecondHopPrb")
	flags.pucch._pucchNumRbs = viper.GetIntSlice("nrrg.pucch._pucchNumRbs")
	flags.pucch._pucchStartSymb = viper.GetIntSlice("nrrg.pucch._pucchStartSymb")
	flags.pucch._pucchNumSymbs = viper.GetIntSlice("nrrg.pucch._pucchNumSymbs")
	//flags.pucch._dsrResId = viper.GetIntSlice("nrrg.pucch._dsrResId")
	flags.pucch.dsrPeriod = viper.GetString("nrrg.pucch.dsrPeriod")
	flags.pucch.dsrOffset = viper.GetInt("nrrg.pucch.dsrOffset")
	flags.pucch._dsrPucchRes = viper.GetInt("nrrg.pucch._dsrPucchRes")

	flags.advanced.bestSsb = viper.GetInt("nrrg.advanced.bestSsb")
	flags.advanced.pdcchSlotSib1 = viper.GetInt("nrrg.advanced.pdcchSlotSib1")
	flags.advanced.prachOccMsg1 = viper.GetInt("nrrg.advanced.prachOccMsg1")
	flags.advanced.pdcchOccMsg2 = viper.GetInt("nrrg.advanced.pdcchOccMsg2")
	flags.advanced.pdcchOccMsg4 = viper.GetInt("nrrg.advanced.pdcchOccMsg4")
	//flags.advanced.dsrRes = viper.GetInt("nrrg.advanced.dsrRes")
}

var w = []int{len("Flag"), len("Type"), len("Current Value"), len("Default Value")}

// var w =[]int{len("Flag"), len("Type"), len("Current Value")}

/*
laPrint performs left-aligned printing.
*/
func laPrint(cmd *cobra.Command, args []string) {
	regGreen.Printf("[INFO]: List of [%v] parameters\n", cmd.Name())
	cmd.Flags().VisitAll(
		func(f *pflag.Flag) {
			if f.Name != "config" && f.Name != "help" {
				if len(f.Name) > w[0] {
					w[0] = len(f.Name)
				}
				if len(f.Value.Type()) > w[1] {
					w[1] = len(f.Value.Type())
				}
				if len(f.Value.String()) > w[2] {
					w[2] = len(f.Value.String())
				}
				if len(f.DefValue) > w[3] {
					w[3] = len(f.DefValue)
				}
			}
		})

	for i := 0; i < len(w); i++ {
		w[i] += 4
	}

	fmt.Printf("%-*v%-*v%-*v%-*v%v\n", w[0], "Flag", w[1], "Type", w[2], "Current Value", w[3], "Default Value", "Modifiable")
	// fmt.Printf("%-*v%-*v%-*v%v\n", w[0], "Flag", w[1], "Type", w[2], "Current Value", "Modifiable")
	cmd.Flags().VisitAll(
		func(f *pflag.Flag) {
			if f.Name != "config" && f.Name != "help" {
				if f.Hidden {
					fmt.Printf("%-*v%-*v%-*v%-*v%v\n", w[0], f.Name, w[1], f.Value.Type(), w[2], f.Value, w[3], f.DefValue, !f.Hidden)
				} else {
					regMagenta.Printf("%-*v%-*v%-*v%-*v%v\n", w[0], f.Name, w[1], f.Value.Type(), w[2], f.Value, w[3], f.DefValue, !f.Hidden)
				}
				// fmt.Printf("%-*v%-*v%-*v%v\n", w[0], f.Name, w[1], f.Value.Type(), w[2], f.Value, !f.Hidden)
			}
		})

	fmt.Println()
}

func initLrbsPuschTp(bwpSize int) []int {
	var lrbsTp []int
	for x := range utils.PyRange(0, utils.CeilInt(math.Log(float64(bwpSize))/math.Log(2)), 1) {
		for y := range utils.PyRange(0, utils.CeilInt(math.Log(float64(bwpSize))/math.Log(3)), 1) {
			for z := range utils.PyRange(0, utils.CeilInt(math.Log(float64(bwpSize))/math.Log(5)), 1) {
				lrbs := int(math.Pow(2, float64(x)) * math.Pow(3, float64(y)) * math.Pow(5, float64(z)))
				if lrbs <= bwpSize {
					lrbsTp = append(lrbsTp, lrbs)
				}
			}
		}
	}

	sort.Ints(lrbsTp)

	return lrbsTp
}
