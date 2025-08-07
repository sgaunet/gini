// Package cmd contains all CLI commands for the gini application.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// delCmd represents the del command.
var delCmd = &cobra.Command{
	Use:   "del",
	Short: "delete a key from an ini file",
	Long:  `delete a key from an ini file`,
	Run: func(_ *cobra.Command, _ []string) {
		if iniFile == "" {
			fmt.Fprintln(os.Stderr, "specify inifile")
			os.Exit(1)
		}
		cfg, err := ini.Load(iniFile)
		if err != nil {
			fmt.Printf("Fail to load file: %v", err)
			os.Exit(1)
		}

		cfg.Section(section).DeleteKey(key)
		err = cfg.SaveTo(iniFile)
		if err != nil {
			fmt.Printf("Fail to save file: %v", err)
			os.Exit(1)
		}
	},
}
