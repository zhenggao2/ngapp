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
)

var (
	gnblist string
	gnblogs string
	gnbtools string
)

// autobipCmd represents the autobip command
var autobipCmd = &cobra.Command{
	Use:   "autobip",
	Short: "Automatic BIP capture/analysis tool",
	Long:  `The autobip module automatically captures and parses BIP.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadAutobipFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()
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
	rootCmd.AddCommand(autobipCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tmpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//tmpCmd.Flags().StringP("trace", "d", "./trace_path", "path containing tti files")
	autobipCmd.Flags().StringVar(&gnblist, "gnblist", "C:/gnb_list.txt", "file containing list of gNBs, format of per line should be:[gnb_id,gnb_ip,gnb_sw,scs,chbw]")
	autobipCmd.Flags().StringVar(&gnblogs, "gnblogs", "C:/gnb_logs", "path of gnb_logs tool")
	autobipCmd.Flags().StringVar(&gnbtools, "gnbtools", "C:/gnb_tools", "path of gNB tools, containing TLDA, generated_luashark, tti-dec-bin and snapshot_tool of each gNB SW load")
	autobipCmd.Flags().StringVar(&wshark, "wshark", "C:/Program Files/Wireshark", "path of tshark")
	autobipCmd.Flags().IntVar(&maxgo, "maxgo", 3, "maximum number of concurrent goroutines(tune me in case of 'out of memory' issue!)[2..numCPU]")
	autobipCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("autobip.gnblist", autobipCmd.Flags().Lookup("gnblist"))
	viper.BindPFlag("autobip.gnblogs", autobipCmd.Flags().Lookup("gnblogs"))
	viper.BindPFlag("autobip.gnbtools", autobipCmd.Flags().Lookup("gnbtools"))
	viper.BindPFlag("autobip.wshark", autobipCmd.Flags().Lookup("wshark"))
	viper.BindPFlag("autobip.maxgo", autobipCmd.Flags().Lookup("maxgo"))
	viper.BindPFlag("autobip.debug", autobipCmd.Flags().Lookup("debug"))
}

func loadAutobipFlags() {
	gnblist = viper.GetString("autobip.gnblist")
	gnblogs = viper.GetString("autobip.gnblogs")
	gnbtools = viper.GetString("autobip.gnbtools")
	wshark = viper.GetString("autobip.wshark")
	maxgo = viper.GetInt("autobip.maxgo")
	debug = viper.GetBool("autobip.debug")
}