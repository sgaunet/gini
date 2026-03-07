package cmd

import (
	"fmt"
	"log/slog"

	"github.com/sgaunet/gini/internal/inifile"
	"github.com/sgaunet/gini/internal/tools"
	"github.com/spf13/cobra"
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
  -f, --file       Path to the INI file to update
  -k, --key        Key name to set
  -s, --section    Section name (use empty string "" for the default section)
  -v, --value      Value to set for the key

Optional flags:
  -c, --create    Create the INI file if it doesn't exist (default: false)`,
	Example: `  # Set a key in a named section
  gini set -f config.ini -s database -k host -v localhost

  # Set a key in the default section
  gini set -f config.ini -s "" -k version -v 1.0.0

  # Create file if it doesn't exist
  gini set -f newconfig.ini -s app -k name -v myapp -c`,
	RunE: func(_ *cobra.Command, _ []string) error {
		slog.Debug("setting key", "file", iniFile, "section", section, "key", key, "value", value)

		if createIniFileIfAbsent && !tools.IsFileExists(iniFile) {
			slog.Debug("creating INI file", "file", iniFile)
			if err := tools.TouchFile(iniFile); err != nil {
				return fmt.Errorf("can't create file: %w", err)
			}
		}

		cfg, lock, err := inifile.ValidateAndLoad(iniFile, section, key, tools.ExclusiveLock)
		if err != nil {
			return fmt.Errorf("set: %w", err)
		}
		defer func() { _ = lock.Unlock() }()

		cfg.Section(section).Key(key).SetValue(value)
		if err := inifile.SaveConfig(cfg, iniFile); err != nil {
			return fmt.Errorf("set: %w", err)
		}
		slog.Debug("key set successfully", "file", iniFile, "section", section, "key", key)
		return nil
	},
}
