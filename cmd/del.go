// Package cmd contains all CLI commands for the gini application.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// delCmd represents the del command.
var delCmd = &cobra.Command{
	Use:   "del",
	Short: "delete a key from an ini file",
	Long:  `delete a key from an ini file`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if iniFile == "" {
			return errNoIniFile
		}
		cfg, err := ini.Load(iniFile)
		if err != nil {
			return fmt.Errorf("fail to load file: %w", err)
		}

		cfg.Section(section).DeleteKey(key)
		err = cfg.SaveTo(iniFile)
		if err != nil {
			return fmt.Errorf("fail to save file: %w", err)
		}
		return nil
	},
}
