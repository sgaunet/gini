package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var iniFile string
var section string
var key string
var createIniFileIfAbsent bool

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gini",
	Short: "Tool to get/set key from an ini file.",
	Long:  `Tool to get/set key from an ini file.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVar(&iniFile, "i", "", "init file to read/update")
	rootCmd.PersistentFlags().StringVar(&key, "k", "", "key to read or update")
	rootCmd.PersistentFlags().StringVar(&section, "s", "", "section of ini file (can be empty)")

	rootCmd.AddCommand(getCmd)

	setCmd.Flags().StringVar(&value, "v", "", "value to set")
	setCmd.Flags().BoolVar(&createIniFileIfAbsent, "c", false, "create file if no present")
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(delCmd)

	rootCmd.AddCommand(delSectionCmd)
}
