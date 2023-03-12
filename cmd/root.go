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
	"os"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/zhenggao2/ngapp/utils"
)

// separate flag per module
var (
	CMD_FLAG_NRRG = 0x1

	CMD_FLAG_AUTO_BIP = 0x1 << 1

	CMD_FLAG_CM       = 0x1 << 2
	CMD_FLAG_CM_DIFF  = 0x1 << 3
	CMD_FLAG_CM_FIND  = 0x1 << 4
	CMD_FLAG_CM_PDCCH = 0x1 << 5
	CMD_FLAG_CM_ALL   = CMD_FLAG_CM | CMD_FLAG_CM_DIFF | CMD_FLAG_CM_FIND | CMD_FLAG_CM_PDCCH

	CMD_FLAG_TTI      = 0x1 << 6
	CMD_FLAG_BIP      = 0x1 << 7
	CMD_FLAG_DDR4     = 0x1 << 8
	CMD_FLAG_L2TRACE  = 0x1 << 9
	CMD_FLAG_LOGS_ALL = CMD_FLAG_TTI | CMD_FLAG_BIP | CMD_FLAG_DDR4 | CMD_FLAG_L2TRACE

	CMD_FLAG_PM     = 0x1 << 10
	CMD_FLAG_KPI    = 0x1 << 11
	CMD_FLAG_PM_ALL = CMD_FLAG_PM | CMD_FLAG_KPI
)

var (
	Logger  = utils.NewZapLogger(fmt.Sprintf("./logs/ngapp_%v.log", time.Now().Format("20060102_150405")))
	cfgFile string
	// maximum number of goroutines. Adjust maxgo in case ngapp has crashed with 'out of memory' error.
	maxgo int
	debug bool
	// cmdFlags = CMD_FLAG_NRRG | CMD_FLAG_AUTO_BIP | CMD_FLAG_CM_ALL | CMD_FLAG_LOGS_ALL | CMD_FLAG_PM_ALL
	cmdFlags = CMD_FLAG_NRRG
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ngapp",
	Short: "",
	Long: `ngapp is a collection of useful applets for 4G and 5G NPO.
Author: zhengwei.gao@yahoo.com
Blog: http://blog.csdn.net/jeffyko`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./ngapp.yaml", "config file (default is $HOME/ngapp.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name "ngapp" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("ngapp")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
		Logger.Info(fmt.Sprintf("Using config file: %v", viper.ConfigFileUsed()))
	}
}
