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
	Long: `Retrieve and print the value of a specific key from an INI file.

The get command reads the specified key from a given section and outputs its value to stdout.
If the key does not exist, the command exits successfully with no output (exit code 0).

Required flags:
  -i    Path to the INI file to read
  -k    Key name to retrieve
  -s    Section name (use empty string "" for the default section)

Examples:
  # Get a key from a named section
  gini get -i config.ini -s database -k host

  # Get a key from the default section (empty string)
  gini get -i config.ini -s "" -k app_name

  # Use in scripts with output capture
  DB_HOST=$(gini get -i config.ini -s database -k host)`,
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
