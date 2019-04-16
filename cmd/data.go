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
	"strconv"
	"strings"

	"github.com/scalog/scalog/data"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func parsePodName(podName string) (string, int, int) {
	splitPodName := strings.Split(podName, "-")
	shardGroup := strings.Join(splitPodName[:len(splitPodName)-1], "-")
	shardID, err := strconv.Atoi(splitPodName[len(splitPodName)-2])
	if err != nil {
		return "", -1, -1
	}
	replicaID, err := strconv.Atoi(splitPodName[len(splitPodName)-1])
	if err != nil {
		replicaID = -1
		shardGroup = ""
		shardID = -1
	}
	return shardGroup, replicaID, shardID
}

func init() {
	rootCmd.AddCommand(dataCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dataCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dataCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Load environment variables into Viper
	bindEnvVar("node_name")
	bindEnvVar("name")
	bindEnvVar("namespace")
	bindEnvVar("pod_ip")
	bindEnvVar("replica_count")
	bindEnvVar("batch_interval")

	// UGLY - We parse metainformation from the name of this pod, since Kubernetes cannot
	// do it itself
	shardGroup, replicaID, shardID := parsePodName(viper.GetString("name"))
	viper.Set("shardGroup", shardGroup)
	viper.Set("shardID", shardID)
	viper.Set("orderPort", 21024)
	viper.Set("id", replicaID)
}

// dataCmd represents the data command
var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("data called")
		data.Start()
	},
}
