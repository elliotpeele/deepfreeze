// Copyright © 2016 Elliot Peele <elliot@bentlogic.net>
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
	"os"
	"path"

	"github.com/elliotpeele/deepfreeze/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var cfgPath string

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup application for securely storing incremental backups",
	Long: `Backup application designed to work with Amazon Glacier in
mind, but can also handle local filesystem based backups.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			os.MkdirAll(cfgPath, 0755)
		}

		logFile, _ := cmd.InheritedFlags().GetString("log-file")
		logFile = os.ExpandEnv(logFile)
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening log file, defaulting to stderr: %s\n", err)
			f = nil
		}

		debug, _ := cmd.InheritedFlags().GetBool("debug")
		log.SetupLogging(f, os.Stderr, debug, debug)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config",
		"$HOME/.config/deepfreeze/deepfreeze.yaml", "Path to the configuration file")

	RootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debug logging")
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))

	RootCmd.PersistentFlags().String("log-file",
		"$HOME/.config/deepfreeze/deepfreeze.log", "log file location")
	viper.BindPFlag("log-file", RootCmd.PersistentFlags().Lookup("log-file"))

	cfgPath = os.ExpandEnv(path.Join("$HOME", ".config", "deepfreeze"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("deepfreeze")               // name of config file (without extension)
	viper.AddConfigPath("$HOME/.config/deepfreeze") // adding home directory as first search path
	viper.AutomaticEnv()                            // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
