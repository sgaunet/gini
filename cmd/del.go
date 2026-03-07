package cmd

import (
	"fmt"
	"log/slog"

	"github.com/sgaunet/gini/internal/inifile"
	"github.com/sgaunet/gini/internal/tools"
	"github.com/spf13/cobra"
)

// delCmd represents the del command.
var delCmd = &cobra.Command{
	Use:   "del",
	Short: "delete a key from an ini file",
	Long: `Delete a specific key from a section in an INI file.

The del command removes the specified key from the given section. By default, if the key
doesn't exist, the operation completes successfully without error. Use --strict to return
an error when the key doesn't exist. The file is saved using atomic writes to prevent corruption.

Note: To delete an entire section with all its keys, use the 'delsection' command instead.

Required flags:
  -f, --file       Path to the INI file to modify
  -k, --key        Key name to delete
  -s, --section    Section name (use empty string "" for the default section)

Optional flags:
  --strict    Fail with error if key doesn't exist`,
	Example: `  # Delete a key from a named section
  gini del -f config.ini -s cache -k ttl

  # Delete a key from the default section
  gini del -f config.ini -s "" -k deprecated_option

  # Fail if key doesn't exist (strict mode)
  gini del --strict -f config.ini -s database -k password

  # Remove a database password
  gini del -f config.ini -s database -k password`,
	RunE: func(_ *cobra.Command, _ []string) error {
		slog.Debug("deleting key", "file", cfg.File, "section", cfg.Section, "key", cfg.Key)

		ini, lock, err := inifile.ValidateAndLoad(cfg.File, cfg.Section, cfg.Key, tools.ExclusiveLock)
		if err != nil {
			return fmt.Errorf("del: %w", err)
		}
		defer func() { _ = lock.Unlock() }()

		if cfg.Strict && !ini.Section(cfg.Section).HasKey(cfg.Key) {
			return fmt.Errorf("key '%s' in section '%s': %w", cfg.Key, cfg.Section, errKeyNotFound)
		}

		ini.Section(cfg.Section).DeleteKey(cfg.Key)
		if err := inifile.SaveConfig(ini, cfg.File); err != nil {
			return fmt.Errorf("del: %w", err)
		}
		slog.Debug("key deleted successfully", "file", cfg.File, "section", cfg.Section, "key", cfg.Key)
		return nil
	},
}
