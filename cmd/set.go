package cmd

import (
	"fmt"
	"log/slog"

	"github.com/sgaunet/gini/internal/inifile"
	"github.com/sgaunet/gini/internal/tools"
	"github.com/spf13/cobra"
)

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
		slog.Debug("setting key", "file", cfg.File, "section", cfg.Section, "key", cfg.Key, "value", cfg.Value)

		if cfg.Create && !tools.IsFileExists(cfg.File) {
			slog.Debug("creating INI file", "file", cfg.File)
			if err := tools.TouchFile(cfg.File); err != nil {
				return fmt.Errorf("can't create file: %w", err)
			}
		}

		ini, lock, err := inifile.ValidateAndLoad(cfg.File, cfg.Section, cfg.Key, tools.ExclusiveLock)
		if err != nil {
			return fmt.Errorf("set: %w", err)
		}
		defer func() { _ = lock.Unlock() }()

		ini.Section(cfg.Section).Key(cfg.Key).SetValue(cfg.Value)
		if err := inifile.SaveConfig(ini, cfg.File); err != nil {
			return fmt.Errorf("set: %w", err)
		}
		slog.Debug("key set successfully", "file", cfg.File, "section", cfg.Section, "key", cfg.Key)
		return nil
	},
}
