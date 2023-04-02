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
	"github.com/zhenggao2/ngapp/nokpm"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	tpm     string
	op      string
	pmpath  string
	pmdb    string
	kpipath string
	btsid   string
	stime   string
	etime   string
)

// pmCmd represents the pm command
var pmCmd = &cobra.Command{
	Use:   "pm",
	Short: "PM analysis tool",
	Long:  `The pm module parses and stores raw PMs or SQL queries.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadPmFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()

		if tpm == "raw" || tpm == "sql" {
			fileInfo, err := ioutil.ReadDir(pmpath)
			if err != nil {
				Logger.Fatal(fmt.Sprintf("Fail to read directory: %s.", pmpath))
				fmt.Printf("Fail to read directory: %s.\n", pmpath)
				return
			}

			// create output directory if necessary
			if err := os.MkdirAll(pmdb, 0775); err != nil {
				panic(fmt.Sprintf("Fail to create directory: %v", err))
			}

			parser := new(nokpm.PmParser)
			parser.Init(Logger, op, pmdb, debug)

			wg := &sync.WaitGroup{}
			for _, file := range fileInfo {
				if !file.IsDir() {
					ext := strings.ToLower(filepath.Ext(file.Name()))
					if (tpm == "raw" && ext == ".xml") || (tpm == "sql" && ext == ".csv") {
						for {
							if runtime.NumGoroutine() >= maxgo {
								time.Sleep(1 * time.Second)
							} else {
								break
							}
						}

						wg.Add(1)
						go func(fn string) {
							defer wg.Done()
							parser.Parse(fn, tpm)
						}(filepath.Join(pmpath, file.Name()))
					}
				}
			}
			wg.Wait()

			if tpm == "raw" {
				parser.ArchiveRawPm()
			}
		} else {
			fmt.Printf("Unsupported tpm[=%s].\n", tpm)
		}
	},
}

// kpiCmd represents the kpi command
var kpiCmd = &cobra.Command{
	Use:   "kpi",
	Short: "KPI reporting tool",
	Long:  `The kpi module parses KPI definitions and generates KPI reports.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadKpiFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		laPrint(cmd, args)
		viper.WriteConfig()

		fileInfo, err := ioutil.ReadDir(kpipath)
		if err != nil {
			Logger.Fatal(fmt.Sprintf("Fail to read directory: %s.", kpipath))
			fmt.Printf("Fail to read directory: %s.\n", kpipath)
			return
		}

		// create output directory if necessary
		if err := os.MkdirAll(filepath.Join(kpipath, "kpi_report"), 0775); err != nil {
			panic(fmt.Sprintf("Fail to create directory: %v", err))
		}

		parser := new(nokpm.KpiParser)
		parser.Init(Logger, op, pmdb, maxgo, debug)
		for _, file := range fileInfo {
			if !file.IsDir() {
				parser.ParseKpiDef(filepath.Join(kpipath, file.Name()))
			}
		}
		parser.LoadPmDb(pmdb, btsid, stime, etime)
		parser.CalcKpi(filepath.Join(kpipath, "kpi_report"))
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
	if cmdFlags&CMD_FLAG_PM != 0 {
		rootCmd.AddCommand(pmCmd)
	}
	if cmdFlags&CMD_FLAG_KPI != 0 {
		rootCmd.AddCommand(kpiCmd)
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// someCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// someCmd.Flags().StringP("trace", "d", "./trace_path", "path containing tti files")
	pmCmd.Flags().StringVar(&tpm, "tpm", "raw", "type of PM(raw PM: .xml, sql queries: .csv)[raw,sql]")
	pmCmd.Flags().StringVar(&op, "op", "cmcc", "name of specific operator[cmcc,twm]")
	pmCmd.Flags().StringVar(&pmpath, "pmpath", "./data", "path containing PM files")
	pmCmd.Flags().StringVar(&pmdb, "pmdb", "./pmdb", "path of PM database")
	pmCmd.Flags().IntVar(&maxgo, "maxgo", 3, "maximum number of concurrent goroutines(tune me in case of 'out of memory' issue!)[2..numCPU]")
	pmCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("pm.tpm", pmCmd.Flags().Lookup("tpm"))
	viper.BindPFlag("pm.op", pmCmd.Flags().Lookup("op"))
	viper.BindPFlag("pm.pmpath", pmCmd.Flags().Lookup("pmpath"))
	viper.BindPFlag("pm.pmdb", pmCmd.Flags().Lookup("pmdb"))
	viper.BindPFlag("pm.maxgo", pmCmd.Flags().Lookup("maxgo"))
	viper.BindPFlag("pm.debug", pmCmd.Flags().Lookup("debug"))

	kpiCmd.Flags().StringVar(&op, "op", "cmcc", "name of specific operator[cmcc,twm]")
	kpiCmd.Flags().StringVar(&kpipath, "kpipath", "./data", "path containing KPI definitions")
	kpiCmd.Flags().StringVar(&pmdb, "pmdb", "./pmdb", "path of PM database")
	kpiCmd.Flags().StringVar(&btsid, "btsid", "123,456", "target BTS IDs(comma seperated) or 'all' as a special value(CAUTION!)")
	kpiCmd.Flags().StringVar(&stime, "stime", "20060102", "start time inclusive[YYYYMMDD]")
	kpiCmd.Flags().StringVar(&etime, "etime", "20060102", "end time inclusive[YYYYMMDD]")
	kpiCmd.Flags().IntVar(&maxgo, "maxgo", 3, "maximum number of concurrent goroutines(tune me in case of 'out of memory' issue!)[2..numCPU]")
	kpiCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")
	viper.BindPFlag("kpi.op", kpiCmd.Flags().Lookup("op"))
	viper.BindPFlag("kpi.kpipath", kpiCmd.Flags().Lookup("kpipath"))
	viper.BindPFlag("kpi.pmdb", kpiCmd.Flags().Lookup("pmdb"))
	viper.BindPFlag("kpi.btsid", kpiCmd.Flags().Lookup("btsid"))
	viper.BindPFlag("kpi.stime", kpiCmd.Flags().Lookup("stime"))
	viper.BindPFlag("kpi.etime", kpiCmd.Flags().Lookup("etime"))
	viper.BindPFlag("kpi.maxgo", kpiCmd.Flags().Lookup("maxgo"))
	viper.BindPFlag("kpi.debug", kpiCmd.Flags().Lookup("debug"))
}

func loadPmFlags() {
	tpm = viper.GetString("pm.tpm")
	op = viper.GetString("pm.op")
	pmpath = viper.GetString("pm.pmpath")
	pmdb = viper.GetString("pm.pmdb")
	maxgo = viper.GetInt("pm.maxgo")
	debug = viper.GetBool("pm.debug")
}

func loadKpiFlags() {
	op = viper.GetString("kpi.op")
	kpipath = viper.GetString("kpi.kpipath")
	pmdb = viper.GetString("kpi.pmdb")
	btsid = viper.GetString("kpi.btsid")
	stime = viper.GetString("kpi.stime")
	etime = viper.GetString("kpi.etime")
	maxgo = viper.GetInt("kpi.maxgo")
	debug = viper.GetBool("kpi.debug")
}
