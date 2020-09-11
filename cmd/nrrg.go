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
	},
}

// confFreqBandCmd represents the 'nrrg conf freqband' command
var confFreqBandCmd = &cobra.Command{
	Use:   "freqband [sub]",
	Short: "",
	Long: `'nrrg conf freqband' can be used to get/set frequency-band related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg conf freqband called")
	},
}

// confSsbGridCmd represents the 'nrrg conf ssbgrid' command
var confSsbGridCmd = &cobra.Command{
	Use:   "ssbgrid [sub]",
	Short: "",
	Long: `'nrrg conf ssbgrid' can be used to get/set SSB-grid related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg conf ssbgrid called")
	},
}

// confSsbBurstCmd represents the 'nrrg conf ssbburst' command
var confSsbBurstCmd = &cobra.Command{
	Use:   "ssbburst [sub]",
	Short: "",
	Long: `'nrrg conf ssbburst' can be used to get/set SSB-burst related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg conf ssbburst called")
	},
}

// confMibCmd represents the 'nrrg conf mib' command
var confMibCmd = &cobra.Command{
	Use:   "mib [sub]",
	Short: "",
	Long: `'nrrg conf mib' can be used to get/set MIB related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg conf mib called")
	},
}

// confCarrierGridCmd represents the 'nrrg conf carriergrid' command
var confCarrierGridCmd = &cobra.Command{
	Use:   "carriergrid [sub]",
	Short: "",
	Long: `'nrrg conf carriergrid' can be used to get/set carrier-grid related network configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg conf carriergrid called")
	},
}

// nrrgSimCmd represents the 'nrrg sim' command
var nrrgSimCmd = &cobra.Command{
	Use:   "sim",
	Short: "",
	Long: `'nrrg sim' can be used to perform NR-Uu simulation.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nrrg sim called")
	},
}

func init() {
	nrrgConfCmd.AddCommand(confFreqBandCmd)
	nrrgConfCmd.AddCommand(confSsbGridCmd)
	nrrgConfCmd.AddCommand(confSsbBurstCmd)
	nrrgConfCmd.AddCommand(confMibCmd)
	nrrgConfCmd.AddCommand(confCarrierGridCmd)
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

	confSsbGridCmd.Flags().StringVar(&ssbScs, "ssbScs",  "30KHz", "SSB subcarrier spacing")
	confSsbGridCmd.Flags().String("ssbPattern", "Case C", "SSB pattern")
	confSsbGridCmd.Flags().IntVar(&kSsb, "kSsb", 0, "k_SSB")
	confSsbGridCmd.Flags().Int("nCrbSsb", 32, "n_CRB_SSB")
	viper.BindPFlag("nrrg.ssbGrid.ssbScs", confSsbGridCmd.Flags().Lookup("ssbScs"))
	viper.BindPFlag("nrrg.ssbGrid.ssbPattern", confSsbGridCmd.Flags().Lookup("ssbPattern"))
	viper.BindPFlag("nrrg.ssbGrid.kSsb", confSsbGridCmd.Flags().Lookup("kSsb"))
	viper.BindPFlag("nrrg.ssbGrid.nCrbSsb", confSsbGridCmd.Flags().Lookup("nCrbSsb"))

	confSsbBurstCmd.Flags().Int("maxL", 8, "max_L")
	confSsbBurstCmd.Flags().StringVar(&inOneGrp, "inOneGroup", "11111111", "inOneGroup of ssb-PositionsInBurst")
	confSsbBurstCmd.Flags().StringVar(&grpPresence, "groupPresence", "", "groupPresence of ssb-PositionsInBurst")
	confSsbBurstCmd.Flags().StringVar(&ssbPeriod, "ssbPeriod", "20ms", "ssb-PeriodicityServingCell")
	viper.BindPFlag("nrrg.ssbBurst.maxL", confSsbBurstCmd.Flags().Lookup("maxL"))
	viper.BindPFlag("nrrg.ssbBurst.inOneGroup", confSsbBurstCmd.Flags().Lookup("inOneGroup"))
	viper.BindPFlag("nrrg.ssbBurst.groupPresence", confSsbBurstCmd.Flags().Lookup("groupPresence"))
	viper.BindPFlag("nrrg.ssbBurst.ssbPeriod", confSsbBurstCmd.Flags().Lookup("ssbPeriod"))

	confMibCmd.Flags().IntVar(&sfn, "sfn", 0, "System frame number(SFN)")
	confMibCmd.Flags().IntVar(&hrf, "hrf", 0, "Half frame bit")
	confMibCmd.Flags().StringVar(&dmrsTypeAPos, "dmrsTypeAPos", "pos2", "dmrs-TypeA-Position")
	confMibCmd.Flags().StringVar(&commonScs, "commonScs", "30KHz", "subCarrierSpacingCommon")
	confMibCmd.Flags().IntVar(&rmsiCoreset0, "rmsiCoreset0", 12, "coresetZero of PDCCH-ConfigSIB1")
	confMibCmd.Flags().IntVar(&rmsiCss0, "rmsiCss0", 0, "searchSpaceZero of PDCCH-ConfigSIB1")
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

	confCarrierGridCmd.Flags().StringVar(&carrierScs, "carrierScs", "30KHz", "subcarrierSpacing of SCS-SpecificCarrier")
	confCarrierGridCmd.Flags().StringVar(&bw, "bw", "100MHz", "Transmission bandwidth(MHz)")
	confCarrierGridCmd.Flags().Int("carrierNumRbs", 273, "carrierBandwidth(N_RB) of SCS-SpecificCarrier")
	confCarrierGridCmd.Flags().Int("offsetToCarrier", 0, "offsetToCarrier of SCS-SpecificCarrier")
	viper.BindPFlag("nrrg.carrierGrid.carrierScs", confCarrierGridCmd.Flags().Lookup("carrierScs"))
	viper.BindPFlag("nrrg.carrierGrid.bw", confCarrierGridCmd.Flags().Lookup("bw"))
	viper.BindPFlag("nrrg.carrierGrid.carrierNumRbs", confCarrierGridCmd.Flags().Lookup("carrierNumRbs"))
	viper.BindPFlag("nrrg.carrierGrid.offsetToCarrier", confCarrierGridCmd.Flags().Lookup("offsetToCarrier"))
}
