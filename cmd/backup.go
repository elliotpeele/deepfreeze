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
	"github.com/elliotpeele/deepfreeze/freezer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := cmd.PersistentFlags().GetString("root")
		if err != nil {
			return err
		}

		dest, err := cmd.PersistentFlags().GetString("dest")
		if err != nil {
			return err
		}

		excludes, err := cmd.PersistentFlags().GetStringSlice("exclude")
		if err != nil {
			return err
		}

		f, err := freezer.New(root, dest, excludes)
		if err != nil {
			return err
		}

		if err := f.Freeze(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(backupCmd)

	backupCmd.PersistentFlags().StringP("root", "r", "", "path to backup")
	viper.BindPFlag("root", backupCmd.PersistentFlags().Lookup("root"))

	backupCmd.PersistentFlags().String("dest", "/var/lib/deepfreeze/",
		"path for storing backup data")
	viper.BindPFlag("dest", backupCmd.PersistentFlags().Lookup("dest"))

	backupCmd.PersistentFlags().StringSliceP("exclude", "e", nil,
		"directory paths to ignore")
	viper.BindPFlag("exclude", backupCmd.PersistentFlags().Lookup("exclude"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// backupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
