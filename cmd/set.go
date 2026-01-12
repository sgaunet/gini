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
	Long: `Add a new key or update an existing key's value in an INI file.

The set command creates or updates a key within the specified section. If the key already
exists, its value is updated. If the key doesn't exist, it is created. The file is saved
using atomic writes to prevent corruption.

Required flags:
  -i    Path to the INI file to update
  -k    Key name to set
  -s    Section name (use empty string "" for the default section)
  -v    Value to set for the key

Optional flags:
  -c    Create the INI file if it doesn't exist (default: false)

Examples:
  # Set a key in a named section
  gini set -i config.ini -s database -k host -v localhost

  # Set a key in the default section
  gini set -i config.ini -s "" -k version -v 1.0.0

  # Create file if it doesn't exist
  gini set -i newconfig.ini -s app -k name -v myapp -c`,
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
		err = tools.AtomicSave(cfg, iniFile)
		if err != nil {
			return fmt.Errorf("fail to save file: %w", err)
		}
		return nil
	},
}
