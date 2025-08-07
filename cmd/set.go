package cmd

import (
	"fmt"
	"os"

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
	Run: func(_ *cobra.Command, _ []string) {
		if iniFile == "" {
			fmt.Fprintln(os.Stderr, "specify inifile")
			os.Exit(1)
		}
		if !tools.IsFileExists(iniFile) && createIniFileIfAbsent {
			err := tools.TouchFile(iniFile)
			if err != nil {
				fmt.Fprintln(os.Stderr, "can't create file : %w")
				os.Exit(1)
			}
		}
		cfg, err := ini.Load(iniFile)
		if err != nil {
			fmt.Printf("Fail to load file: %v", err)
			os.Exit(1)
		}
		// Classic read of values, default section can be represented as empty string
		cfg.Section(section).Key(key).SetValue(value)
		err = cfg.SaveTo(iniFile)
		if err != nil {
			fmt.Printf("Fail to save file: %v", err)
			os.Exit(1)
		}
	},
}
