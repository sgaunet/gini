package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

var (
	errNoIniFile      = errors.New("specify inifile")
	errKeyNotFound    = errors.New("key not found")
	errSectionNotFound = errors.New("section not found")
)

var iniFile string
var section string
var key string
var createIniFileIfAbsent bool
var strict bool

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gini",
	Short: "Tool to get/set key from an ini file.",
	Long: `gini is a CLI utility for performing basic operations on INI configuration files.

It provides commands to read, write, update, and delete keys and sections in INI files.
All operations use atomic writes to prevent file corruption during updates.

Common flags:
  -i, --i string    Path to the INI file (required for all commands)
  -s, --s string    Section name (use empty string for default section)
  -k, --k string    Key name within the section

Examples:
  gini get -i config.ini -s database -k host
  gini set -i config.ini -s database -k port -v 5432
  gini del -i config.ini -s cache -k ttl
  gini delsection -i config.ini -s deprecated`,
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
	rootCmd.PersistentFlags().StringVar(&iniFile, "i", "", "INI file to read/update")
	rootCmd.PersistentFlags().StringVar(&key, "k", "", "key to read or update")
	rootCmd.PersistentFlags().StringVar(&section, "s", "", "section of ini file (can be empty)")
	rootCmd.PersistentFlags().BoolVar(&strict, "strict", false, "fail with error if key/section doesn't exist")

	rootCmd.AddCommand(getCmd)

	setCmd.Flags().StringVar(&value, "v", "", "value to set")
	setCmd.Flags().BoolVar(&createIniFileIfAbsent, "c", false, "create file if no present")
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(delCmd)

	rootCmd.AddCommand(delSectionCmd)
}
