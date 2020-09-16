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

)

// nrrgCmd represents the nrrg command
var nrrgCmd = &cobra.Command{
	Use:   "nrrg [sub]",
	Short: "",
	Long: `nrrg generates NR(new radio) resource grid according to network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.WriteConfig()
	},
}

// nrrgConfCmd represents the 'nrrg conf' command
var nrrgConfCmd = &cobra.Command{
	Use:   "conf [sub]",
	Short: "",
	Long: `'nrrg conf' can be used to get/set network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg conf called")
		viper.WriteConfig()
	},
}

// confFreqBandCmd represents the 'nrrg conf freqband' command
var confFreqBandCmd = &cobra.Command{
	Use:   "freqband [sub]",
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
	Use:   "ssbgrid [sub]",
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
	Use:   "ssbburst [sub]",
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
	Use:   "mib [sub]",
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
	Use:   "carriergrid [sub]",
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
	Use:   "commonsetting [sub]",
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
	Use:   "css0 [sub]",
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
	Use:   "coreset1 [sub]",
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
	Use:   "uss [sub]",
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
	Use:   "dci10 [sub]",
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
	Use:   "dci11 [sub]",
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
	Use:   "msg3 [sub]",
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
	Use:   "dci01 [sub]",
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
	Use:   "bwp [sub]",
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
	Use:   "rach [sub]",
	Short: "",
	Long: `'nrrg conf rach' can be used to get/set random access related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confDmrsRaCmd represents the 'nrrg conf dmrsra' command
var confDmrsRaCmd = &cobra.Command{
	Use:   "dmrsra [sub]",
	Short: "",
	Long: `'nrrg conf dmrsra' can be used to get/set DMRS of SIB1/Msg2/Msg4/Msg3 related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confDmrsPdschCmd represents the 'nrrg conf dmrspdsch' command
var confDmrsPdschCmd = &cobra.Command{
	Use:   "dmrspdsch [sub]",
	Short: "",
	Long: `'nrrg conf dmrspdsch' can be used to get/set DMRS/PTRS of PDSCH related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Flag | ActualValue | DefaultValue\n")
		cmd.Flags().VisitAll(func (f *pflag.Flag) { if f.Name != "config" && f.Name != "help" {fmt.Printf("%v | %v | %v\n", f.Name, f.Value, f.DefValue)}})
		viper.WriteConfig()
	},
}

// confDmrsPuschCmd represents the 'nrrg conf dmrspusch' command
var confDmrsPuschCmd = &cobra.Command{
	Use:   "dmrspusch [sub]",
	Short: "",
	Long: `'nrrg conf dmrspusch' can be used to get/set DMRS/PTRS of PUSCH related network configurations.`,
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
	nrrgConfCmd.AddCommand(confDmrsRaCmd)
	nrrgConfCmd.AddCommand(confDmrsPdschCmd)
	nrrgConfCmd.AddCommand(confDmrsPuschCmd)
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
	confFreqBandCmd.Flags().String("duplexMode", "TDD", "Duplex mode")
	confFreqBandCmd.Flags().Int("maxDlFreq", 2690, "Maximum DL frequency(MHz)")
	confFreqBandCmd.Flags().String("freqRange", "FR1", "Frequency range(FR1/FR2)")
	confFreqBandCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.freqBand.opBand", confFreqBandCmd.Flags().Lookup("opBand"))
	viper.BindPFlag("nrrg.freqBand.duplexMode", confFreqBandCmd.Flags().Lookup("duplexMode"))
	viper.BindPFlag("nrrg.freqBand.maxDlFreq", confFreqBandCmd.Flags().Lookup("maxDlFreq"))
	viper.BindPFlag("nrrg.freqBand.freqRange", confFreqBandCmd.Flags().Lookup("freqRange"))
	confFreqBandCmd.Flags().MarkHidden("duplexMode")
	confFreqBandCmd.Flags().MarkHidden("maxDlFreq")
	confFreqBandCmd.Flags().MarkHidden("freqRange")

	confSsbGridCmd.Flags().StringVar(&ssbScs, "ssbScs",  "30KHz", "SSB subcarrier spacing")
	confSsbGridCmd.Flags().String("ssbPattern", "Case C", "SSB pattern")
	confSsbGridCmd.Flags().IntVar(&kSsb, "kSsb", 0, "k_SSB[0..23]")
	confSsbGridCmd.Flags().Int("nCrbSsb", 32, "n_CRB_SSB")
	confSsbGridCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.ssbGrid.ssbScs", confSsbGridCmd.Flags().Lookup("ssbScs"))
	viper.BindPFlag("nrrg.ssbGrid.ssbPattern", confSsbGridCmd.Flags().Lookup("ssbPattern"))
	viper.BindPFlag("nrrg.ssbGrid.kSsb", confSsbGridCmd.Flags().Lookup("kSsb"))
	viper.BindPFlag("nrrg.ssbGrid.nCrbSsb", confSsbGridCmd.Flags().Lookup("nCrbSsb"))
	confSsbGridCmd.Flags().MarkHidden("ssbPattern")
	confSsbGridCmd.Flags().MarkHidden("nCrbSsb")

	confSsbBurstCmd.Flags().Int("maxL", 8, "max_L")
	confSsbBurstCmd.Flags().StringVar(&inOneGrp, "inOneGroup", "11111111", "inOneGroup of ssb-PositionsInBurst")
	confSsbBurstCmd.Flags().StringVar(&grpPresence, "groupPresence", "", "groupPresence of ssb-PositionsInBurst")
	confSsbBurstCmd.Flags().StringVar(&ssbPeriod, "ssbPeriod", "20ms", "ssb-PeriodicityServingCell[5ms,10ms,20ms,40ms,80ms,160ms]")
	confSsbBurstCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.ssbBurst.maxL", confSsbBurstCmd.Flags().Lookup("maxL"))
	viper.BindPFlag("nrrg.ssbBurst.inOneGroup", confSsbBurstCmd.Flags().Lookup("inOneGroup"))
	viper.BindPFlag("nrrg.ssbBurst.groupPresence", confSsbBurstCmd.Flags().Lookup("groupPresence"))
	viper.BindPFlag("nrrg.ssbBurst.ssbPeriod", confSsbBurstCmd.Flags().Lookup("ssbPeriod"))
	confSsbBurstCmd.Flags().MarkHidden("maxL")

	confMibCmd.Flags().IntVar(&sfn, "sfn", 0, "System frame number(SFN)[0..1023]")
	confMibCmd.Flags().IntVar(&hrf, "hrf", 0, "Half frame bit[0,1]")
	confMibCmd.Flags().StringVar(&dmrsTypeAPos, "dmrsTypeAPos", "pos2", "dmrs-TypeA-Position[pos2,pos3]")
	confMibCmd.Flags().StringVar(&commonScs, "commonScs", "30KHz", "subCarrierSpacingCommon")
	confMibCmd.Flags().IntVar(&rmsiCoreset0, "rmsiCoreset0", 12, "coresetZero of PDCCH-ConfigSIB1[0..15]")
	confMibCmd.Flags().IntVar(&rmsiCss0, "rmsiCss0", 0, "searchSpaceZero of PDCCH-ConfigSIB1[0..15]")
	confMibCmd.Flags().Int("coreset0MultiplexingPat", 1, "Multiplexing pattern of CORESET0")
	confMibCmd.Flags().Int("coreset0NumRbs", 48, "Number of PRBs of CORESET0")
	confMibCmd.Flags().Int("coreset0NumSymbs", 1, "Number of OFDM symbols of CORESET0")
	confMibCmd.Flags().IntSlice("coreset0OffsetList", []int{16}, "List of offset of CORESET0")
	confMibCmd.Flags().Int("coreset0Offset", 16, "Offset of CORESET0")
	confMibCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.mib.sfn", confMibCmd.Flags().Lookup("sfn"))
	viper.BindPFlag("nrrg.mib.hrf", confMibCmd.Flags().Lookup("hrf"))
	viper.BindPFlag("nrrg.mib.dmrsTypeAPos", confMibCmd.Flags().Lookup("dmrsTypeAPos"))
	viper.BindPFlag("nrrg.mib.commonScs", confMibCmd.Flags().Lookup("commonScs"))
	viper.BindPFlag("nrrg.mib.rmsiCoreset0", confMibCmd.Flags().Lookup("rmsiCoreset0"))
	viper.BindPFlag("nrrg.mib.rmsiCss0", confMibCmd.Flags().Lookup("rmsiCss0"))
	viper.BindPFlag("nrrg.mib.coreset0MultiplexingPat", confMibCmd.Flags().Lookup("coreset0MultiplexingPat"))
	viper.BindPFlag("nrrg.mib.coreset0NumRbs", confMibCmd.Flags().Lookup("coreset0NumRbs"))
	viper.BindPFlag("nrrg.mib.coreset0NumSymbs", confMibCmd.Flags().Lookup("coreset0NumSymbs"))
	viper.BindPFlag("nrrg.mib.coreset0OffsetList", confMibCmd.Flags().Lookup("coreset0OffsetList"))
	viper.BindPFlag("nrrg.mib.coreset0Offset", confMibCmd.Flags().Lookup("coreset0Offset"))
	confMibCmd.Flags().MarkHidden("coreset0MultiplexingPat")
	confMibCmd.Flags().MarkHidden("coreset0NumRbs")
	confMibCmd.Flags().MarkHidden("coreset0NumSymbs")
	confMibCmd.Flags().MarkHidden("coreset0OffsetList")
	confMibCmd.Flags().MarkHidden("coreset0Offset")

	confCarrierGridCmd.Flags().StringVar(&carrierScs, "carrierScs", "30KHz", "subcarrierSpacing of SCS-SpecificCarrier")
	confCarrierGridCmd.Flags().StringVar(&bw, "bw", "100MHz", "Transmission bandwidth(MHz)")
	confCarrierGridCmd.Flags().Int("carrierNumRbs", 273, "carrierBandwidth(N_RB) of SCS-SpecificCarrier")
	confCarrierGridCmd.Flags().Int("offsetToCarrier", 0, "offsetToCarrier of SCS-SpecificCarrier")
	confCarrierGridCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.carrierGrid.carrierScs", confCarrierGridCmd.Flags().Lookup("carrierScs"))
	viper.BindPFlag("nrrg.carrierGrid.bw", confCarrierGridCmd.Flags().Lookup("bw"))
	viper.BindPFlag("nrrg.carrierGrid.carrierNumRbs", confCarrierGridCmd.Flags().Lookup("carrierNumRbs"))
	viper.BindPFlag("nrrg.carrierGrid.offsetToCarrier", confCarrierGridCmd.Flags().Lookup("offsetToCarrier"))
	confCarrierGridCmd.Flags().MarkHidden("carrierNumRbs")
	confCarrierGridCmd.Flags().MarkHidden("offsetToCarrier")

	confCommonSettingCmd.Flags().IntVar(&pci, "pci", 0, "Physical cell identity[0..1007]")
	confCommonSettingCmd.Flags().StringVar(&numUeAp, "numUeAp", "2T", "Number of UE antennas[1T,2T,4T]")
	confCommonSettingCmd.Flags().String("refScs", "30KHz", "referenceSubcarrierSpacing of TDD-UL-DL-ConfigCommon")
	confCommonSettingCmd.Flags().StringSliceVar(&patPeriod, "patPeriod",  []string{"5ms"}, "dl-UL-TransmissionPeriodicity of TDD-UL-DL-ConfigCommon[0.5ms,0.625ms,1ms,1.25ms,2ms,2.5ms,3ms,4ms,5ms,10ms]")
	confCommonSettingCmd.Flags().IntSliceVar(&patNumDlSlots, "patNumDlSlots",  []int{7}, "nrofDownlinkSlot of TDD-UL-DL-ConfigCommon[0..80]")
	confCommonSettingCmd.Flags().IntSliceVar(&patNumDlSymbs, "patNumDlSymbs",  []int{6}, "nrofDownlinkSymbols of TDD-UL-DL-ConfigCommon[0..13]")
	confCommonSettingCmd.Flags().IntSliceVar(&patNumUlSymbs, "patNumUlSymbs",  []int{4}, "nrofUplinkSymbols of TDD-UL-DL-ConfigCommon[0..13]")
	confCommonSettingCmd.Flags().IntSliceVar(&patNumUlSlots, "patNumUlSlots",  []int{2}, "nrofUplinkSlots of TDD-UL-DL-ConfigCommon[0..80]")
	confCommonSettingCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.commonsetting.pci", confCommonSettingCmd.Flags().Lookup("pci"))
	viper.BindPFlag("nrrg.commonsetting.numUeAp", confCommonSettingCmd.Flags().Lookup("numUeAp"))
	viper.BindPFlag("nrrg.commonsetting.refScs", confCommonSettingCmd.Flags().Lookup("refScs"))
	viper.BindPFlag("nrrg.commonsetting.patPeriod", confCommonSettingCmd.Flags().Lookup("patPeriod"))
	viper.BindPFlag("nrrg.commonsetting.patNumDlSlots", confCommonSettingCmd.Flags().Lookup("patNumDlSlots"))
	viper.BindPFlag("nrrg.commonsetting.patNumDlSymbs", confCommonSettingCmd.Flags().Lookup("patNumDlSymbs"))
	viper.BindPFlag("nrrg.commonsetting.patNumUlSymbs", confCommonSettingCmd.Flags().Lookup("patNumUlSymbs"))
	viper.BindPFlag("nrrg.commonsetting.patNumUlSlots", confCommonSettingCmd.Flags().Lookup("patNumUlSlots"))
	confCommonSettingCmd.Flags().MarkHidden("refScs")

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

	confDci10Cmd.Flags().StringSlice("rnti", []string{"SI-RNTI", "RA-RNTI", "TC-RNTI"}, "RNTI for DCI 1_0")
	confDci10Cmd.Flags().IntSlice("muPdcch", []int{1, 1, 1}, "Subcarrier spacing of PDCCH[0..3]")
	confDci10Cmd.Flags().IntSlice("muPdsch", []int{1, 1, 1}, "Subcarrier spacing of PDSCH[0..3]")
	confDci10Cmd.Flags().IntSliceVar(&dci10TdRa, "dci10TdRa", []int{10, 10, 10}, "Time-domain-resource-assignment field of DCI 1_0[0..15]")
	confDci10Cmd.Flags().StringSlice("tdMappingType", []string{"typeB", "typeB", "typeB"}, "Mapping type for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSlice("tdK0", []int{0, 0, 0}, "Slot offset K0 for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSlice("tdSliv", []int{26, 26, 26}, "SLIV for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSlice("tdStartSymb", []int{12, 12, 12}, "Starting symbol S for PDSCH time-domain allocation")
	confDci10Cmd.Flags().IntSlice("tdNumSymbs", []int{2, 2, 2}, "Number of OFDM symbols L for PDSCH time-domain allocation")
	confDci10Cmd.Flags().StringSlice("fdRaType", []string{"raType1", "raType1", "raType1"}, "resourceAllocation for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().StringSlice("fdRa", []string{"00001011111", "00001011111", "00001011111"}, "Frequency-domain-resource-assignment field of DCI 1_0")
	confDci10Cmd.Flags().IntSliceVar(&dci10FdStartRb, "dci10FdStartRb", []int{0, 0, 0}, "RB_start of RIV for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&dci10FdNumRbs, "dci10FdNumRbs", []int{48, 48, 48}, "L_RBs of RIV for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().StringSliceVar(&dci10FdVrbPrbMappingType, "dci10FdVrbPrbMappingType", []string{"interleaved", "interleaved", "interleaved"}, "VRB-to-PRB-mapping field of DCI 1_0")
	confDci10Cmd.Flags().StringSlice("fdBundleSize", []string{"n2", "n2", "n2"}, "L(vrb-ToPRB-Interleaver) for PDSCH frequency-domain allocation")
	confDci10Cmd.Flags().IntSliceVar(&dci10McsCw0, "dci10McsCw0", []int{2, 2, 2}, "Modulation-and-coding-scheme field of DCI 1_0[0..9]")
	confDci10Cmd.Flags().IntSlice("tbs", []int{408, 408, 408}, "Transport block size(bits) for PDSCH")
	confDci10Cmd.Flags().IntVar(&dci10Msg2TbScaling, "dci10Msg2TbScaling", 0, "TB-scaling field of DCI 1_0 scheduling Msg2[0..2]")
	confDci10Cmd.Flags().IntVar(&dci10Msg4DeltaPri, "dci10Msg4DeltaPri", 1, "PUCCH-resource-indicator field of DCI 1_0 scheduling Msg4[0..7]")
	confDci10Cmd.Flags().IntVar(&dci10Msg4TdK1, "dci10Msg4TdK1", 6, "PDSCH-to-HARQ_feedback-timing-indicator(K1) field of DCI 1_0 scheduling Msg4[0..7]")
	confDci10Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dci10.rnti", confDci10Cmd.Flags().Lookup("rnti"))
	viper.BindPFlag("nrrg.dci10.muPdcch", confDci10Cmd.Flags().Lookup("muPdcch"))
	viper.BindPFlag("nrrg.dci10.muPdsch", confDci10Cmd.Flags().Lookup("muPdsch"))
	viper.BindPFlag("nrrg.dci10.dci10TdRa", confDci10Cmd.Flags().Lookup("dci10TdRa"))
	viper.BindPFlag("nrrg.dci10.tdMappingType", confDci10Cmd.Flags().Lookup("tdMappingType"))
	viper.BindPFlag("nrrg.dci10.tdK0", confDci10Cmd.Flags().Lookup("tdK0"))
	viper.BindPFlag("nrrg.dci10.tdSliv", confDci10Cmd.Flags().Lookup("tdSliv"))
	viper.BindPFlag("nrrg.dci10.tdStartSymb", confDci10Cmd.Flags().Lookup("tdStartSymb"))
	viper.BindPFlag("nrrg.dci10.tdNumSymbs", confDci10Cmd.Flags().Lookup("tdNumSymbs"))
	viper.BindPFlag("nrrg.dci10.fdRaType", confDci10Cmd.Flags().Lookup("fdRaType"))
	viper.BindPFlag("nrrg.dci10.fdRa", confDci10Cmd.Flags().Lookup("fdRa"))
	viper.BindPFlag("nrrg.dci10.dci10FdStartRb", confDci10Cmd.Flags().Lookup("dci10FdStartRb"))
	viper.BindPFlag("nrrg.dci10.dci10FdNumRbs", confDci10Cmd.Flags().Lookup("dci10FdNumRbs"))
	viper.BindPFlag("nrrg.dci10.dci10FdVrbPrbMappingType", confDci10Cmd.Flags().Lookup("dci10FdVrbPrbMappingType"))
	viper.BindPFlag("nrrg.dci10.fdBundleSize", confDci10Cmd.Flags().Lookup("fdBundleSize"))
	viper.BindPFlag("nrrg.dci10.dci10McsCw0", confDci10Cmd.Flags().Lookup("dci10McsCw0"))
	viper.BindPFlag("nrrg.dci10.tbs", confDci10Cmd.Flags().Lookup("tbs"))
	viper.BindPFlag("nrrg.dci10.dci10Msg2TbScaling", confDci10Cmd.Flags().Lookup("dci10Msg2TbScaling"))
	viper.BindPFlag("nrrg.dci10.dci10Msg4DeltaPri", confDci10Cmd.Flags().Lookup("dci10Msg4DeltaPri"))
	viper.BindPFlag("nrrg.dci10.dci10Msg4TdK1", confDci10Cmd.Flags().Lookup("dci10Msg4TdK1"))
	confDci10Cmd.Flags().MarkHidden("rnti")
	confDci10Cmd.Flags().MarkHidden("muPdcch")
	confDci10Cmd.Flags().MarkHidden("muPdsch")
	confDci10Cmd.Flags().MarkHidden("tdMappingType")
	confDci10Cmd.Flags().MarkHidden("tdK0")
	confDci10Cmd.Flags().MarkHidden("tdSliv")
	confDci10Cmd.Flags().MarkHidden("tdStartSymb")
	confDci10Cmd.Flags().MarkHidden("tdNumSymbs")
	confDci10Cmd.Flags().MarkHidden("fdRaType")
	confDci10Cmd.Flags().MarkHidden("fdRa")
	confDci10Cmd.Flags().MarkHidden("fdBundleSize")
	confDci10Cmd.Flags().MarkHidden("tbs")

	confDci11Cmd.Flags().String("rnti", "C-RNTI", "RNTI for DCI 1_1")
	confDci11Cmd.Flags().Int("muPdcch", 1, "Subcarrier spacing of PDCCH[0..3]")
	confDci11Cmd.Flags().Int("muPdsch", 1, "Subcarrier spacing of PDSCH[0..3]")
	confDci11Cmd.Flags().Int("actBwp", 1, "Active DL bandwidth part of PDSCH[0..1]")
	confDci11Cmd.Flags().Int("indicatedBwp", 1, "Bandwidth-part-indicator field of DCI 1_1[0..1]")
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
	confDci11Cmd.Flags().Int("tbs", 1277992, "Transport block size(bits) for PDSCH")
	confDci11Cmd.Flags().IntVar(&dci11DeltaPri, "dci11DeltaPri", 1, "PUCCH-resource-indicator field of DCI 1_1[0..4]")
	confDci11Cmd.Flags().IntVar(&dci11TdK1, "dci11TdK1", 2, "PDSCH-to-HARQ_feedback-timing-indicator(K1) field of DCI 1_1[0..7]")
	confDci11Cmd.Flags().IntVar(&dci11AntPorts, "dci11AntPorts", 10, "Antenna_port(s) field of DCI 1_1[0..15]")
	confDci11Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dci11.rnti", confDci11Cmd.Flags().Lookup("rnti"))
	viper.BindPFlag("nrrg.dci11.muPdcch", confDci11Cmd.Flags().Lookup("muPdcch"))
	viper.BindPFlag("nrrg.dci11.muPdsch", confDci11Cmd.Flags().Lookup("muPdsch"))
	viper.BindPFlag("nrrg.dci11.actBwp", confDci11Cmd.Flags().Lookup("actBwp"))
	viper.BindPFlag("nrrg.dci11.indicatedBwp", confDci11Cmd.Flags().Lookup("indicatedBwp"))
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
	viper.BindPFlag("nrrg.dci11.tbs", confDci11Cmd.Flags().Lookup("tbs"))
	viper.BindPFlag("nrrg.dci11.dci11DeltaPri", confDci11Cmd.Flags().Lookup("dci11DeltaPri"))
	viper.BindPFlag("nrrg.dci11.dci11TdK1", confDci11Cmd.Flags().Lookup("dci11TdK1"))
	viper.BindPFlag("nrrg.dci11.dci11AntPorts", confDci11Cmd.Flags().Lookup("dci11AntPorts"))
	confDci11Cmd.Flags().MarkHidden("rnti")
	confDci11Cmd.Flags().MarkHidden("muPdcch")
	confDci11Cmd.Flags().MarkHidden("muPdsch")
	confDci11Cmd.Flags().MarkHidden("actBwp")
	confDci11Cmd.Flags().MarkHidden("indicatedBwp")
	confDci11Cmd.Flags().MarkHidden("tbs")

	confMsg3Cmd.Flags().Int("muPusch", 1, "Subcarrier spacing of PUSCH[0..3]")
	confMsg3Cmd.Flags().IntVar(&msg3TdRa, "msg3TdRa", 6, "PUSCH-time-resource-allocation field of RAR UL grant scheduling Msg3[0..15]")
	confMsg3Cmd.Flags().String("tdMappingType", "typeB", "Mapping type for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().Int("tdK2", 1, "Slot offset K2 for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().Int("tdDelta", 3, "Slot offset delta for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().Int("tdSliv", 74, "SLIV for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().Int("tdStartSymb", 4, "Starting symbol S for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().Int("tdNumSymbs", 6, "Number of OFDM symbols L for Msg3 PUSCH time-domain allocation")
	confMsg3Cmd.Flags().String("fdRaType", "raType1", "resourceAllocation for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().StringVar(&msg3FdFreqHop, "msg3FdFreqHop", "enabled", "Frequency-hopping-flag field of RAR UL grant scheduling Msg3[disabled,enabled]")
	confMsg3Cmd.Flags().StringVar(&msg3FdRa, "msg3FdRa", "0100000100001101", "PUSCH-frequency-resource-allocation field of RAR UL grant scheduling Msg3")
	confMsg3Cmd.Flags().IntVar(&msg3FdStartRb, "msg3FdStartRb", 0, "RB_start of RIV for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().IntVar(&msg3FdNumRbs, "msg3FdNumRbs", 62, "L_RBs of RIV for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().Int("fdSecondHopFreqOff", 68, "Frequency offset of second hop for Msg3 PUSCH frequency-domain allocation")
	confMsg3Cmd.Flags().IntVar(&msg3McsCw0, "msg3McsCw0", 2, "MCS field of RAR UL grant scheduling Msg3[0..28]")
	confMsg3Cmd.Flags().Int("tbs", 1544, "Transport block size(bits) for PUSCH")
	confMsg3Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.msg3.muPusch", confMsg3Cmd.Flags().Lookup("muPusch"))
	viper.BindPFlag("nrrg.msg3.msg3TdRa", confMsg3Cmd.Flags().Lookup("msg3TdRa"))
	viper.BindPFlag("nrrg.msg3.tdMappingType", confMsg3Cmd.Flags().Lookup("tdMappingType"))
	viper.BindPFlag("nrrg.msg3.tdK2", confMsg3Cmd.Flags().Lookup("tdK2"))
	viper.BindPFlag("nrrg.msg3.tdDelta", confMsg3Cmd.Flags().Lookup("tdDelta"))
	viper.BindPFlag("nrrg.msg3.tdSliv", confMsg3Cmd.Flags().Lookup("tdSliv"))
	viper.BindPFlag("nrrg.msg3.tdStartSymb", confMsg3Cmd.Flags().Lookup("tdStartSymb"))
	viper.BindPFlag("nrrg.msg3.tdNumSymbs", confMsg3Cmd.Flags().Lookup("tdNumSymbs"))
	viper.BindPFlag("nrrg.msg3.fdRaType", confMsg3Cmd.Flags().Lookup("fdRaType"))
	viper.BindPFlag("nrrg.msg3.msg3FdFreqHop", confMsg3Cmd.Flags().Lookup("msg3FdFreqHop"))
	viper.BindPFlag("nrrg.msg3.msg3FdRa", confMsg3Cmd.Flags().Lookup("msg3FdRa"))
	viper.BindPFlag("nrrg.msg3.msg3FdStartRb", confMsg3Cmd.Flags().Lookup("msg3FdStartRb"))
	viper.BindPFlag("nrrg.msg3.msg3FdNumRbs", confMsg3Cmd.Flags().Lookup("msg3FdNumRbs"))
	viper.BindPFlag("nrrg.msg3.fdSecondHopFreqOff", confMsg3Cmd.Flags().Lookup("fdSecondHopFreqOff"))
	viper.BindPFlag("nrrg.msg3.msg3McsCw0", confMsg3Cmd.Flags().Lookup("msg3McsCw0"))
	viper.BindPFlag("nrrg.msg3.tbs", confMsg3Cmd.Flags().Lookup("tbs"))
	confMsg3Cmd.Flags().MarkHidden("muPusch")
	confMsg3Cmd.Flags().MarkHidden("tdMappingType")
	confMsg3Cmd.Flags().MarkHidden("tdK2")
	confMsg3Cmd.Flags().MarkHidden("tdDelta")
	confMsg3Cmd.Flags().MarkHidden("tdSliv")
	confMsg3Cmd.Flags().MarkHidden("tdStartSymb")
	confMsg3Cmd.Flags().MarkHidden("tdNumSymbs")
	confMsg3Cmd.Flags().MarkHidden("fdRaType")
	confMsg3Cmd.Flags().MarkHidden("fdSecondHopFreqOff")
	confMsg3Cmd.Flags().MarkHidden("tbs")

	confDci01Cmd.Flags().String("rnti", "C-RNTI", "RNTI for DCI 0_1")
	confDci01Cmd.Flags().Int("muPdcch", 1, "Subcarrier spacing of PDCCH[0..3]")
	confDci01Cmd.Flags().Int("muPusch", 1, "Subcarrier spacing of PUSCH[0..3]")
	confDci01Cmd.Flags().Int("actBwp", 1, "Active UL bandwidth part of PUSCH[0..1]")
	confDci01Cmd.Flags().Int("indicatedBwp", 1, "Bandwidth-part-indicator field of DCI 0_1[0..1]")
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
	confDci01Cmd.Flags().Int("tbs", 475584, "Transport block size(bits) for PUSCH")
	confDci01Cmd.Flags().IntVar(&dci01CbTpmiNumLayers, "dci01CbTpmiNumLayers", 2, "Precoding-information-and-number-of-layers field of DCI 0_1[0..63]")
	confDci01Cmd.Flags().StringVar(&dci01Sri, "dci01Sri", "", "SRS-resource-indicator field of DCI 0_1")
	confDci01Cmd.Flags().IntVar(&dci01AntPorts, "dci01AntPorts", 0, "Antenna_port(s) field of DCI 0_1[0..7]")
	confDci01Cmd.Flags().IntVar(&dci01PtrsDmrsMap, "dci01PtrsDmrsMap", 0, "PTRS-DMRS-association field of DCI 0_1[0..3]")
	confDci01Cmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dci01.rnti", confDci01Cmd.Flags().Lookup("rnti"))
	viper.BindPFlag("nrrg.dci01.muPdcch", confDci01Cmd.Flags().Lookup("muPdcch"))
	viper.BindPFlag("nrrg.dci01.muPusch", confDci01Cmd.Flags().Lookup("muPusch"))
	viper.BindPFlag("nrrg.dci01.actBwp", confDci01Cmd.Flags().Lookup("actBwp"))
	viper.BindPFlag("nrrg.dci01.indicatedBwp", confDci01Cmd.Flags().Lookup("indicatedBwp"))
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
	viper.BindPFlag("nrrg.dci01.tbs", confDci01Cmd.Flags().Lookup("tbs"))
	viper.BindPFlag("nrrg.dci01.dci01CbTpmiNumLayers", confDci01Cmd.Flags().Lookup("dci01CbTpmiNumLayers"))
	viper.BindPFlag("nrrg.dci01.dci01Sri", confDci01Cmd.Flags().Lookup("dci01Sri"))
	viper.BindPFlag("nrrg.dci01.dci01AntPorts", confDci01Cmd.Flags().Lookup("dci01AntPorts"))
	viper.BindPFlag("nrrg.dci01.dci01PtrsDmrsMap", confDci01Cmd.Flags().Lookup("dci01PtrsDmrsMap"))
	confDci01Cmd.Flags().MarkHidden("rnti")
	confDci01Cmd.Flags().MarkHidden("muPdcch")
	confDci01Cmd.Flags().MarkHidden("muPusch")
	confDci01Cmd.Flags().MarkHidden("actBwp")
	confDci01Cmd.Flags().MarkHidden("indicatedBwp")
	confDci01Cmd.Flags().MarkHidden("tbs")

	confBwpCmd.Flags().StringSlice("bwpType", []string{"iniDlBwp", "dedDlBwp", "iniUlBwp", "dedUlBwp"}, "BWP type")
	confBwpCmd.Flags().IntSlice("bwpId", []int{0, 1, 0, 1}, "bwp-Id of BWP-Uplink or BWP-Downlink IE")
	confBwpCmd.Flags().StringSlice("bwpScs", []string{"30KHz", "30KHz", "30KHz", "30KHz"}, "subcarrierSpacing of BWP IE")
	confBwpCmd.Flags().StringSlice("bwpCp", []string{"normal", "normal", "normal", "normal"}, "cyclicPrefix of BWP IE")
	confBwpCmd.Flags().IntSliceVar(&bwpLocAndBw, "bwpLocAndBw", []int{12925, 1099, 1099, 1099}, "locationAndBandwidth of BWP IE")
	confBwpCmd.Flags().IntSliceVar(&bwpStartRb, "bwpStartRb", []int{0, 0, 0, 0}, "RB_start of BWP")
	confBwpCmd.Flags().IntSliceVar(&bwpNumRbs, "bwpNumRbs", []int{48, 273, 273, 273}, "L_RBs of BWP")
	confBwpCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.bwp.bwpType", confBwpCmd.Flags().Lookup("bwpType"))
	viper.BindPFlag("nrrg.bwp.bwpId", confBwpCmd.Flags().Lookup("bwpId"))
	viper.BindPFlag("nrrg.bwp.bwpScs", confBwpCmd.Flags().Lookup("bwpScs"))
	viper.BindPFlag("nrrg.bwp.bwpCp", confBwpCmd.Flags().Lookup("bwpCp"))
	viper.BindPFlag("nrrg.bwp.bwpLocAndBw", confBwpCmd.Flags().Lookup("bwpLocAndBw"))
	viper.BindPFlag("nrrg.bwp.bwpStartRb", confBwpCmd.Flags().Lookup("bwpStartRb"))
	viper.BindPFlag("nrrg.bwp.bwpNumRbs", confBwpCmd.Flags().Lookup("bwpNumRbs"))
	confBwpCmd.Flags().MarkHidden("bwpType")
	confBwpCmd.Flags().MarkHidden("bwpId")
	confBwpCmd.Flags().MarkHidden("bwpScs")
	confBwpCmd.Flags().MarkHidden("bwpCp")

	confRachCmd.Flags().IntVar(&prachConfId, "prachConfId", 148, "prach-ConfigurationIndex of RACH-ConfigGeneric[0..255]")
	confRachCmd.Flags().String("raFormat", "B4", "Preamble format")
	confRachCmd.Flags().Int("raX", 2, "The x in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().IntSlice("raY", []int{1}, "The y in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().IntSlice("raSubfNumFr1SlotNumFr2", []int{9}, "The Subframe-number in 3GPP TS 38.211 Table 6.3.3.2-2 and Table 6.3.3.2-3, or the Slot-number in Table 6.3.3.2-4")
	confRachCmd.Flags().Int("raStartingSymb", 0, "The Starting-symbol in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().Int("raNumSlotsPerSubfFr1Per60KSlotFr2", 1, "The Number-of-PRACH-slots-within-a-subframe in 3GPP TS 38.211 Table 6.3.3.2-2 and Table 6.3.3.2-3, or the Number-of-PRACH-slots-within-a-60-kHz-slot in Table 6.3.3.2-4")
	confRachCmd.Flags().Int("raNumOccasionsPerSlot", 1, "The Number-of-time-domain-PRACH-occasions-within-a-PRACH-slot in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().Int("raDuration", 12, "The PRACH-duration in 3GPP TS 38.211 Table 6.3.3.2-2, Table 6.3.3.2-3 and Table 6.3.3.2-4")
	confRachCmd.Flags().StringVar(&msg1Scs, "msg1Scs", "30KHz", "msg1-SubcarrierSpacing of RACH-ConfigCommon")
	confRachCmd.Flags().IntVar(&msg1Fdm, "msg1Fdm", 1, "msg1-FDM of RACH-ConfigGeneric[1,2,4,8]")
	confRachCmd.Flags().IntVar(&msg1FreqStart, "msg1FreqStart", 0, "msg1-FrequencyStart of RACH-ConfigGeneric[0..274]")
	confRachCmd.Flags().StringVar(&raRespWin, "raRespWin", "sl20", "ra-ResponseWindow of RACH-ConfigGeneric[sl1,sl2,sl4,sl8,sl10,sl20,sl40,sl80]")
	confRachCmd.Flags().IntVar(&totNumPreambs, "totNumPreambs", 64, "totalNumberOfRA-Preambles of RACH-ConfigCommon[1..64]")
	confRachCmd.Flags().StringVar(&ssbPerRachOccasion, "ssbPerRachOccasion", "one", "ssb-perRACH-Occasion of RACH-ConfigGeneric[oneEighth,oneFourth,oneHalf,one,two,four,eight,sixteen]")
	confRachCmd.Flags().IntVar(&cbPreambsPerSsb, "cbPreambsPerSsb", 64, "cb-PreamblesPerSSB of RACH-ConfigCommon[depends on ssbPerRachOccasion]")
	confRachCmd.Flags().StringVar(&contResTimer, "contResTimer", "sf64", "ra-ContentionResolutionTimer of RACH-ConfigGeneric[sf8,sf16,sf24,sf32,sf40,sf48,sf56,sf64]")
	confRachCmd.Flags().StringVar(&msg3Tp, "msg3Tp", "disabled", "msg3-transformPrecoder of RACH-ConfigGeneric[disabled,enabled]")
	confRachCmd.Flags().Int("raLen", 139, "L_RA of 3GPP TS 38.211 Table 6.3.3.1-1 and Table 6.3.3.1-2")
	confRachCmd.Flags().Int("raNumRbs", 12, "Allocation-expressed-in-number-of-RBs-for-PUSCH of 3GPP TS 38.211 Table 6.3.3.2-1")
	confRachCmd.Flags().Int("raKBar", 2, "k_bar of 3GPP TS 38.211 Table 6.3.3.2-1")
	confRachCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.rach.prachConfId", confRachCmd.Flags().Lookup("prachConfId"))
	viper.BindPFlag("nrrg.rach.raFormat", confRachCmd.Flags().Lookup("raFormat"))
	viper.BindPFlag("nrrg.rach.raX", confRachCmd.Flags().Lookup("raX"))
	viper.BindPFlag("nrrg.rach.raY", confRachCmd.Flags().Lookup("raY"))
	viper.BindPFlag("nrrg.rach.raSubfNumFr1SlotNumFr2", confRachCmd.Flags().Lookup("raSubfNumFr1SlotNumFr2"))
	viper.BindPFlag("nrrg.rach.raStartingSymb", confRachCmd.Flags().Lookup("raStartingSymb"))
	viper.BindPFlag("nrrg.rach.raNumSlotsPerSubfFr1Per60KSlotFr2", confRachCmd.Flags().Lookup("raNumSlotsPerSubfFr1Per60KSlotFr2"))
	viper.BindPFlag("nrrg.rach.raNumOccasionsPerSlot", confRachCmd.Flags().Lookup("raNumOccasionsPerSlot"))
	viper.BindPFlag("nrrg.rach.raDuration", confRachCmd.Flags().Lookup("raDuration"))
	viper.BindPFlag("nrrg.rach.msg1Scs", confRachCmd.Flags().Lookup("msg1Scs"))
	viper.BindPFlag("nrrg.rach.msg1Fdm", confRachCmd.Flags().Lookup("msg1Fdm"))
	viper.BindPFlag("nrrg.rach.msg1FreqStart", confRachCmd.Flags().Lookup("msg1FreqStart"))
	viper.BindPFlag("nrrg.rach.raRespWin", confRachCmd.Flags().Lookup("raRespWin"))
	viper.BindPFlag("nrrg.rach.totNumPreambs", confRachCmd.Flags().Lookup("totNumPreambs"))
	viper.BindPFlag("nrrg.rach.ssbPerRachOccasion", confRachCmd.Flags().Lookup("ssbPerRachOccasion"))
	viper.BindPFlag("nrrg.rach.cbPreambsPerSsb", confRachCmd.Flags().Lookup("cbPreambsPerSsb"))
	viper.BindPFlag("nrrg.rach.contResTimer", confRachCmd.Flags().Lookup("contResTimer"))
	viper.BindPFlag("nrrg.rach.msg3Tp", confRachCmd.Flags().Lookup("msg3Tp"))
	viper.BindPFlag("nrrg.rach.raLen", confRachCmd.Flags().Lookup("raLen"))
	viper.BindPFlag("nrrg.rach.raNumRbs", confRachCmd.Flags().Lookup("raNumRbs"))
	viper.BindPFlag("nrrg.rach.raKBar", confRachCmd.Flags().Lookup("raKBar"))
	confRachCmd.Flags().MarkHidden("raFormat")
	confRachCmd.Flags().MarkHidden("raX")
	confRachCmd.Flags().MarkHidden("raY")
	confRachCmd.Flags().MarkHidden("raSubfNumFr1SlotNumFr2")
	confRachCmd.Flags().MarkHidden("raStartingSymb")
	confRachCmd.Flags().MarkHidden("raNumSlotsPerSubfFr1Per60KSlotFr2")
	confRachCmd.Flags().MarkHidden("raNumOccasionsPerSlot")
	confRachCmd.Flags().MarkHidden("raDuration")
	confRachCmd.Flags().MarkHidden("raLen")
	confRachCmd.Flags().MarkHidden("raNumRbs")
	confRachCmd.Flags().MarkHidden("raKBar")

	confDmrsRaCmd.Flags().StringSlice("schInfo", []string{"SIB1", "Msg2", "Msg4", "Msg3"}, "Information of UL/DL-SCH")
	confDmrsRaCmd.Flags().StringSlice("dmrsType", []string{"type1", "type1", "type1", "type1"}, "dmrs-Type as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	confDmrsRaCmd.Flags().StringSlice("dmrsAddPos", []string{"pos0", "pos0", "pos0", "pos1"}, "dmrs-AdditionalPosition as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	confDmrsRaCmd.Flags().StringSlice("maxLength", []string{"len1", "len1", "len1", "len1"}, "maxLength as in DMRS-UplinkConfig/DMRS-DownlinkConfig")
	confDmrsRaCmd.Flags().IntSlice("dmrsPorts", []int{1000, 1000, 1000, 0}, "DMRS antenna ports")
	confDmrsRaCmd.Flags().IntSlice("cdmGroupsWoData", []int{1, 1, 1, 2}, "CDM group(s) without data")
	confDmrsRaCmd.Flags().IntSlice("numFrontLoadSymbs", []int{1, 1, 1, 1}, "Number of front-load DMRS symbols")
	// tdL for SIB1/Msg2/Msg4 is semicolon separated
	// tdL for Msg3 is semicolon separated if msg3FreqHop is disabled, otherwise, tdL is colon separated for each hop
	confDmrsRaCmd.Flags().StringSlice("tdL", []string{"0", "0", "0", "0:0"}, "Time-domain locations for DMRS")
	// fdK indicates REs per PRB for DMRS
	confDmrsRaCmd.Flags().StringSlice("fdK", []string{"101010101010", "101010101010", "101010101010", "111111111111"}, "Frequency-domain locations of DMRS")
	confDmrsRaCmd.Flags().SortFlags = false
	viper.BindPFlag("nrrg.dmrsra.schInfo", confDmrsRaCmd.Flags().Lookup("schInfo"))
	viper.BindPFlag("nrrg.dmrsra.dmrsType", confDmrsRaCmd.Flags().Lookup("dmrsType"))
	viper.BindPFlag("nrrg.dmrsra.dmrsAddPos", confDmrsRaCmd.Flags().Lookup("dmrsAddPos"))
	viper.BindPFlag("nrrg.dmrsra.maxLength", confDmrsRaCmd.Flags().Lookup("maxLength"))
	viper.BindPFlag("nrrg.dmrsra.dmrsPorts", confDmrsRaCmd.Flags().Lookup("dmrsPorts"))
	viper.BindPFlag("nrrg.dmrsra.cdmGroupsWoData", confDmrsRaCmd.Flags().Lookup("cdmGroupsWoData"))
	viper.BindPFlag("nrrg.dmrsra.numFrontLoadSymbs", confDmrsRaCmd.Flags().Lookup("numFrontLoadSymbs"))
	viper.BindPFlag("nrrg.dmrsra.tdL", confDmrsRaCmd.Flags().Lookup("tdL"))
	viper.BindPFlag("nrrg.dmrsra.fdK", confDmrsRaCmd.Flags().Lookup("fdK"))
	confDmrsRaCmd.Flags().MarkHidden("schInfo")
	confDmrsRaCmd.Flags().MarkHidden("dmrsType")
	confDmrsRaCmd.Flags().MarkHidden("dmrsAddPos")
	confDmrsRaCmd.Flags().MarkHidden("maxLength")
	confDmrsRaCmd.Flags().MarkHidden("dmrsPorts")
	confDmrsRaCmd.Flags().MarkHidden("cdmGroupsWoData")
	confDmrsRaCmd.Flags().MarkHidden("numFrontLoadSymbs")
	confDmrsRaCmd.Flags().MarkHidden("tdL")
	confDmrsRaCmd.Flags().MarkHidden("fdK")
}
