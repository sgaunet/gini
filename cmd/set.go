package cmd

import (
	"fmt"

	"github.com/sgaunet/gini/internal/tools"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var value string

// setCmd represents the set command.
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "add/update key/value",
	Long:  `add/update key/value in the desired section (can be empty)`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if iniFile == "" {
			return errNoIniFile
		}
		if !tools.IsFileExists(iniFile) && createIniFileIfAbsent {
			err := tools.TouchFile(iniFile)
			if err != nil {
				return fmt.Errorf("can't create file: %w", err)
			}
		}
		cfg, err := ini.Load(iniFile)
		if err != nil {
			return fmt.Errorf("fail to load file: %w", err)
		}
		// Classic read of values, default section can be represented as empty string
		cfg.Section(section).Key(key).SetValue(value)
		err = cfg.SaveTo(iniFile)
		if err != nil {
			return fmt.Errorf("fail to save file: %w", err)
		}
		return nil
	},
}
