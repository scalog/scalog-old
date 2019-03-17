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
	"fmt"

	"github.com/scalog/scalog/order"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// orderCmd represents the order command
var orderCmd = &cobra.Command{
	Use:   "order",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("order called")
		order.Start()
	},
}

func init() {
	rootCmd.AddCommand(orderCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// orderCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// orderCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	orderCmd.PersistentFlags().Int("raftNodeID", 1, "ID of raft node in raft cluster")
	viper.BindPFlag("raftNodeID", orderCmd.PersistentFlags().Lookup("raftNodeID"))

	// Raft must listen on a port different than the one we serve gRPC requests on
	viper.SetDefault("raftPort", 1337)

	viper.BindEnv("name")
	viper.BindEnv("namespace")
	viper.BindEnv("pod_ip")
	// The number of replicas running in the ordering layer
	viper.BindEnv("raft_cluster_size")
	// UID of this pod
	viper.BindEnv("uid")
}
