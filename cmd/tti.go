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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zhenggao2/ngapp/ttitrace"
	"github.com/zhenggao2/ngapp/l2trace"
)

var (
	tlog string
	py2 string
	tlda string
	luashark string
	tshark string
	trace   string
	pattern string
	rat     string
	scs     string
	filter  string
	debug   bool
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "L2TtiTrace/L2Trace/BIP log analysis tool",
	Long:  `The logs module parses and aggregates L2TtiTrace/L2Trace/BIP log of Nokia gNB.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadTtiFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()

		if tlog == "l2tti" {
			tti := new(ttitrace.L2TtiTraceParser)
			tti.Init(Logger, trace, pattern, rat, scs, filter, debug)
			tti.Exec()
		} else if tlog == "l2trace" {
			parser := new(l2trace.L2TraceParser)
			parser.Init(Logger, py2, tlda, luashark, tshark, trace, pattern, debug)
			parser.Exec()
		} else if tlog == "bip" {
			// TODO
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
	//logsCmd.Flags().StringP("trace", "d", "", "path containing tti files")
	logsCmd.Flags().StringVar(&tlog, "tlog", "l2tti", "type of traces[l2tti,l2trace,bip] ")
	logsCmd.Flags().StringVar(&py2, "py2", "C:/Python27", "path of Python2")
	logsCmd.Flags().StringVar(&tlda, "tlda", "C:/TLDA", "path of TLDA")
	logsCmd.Flags().StringVar(&luashark, "luashark", "C:/luashark", "path of luashark scripts")
	logsCmd.Flags().StringVar(&tshark, "tshark", "C:/Program Files/Wireshark", "path of Tshark")
	logsCmd.Flags().StringVar(&trace, "trace", "./data", "path containing trace files")
	logsCmd.Flags().StringVar(&pattern, "pattern", ".csv", "pattern of trace files[.csv,.pcap,.dat]")
	logsCmd.Flags().StringVar(&rat, "rat", "nr", "RAT info of traces[nr]")
	logsCmd.Flags().StringVar(&scs, "scs", "30khz", "NRCELLGRP/scs setting[15khz,30khz,120khz]")
	logsCmd.Flags().StringVar(&filter, "filter", "both", "ul/dl tti filter[ul,dl,both]")
	logsCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("logs.tlog", logsCmd.Flags().Lookup("tlog"))
	viper.BindPFlag("logs.py2", logsCmd.Flags().Lookup("py2"))
	viper.BindPFlag("logs.tlda", logsCmd.Flags().Lookup("tlda"))
	viper.BindPFlag("logs.luashark", logsCmd.Flags().Lookup("luashark"))
	viper.BindPFlag("logs.tshark", logsCmd.Flags().Lookup("tshark"))
	viper.BindPFlag("logs.trace", logsCmd.Flags().Lookup("trace"))
	viper.BindPFlag("logs.pattern", logsCmd.Flags().Lookup("pattern"))
	viper.BindPFlag("logs.rat", logsCmd.Flags().Lookup("rat"))
	viper.BindPFlag("logs.scs", logsCmd.Flags().Lookup("scs"))
	viper.BindPFlag("logs.filter", logsCmd.Flags().Lookup("filter"))
	viper.BindPFlag("logs.debug", logsCmd.Flags().Lookup("debug"))
}

func loadTtiFlags() {
	tlog = viper.GetString("logs.tlog")
	py2 = viper.GetString("logs.py2")
	tlda = viper.GetString("logs.tlda")
	luashark = viper.GetString("logs.luashark")
	tshark = viper.GetString("logs.tshark")
	trace = viper.GetString("logs.trace")
	pattern = viper.GetString("logs.pattern")
	rat = viper.GetString("logs.rat")
	scs = viper.GetString("logs.scs")
	filter = viper.GetString("logs.filter")
	debug = viper.GetBool("logs.debug")
}
