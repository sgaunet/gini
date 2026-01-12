// Package cmd contains all CLI commands for the gini application.
package cmd

import (
	"fmt"

	"github.com/sgaunet/gini/internal/tools"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
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
  -i    Path to the INI file to modify
  -k    Key name to delete
  -s    Section name (use empty string "" for the default section)

Optional flags:
  --strict    Fail with error if key doesn't exist

Examples:
  # Delete a key from a named section
  gini del -i config.ini -s cache -k ttl

  # Delete a key from the default section
  gini del -i config.ini -s "" -k deprecated_option

  # Fail if key doesn't exist (strict mode)
  gini del --strict -i config.ini -s database -k password

  # Remove a database password
  gini del -i config.ini -s database -k password`,
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

		if strict && !cfg.Section(section).HasKey(key) {
			return fmt.Errorf("key '%s' not found in section '%s'", key, section)
		}

		cfg.Section(section).DeleteKey(key)
		err = tools.AtomicSave(cfg, iniFile)
		if err != nil {
			return fmt.Errorf("fail to save file: %w", err)
		}
		return nil
	},
}
