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
package main

import (
	"fmt"
	"github.com/unidoc/unioffice/common/license"
	"github.com/zhenggao2/ngapp/cmd"
	"runtime"
)

var swVersion = "v0.21.081301"

func init() {
	err := license.SetMeteredKey("fb1b3cb24189879d60454c44462d70091e9cdb168e21205c9aa38cf0e413891d")
	if err != nil {
		// fmt.Printf("ERROR: Failed to set metered key: %v\n", err)
		// fmt.Printf("Make sure to get a valid key from https://cloud.unidoc.io\n")
		panic(err)
	}
}

func main() {
	fmt.Printf("ngapp version: %s, built with: %s\n\n", swVersion, runtime.Version())
	// runtime.GOMAXPROCS(runtime.NumCPU() / 2)

	lic := license.GetLicenseKey()
	if lic == nil {
		// fmt.Printf("Failed retrieving license key")
		return
	}

	/*
		// GetMeteredState freshly checks the state, contacting the licensing server.
		state, err := license.GetMeteredState()
		if err != nil {
			fmt.Printf("ERROR getting metered state: %+v\n", err)
			panic(err)
		}
		fmt.Printf("State: %+v\n", state)
		if state.OK {
			fmt.Printf("State is OK\n")
		} else {
			fmt.Printf("State is not OK\n")
		}
		fmt.Printf("Credits: %v\n", state.Credits)
		fmt.Printf("Used credits: %v\n", state.Used)
	*/

	cmd.Execute()
}
