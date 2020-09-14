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
	msg3FdSecondHopFreqOff int
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
)

// nrrgCmd represents the nrrg command
var nrrgCmd = &cobra.Command{
	Use:   "nrrg [sub]",
	Short: "",
	Long: `nrrg generates NR(new radio) resource grid according to network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg called")
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
	nrrgConfCmd.AddCommand(confDci01Cmd)
	nrrgConfCmd.AddCommand(confDci11Cmd)
	nrrgConfCmd.AddCommand(confMsg3Cmd)
	nrrgConfCmd.AddCommand(confDci01Cmd)
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
	viper.BindPFlag("nrrg.css0.css0AggLevel", confCss0Cmd.Flags().Lookup("css0AggLevel"))
	viper.BindPFlag("nrrg.css0.css0NumCandidates", confCss0Cmd.Flags().Lookup("css0NumCandidates"))
}
