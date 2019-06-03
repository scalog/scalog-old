// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/scalog/scalog/client/examples/bench"

	"github.com/spf13/cobra"
)

// benchCmd represents the bench command
var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Scalog client-side benchmarks",
	Long:  `Scalog client-side benchmarks`,
	Run: func(cmd *cobra.Command, args []string) {
		num, err := cmd.Flags().GetInt32("num")
		if err != nil {
			panic(err)
		}
		size, err := cmd.Flags().GetInt32("size")
		if err != nil {
			panic(err)
		}
		b, err := bench.NewBench(num, size)
		if err != nil {
			panic(err)
		}
		err = b.Start()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(benchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// benchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	benchCmd.Flags().Int32P("num", "n", 128, "Number of append operations")
	benchCmd.Flags().Int32P("size", "s", 2048, "Size of each append operation")
}
