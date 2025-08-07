package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// getCmd represents the get command.
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "retrieve a key from an ini file",
	Long:  `retrieve a key from an ini file`,
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

		if cfg.Section(section).HasKey(key) {
			fmt.Println(cfg.Section(section).Key(key).String())
		}
	},
}
