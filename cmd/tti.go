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
)

var (
	dir     string
	pattern string
	rat     string
	scs     string
	filter string
	debug   bool
)

// ttiCmd represents the tti command
var ttiCmd = &cobra.Command{
	Use:   "tti",
	Short: "L2 TTI trace tool",
	Long:  `tti parses and aggregates MAC TTI trace of LTE/NR.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadTtiFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()

		tti := new(ttitrace.TtiParser)
		tti.Init(Logger, dir, pattern, rat, scs, filter, debug)
		tti.Exec()
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
	rootCmd.AddCommand(ttiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ttiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//ttiCmd.Flags().StringP("dir", "d", "", "directory containing tti files")
	ttiCmd.Flags().StringVarP(&dir, "dir", "d", "./data", "directory containing tti files")
	ttiCmd.Flags().StringVarP(&pattern, "pattern", "p", ".csv", "pattern of tti files")
	ttiCmd.Flags().StringVar(&rat, "rat", "nr", "RAT info of MAC TTI traces[nr]")
	ttiCmd.Flags().StringVar(&scs, "scs", "30khz", "NRCELLGRP/scs setting[15khz,30khz,120khz]")
	ttiCmd.Flags().StringVar(&filter, "filter", "both", "ul/dl tti filter[ul,dl,both]")
	ttiCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("tti.dir", ttiCmd.Flags().Lookup("dir"))
	viper.BindPFlag("tti.pattern", ttiCmd.Flags().Lookup("pattern"))
	viper.BindPFlag("tti.rat", ttiCmd.Flags().Lookup("rat"))
	viper.BindPFlag("tti.scs", ttiCmd.Flags().Lookup("scs"))
	viper.BindPFlag("tti.filter", ttiCmd.Flags().Lookup("filter"))
	viper.BindPFlag("tti.debug", ttiCmd.Flags().Lookup("debug"))
}

func loadTtiFlags() {
	dir = viper.GetString("tti.dir")
	pattern = viper.GetString("tti.pattern")
	rat = viper.GetString("tti.rat")
	scs = viper.GetString("tti.scs")
	filter = viper.GetString("tti.filter")
	debug = viper.GetBool("tti.debug")
}
