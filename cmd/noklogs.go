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
	"github.com/zhenggao2/ngapp/ddr4trace"
	"github.com/zhenggao2/ngapp/ttitrace"
	"github.com/zhenggao2/ngapp/l2trace"
	"github.com/zhenggao2/ngapp/biptrace"
)

var (
	tlog     string
	py2      string
	py3 string
	tlda     string
	snaptool string
	luashark string
	wshark   string
	trace    string
	pattern  string
	rat      string
	scs      string
	chbw string
	nap int
	filter   string
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "L2TtiTrace/L2Trace/BIP/DDR4 analysis tool",
	Long:  `The logs module parses L2TtiTrace/L2Trace/BIP/DDR4.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadLogsFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()

		if tlog == "l2tti" && (pattern == ".csv" || pattern == ".bin") {
			// .bin is raw L2TtiTrace from either Snapshot or gnb_logs
			// .csv is output from L2TtiTrace EventDecoder
			tti := new(ttitrace.L2TtiTraceParser)
			tti.Init(Logger, trace, pattern, rat, scs, filter, maxgo, debug)
			tti.Exec()
		} else if tlog == "l2trace" && (pattern == ".dat" || pattern == ".pcap") {
			// .dat is raw L2Trace from Snapshot
			// .pcap is raw L2Trace from gnb_logs or DCAP
			parser := new(l2trace.L2TraceParser)
			parser.Init(Logger, py2, tlda, luashark, wshark, trace, pattern, maxgo, debug)
			parser.Exec()
		} else if tlog == "bip" && pattern == ".pcap" {
			// .pcap is raw BIP from gnb_logs
			parser := new(biptrace.BipTraceParser)
			parser.Init(Logger, luashark, wshark, trace, pattern, maxgo, debug)
			parser.Exec()
		} else if tlog == "ddr4" && pattern == ".bin"{
			// .bin is raw DDR4 from WebEM or gnb_logs
			parser := new(ddr4trace.Ddr4TraceParser)
			parser.Init(Logger, py3, snaptool, trace, pattern, maxgo, debug)
			parser.Exec()
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
	rootCmd.AddCommand(logsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//logsCmd.Flags().StringP("trace", "d", "./trace_path", "path containing tti files")
	logsCmd.Flags().StringVar(&tlog, "tlog", "l2tti", "type of traces[l2tti,l2trace,bip,ddr4]")
	logsCmd.Flags().StringVar(&py2, "py2", "C:/Python27", "path of Python2")
	logsCmd.Flags().StringVar(&py3, "py3", "C:/Python38", "path of Python3")
	logsCmd.Flags().StringVar(&tlda, "tlda", "C:/TLDA", "path of TLDA")
	logsCmd.Flags().StringVar(&snaptool, "snaptool", "C:/snapshot_tool", "path of Loki snapshot tool")
	logsCmd.Flags().StringVar(&luashark, "luashark", "C:/luashark", "path of luashark scripts")
	logsCmd.Flags().StringVar(&wshark, "wshark", "C:/Program Files/Wireshark", "path of Tshark")
	logsCmd.Flags().StringVar(&trace, "trace", "./data", "path containing trace files")
	logsCmd.Flags().StringVar(&pattern, "pattern", ".csv", "pattern of trace files[.csv,.pcap,.dat,.bin]")
	logsCmd.Flags().StringVar(&rat, "rat", "nr", "RAT info of traces[nr]")
	logsCmd.Flags().StringVar(&scs, "scs", "30khz", "NRCELLGRP/scs setting[15khz,30khz,120khz]")
	logsCmd.Flags().StringVar(&chbw, "chbw", "30MHz", "NRCELL/chBw or NRCELL_FDD/chBwDl(chBwUl) setting[20MHz,30MHz,100MHz]")
	logsCmd.Flags().IntVar(&nap, "nap", 4, "Number of antenna ports of gNB[2,4]")
	logsCmd.Flags().StringVar(&filter, "filter", "both", "ul/dl tti filter[ul,dl,both]")
	logsCmd.Flags().IntVar(&maxgo, "maxgo", 3, "maximum number of concurrent goroutines(tune me in case of 'out of memory' issue!)[2..numCPU]")
	logsCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("logs.tlog", logsCmd.Flags().Lookup("tlog"))
	viper.BindPFlag("logs.py2", logsCmd.Flags().Lookup("py2"))
	viper.BindPFlag("logs.py3", logsCmd.Flags().Lookup("py3"))
	viper.BindPFlag("logs.tlda", logsCmd.Flags().Lookup("tlda"))
	viper.BindPFlag("logs.snaptool", logsCmd.Flags().Lookup("snaptool"))
	viper.BindPFlag("logs.luashark", logsCmd.Flags().Lookup("luashark"))
	viper.BindPFlag("logs.wshark", logsCmd.Flags().Lookup("wshark"))
	viper.BindPFlag("logs.trace", logsCmd.Flags().Lookup("trace"))
	viper.BindPFlag("logs.pattern", logsCmd.Flags().Lookup("pattern"))
	viper.BindPFlag("logs.rat", logsCmd.Flags().Lookup("rat"))
	viper.BindPFlag("logs.scs", logsCmd.Flags().Lookup("scs"))
	viper.BindPFlag("logs.chbw", logsCmd.Flags().Lookup("chbw"))
	viper.BindPFlag("logs.nap", logsCmd.Flags().Lookup("nap"))
	viper.BindPFlag("logs.filter", logsCmd.Flags().Lookup("filter"))
	viper.BindPFlag("logs.maxgo", logsCmd.Flags().Lookup("maxgo"))
	viper.BindPFlag("logs.debug", logsCmd.Flags().Lookup("debug"))
}

func loadLogsFlags() {
	tlog = viper.GetString("logs.tlog")
	py2 = viper.GetString("logs.py2")
	py3 = viper.GetString("logs.py3")
	tlda = viper.GetString("logs.tlda")
	snaptool = viper.GetString("logs.snaptool")
	luashark = viper.GetString("logs.luashark")
	wshark = viper.GetString("logs.wshark")
	trace = viper.GetString("logs.trace")
	pattern = viper.GetString("logs.pattern")
	rat = viper.GetString("logs.rat")
	scs = viper.GetString("logs.scs")
	chbw = viper.GetString("logs.chbw")
	nap = viper.GetInt("logs.nap")
	filter = viper.GetString("logs.filter")
	maxgo = viper.GetInt("logs.maxgo")
	debug = viper.GetBool("logs.debug")
}
