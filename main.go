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

package main

import (
	"github.com/scalog/scalog/cmd"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("port", 21024)

	viper.SetConfigName("scalog")        // name of config file (without extension)
	viper.AddConfigPath("/etc/scalog/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.scalog") // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	err := viper.ReadInConfig()          // Find and read the config file
	if err != nil {                      // Handle errors reading the config file
		// Do nothing.
		// If the configuration file doesn't exist, use value passed from command line.
	}
}

func main() {
	cmd.Execute()
}
