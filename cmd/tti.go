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
	debug   bool
)

// ttiCmd represents the tti command
var ttiCmd = &cobra.Command{
	Use:   "tti",
	Short: "",
	Long:  `tti parses and aggregates MAC TTI trace of LTE/NR.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(dir) == 0 {
			cmd.Flags().Set("dir", viper.GetString("tti.dir"))
		} else {
			viper.WriteConfig()
		}

		// fmt.Println(dir)
		// fmt.Println(pattern)
		// fmt.Println(viper.Get("tti_dir"))

		tti := new(ttitrace.TtiParser)
		tti.Init(Logger, dir, pattern, rat, scs, debug)
		tti.Exec()
	},
}

func init() {
	rootCmd.AddCommand(ttiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ttiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//ttiCmd.Flags().StringP("dir", "d", "", "directory containing tti files")
	//ttiCmd.Flags().StringP("pattern", "p", ".csv", "pattern of tti files")
	ttiCmd.Flags().StringVarP(&dir, "dir", "d", "./data", "directory containing tti files")
	ttiCmd.Flags().StringVarP(&pattern, "pattern", "p", ".csv", "pattern of tti files")
	ttiCmd.Flags().StringVar(&rat, "rat", "nr", "RAT info of MAC TTI traces[nr]")
	ttiCmd.Flags().StringVar(&scs, "scs", "30khz", "NRCELLGRP/scs setting[15khz,30khz,120khz]")
	ttiCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("tti.dir", ttiCmd.Flags().Lookup("dir"))
}
