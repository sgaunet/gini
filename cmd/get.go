package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "retrieve a key from an ini file",
	Long:  `retrieve a key from an ini file`,
	Run: func(cmd *cobra.Command, args []string) {
		if iniFile == "" {
			fmt.Fprintln(os.Stderr, "specify inifile")
			os.Exit(1)
		}
		cfg, err := ini.Load(iniFile)
		if err != nil {
			fmt.Printf("Fail to load file: %v", err)
			os.Exit(1)
		}

		// Classic read of values, default section can be represented as empty string
		fmt.Println(cfg.Section(section).Key(key).String())
	},
}
