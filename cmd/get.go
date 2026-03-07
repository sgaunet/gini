package cmd

import (
	"fmt"
	"log/slog"

	"github.com/sgaunet/gini/internal/inifile"
	"github.com/sgaunet/gini/internal/tools"
	"github.com/spf13/cobra"
)

// getCmd represents the get command.
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "retrieve a key from an ini file",
	Long: `Retrieve and print the value of a specific key from an INI file.

The get command reads the specified key from a given section and outputs its value to stdout.
By default, if the key does not exist, the command exits successfully with no output (exit code 0).
Use --strict to return an error when the key doesn't exist.

Required flags:
  -f, --file       Path to the INI file to read
  -k, --key        Key name to retrieve
  -s, --section    Section name (use empty string "" for the default section)

Optional flags:
  --strict    Fail with error if key doesn't exist`,
	Example: `  # Get a key from a named section
  gini get -f config.ini -s database -k host

  # Get a key from the default section (empty string)
  gini get -f config.ini -s "" -k app_name

  # Fail if key doesn't exist (strict mode)
  gini get --strict -f config.ini -s database -k host

  # Use in scripts with output capture
  DB_HOST=$(gini get -f config.ini -s database -k host)`,
	RunE: func(_ *cobra.Command, _ []string) error {
		slog.Debug("loading INI file", "file", iniFile, "section", section, "key", key)

		cfg, lock, err := inifile.ValidateAndLoad(iniFile, section, key, tools.SharedLock)
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}
		defer func() { _ = lock.Unlock() }()

		if cfg.Section(section).HasKey(key) {
			v := cfg.Section(section).Key(key).String()
			slog.Debug("key found", "section", section, "key", key, "value", v)
			fmt.Println(v)
			return nil
		}

		slog.Debug("key not found", "section", section, "key", key)
		if strict {
			return fmt.Errorf("key '%s' in section '%s': %w", key, section, errKeyNotFound)
		}
		return nil
	},
}
