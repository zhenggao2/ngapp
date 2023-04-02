/*
Copyright © 2020 Zhengwei Gao<28912001@qq.com>

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
	"github.com/zhenggao2/ngapp/biptrace"
	"github.com/zhenggao2/ngapp/ddr4trace"
	"github.com/zhenggao2/ngapp/l2trace"
	"github.com/zhenggao2/ngapp/ttitrace"
)

var (
	tlog     string
	py2      string
	py3      string
	tlda     string
	snaptool string
	luashark string
	wshark   string
	trace    string
	pattern  string
	rat      string
	scs      string
	chbw     string
	gain     int
	filter   string
	ttidec   string
)

// ttiCmd represents the tti command
var ttiCmd = &cobra.Command{
	Use:   "tti",
	Short: "L2TtiTrace analysis tool",
	Long:  `The tti module decodes/parses L2TtiTrace.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadTtiFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()

		if tlog == "l2tti" && (pattern == ".csv" || pattern == ".bin") {
			// .bin is raw L2TtiTrace from either Snapshot or gnb_logs
			// .csv is output from L2TtiTrace EventDecoder
			tti := new(ttitrace.L2TtiTraceParser)
			tti.Init(Logger, py3, ttidec, trace, pattern, rat, scs, filter, maxgo, debug)
			tti.Exec()
		} else {
			fmt.Printf("Unsupported tlog[=%s] or pattern[=%s].\n", tlog, pattern)
		}
	},
}

// bipCmd represents the bip command
var bipCmd = &cobra.Command{
	Use:   "bip",
	Short: "BIP analysis tool",
	Long:  `The bip module parses BIP and calculates noisePower.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadBipFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()

		if tlog == "bip" && pattern == ".pcap" {
			// .pcap is raw BIP from gnb_logs
			bip := new(biptrace.BipTraceParser)
			bip.Init(Logger, luashark, wshark, trace, pattern, scs, chbw, maxgo, debug)
			bip.Exec()
		} else {
			fmt.Printf("Unsupported tlog[=%s] or pattern[=%s].\n", tlog, pattern)
		}
	},
}

// ddr4Cmd represents the ddr4 command
var ddr4Cmd = &cobra.Command{
	Use:   "ddr4",
	Short: "DDR4 analysis tool",
	Long:  `The ddr4 module parses DDR4 and calculates RSSI.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadDdr4Flags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()

		if tlog == "ddr4" && pattern == ".bin" {
			// .bin is raw DDR4 from WebEM or gnb_logs
			ddr4 := new(ddr4trace.Ddr4TraceParser)
			ddr4.Init(Logger, py3, snaptool, trace, pattern, scs, chbw, filter, maxgo, gain, debug)
			ddr4.Exec()
		} else {
			fmt.Printf("Unsupported tlog[=%s] or pattern[=%s].\n", tlog, pattern)
		}
	},
}

// l2traceCmd represents the l2trace command
var l2traceCmd = &cobra.Command{
	Use:   "l2trace",
	Short: "L2Trace analysis tool",
	Long:  `The l2trace module parses L2Trace.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadL2TraceFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()

		if tlog == "l2trace" && (pattern == ".dat" || pattern == ".pcap") {
			// .dat is raw L2Trace from Snapshot
			// .pcap is raw L2Trace from gnb_logs or DCAP
			l2trace := new(l2trace.L2TraceParser)
			l2trace.Init(Logger, py2, tlda, luashark, wshark, trace, pattern, maxgo, debug)
			l2trace.Exec()
		} else {
			fmt.Printf("Unsupported tlog[=%s] or pattern[=%s].\n", tlog, pattern)
		}
	},
}

// from <Effective Go>
/*
The init function

Finally, each source file can define its own niladic init function to set up whatever state is required. (Actually each file can have multiple init functions.)
And finally means finally: init is called after all the variable declarations in the package have evaluated their initializers, and those are evaluated only after all the imported packages have been initialized.

Besides initializations that cannot be expressed as declarations, a common use of init functions is to verify or repair correctness of the program state before real execution begins.
*/
func init() {
	if cmdFlags&CMD_FLAG_TTI != 0 {
		rootCmd.AddCommand(ttiCmd)
	}
	if cmdFlags&CMD_FLAG_BIP != 0 {
		rootCmd.AddCommand(bipCmd)
	}
	if cmdFlags&CMD_FLAG_DDR4 != 0 {
		rootCmd.AddCommand(ddr4Cmd)
	}
	if cmdFlags&CMD_FLAG_L2TRACE != 0 {
		rootCmd.AddCommand(l2traceCmd)
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cmd.Flags().StringP("trace", "d", "./trace_path", "path containing tti files")

	// tti.Init(Logger, trace, pattern, rat, scs, filter, maxgo, debug)
	ttiCmd.Flags().StringVar(&tlog, "tlog", "l2tti", "type of traces[l2tti,l2trace,bip,ddr4]")
	ttiCmd.Flags().StringVar(&py3, "py3", "C:/Python38", "path of Python3")
	ttiCmd.Flags().StringVar(&ttidec, "ttidec", "C:/tti-dec-bin", "path of tti-dec-bin tool")
	ttiCmd.Flags().StringVar(&trace, "trace", "./data", "path containing trace files")
	ttiCmd.Flags().StringVar(&pattern, "pattern", ".csv", "pattern of trace files[.csv,.pcap,.dat,.bin]")
	ttiCmd.Flags().StringVar(&rat, "rat", "nr", "RAT info of traces[nr]")
	ttiCmd.Flags().StringVar(&scs, "scs", "30k", "NRCELLGRP/scs setting[15k,30k,120k]")
	ttiCmd.Flags().StringVar(&filter, "filter", "both", "ul/dl tti filter[ul,dl,both]")
	ttiCmd.Flags().IntVar(&maxgo, "maxgo", 3, "maximum number of concurrent goroutines(tune me in case of 'out of memory' issue!)[2..numCPU]")
	ttiCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("tti.tlog", ttiCmd.Flags().Lookup("tlog"))
	viper.BindPFlag("tti.py3", ttiCmd.Flags().Lookup("py3"))
	viper.BindPFlag("tti.ttidec", ttiCmd.Flags().Lookup("ttidec"))
	viper.BindPFlag("tti.trace", ttiCmd.Flags().Lookup("trace"))
	viper.BindPFlag("tti.pattern", ttiCmd.Flags().Lookup("pattern"))
	viper.BindPFlag("tti.rat", ttiCmd.Flags().Lookup("rat"))
	viper.BindPFlag("tti.scs", ttiCmd.Flags().Lookup("scs"))
	viper.BindPFlag("tti.filter", ttiCmd.Flags().Lookup("filter"))
	viper.BindPFlag("tti.maxgo", ttiCmd.Flags().Lookup("maxgo"))
	viper.BindPFlag("tti.debug", ttiCmd.Flags().Lookup("debug"))

	// bip.Init(Logger, luashark, wshark, trace, pattern, scs, chbw, maxgo, debug)
	bipCmd.Flags().StringVar(&tlog, "tlog", "l2tti", "type of traces[l2tti,l2trace,bip,ddr4]")
	bipCmd.Flags().StringVar(&luashark, "luashark", "C:/luashark", "path of luashark scripts")
	bipCmd.Flags().StringVar(&wshark, "wshark", "C:/Program Files/Wireshark", "path of tshark")
	bipCmd.Flags().StringVar(&trace, "trace", "./data", "path containing trace files")
	bipCmd.Flags().StringVar(&pattern, "pattern", ".csv", "pattern of trace files[.csv,.pcap,.dat,.bin]")
	bipCmd.Flags().StringVar(&scs, "scs", "30k", "NRCELLGRP/scs setting[15k,30k,120k]")
	bipCmd.Flags().StringVar(&chbw, "chbw", "30m", "NRCELL/chBw or NRCELL_FDD/chBwDl(chBwUl) setting[20m,30m,100m]")
	bipCmd.Flags().IntVar(&maxgo, "maxgo", 3, "maximum number of concurrent goroutines(tune me in case of 'out of memory' issue!)[2..numCPU]")
	bipCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("bip.tlog", bipCmd.Flags().Lookup("tlog"))
	viper.BindPFlag("bip.luashark", bipCmd.Flags().Lookup("luashark"))
	viper.BindPFlag("bip.wshark", bipCmd.Flags().Lookup("wshark"))
	viper.BindPFlag("bip.trace", bipCmd.Flags().Lookup("trace"))
	viper.BindPFlag("bip.pattern", bipCmd.Flags().Lookup("pattern"))
	viper.BindPFlag("bip.scs", bipCmd.Flags().Lookup("scs"))
	viper.BindPFlag("bip.chbw", bipCmd.Flags().Lookup("chbw"))
	viper.BindPFlag("bip.maxgo", bipCmd.Flags().Lookup("maxgo"))
	viper.BindPFlag("bip.debug", bipCmd.Flags().Lookup("debug"))

	// ddr4.Init(Logger, py3, snaptool, trace, pattern, scs, chbw, filter, maxgo, gain, debug)
	ddr4Cmd.Flags().StringVar(&tlog, "tlog", "l2tti", "type of traces[l2tti,l2trace,bip,ddr4]")
	ddr4Cmd.Flags().StringVar(&py3, "py3", "C:/Python38", "path of Python3")
	ddr4Cmd.Flags().StringVar(&snaptool, "snaptool", "C:/snapshot_tool", "path of Loki snapshot tool")
	ddr4Cmd.Flags().StringVar(&trace, "trace", "./data", "path containing trace files")
	ddr4Cmd.Flags().StringVar(&pattern, "pattern", ".csv", "pattern of trace files[.csv,.pcap,.dat,.bin]")
	ddr4Cmd.Flags().StringVar(&scs, "scs", "30k", "NRCELLGRP/scs setting[15k,30k,120k]")
	ddr4Cmd.Flags().StringVar(&chbw, "chbw", "30m", "NRCELL/chBw or NRCELL_FDD/chBwDl(chBwUl) setting[20m,30m,100m]")
	ddr4Cmd.Flags().IntVar(&gain, "gain", 105, "gain assumption in unit of dB")
	ddr4Cmd.Flags().StringVar(&filter, "filter", "both", "ul/dl tti filter[ul,dl,both]")
	ddr4Cmd.Flags().IntVar(&maxgo, "maxgo", 3, "maximum number of concurrent goroutines(tune me in case of 'out of memory' issue!)[2..numCPU]")
	ddr4Cmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("ddr4.tlog", ddr4Cmd.Flags().Lookup("tlog"))
	viper.BindPFlag("ddr4.py3", ddr4Cmd.Flags().Lookup("py3"))
	viper.BindPFlag("ddr4.snaptool", ddr4Cmd.Flags().Lookup("snaptool"))
	viper.BindPFlag("ddr4.trace", ddr4Cmd.Flags().Lookup("trace"))
	viper.BindPFlag("ddr4.pattern", ddr4Cmd.Flags().Lookup("pattern"))
	viper.BindPFlag("ddr4.scs", ddr4Cmd.Flags().Lookup("scs"))
	viper.BindPFlag("ddr4.chbw", ddr4Cmd.Flags().Lookup("chbw"))
	viper.BindPFlag("ddr4.gain", ddr4Cmd.Flags().Lookup("gain"))
	viper.BindPFlag("ddr4.filter", ddr4Cmd.Flags().Lookup("filter"))
	viper.BindPFlag("ddr4.maxgo", ddr4Cmd.Flags().Lookup("maxgo"))
	viper.BindPFlag("ddr4.debug", ddr4Cmd.Flags().Lookup("debug"))

	// l2trace.Init(Logger, py2, tlda, luashark, wshark, trace, pattern, maxgo, debug)
	l2traceCmd.Flags().StringVar(&tlog, "tlog", "l2tti", "type of traces[l2tti,l2trace,bip,ddr4]")
	l2traceCmd.Flags().StringVar(&py2, "py2", "C:/Python27", "path of Python2")
	l2traceCmd.Flags().StringVar(&tlda, "tlda", "C:/TLDA", "path of TLDA")
	l2traceCmd.Flags().StringVar(&luashark, "luashark", "C:/luashark", "path of luashark scripts")
	l2traceCmd.Flags().StringVar(&wshark, "wshark", "C:/Program Files/Wireshark", "path of tshark")
	l2traceCmd.Flags().StringVar(&trace, "trace", "./data", "path containing trace files")
	l2traceCmd.Flags().StringVar(&pattern, "pattern", ".csv", "pattern of trace files[.csv,.pcap,.dat,.bin]")
	l2traceCmd.Flags().IntVar(&maxgo, "maxgo", 3, "maximum number of concurrent goroutines(tune me in case of 'out of memory' issue!)[2..numCPU]")
	l2traceCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("l2trace.tlog", l2traceCmd.Flags().Lookup("tlog"))
	viper.BindPFlag("l2trace.py2", l2traceCmd.Flags().Lookup("py2"))
	viper.BindPFlag("l2trace.tlda", l2traceCmd.Flags().Lookup("tlda"))
	viper.BindPFlag("l2trace.luashark", l2traceCmd.Flags().Lookup("luashark"))
	viper.BindPFlag("l2trace.wshark", l2traceCmd.Flags().Lookup("wshark"))
	viper.BindPFlag("l2trace.trace", l2traceCmd.Flags().Lookup("trace"))
	viper.BindPFlag("l2trace.pattern", l2traceCmd.Flags().Lookup("pattern"))
	viper.BindPFlag("l2trace.maxgo", l2traceCmd.Flags().Lookup("maxgo"))
	viper.BindPFlag("l2trace.debug", l2traceCmd.Flags().Lookup("debug"))
}

func loadTtiFlags() {
	// tti.Init(Logger, py3, ttidec, trace, pattern, rat, scs, filter, maxgo, debug)
	tlog = viper.GetString("tti.tlog")
	py3 = viper.GetString("tti.py3")
	ttidec = viper.GetString("tti.ttidec")
	trace = viper.GetString("tti.trace")
	pattern = viper.GetString("tti.pattern")
	rat = viper.GetString("tti.rat")
	scs = viper.GetString("tti.scs")
	filter = viper.GetString("tti.filter")
	maxgo = viper.GetInt("tti.maxgo")
	debug = viper.GetBool("tti.debug")
}

func loadBipFlags() {
	// bip.Init(Logger, luashark, wshark, trace, pattern, scs, chbw, maxgo, debug)
	tlog = viper.GetString("bip.tlog")
	luashark = viper.GetString("bip.luashark")
	wshark = viper.GetString("bip.wshark")
	trace = viper.GetString("bip.trace")
	pattern = viper.GetString("bip.pattern")
	scs = viper.GetString("bip.scs")
	chbw = viper.GetString("bip.chbw")
	maxgo = viper.GetInt("bip.maxgo")
	debug = viper.GetBool("bip.debug")
}

func loadDdr4Flags() {
	// ddr4.Init(Logger, py3, snaptool, trace, pattern, scs, chbw, filter, maxgo, gain, debug)
	tlog = viper.GetString("ddr4.tlog")
	py3 = viper.GetString("ddr4.py3")
	snaptool = viper.GetString("ddr4.snaptool")
	trace = viper.GetString("ddr4.trace")
	pattern = viper.GetString("ddr4.pattern")
	scs = viper.GetString("ddr4.scs")
	chbw = viper.GetString("ddr4.chbw")
	gain = viper.GetInt("ddr4.gain")
	filter = viper.GetString("ddr4.filter")
	maxgo = viper.GetInt("ddr4.maxgo")
	debug = viper.GetBool("ddr4.debug")
}

func loadL2TraceFlags() {
	// l2trace.Init(Logger, py2, tlda, luashark, wshark, trace, pattern, maxgo, debug)
	tlog = viper.GetString("l2trace.tlog")
	py2 = viper.GetString("l2trace.py2")
	tlda = viper.GetString("l2trace.tlda")
	luashark = viper.GetString("l2trace.luashark")
	wshark = viper.GetString("l2trace.wshark")
	trace = viper.GetString("l2trace.trace")
	pattern = viper.GetString("l2trace.pattern")
	maxgo = viper.GetInt("l2trace.maxgo")
	debug = viper.GetBool("l2trace.debug")
}
