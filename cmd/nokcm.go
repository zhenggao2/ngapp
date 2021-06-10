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
	"github.com/zhenggao2/ngapp/nokcm"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	tcm    string
	cmpath string
)

// cmCmd represents the cm command
var cmCmd = &cobra.Command{
	Use:   "cm",
	Short: "CM analysis tool",
	Long:  `The cm module parses SCFC/Vendor/FrequencyHistory/IMS2 or generates CMCC-CM.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadCmFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()

		if tcm == "scfc" || tcm == "vendor" || tcm == "freqhist" {
			fileInfo, err := ioutil.ReadDir(cmpath)
			if err != nil {
				Logger.Fatal(fmt.Sprintf("Fail to read directory: %s.", cmpath))
				fmt.Printf("Fail to read directory: %s.\n", cmpath)
				return
			}

			// recreate output directory if necessary
			out := path.Join(cmpath, fmt.Sprintf("parsed_%s", tcm))
			os.RemoveAll(out)
			if err := os.MkdirAll(out, 0775); err != nil {
				panic(fmt.Sprintf("Fail to create directory: %v", err))
			}

			parser := new(nokcm.XmlParser)
			parser.Init(Logger, out, debug)
			wg := &sync.WaitGroup{}
			for _, file := range fileInfo {
				if !file.IsDir() && strings.ToLower(path.Ext(file.Name())) == ".xml" {
					for {
						if runtime.NumGoroutine() >= maxgo {
							time.Sleep(1 * time.Second)
						} else {
							break
						}
					}

					xml := path.Join(cmpath, file.Name())
					wg.Add(1)
					go func(fn string) {
						defer wg.Done()
						parser.Parse(fn, tcm)
					} (xml)
				}
			}
			wg.Wait()
		} else if tcm == "ims2" {
			fileInfo, err := ioutil.ReadDir(cmpath)
			if err != nil {
				Logger.Fatal(fmt.Sprintf("Fail to read directory: %s.", cmpath))
				fmt.Printf("Fail to read directory: %s.\n", cmpath)
				return
			}

			parser := new(nokcm.Ims2Parser)
			parser.Init(Logger, debug)

			wg := &sync.WaitGroup{}
			for _, file := range fileInfo {
				if !file.IsDir() && strings.ToLower(path.Ext(file.Name())) == ".ims2" {
					for {
						if runtime.NumGoroutine() >= maxgo {
							time.Sleep(1 * time.Second)
						} else {
							break
						}
					}

					ims2 := path.Join(cmpath, file.Name())
					wg.Add(1)
					go func(fn string) {
						defer wg.Done()
						parser.Parse(fn)
					} (ims2)
				}
			}
			wg.Wait()
		} else if tcm == "cmcc" {
			// TODO
		} else {
			fmt.Printf("Unsupported tcm[=%s].\n", tcm)
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
	rootCmd.AddCommand(cmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// someCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// someCmd.Flags().StringP("trace", "d", "./trace_path", "path containing tti files")
	cmCmd.Flags().StringVar(&tcm, "tcm", "scfc", "type of CM[scfc,vendor,freqhist,ims2,cmcc]")
	cmCmd.Flags().StringVar(&cmpath, "cmpath", "./data", "path containing CM files")
	cmCmd.Flags().IntVar(&maxgo, "maxgo", 3, "maximum number of concurrent goroutines(tune me in case of 'out of memory' issue!)[2..numCPU]")
	cmCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("cm.tcm", cmCmd.Flags().Lookup("tcm"))
	viper.BindPFlag("cm.cmpath", cmCmd.Flags().Lookup("cmpath"))
	viper.BindPFlag("cm.maxgo", cmCmd.Flags().Lookup("maxgo"))
	viper.BindPFlag("cm.debug", cmCmd.Flags().Lookup("debug"))
}

func loadCmFlags() {
	tcm = viper.GetString("cm.tcm")
	cmpath = viper.GetString("cm.cmpath")
	maxgo = viper.GetInt("cm.maxgo")
	debug = viper.GetBool("cm.debug")
}

