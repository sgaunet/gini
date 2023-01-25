package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

// delSectionCmd represents the get command
var delSectionCmd = &cobra.Command{
	Use:   "delsection",
	Short: "delete all keys of a section",
	Long:  `delete all keys of a section`,
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

		// for _, k := range cfg.Section(section).Keys() {
		// 	cfg.Section(section).DeleteKey(k.Name())
		// }
		cfg.DeleteSection(section)
		err = cfg.SaveTo(iniFile)
		if err != nil {
			fmt.Printf("Fail to save file: %v", err)
			os.Exit(1)
		}
	},
}
