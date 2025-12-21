package cmd

import (
	"fmt"

	"github.com/sgaunet/gini/internal/tools"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// getCmd represents the get command.
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "retrieve a key from an ini file",
	Long:  `retrieve a key from an ini file`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if iniFile == "" {
			return errNoIniFile
		}
		if err := tools.ValidateKey(key); err != nil {
			return fmt.Errorf("invalid key: %w", err)
		}
		if err := tools.ValidateSection(section); err != nil {
			return fmt.Errorf("invalid section: %w", err)
		}

		cfg, err := ini.Load(iniFile)
		if err != nil {
			return fmt.Errorf("fail to load file: %w", err)
		}

		if cfg.Section(section).HasKey(key) {
			fmt.Println(cfg.Section(section).Key(key).String())
		}
		return nil
	},
}
