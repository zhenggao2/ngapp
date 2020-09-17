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
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// operating band
	opBand string

	// ssb grid
	ssbScs string
	kSsb int

	// ssb burst
	inOneGrp string
	grpPresence string
	ssbPeriod string

	// mib
	sfn int
	hrf int
	dmrsTypeAPos string
	commonScs string
	rmsiCoreset0 int
	rmsiCss0 int

	// carrier grid
	carrierScs string
	bw string

	// common setting
	pci int
	numUeAp string
	// common setting - tdd ul/dl config common
	patPeriod []string
	patNumDlSlots []int
	patNumDlSymbs []int
	patNumUlSymbs []int
	patNumUlSlots []int

	// CSS0
	css0AggLevel int
	css0NumCandidates string

	// CORESET1
	coreset1FreqRes string
	// TODO: rename coreset1NumSymbs to coreset1Duration
	coreset1NumSymbs int
	coreset1CceRegMap string
	coreset1RegBundleSize string
	coreset1InterleaverSize string
	coreset1ShiftInd int
	// coreset1PrecoderGranularity string

	// USS
	ussPeriod string
	ussOffset int
	ussDuration int
	ussFirstSymbs string
	ussAggLevel int
	ussNumCandidates string

	// DCI 1_0 scheduling Sib1/Msg2/Msg4 with SI-RNTI/RA-RNTI/TC-RNTI
	dci10TdRa []int
	dci10FdStartRb []int
	dci10FdNumRbs []int
	dci10FdVrbPrbMappingType []string
	dci10McsCw0 []int
	dci10Msg2TbScaling int
	dci10Msg4DeltaPri int
	dci10Msg4TdK1 int

	// DCI 1_1 scheduling PDSCH with C-RNTI
	dci11TdRa int
	dci11TdMappingType string
	dci11TdK0 int
	dci11TdSliv int
	dci11TdStartSymb int
	dci11TdNumSymbs int
	dci11FdRaType string
	dci11FdRa string
	dci11FdStartRb int
	dci11FdNumRbs int
	dci11FdVrbPrbMappingType string
	dci11FdBundleSize string
	dci11McsCw0 int
	dci11McsCw1 int
	dci11DeltaPri int
	dci11TdK1 int
	dci11AntPorts int

	// Msg3 PUSCH scheduled by UL grant in RAR(Msg2)
	msg3TdRa int
	msg3FdFreqHop string
	msg3FdRa string
	msg3FdStartRb int
	msg3FdNumRbs int
	msg3McsCw0 int

	// DCI 0_1 scheduling PUSCH with C-RNTI
	dci01TdRa int
	dci01TdMappingType string
	dci01TdK2 int
	dci01TdSliv int
	dci01TdStartSymb int
	dci01TdNumSymbs int
	dci01FdRaType string
	dci01FdFreqHop string
	dci01FdRa string
	dci01FdStartRb int
	dci01FdNumRbs int
	dci01McsCw0 int
	dci01CbTpmiNumLayers int
	dci01Sri string
	dci01AntPorts int
	dci01PtrsDmrsMap int

	// initial/dedicated UL/DL BWP
	bwpLocAndBw []int
	bwpStartRb []int
	bwpNumRbs []int

	// random access
	prachConfId int
	msg1Scs string
	msg1Fdm int
	msg1FreqStart int
	raRespWin string
	totNumPreambs int
	ssbPerRachOccasion string
	cbPreambsPerSsb int
	contResTimer string
	msg3Tp string

	// DMRS for PDSCH
	pdschDmrsType string
	pdschDmrsAddPos string
	pdschMaxLength string

	// PTRS for PDSCH
	pdschPtrsEnabled bool
	pdschPtrsTimeDensity int
	pdschPtrsFreqDensity int
	pdschPtrsReOffset string

	// DMRS for PUSCH
	puschDmrsType string
	puschDmrsAddPos string
	puschMaxLength string

	// PTRS for PUSCH
	puschPtrsEnabled bool
	puschPtrsTimeDensity int
	puschPtrsFreqDensity int
	puschPtrsReOffset string
	puschPtrsMaxNumPorts string
	puschPtrsTimeDensityTp int
	puschPtrsGrpPatternTp string

	// PDSCH-config and PDSCH-ServingCellConfig
	pdschAggFactor string
	pdschRbgCfg string
	pdschMcsTable string
	pdschXOh string

	// PUSCH-config and PUSCH-ServingCellConfig
	puschTxCfg string
	puschCbSubset string
	puschCbMaxRankNonCbMaxLayers int
	puschFreqHopOffset int
	puschTp string
	puschAggFactor string
	puschRbgCfg string
	puschMcsTable string
	puschXOh string

	// NZP-CSI-RS resource
	nzpCsiRsFreqAllocRow string
	nzpCsiRsFreqAllocBits string
	nzpCsiRsNumPorts string
	nzpCsiRsCdmType string
	nzpCsiRsDensity string
	nzpCsiRsFirstSymb int
	nzpCsiRsFirstSymb2 int
	nzpCsiRsStartRb int
	nzpCsiRsNumRbs int
	nzpCsiRsPeriod string
	nzpCsiRsOffset int

	// TRS resource

	// CSI-IM resource

	// CSI-ResourceConfig and CSI-ReportConfig

	// SRS resource

	// PUCCH resource

	// PUCCH-FormatConfig

	// DSR resource
)

// nrrgCmd represents the nrrg command
var nrrgCmd = &cobra.Command{
	Use:   "nrrg",
	Short: "",
	Long: `nrrg generates NR(new radio) resource grid according to network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.WriteConfig()
	},
}

// nrrgConfCmd represents the 'nrrg conf' command
var nrrgConfCmd = &cobra.Command{
	Use:   "conf",
	Short: "",
	Long: `'nrrg conf' can be used to get/set network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg conf called")
		viper.WriteConfig()
	},
}

// confFreqBandCmd represents the 'nrrg conf freqband' command
var confFreqBandCmd = &cobra.Command{
	Use:   "freqband",
	Short: "",
	Long: `'nrrg conf freqband' can be used to get/set frequency-band related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confSsbGridCmd represents the 'nrrg conf ssbgrid' command
var confSsbGridCmd = &cobra.Command{
	Use:   "ssbgrid",
	Short: "",
	Long: `'nrrg conf ssbgrid' can be used to get/set SSB-grid related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confSsbBurstCmd represents the 'nrrg conf ssbburst' command
var confSsbBurstCmd = &cobra.Command{
	Use:   "ssbburst",
	Short: "",
	Long: `'nrrg conf ssbburst' can be used to get/set SSB-burst related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confMibCmd represents the 'nrrg conf mib' command
var confMibCmd = &cobra.Command{
	Use:   "mib",
	Short: "",
	Long: `'nrrg conf mib' can be used to get/set MIB related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confCarrierGridCmd represents the 'nrrg conf carriergrid' command
var confCarrierGridCmd = &cobra.Command{
	Use:   "carriergrid",
	Short: "",
	Long: `'nrrg conf carriergrid' can be used to get/set carrier-grid related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confCommonSettingCmd represents the 'nrrg conf commonsetting' command
var confCommonSettingCmd = &cobra.Command{
	Use:   "commonsetting",
	Short: "",
	Long: `'nrrg conf commonsetting' can be used to get/set common-setting related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confCss0Cmd represents the 'nrrg conf css0' command
var confCss0Cmd = &cobra.Command{
	Use:   "css0",
	Short: "",
	Long: `'nrrg conf css0' can be used to get/set Common search space(CSS0) related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confCoreset1Cmd represents the 'nrrg conf coreset1' command
var confCoreset1Cmd = &cobra.Command{
	Use:   "coreset1",
	Short: "",
	Long: `'nrrg conf coreset1' can be used to get/set CORESET1 related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confUssCmd represents the 'nrrg conf uss' command
var confUssCmd = &cobra.Command{
	Use:   "uss",
	Short: "",
	Long: `'nrrg conf uss' can be used to get/set UE-specific search space related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confDci10Cmd represents the 'nrrg conf dci10' command
var confDci10Cmd = &cobra.Command{
	Use:   "dci10",
	Short: "",
	Long: `'nrrg conf dci10' can be used to get/set DCI 1_0 (scheduling SIB1/Msg2/Msg4) related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confDci11Cmd represents the 'nrrg conf dci11' command
var confDci11Cmd = &cobra.Command{
	Use:   "dci11",
	Short: "",
	Long: `'nrrg conf dci11' can be used to get/set DCI 1_1(scheduling PDSCH with C-RNTI) related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confMsg3Cmd represents the 'nrrg conf msg3' command
var confMsg3Cmd = &cobra.Command{
	Use:   "msg3",
	Short: "",
	Long: `'nrrg conf msg3' can be used to get/set Msg3(scheduled by UL grant in RAR) related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confDci01Cmd represents the 'nrrg conf dci01' command
var confDci01Cmd = &cobra.Command{
	Use:   "dci01",
	Short: "",
	Long: `'nrrg conf dci01' can be used to get/set DCI 0_1(scheduling PUSCH with C-RNTI) related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confBwpCmd represents the 'nrrg conf bwp' command
var confBwpCmd = &cobra.Command{
	Use:   "bwp",
	Short: "",
	Long: `'nrrg conf bwp' can be used to get/set generic BWP related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confRachCmd represents the 'nrrg conf rach' command
var confRachCmd = &cobra.Command{
	Use:   "rach",
	Short: "",
	Long: `'nrrg conf rach' can be used to get/set random access related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confDmrsCommonCmd represents the 'nrrg conf dmrscommon' command
var confDmrsCommonCmd = &cobra.Command{
	Use:   "dmrscommon",
	Short: "",
	Long: `'nrrg conf dmrscommon' can be used to get/set DMRS of SIB1/Msg2/Msg4/Msg3 related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confDmrsPdschCmd represents the 'nrrg conf dmrspdsch' command
var confDmrsPdschCmd = &cobra.Command{
	Use:   "dmrspdsch",
	Short: "",
	Long: `'nrrg conf dmrspdsch' can be used to get/set DMRS of PDSCH related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confPtrsPdschCmd represents the 'nrrg conf ptrspdsch' command
var confPtrsPdschCmd = &cobra.Command{
	Use:   "ptrspdsch",
	Short: "",
	Long: `'nrrg conf ptrspdsch' can be used to get/set PTRS of PDSCH related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confDmrsPuschCmd represents the 'nrrg conf dmrspusch' command
var confDmrsPuschCmd = &cobra.Command{
	Use:   "dmrspusch",
	Short: "",
	Long: `'nrrg conf dmrspusch' can be used to get/set DMRS of PUSCH related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confPtrsPuschCmd represents the 'nrrg conf ptrspusch' command
var confPtrsPuschCmd = &cobra.Command{
	Use:   "ptrspusch",
	Short: "",
	Long: `'nrrg conf ptrspusch' can be used to get/set PTRS of PUSCH related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confPdschCmd represents the 'nrrg conf pdsch' command
var confPdschCmd = &cobra.Command{
	Use:   "pdsch",
	Short: "",
	Long: `'nrrg conf pdsch' can be used to get/set PDSCH-config or PDSCH-ServingCellConfig related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confPuschCmd represents the 'nrrg conf pusch' command
var confPuschCmd = &cobra.Command{
	Use:   "pusch",
	Short: "",
	Long: `'nrrg conf pusch' can be used to get/set PUSCH-config or PUSCH-ServingCellConfig related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confNzpCsiRsCmd represents the 'nrrg conf nzpcsirs' command
var confNzpCsiRsCmd = &cobra.Command{
	Use:   "nzpcsirs",
	Short: "",
	Long: `'nrrg conf nzpcsirs' can be used to get/set NZP-CSI-RS resource related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}


// TODO: add more subcmd here!!!

// nrrgSimCmd represents the 'nrrg sim' command
var nrrgSimCmd = &cobra.Command{
	Use:   "sim",
	Short: "",
	Long: `'nrrg sim' can be used to perform NR-Uu simulation.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg sim called")
		viper.WriteConfig()
	},
}

func init() {
	nrrgConfCmd.AddCommand(confFreqBandCmd)
	nrrgConfCmd.AddCommand(confSsbGridCmd)
	nrrgConfCmd.AddCommand(confSsbBurstCmd)
	nrrgConfCmd.AddCommand(confMibCmd)
	nrrgConfCmd.AddCommand(confCarrierGridCmd)
	nrrgConfCmd.AddCommand(confCommonSettingCmd)
	nrrgConfCmd.AddCommand(confCss0Cmd)
	nrrgConfCmd.AddCommand(confCoreset1Cmd)
	nrrgConfCmd.AddCommand(confUssCmd)
	nrrgConfCmd.AddCommand(confDci10Cmd)
	nrrgConfCmd.AddCommand(confDci11Cmd)
	nrrgConfCmd.AddCommand(confMsg3Cmd)
	nrrgConfCmd.AddCommand(confDci01Cmd)
	nrrgConfCmd.AddCommand(confBwpCmd)
	nrrgConfCmd.AddCommand(confRachCmd)
	nrrgConfCmd.AddCommand(confDmrsCommonCmd)
	nrrgConfCmd.AddCommand(confDmrsPdschCmd)
	nrrgConfCmd.AddCommand(confPtrsPdschCmd)
	nrrgConfCmd.AddCommand(confDmrsPuschCmd)
	nrrgConfCmd.AddCommand(confPtrsPuschCmd)
	nrrgConfCmd.AddCommand(confPdschCmd)
	nrrgConfCmd.AddCommand(confPuschCmd)
	nrrgConfCmd.AddCommand(confNzpCsiRsCmd)
	nrrgCmd.AddCommand(nrrgConfCmd)
	nrrgCmd.AddCommand(nrrgSimCmd)
	rootCmd.AddCommand(nrrgCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nrrgCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	confFreqBandCmd.Flags().StringVar(&opBand, "opBand", "n41", "Operating band")
	confFreqBandCmd.Flags().String("_duplexMode", "TDD", "Duplex mode")
	confFreqBandCmd.Flags().Int("_maxDlFreq", 2690, "Maximum DL frequency(MHz)")
	confFreqBandCmd.Flags().String("_freqRange", "FR1", "Frequency range(FR1/FR2)")
	confFreqBandCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.freqBand.opBand", confFreqBandCmd.Flags().Lookup("opBand"))
	viper.BindPFlag("nrrg.freqBand._duplexMode", confFreqBandCmd.Flags().Lookup("_duplexMode"))
	viper.BindPFlag("nrrg.freqBand._maxDlFreq", confFreqBandCmd.Flags().Lookup("_maxDlFreq"))
	viper.BindPFlag("nrrg.freqBand._freqRange", confFreqBandCmd.Flags().Lookup("_freqRange"))
	confFreqBandCmd.Flags().MarkHidden("_duplexMode")
	confFreqBandCmd.Flags().MarkHidden("_maxDlFreq")
	confFreqBandCmd.Flags().MarkHidden("_freqRange")

	confSsbGridCmd.Flags().StringVar(&ssbScs, "ssbScs",  "30KHz", "SSB subcarrier spacing")
	confSsbGridCmd.Flags().String("_ssbPattern", "Case C", "SSB pattern")
	confSsbGridCmd.Flags().IntVar(&kSsb, "kSsb", 0, "k_SSB[0..23]")
	confSsbGridCmd.Flags().Int("_nCrbSsb", 32, "n_CRB_SSB")
	confSsbGridCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.ssbGrid.ssbScs", confSsbGridCmd.Flags().Lookup("ssbScs"))
	viper.BindPFlag("nrrg.ssbGrid._ssbPattern", confSsbGridCmd.Flags().Lookup("_ssbPattern"))
	viper.BindPFlag("nrrg.ssbGrid.kSsb", confSsbGridCmd.Flags().Lookup("kSsb"))
	viper.BindPFlag("nrrg.ssbGrid._nCrbSsb", confSsbGridCmd.Flags().Lookup("_nCrbSsb"))
	confSsbGridCmd.Flags().MarkHidden("_ssbPattern")
	confSsbGridCmd.Flags().MarkHidden("_nCrbSsb")

	confSsbBurstCmd.Flags().Int("_maxL", 8, "max_L")
	confSsbBurstCmd.Flags().StringVar(&inOneGrp, "inOneGroup", "11111111", "inOneGroup of ssb-PositionsInBurst")
	confSsbBurstCmd.Flags().StringVar(&grpPresence, "groupPresence", "", "groupPresence of ssb-PositionsInBurst")
	confSsbBurstCmd.Flags().StringVar(&ssbPeriod, "ssbPeriod", "20ms", "ssb-PeriodicityServingCell[5ms,10ms,20ms,40ms,80ms,160ms]")
	confSsbBurstCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.ssbBurst._maxL", confSsbBurstCmd.Flags().Lookup("_maxL"))
	viper.BindPFlag("nrrg.ssbBurst.inOneGroup", confSsbBurstCmd.Flags().Lookup("inOneGroup"))
	viper.BindPFlag("nrrg.ssbBurst.groupPresence", confSsbBurstCmd.Flags().Lookup("groupPresence"))
	viper.BindPFlag("nrrg.ssbBurst.ssbPeriod", confSsbBurstCmd.Flags().Lookup("ssbPeriod"))
	confSsbBurstCmd.Flags().MarkHidden("_maxL")

	confMibCmd.Flags().IntVar(&sfn, "sfn", 0, "System frame number(SFN)[0..1023]")
	confMibCmd.Flags().IntVar(&hrf, "hrf", 0, "Half frame bit[0,1]")
	confMibCmd.Flags().StringVar(&dmrsTypeAPos, "dmrsTypeAPos", "pos2", "dmrs-TypeA-Position[pos2,pos3]")
	confMibCmd.Flags().StringVar(&commonScs, "commonScs", "30KHz", "subCarrierSpacingCommon")
	confMibCmd.Flags().IntVar(&rmsiCoreset0, "rmsiCoreset0", 12, "coresetZero of PDCCH-ConfigSIB1[0..15]")
	confMibCmd.Flags().IntVar(&rmsiCss0, "rmsiCss0", 0, "searchSpaceZero of PDCCH-ConfigSIB1[0..15]")
	confMibCmd.Flags().Int("_coreset0MultiplexingPat", 1, "Multiplexing pattern of CORESET0")
	confMibCmd.Flags().Int("_coreset0NumRbs", 48, "Number of PRBs of CORESET0")
	confMibCmd.Flags().Int("_coreset0NumSymbs", 1, "Number of OFDM symbols of CORESET0")
	confMibCmd.Flags().IntSlice("_coreset0OffsetList", []int{16}, "List of offset of CORESET0")
	confMibCmd.Flags().Int("_coreset0Offset", 16, "Offset of CORESET0")
	confMibCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.mib.sfn", confMibCmd.Flags().Lookup("sfn"))
	viper.BindPFlag("nrrg.mib.hrf", confMibCmd.Flags().Lookup("hrf"))
	viper.BindPFlag("nrrg.mib.dmrsTypeAPos", confMibCmd.Flags().Lookup("dmrsTypeAPos"))
	viper.BindPFlag("nrrg.mib.commonScs", confMibCmd.Flags().Lookup("commonScs"))
	viper.BindPFlag("nrrg.mib.rmsiCoreset0", confMibCmd.Flags().Lookup("rmsiCoreset0"))
	viper.BindPFlag("nrrg.mib.rmsiCss0", confMibCmd.Flags().Lookup("rmsiCss0"))
	viper.BindPFlag("nrrg.mib._coreset0MultiplexingPat", confMibCmd.Flags().Lookup("_coreset0MultiplexingPat"))
	viper.BindPFlag("nrrg.mib._coreset0NumRbs", confMibCmd.Flags().Lookup("_coreset0NumRbs"))
	viper.BindPFlag("nrrg.mib._coreset0NumSymbs", confMibCmd.Flags().Lookup("_coreset0NumSymbs"))
	viper.BindPFlag("nrrg.mib._coreset0OffsetList", confMibCmd.Flags().Lookup("_coreset0OffsetList"))
	viper.BindPFlag("nrrg.mib._coreset0Offset", confMibCmd.Flags().Lookup("_coreset0Offset"))
	confMibCmd.Flags().MarkHidden("_coreset0MultiplexingPat")
	confMibCmd.Flags().MarkHidden("_coreset0NumRbs")
	confMibCmd.Flags().MarkHidden("_coreset0NumSymbs")
	confMibCmd.Flags().MarkHidden("_coreset0OffsetList")
	confMibCmd.Flags().MarkHidden("_coreset0Offset")

	confCarrierGridCmd.Flags().StringVar(&carrierScs, "carrierScs", "30KHz", "subcarrierSpacing of SCS-SpecificCarrier")
	confCarrierGridCmd.Flags().StringVar(&bw, "bw", "100MHz", "Transmission bandwidth(MHz)")
	confCarrierGridCmd.Flags().Int("_carrierNumRbs", 273, "carrierBandwidth(N_RB) of SCS-SpecificCarrier")
	confCarrierGridCmd.Flags().Int("_offsetToCarrier", 0, "_offsetToCarrier of SCS-SpecificCarrier")
	confCarrierGridCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.carrierGrid.carrierScs", confCarrierGridCmd.Flags().Lookup("carrierScs"))
	viper.BindPFlag("nrrg.carrierGrid.bw", confCarrierGridCmd.Flags().Lookup("bw"))
	viper.BindPFlag("nrrg.carrierGrid._carrierNumRbs", confCarrierGridCmd.Flags().Lookup("_carrierNumRbs"))
	viper.BindPFlag("nrrg.carrierGrid._offsetToCarrier", confCarrierGridCmd.Flags().Lookup("_offsetToCarrier"))
	confCarrierGridCmd.Flags().MarkHidden("_carrierNumRbs")
	confCarrierGridCmd.Flags().MarkHidden("_offsetToCarrier")

	confCommonSettingCmd.Flags().IntVar(&pci, "pci", 0, "Physical cell identity[0..1007]")
	confCommonSettingCmd.Flags().StringVar(&numUeAp, "numUeAp", "2T", "Number of UE antennas[1T,2T,4T]")
	confCommonSettingCmd.Flags().String("_refScs", "30KHz", "referenceSubcarrierSpacing of TDD-UL-DL-ConfigCommon")
	confCommonSettingCmd.Flags().StringSliceVar(&patPeriod, "patPeriod",  []string{"5ms"}, "dl-UL-TransmissionPeriodicity of TDD-UL-DL-ConfigCommon[0.5ms,0.625ms,1ms,1.25ms,2ms,2.5ms,3ms,4ms,5ms,10ms]")
	confCommonSettingCmd.Flags().IntSliceVar(&patNumDlSlots, "patNumDlSlots",  []int{7}, "nrofDownlinkSlot of TDD-UL-DL-ConfigCommon[0..80]")
	confCommonSettingCmd.Flags().IntSliceVar(&patNumDlSymbs, "patNumDlSymbs",  []int{6}, "nrofDownlinkSymbols of TDD-UL-DL-ConfigCommon[0..13]")
	confCommonSettingCmd.Flags().IntSliceVar(&patNumUlSymbs, "patNumUlSymbs",  []int{4}, "nrofUplinkSymbols of TDD-UL-DL-ConfigCommon[0..13]")
	confCommonSettingCmd.Flags().IntSliceVar(&patNumUlSlots, "patNumUlSlots",  []int{2}, "nrofUplinkSlots of TDD-UL-DL-ConfigCommon[0..80]")
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

	confCss0Cmd.Flags().IntVar(&css0AggLevel, "css0AggLevel", 4, "CCE aggregation level of CSS0[4,8,16]")
	confCss0Cmd.Flags().StringVar(&css0NumCandidates, "css0NumCandidates", "n4", "Number of PDCCH candidates of CSS0[n1,n2,n4]")
	confCss0Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.css0.css0AggLevel", confCss0Cmd.Flags().Lookup("css0AggLevel"))
	viper.BindPFlag("nrrg.css0.css0NumCandidates", confCss0Cmd.Flags().Lookup("css0NumCandidates"))

	confCoreset1Cmd.Flags().StringVar(&coreset1FreqRes, "coreset1FreqRes", "111111111111111111111111111111111111111111111", "frequencyDomainResources of ControlResourceSet")
	confCoreset1Cmd.Flags().IntVar(&coreset1NumSymbs, "coreset1Duration", 1, "duration of ControlResourceSet[1..3]")
	confCoreset1Cmd.Flags().StringVar(&coreset1CceRegMap, "coreset1CceRegMap", "interleaved", "cce-REG-MappingType of ControlResourceSet[1..3]")
	confCoreset1Cmd.Flags().StringVar(&coreset1RegBundleSize, "coreset1RegBundleSize", "n2", "reg-BundleSize of ControlResourceSet[n2,n6]")
	confCoreset1Cmd.Flags().StringVar(&coreset1InterleaverSize, "coreset1InterleaverSize", "n2", "interleaverSize of ControlResourceSet[n2,n3,n6]")
	confCoreset1Cmd.Flags().IntVar(&coreset1ShiftInd, "coreset1ShiftInd", 0, "shiftIndex of ControlResourceSet[0..274]")
	confCoreset1Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.coreset1.coreset1FreqRes", confCoreset1Cmd.Flags().Lookup("coreset1FreqRes"))
	viper.BindPFlag("nrrg.coreset1.coreset1Duration", confCoreset1Cmd.Flags().Lookup("coreset1Duration"))
	viper.BindPFlag("nrrg.coreset1.coreset1CceRegMap", confCoreset1Cmd.Flags().Lookup("coreset1CceRegMap"))
	viper.BindPFlag("nrrg.coreset1.coreset1RegBundleSize", confCoreset1Cmd.Flags().Lookup("coreset1RegBundleSize"))
	viper.BindPFlag("nrrg.coreset1.coreset1InterleaverSize", confCoreset1Cmd.Flags().Lookup("coreset1InterleaverSize"))
	viper.BindPFlag("nrrg.coreset1.coreset1ShiftInd", confCoreset1Cmd.Flags().Lookup("coreset1ShiftInd"))

	confUssCmd.Flags().StringVar(&ussPeriod, "ussPeriod", "sl1", "monitoringSlotPeriodicity of SearchSpace[sl1,sl2,sl4,sl5,sl8,sl10,sl16,sl20,sl40,sl80,sl160,sl320,sl640,sl1280,sl2560]")
	confUssCmd.Flags().IntVar(&ussOffset, "ussOffset", 0, "monitoringSlotOffset of SearchSpace[0..ussPeriod-1]")
	confUssCmd.Flags().IntVar(&ussDuration, "ussDuration", 1, "duration of SearchSpace[1 or 2..ussPeriod-1]")
	confUssCmd.Flags().StringVar(&ussFirstSymbs, "ussFirstSymbs", "10101010101010", "monitoringSymbolsWithinSlot of SearchSpace")
	confUssCmd.Flags().IntVar(&ussAggLevel, "ussAggLevel", 4, "aggregationLevel of SearchSpace[1,2,4,8,16]")
	confUssCmd.Flags().StringVar(&ussNumCandidates, "ussNumCandidates", "n1", "nrofCandidates of SearchSpace[n1,n2,n3,n4,n5,n6,n8]")
	confUssCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.uss.ussPeriod", confUssCmd.Flags().Lookup("ussPeriod"))
	viper.BindPFlag("nrrg.uss.ussOffset", confUssCmd.Flags().Lookup("ussOffset"))
	viper.BindPFlag("nrrg.uss.ussDuration", confUssCmd.Flags().Lookup("ussDuration"))
	viper.BindPFlag("nrrg.uss.ussFirstSymbs", confUssCmd.Flags().Lookup("ussFirstSymbs"))
	viper.BindPFlag("nrrg.uss.ussAggLevel", confUssCmd.Flags().Lookup("ussAggLevel"))
	viper.BindPFlag("nrrg.uss.ussNumCandidates", confUssCmd.Flags().Lookup("ussNumCandidates"))

	confDci10Cmd.Flags().StringSlice("_rnti", []string{"SI-RNTI", "RA-RNTI", "TC-RNTI"}, "RNTI for DCI 1_0")
	confDci10Cmd.Flags().IntSlice("_muPdcch", []int{1, 1, 1}, "Subcarrier spacing of PDCCH[0..3]")
	confDci10Cmd.Flags().IntSlice("_muPdsch", []int{1, 1, 1}, "Subcarrier spacing of PDSCH[0..3]")
	confDci10Cmd.Flags().IntSliceVar(&dci10TdRa, "dci10TdRa", []int{10, 10, 10}, "Time-domain-resource-assignment field of DCI 1_0[0..15]")
	confDci10Cmd.Flags().StringSlice("_tdMappingType", []string{"typeB", "typeB", "typeB"}, "Mapping type for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSlice("_tdK0", []int{0, 0, 0}, "Slot offset K0 for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSlice("_tdSliv", []int{26, 26, 26}, "SLIV for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSlice("_tdStartSymb", []int{12, 12, 12}, "Starting symbol S for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSlice("_tdNumSymbs", []int{2, 2, 2}, "Number of OFDM symbols L for PDSCH time-domain allocation")
	confDci10Cmd.Flags().StringSlice("_fdRaType", []string{"raType1", "raType1", "raType1"}, "resourceAllocation for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().StringSlice("_fdRa", []string{"00001011111", "00001011111", "00001011111"}, "Frequency-domain-resource-assignment field of DCI 1_0")
	confDci10Cmd.Flags().IntSliceVar(&dci10FdStartRb, "dci10FdStartRb", []int{0, 0, 0}, "RB_start of RIV for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&dci10FdNumRbs, "dci10FdNumRbs", []int{48, 48, 48}, "L_RBs of RIV for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().StringSliceVar(&dci10FdVrbPrbMappingType, "dci10FdVrbPrbMappingType", []string{"interleaved", "interleaved", "interleaved"}, "VRB-to-PRB-mapping field of DCI 1_0")
	confDci10Cmd.Flags().StringSlice("_fdBundleSize", []string{"n2", "n2", "n2"}, "L(vrb-ToPRB-Interleaver) for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&dci10McsCw0, "dci10McsCw0", []int{2, 2, 2}, "Modulation-and-coding-scheme field of DCI 1_0[0..9]")
	confDci10Cmd.Flags().IntSlice("_tbs", []int{408, 408, 408}, "Transport block size(bits) for PDSCH")
	confDci10Cmd.Flags().IntVar(&dci10Msg2TbScaling, "dci10Msg2TbScaling", 0, "TB-scaling field of DCI 1_0 scheduling Msg2[0..2]")
	confDci10Cmd.Flags().IntVar(&dci10Msg4DeltaPri, "dci10Msg4DeltaPri", 1, "PUCCH-resource-indicator field of DCI 1_0 scheduling Msg4[0..7]")
	confDci10Cmd.Flags().IntVar(&dci10Msg4TdK1, "dci10Msg4TdK1", 6, "PDSCH-to-HARQ_feedback-timing-indicator(K1) field of DCI 1_0 scheduling Msg4[0..7]")
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
	confDci10Cmd.Flags().MarkHidden("_fdRa")
	confDci10Cmd.Flags().MarkHidden("_fdBundleSize")
	confDci10Cmd.Flags().MarkHidden("_tbs")

	confDci11Cmd.Flags().String("_rnti", "C-RNTI", "RNTI for DCI 1_1")
	confDci11Cmd.Flags().Int("_muPdcch", 1, "Subcarrier spacing of PDCCH[0..3]")
	confDci11Cmd.Flags().Int("_muPdsch", 1, "Subcarrier spacing of PDSCH[0..3]")
	confDci11Cmd.Flags().Int("_actBwp", 1, "Active DL bandwidth part of PDSCH[0..1]")
	confDci11Cmd.Flags().Int("_indicatedBwp", 1, "Bandwidth-part-indicator field of DCI 1_1[0..1]")
	confDci11Cmd.Flags().IntVar(&dci11TdRa, "dci11TdRa", 16, "Time-domain-resource-assignment field of DCI 1_1[0..15 or 16]")
	confDci11Cmd.Flags().StringVar(&dci11TdMappingType, "dci11TdMappingType", "typeA", "Mapping type for PDSCH time-domain allocation[typeA,typeB]")
	confDci11Cmd.Flags().IntVar(&dci11TdK0, "dci11TdK0", 0, "Slot offset K0 for PDSCH time-domain allocation")
	confDci11Cmd.Flags().IntVar(&dci11TdSliv, "dci11TdSliv", 27, "SLIV for PDSCH time-domain allocation")
	confDci11Cmd.Flags().IntVar(&dci11TdStartSymb, "dci11TdStartSymb", 0, "Starting symbol S for PDSCH time-domain allocation")
	confDci11Cmd.Flags().IntVar(&dci11TdNumSymbs, "dci11TdNumSymbs", 14, "Number of OFDM symbols L for PDSCH time-domain allocation")
	confDci11Cmd.Flags().StringVar(&dci11FdRaType, "dci11FdRaType", "raType1", "resourceAllocation for PDSCH frequency-domain allocation[raType0,raType1]")
	confDci11Cmd.Flags().StringVar(&dci11FdRa, "dci11FdRa", "0000001000100001", "Frequency-domain-resource-assignment field of DCI 1_1")
	confDci11Cmd.Flags().IntVar(&dci11FdStartRb, "dci11FdStartRb", 0, "RB_start of RIV for PDSCH frequency-domain allocation")
	confDci11Cmd.Flags().IntVar(&dci11FdNumRbs, "dci11FdNumRbs", 273, "L_RBs of RIV for PDSCH frequency-domain allocation")
	confDci11Cmd.Flags().StringVar(&dci11FdVrbPrbMappingType, "dci11FdVrbPrbMappingType", "interleaved", "VRB-to-PRB-mapping field of DCI 1_1[nonInterleaved,interleaved]")
	confDci11Cmd.Flags().StringVar(&dci11FdBundleSize, "dci11FdBundleSize", "n2", "L(vrb-ToPRB-Interleaver) for PDSCH frequency-domain allocation[n2,n4]")
	confDci11Cmd.Flags().IntVar(&dci11McsCw0, "dci11McsCw0", 27, "Modulation-and-coding-scheme-cw0 field of DCI 1_1[-1 or 0..28]")
	confDci11Cmd.Flags().IntVar(&dci11McsCw1, "dci11McsCw1", -1, "Modulation-and-coding-scheme-cw1 field of DCI 1_1[-1 or 0..28]")
	confDci11Cmd.Flags().Int("_tbs", 1277992, "Transport block size(bits) for PDSCH")
	confDci11Cmd.Flags().IntVar(&dci11DeltaPri, "dci11DeltaPri", 1, "PUCCH-resource-indicator field of DCI 1_1[0..4]")
	confDci11Cmd.Flags().IntVar(&dci11TdK1, "dci11TdK1", 2, "PDSCH-to-HARQ_feedback-timing-indicator(K1) field of DCI 1_1[0..7]")
	confDci11Cmd.Flags().IntVar(&dci11AntPorts, "dci11AntPorts", 10, "Antenna_port(s) field of DCI 1_1[0..15]")
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
	confDci11Cmd.Flags().MarkHidden("_tbs")

	confMsg3Cmd.Flags().Int("_muPusch", 1, "Subcarrier spacing of PUSCH[0..3]")
	confMsg3Cmd.Flags().IntVar(&msg3TdRa, "msg3TdRa", 6, "PUSCH-time-resource-allocation field of RAR UL grant scheduling Msg3[0..15]")
	confMsg3Cmd.Flags().String("_tdMappingType", "typeB", "Mapping type for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().Int("_tdK2", 1, "Slot offset K2 for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().Int("_tdDelta", 3, "Slot offset delta for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().Int("_tdSliv", 74, "SLIV for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().Int("_tdStartSymb", 4, "Starting symbol S for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().Int("_tdNumSymbs", 6, "Number of OFDM symbols L for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().String("_fdRaType", "raType1", "resourceAllocation for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().StringVar(&msg3FdFreqHop, "msg3FdFreqHop", "enabled", "Frequency-hopping-flag field of RAR UL grant scheduling Msg3[disabled,enabled]")
	confMsg3Cmd.Flags().StringVar(&msg3FdRa, "msg3FdRa", "0100000100001101", "PUSCH-frequency-resource-allocation field of RAR UL grant scheduling Msg3")
	confMsg3Cmd.Flags().IntVar(&msg3FdStartRb, "msg3FdStartRb", 0, "RB_start of RIV for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().IntVar(&msg3FdNumRbs, "msg3FdNumRbs", 62, "L_RBs of RIV for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().Int("_fdSecondHopFreqOff", 68, "Frequency offset of second hop for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().IntVar(&msg3McsCw0, "msg3McsCw0", 2, "MCS field of RAR UL grant scheduling Msg3[0..28]")
	confMsg3Cmd.Flags().Int("_tbs", 1544, "Transport block size(bits) for PUSCH")
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

	confDci01Cmd.Flags().String("_rnti", "C-RNTI", "RNTI for DCI 0_1")
	confDci01Cmd.Flags().Int("_muPdcch", 1, "Subcarrier spacing of PDCCH[0..3]")
	confDci01Cmd.Flags().Int("_muPusch", 1, "Subcarrier spacing of PUSCH[0..3]")
	confDci01Cmd.Flags().Int("_actBwp", 1, "Active UL bandwidth part of PUSCH[0..1]")
	confDci01Cmd.Flags().Int("_indicatedBwp", 1, "Bandwidth-part-indicator field of DCI 0_1[0..1]")
	confDci01Cmd.Flags().IntVar(&dci01TdRa, "dci01TdRa", 16, "Time-domain-resource-assignment field of DCI 0_1[0..15 or 16]")
	confDci01Cmd.Flags().StringVar(&dci01TdMappingType, "dci01TdMappingType", "typeA", "Mapping type for PUSCH time-domain allocation[typeA,typeB]")
	confDci01Cmd.Flags().IntVar(&dci01TdK2, "dci01TdK2", 1, "Slot offset K2 for PUSCH time-domain allocation[0..32]")
	confDci01Cmd.Flags().IntVar(&dci01TdSliv, "dci01TdSliv", 27, "SLIV for PUSCH time-domain allocation")
	confDci01Cmd.Flags().IntVar(&dci01TdStartSymb, "dci01TdStartSymb", 0, "Starting symbol S for PUSCH time-domain allocation")
	confDci01Cmd.Flags().IntVar(&dci01TdNumSymbs, "dci01TdNumSymbs", 14, "Number of OFDM symbols L for PUSCH time-domain allocation")
	confDci01Cmd.Flags().StringVar(&dci01FdRaType, "dci01FdRaType", "raType1", "resourceAllocation for PUSCH frequency-domain allocation[raType0,raType1]")
	confDci01Cmd.Flags().StringVar(&dci01FdFreqHop, "dci01FdFreqHop", "disabled", "Frequency-hopping-flag field for DCI 0_1[disabled,intraSlot,interSlot]")
	confDci01Cmd.Flags().StringVar(&dci01FdRa, "dci01FdRa", "0000001000100001", "Frequency-domain-resource-assignment field of DCI 0_1")
	confDci01Cmd.Flags().IntVar(&dci01FdStartRb, "dci01FdStartRb", 0, "RB_start of RIV for PUSCH frequency-domain allocation")
	confDci01Cmd.Flags().IntVar(&dci01FdNumRbs, "dci01FdNumRbs", 273, "L_RBs of RIV for PUSCH frequency-domain allocation")
	confDci01Cmd.Flags().IntVar(&dci01McsCw0, "dci01McsCw0", 28, "Modulation-and-coding-scheme-cw0 field of DCI 0_1[0..28]")
	confDci01Cmd.Flags().Int("_tbs", 475584, "Transport block size(bits) for PUSCH")
	confDci01Cmd.Flags().IntVar(&dci01CbTpmiNumLayers, "dci01CbTpmiNumLayers", 2, "Precoding-information-and-number-of-layers field of DCI 0_1[0..63]")
	confDci01Cmd.Flags().StringVar(&dci01Sri, "dci01Sri", "", "SRS-resource-indicator field of DCI 0_1")
	confDci01Cmd.Flags().IntVar(&dci01AntPorts, "dci01AntPorts", 0, "Antenna_port(s) field of DCI 0_1[0..7]")
	confDci01Cmd.Flags().IntVar(&dci01PtrsDmrsMap, "dci01PtrsDmrsMap", 0, "PTRS-DMRS-association field of DCI 0_1[0..3]")
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

	confBwpCmd.Flags().StringSlice("_bwpType", []string{"iniDlBwp", "dedDlBwp", "iniUlBwp", "dedUlBwp"}, "BWP type")
	confBwpCmd.Flags().IntSlice("_bwpId", []int{0, 1, 0, 1}, "bwp-Id of BWP-Uplink or BWP-Downlink")
	confBwpCmd.Flags().StringSlice("_bwpScs", []string{"30KHz", "30KHz", "30KHz", "30KHz"}, "subcarrierSpacing of BWP")
	confBwpCmd.Flags().StringSlice("_bwpCp", []string{"normal", "normal", "normal", "normal"}, "cyclicPrefix of BWP")
	confBwpCmd.Flags().IntSliceVar(&bwpLocAndBw, "bwpLocAndBw", []int{12925, 1099, 1099, 1099}, "locationAndBandwidth of BWP")
	confBwpCmd.Flags().IntSliceVar(&bwpStartRb, "bwpStartRb", []int{0, 0, 0, 0}, "RB_start of BWP")
	confBwpCmd.Flags().IntSliceVar(&bwpNumRbs, "bwpNumRbs", []int{48, 273, 273, 273}, "L_RBs of BWP")
	confBwpCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.bwp._bwpType", confBwpCmd.Flags().Lookup("_bwpType"))
	viper.BindPFlag("nrrg.bwp._bwpId", confBwpCmd.Flags().Lookup("_bwpId"))
	viper.BindPFlag("nrrg.bwp._bwpScs", confBwpCmd.Flags().Lookup("_bwpScs"))
	viper.BindPFlag("nrrg.bwp._bwpCp", confBwpCmd.Flags().Lookup("_bwpCp"))
	viper.BindPFlag("nrrg.bwp.bwpLocAndBw", confBwpCmd.Flags().Lookup("bwpLocAndBw"))
	viper.BindPFlag("nrrg.bwp.bwpStartRb", confBwpCmd.Flags().Lookup("bwpStartRb"))
	viper.BindPFlag("nrrg.bwp.bwpNumRbs", confBwpCmd.Flags().Lookup("bwpNumRbs"))
	confBwpCmd.Flags().MarkHidden("_bwpType")
	confBwpCmd.Flags().MarkHidden("_bwpId")
	confBwpCmd.Flags().MarkHidden("_bwpScs")
	confBwpCmd.Flags().MarkHidden("_bwpCp")

	confRachCmd.Flags().IntVar(&prachConfId, "prachConfId", 148, "prach-ConfigurationIndex of RACH-ConfigGeneric[0..255]")
	confRachCmd.Flags().String("_raFormat", "B4", "Preamble format")
	confRachCmd.Flags().Int("_raX", 2, "The x in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().IntSlice("_raY", []int{1}, "The y in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().IntSlice("_raSubfNumFr1SlotNumFr2", []int{9}, "The Subframe-number in 3GPP TS 38.211 Table 6.3.3.2-2 and Table 6.3.3.2-3, or the Slot-number in Table 6.3.3.2-4")
	confRachCmd.Flags().Int("_raStartingSymb", 0, "The Starting-symbol in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().Int("_raNumSlotsPerSubfFr1Per60KSlotFr2", 1, "The Number-of-PRACH-slots-within-a-subframe in 3GPP TS 38.211 Table 6.3.3.2-2 and Table 6.3.3.2-3, or the Number-of-PRACH-slots-within-a-60-kHz-slot in Table 6.3.3.2-4")
	confRachCmd.Flags().Int("_raNumOccasionsPerSlot", 1, "The Number-of-time-domain-PRACH-occasions-within-a-PRACH-slot in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().Int("_raDuration", 12, "The PRACH-duration in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().StringVar(&msg1Scs, "msg1Scs", "30KHz", "msg1-SubcarrierSpacing of RACH-ConfigCommon")
	confRachCmd.Flags().IntVar(&msg1Fdm, "msg1Fdm", 1, "msg1-FDM of RACH-ConfigGeneric[1,2,4,8]")
	confRachCmd.Flags().IntVar(&msg1FreqStart, "msg1FreqStart", 0, "msg1-FrequencyStart of RACH-ConfigGeneric[0..274]")
	confRachCmd.Flags().StringVar(&raRespWin, "raRespWin", "sl20", "ra-ResponseWindow of RACH-ConfigGeneric[sl1,sl2,sl4,sl8,sl10,sl20,sl40,sl80]")
	confRachCmd.Flags().IntVar(&totNumPreambs, "totNumPreambs", 64, "totalNumberOfRA-Preambles of RACH-ConfigCommon[1..64]")
	confRachCmd.Flags().StringVar(&ssbPerRachOccasion, "ssbPerRachOccasion", "one", "ssb-perRACH-Occasion of RACH-ConfigGeneric[oneEighth,oneFourth,oneHalf,one,two,four,eight,sixteen]")
	confRachCmd.Flags().IntVar(&cbPreambsPerSsb, "cbPreambsPerSsb", 64, "cb-PreamblesPerSSB of RACH-ConfigCommon[depends on ssbPerRachOccasion]")
	confRachCmd.Flags().StringVar(&contResTimer, "contResTimer", "sf64", "ra-ContentionResolutionTimer of RACH-ConfigGeneric[sf8,sf16,sf24,sf32,sf40,sf48,sf56,sf64]")
	confRachCmd.Flags().StringVar(&msg3Tp, "msg3Tp", "disabled", "msg3-transformPrecoder of RACH-ConfigGeneric[disabled,enabled]")
	confRachCmd.Flags().Int("_raLen", 139, "L_RA of 3GPP TS 38.211 Table 6.3.3.1-1 and Table 6.3.3.1-2")
	confRachCmd.Flags().Int("_raNumRbs", 12, "Allocation-expressed-in-number-of-RBs-for-PUSCH of 3GPP TS 38.211 Table 6.3.3.2-1")
	confRachCmd.Flags().Int("_raKBar", 2, "k_bar of 3GPP TS 38.211 Table 6.3.3.2-1")
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

	confDmrsCommonCmd.Flags().StringSlice("_schInfo", []string{"SIB1", "Msg2", "Msg4", "Msg3"}, "Information of UL/DL-SCH")
	confDmrsCommonCmd.Flags().StringSlice("_dmrsType", []string{"type1", "type1", "type1", "type1"}, "dmrs-Type as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	confDmrsCommonCmd.Flags().StringSlice("_dmrsAddPos", []string{"pos0", "pos0", "pos0", "pos1"}, "dmrs-AdditionalPosition as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	confDmrsCommonCmd.Flags().StringSlice("_maxLength", []string{"len1", "len1", "len1", "len1"}, "maxLength as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	confDmrsCommonCmd.Flags().IntSlice("_dmrsPorts", []int{1000, 1000, 1000, 0}, "DMRS antenna ports")
	confDmrsCommonCmd.Flags().IntSlice("_cdmGroupsWoData", []int{1, 1, 1, 2}, "CDM group(s) without data")
	confDmrsCommonCmd.Flags().IntSlice("_numFrontLoadSymbs", []int{1, 1, 1, 1}, "Number of front-load DMRS symbols")
	// _tdL for SIB1/Msg2/Msg4 is comma(',') separated
	// _tdL for Msg3 is comma(',') separated if msg3FreqHop is disabled, otherwise, _tdL is underscore('_') separated for each hop
	confDmrsCommonCmd.Flags().StringSlice("_tdL", []string{"0", "0", "0", "0_0"}, "Time-domain locations for DMRS")
	// _fdK indicates REs per PRB for DMRS
	confDmrsCommonCmd.Flags().StringSlice("_fdK", []string{"101010101010", "101010101010", "101010101010", "111111111111"}, "Frequency-domain locations of DMRS")
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

	confDmrsPdschCmd.Flags().StringVar(&pdschDmrsType, "pdschDmrsType", "type1", "dmrs-Type as in DMRS-DownlinkConfig[type1,type2]")
	confDmrsPdschCmd.Flags().StringVar(&pdschDmrsAddPos, "pdschDmrsAddPos", "pos0", "dmrs-additionalPosition as in DMRS-DownlinkConfig[pos0,pos1,pos2,pos3]")
	confDmrsPdschCmd.Flags().StringVar(&pdschMaxLength, "pdschMaxLength", "len1", "maxLength as in DMRS-DownlinkConfig[len1,len2]")
	confDmrsPdschCmd.Flags().IntSlice("_dmrsPorts", []int{1000, 1001, 1002, 1003}, "DMRS antenna ports")
	confDmrsPdschCmd.Flags().Int("_cdmGroupsWoData", 2, "CDM group(s) without data")
	confDmrsPdschCmd.Flags().Int("_numFrontLoadSymbs", 1, "Number of front-load DMRS symbols")
	// _tdL is semicolon separated
	confDmrsPdschCmd.Flags().String("_tdL", "2", "Time-domain locations for DMRS")
	// _fdK indicates REs per PRB for DMRS
	confDmrsPdschCmd.Flags().String("_fdK", "111111111111", "Frequency-domain locations of DMRS")
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

	confPtrsPdschCmd.Flags().BoolVar(&pdschPtrsEnabled, "pdschPtrsEnabled", true, "Enable PTRS of PDSCH[false,true]")
	confPtrsPdschCmd.Flags().IntVar(&pdschPtrsTimeDensity, "pdschPtrsTimeDensity", 1, "The L_PTRS deduced from timeDensity of PTRS-DownlinkConfig[1,2,4]")
	confPtrsPdschCmd.Flags().IntVar(&pdschPtrsFreqDensity, "pdschPtrsFreqDensity", 2, "The K_PTRS deduced from frequencyDensity of PTRS-DownlinkConfig[2,4]")
	confPtrsPdschCmd.Flags().StringVar(&pdschPtrsReOffset, "pdschPtrsReOffset", "offset00", "resourceElementOffset of PTRS-DownlinkConfig[offset00,offset01,offset10,offset11]")
	confPtrsPdschCmd.Flags().IntSlice("_dmrsPorts", []int{1000}, "Associated DMRS antenna ports")
	confPtrsPdschCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.ptrspdsch.pdschPtrsEnabled", confPtrsPdschCmd.Flags().Lookup("pdschPtrsEnabled"))
	viper.BindPFlag("nrrg.ptrspdsch.pdschPtrsTimeDensity", confPtrsPdschCmd.Flags().Lookup("pdschPtrsTimeDensity"))
	viper.BindPFlag("nrrg.ptrspdsch.pdschPtrsFreqDensity", confPtrsPdschCmd.Flags().Lookup("pdschPtrsFreqDensity"))
	viper.BindPFlag("nrrg.ptrspdsch.pdschPtrsReOffset", confPtrsPdschCmd.Flags().Lookup("pdschPtrsReOffset"))
	viper.BindPFlag("nrrg.ptrspdsch._dmrsPorts", confPtrsPdschCmd.Flags().Lookup("_dmrsPorts"))
	confPtrsPdschCmd.Flags().MarkHidden("_dmrsPorts")

	confDmrsPuschCmd.Flags().StringVar(&puschDmrsType, "puschDmrsType", "type1", "dmrs-Type as in DMRS-UplinkConfig[type1,type2]")
	confDmrsPuschCmd.Flags().StringVar(&puschDmrsAddPos, "puschDmrsAddPos", "pos0", "dmrs-additionalPosition as in DMRS-UplinkConfig[pos0,pos1,pos2,pos3]")
	confDmrsPuschCmd.Flags().StringVar(&puschMaxLength, "puschMaxLength", "len1", "maxLength as in DMRS-UplinkConfig[len1,len2]")
	confDmrsPuschCmd.Flags().IntSlice("_dmrsPorts", []int{0, 1}, "DMRS antenna ports")
	confDmrsPuschCmd.Flags().Int("_cdmGroupsWoData", 1, "CDM group(s) without data")
	confDmrsPuschCmd.Flags().Int("_numFrontLoadSymbs", 1, "Number of front-load DMRS symbols")
	// _tdL is semicolon separated
	confDmrsPuschCmd.Flags().String("_tdL", "2", "Time-domain locations for DMRS")
	// _fdK indicates REs per PRB for DMRS
	confDmrsPuschCmd.Flags().String("_fdK", "101010101010", "Frequency-domain locations of DMRS")
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

	confPtrsPuschCmd.Flags().BoolVar(&puschPtrsEnabled, "puschPtrsEnabled", true, "Enable PTRS of PDSCH[false,true]")
	confPtrsPuschCmd.Flags().IntVar(&puschPtrsTimeDensity, "puschPtrsTimeDensity", 1, "The L_PTRS deduced from timeDensity of PTRS-UplinkConfig for CP-OFDM[1,2,4]")
	confPtrsPuschCmd.Flags().IntVar(&puschPtrsFreqDensity, "puschPtrsFreqDensity", 2, "The K_PTRS deduced from frequencyDensity of PTRS-UplinkConfig for CP-OFDM[2,4]")
	confPtrsPuschCmd.Flags().StringVar(&puschPtrsReOffset, "puschPtrsReOffset", "offset00", "resourceElementOffset of PTRS-UplinkConfig for CP-OFDM[offset00,offset01,offset10,offset11]")
	confPtrsPuschCmd.Flags().StringVar(&puschPtrsMaxNumPorts, "puschPtrsMaxNumPorts", "n1", "maxNrofPorts of PTRS-UplinkConfig for CP-OFDM[n1,n2]")
	confPtrsPuschCmd.Flags().IntSlice("_dmrsPorts", []int{0}, "Associated DMRS antenna ports for CP-OFDM")
	confPtrsPuschCmd.Flags().IntVar(&puschPtrsTimeDensityTp, "puschPtrsTimeDensityTp", 1, "The L_PTRS deduced from timeDensityTransformPrecoding of PTRS-UplinkConfig for DFS-S-OFDM[1,2]")
	confPtrsPuschCmd.Flags().StringVar(&puschPtrsGrpPatternTp, "puschPtrsGrpPatternTp", "p0", "The Scheduled-bandwidth column index of 3GPP TS 38.214 Table 6.2.3.2-1, deduced from sampleDensity of PTRS-UplinkConfig for DFS-S-OFDM[p0,p1,p2,p3,p4]")
	confPtrsPuschCmd.Flags().Int("_numGrpsTp", 2, "The Number-of-PT-RS-groups of 3GPP TS 38.214 Table 6.2.3.2-1, deduced from sampleDensity of PTRS-UplinkConfig for DFS-S-OFDM")
	confPtrsPuschCmd.Flags().Int("_samplesPerGrpTp", 2, "The Number-of-samples-per-PT-RS-group of 3GPP TS 38.214 Table 6.2.3.2-1, deduced from sampleDensity of PTRS-UplinkConfig for DFS-S-OFDM")
	confPtrsPuschCmd.Flags().IntSlice("_dmrsPortsTp", []int{}, "Associated DMRS antenna ports for DFT-S-OFDM")
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

	confPdschCmd.Flags().StringVar(&pdschAggFactor, "pdschAggFactor", "n1", "pdsch-AggregationFactor of PDSCH-Config[n1,n2,n4,n8]")
	confPdschCmd.Flags().StringVar(&pdschRbgCfg, "pdschRbgCfg", "config1", "rbg-Size of PDSCH-Config[config1,config2]")
	confPdschCmd.Flags().Int("_rbgSize", 16, "RBG size of PDSCH resource allocation type 0")
	confPdschCmd.Flags().StringVar(&pdschMcsTable, "pdschMcsTable", "qam256", "mcs-Table of PDSCH-Config[qam64,qam256,qam64LowSE]")
	confPdschCmd.Flags().StringVar(&pdschXOh, "pdschXOh", "xOh0", "xOverhead of PDSCH-ServingCellConfig[xOh0,xOh6,xOh12,xOh18]")
	confPdschCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.pdsch.pdschAggFactor", confPdschCmd.Flags().Lookup("pdschAggFactor"))
	viper.BindPFlag("nrrg.pdsch.pdschRbgCfg", confPdschCmd.Flags().Lookup("pdschRbgCfg"))
	viper.BindPFlag("nrrg.pdsch._rbgSize", confPdschCmd.Flags().Lookup("_rbgSize"))
	viper.BindPFlag("nrrg.pdsch.pdschMcsTable", confPdschCmd.Flags().Lookup("pdschMcsTable"))
	viper.BindPFlag("nrrg.pdsch.pdschXOh", confPdschCmd.Flags().Lookup("pdschXOh"))
	confPdschCmd.Flags().MarkHidden("_rbgSize")

	confPuschCmd.Flags().StringVar(&puschTxCfg, "puschTxCfg", "codebook", "txConfig of PUSCH-Config[codebook,nonCodebook]")
	confPuschCmd.Flags().StringVar(&puschCbSubset, "puschCbSubset", "fullyAndPartialAndNonCoherent", "codebookSubset of PUSCH-Config[fullyAndPartialAndNonCoherent,partialAndNonCoherent,nonCoherent]")
	confPuschCmd.Flags().IntVar(&puschCbMaxRankNonCbMaxLayers, "puschCbMaxRankNonCbMaxLayers", 2, "maxRank of PUSCH-Config or maxMIMO-Layers of PUSCH-ServingCellConfig[1..4]")
	confPuschCmd.Flags().IntVar(&puschFreqHopOffset, "puschFreqHopOffset", 0, "frequencyHoppingOffsetLists of PUSCH-Config[0..274]")
	confPuschCmd.Flags().StringVar(&puschTp, "puschTp", "disabled", "transformPrecoder of PUSCH-Config[disabled,enabled]")
	confPuschCmd.Flags().StringVar(&puschAggFactor, "puschAggFactor", "n1", "pusch-AggregationFactor of PUSCH-Config[n1,n2,n4,n8]")
	confPuschCmd.Flags().StringVar(&puschRbgCfg, "puschRbgCfg", "config1", "rbg-Size of PUSCH-Config[config1,config2]")
	confPuschCmd.Flags().Int("_rbgSize", 16, "RBG size of PUSCH resource allocation type 0")
	confPuschCmd.Flags().StringVar(&puschMcsTable, "puschMcsTable", "qam64", "mcs-Table of PUSCH-Config[qam64,qam256,qam64LowSE]")
	confPuschCmd.Flags().StringVar(&puschXOh, "puschXOh", "xOh0", "xOverhead of PUSCH-ServingCellConfig[xOh0,xOh6,xOh12,xOh18]")
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

	confNzpCsiRsCmd.Flags().Int("_resSetId", 0, "nzp-CSI-ResourceSetId of NZP-CSI-RS-ResourceSet")
	confNzpCsiRsCmd.Flags().Bool("_trsInfo", false, "trs-Info of NZP-CSI-RS-ResourceSet")
	confNzpCsiRsCmd.Flags().Int("_resId", 0, "nzp-CSI-RS-ResourceId of NZP-CSI-RS-Resource")
	confNzpCsiRsCmd.Flags().StringVar(&nzpCsiRsFreqAllocRow, "nzpCsiRsFreqAllocRow", "row4", "The row of frequencyDomainAllocation of CSI-RS-ResourceMapping[row1,row2,row4,other]")
	confNzpCsiRsCmd.Flags().StringVar(&nzpCsiRsFreqAllocBits, "nzpCsiRsFreqAllocBits", "001", "The bit-string of frequencyDomainAllocation of CSI-RS-ResourceMapping")
	confNzpCsiRsCmd.Flags().StringVar(&nzpCsiRsNumPorts, "nzpCsiRsNumPorts", "p4", "nrofPorts of CSI-RS-ResourceMapping[p1,p2,p4,p8,p12,p16,p24,p32]")
	confNzpCsiRsCmd.Flags().StringVar(&nzpCsiRsCdmType, "nzpCsiRsCdmType", "fd-CDM2", "cdm-Type of CSI-RS-ResourceMapping[noCDM,fd-CDM2,cdm4-FD2-TD2,cdm8-FD2-TD4]")
	confNzpCsiRsCmd.Flags().StringVar(&nzpCsiRsDensity, "nzpCsiRsDensity", "one", "density of CSI-RS-ResourceMapping[evenPRBs,oddPRBs,one,three]")
	confNzpCsiRsCmd.Flags().IntVar(&nzpCsiRsFirstSymb, "nzpCsiRsFirstSymb", 1, "firstOFDMSymbolInTimeDomain of CSI-RS-ResourceMapping[0..13]")
	confNzpCsiRsCmd.Flags().IntVar(&nzpCsiRsFirstSymb2, "nzpCsiRsFirstSymb2", -1, "firstOFDMSymbolInTimeDomain2 of CSI-RS-ResourceMapping[-1 or 0..13]")
	confNzpCsiRsCmd.Flags().IntVar(&nzpCsiRsStartRb, "nzpCsiRsStartRb", 0, "startingRB of CSI-FrequencyOccupation[0..274]")
	confNzpCsiRsCmd.Flags().IntVar(&nzpCsiRsNumRbs, "nzpCsiRsNumRbs", 276, "nrofRBs of CSI-FrequencyOccupation[24..276]")
	confNzpCsiRsCmd.Flags().StringVar(&nzpCsiRsPeriod, "nzpCsiRsPeriod", "slots20", "periodicityAndOffset of NZP-CSI-RS-Resource[slots4,slots5,slots8,slots10,slots16,slots20,slots32,slots40,slots64,slots80,slots160,slots320,slots640]")
	confNzpCsiRsCmd.Flags().IntVar(&nzpCsiRsOffset, "nzpCsiRsOffset", 10, "periodicityAndOffset of NZP-CSI-RS-Resource[0..period-1]")
	confNzpCsiRsCmd.Flags().Int("_row", 4, "The Row of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().StringSlice("_kBarLBar", []string{"0,0","2,0"}, "The constants deduced from (k_bar, l_bar) of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().IntSlice("_ki", []int{0, 0}, "The index ki deduced from (k_bar, l_bar) of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().IntSlice("_li", []int{0, 0}, "The index li deduced from (k_bar, l_bar) of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().IntSlice("_cdmGrpIndj", []int{0, 1}, "The CDM-group-index-j of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().IntSlice("_kap", []int{0, 1}, "The k_ap of 3GPP TS 38.211 Table 7.4.1.5.3-1")
	confNzpCsiRsCmd.Flags().IntSlice("_lap", []int{0}, "The l_ap of 3GPP TS 38.211 Table 7.4.1.5.3-1")
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
