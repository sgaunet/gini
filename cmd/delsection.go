package cmd

import (
	"fmt"
	"log/slog"

	"github.com/sgaunet/gini/internal/inifile"
	"github.com/sgaunet/gini/internal/tools"
	"github.com/spf13/cobra"
)

// delSectionCmd represents the delsection command.
var delSectionCmd = &cobra.Command{
	Use:   "delsection",
	Short: "delete all keys of a section",
	Long: `Delete an entire section and all its keys from an INI file.

The delsection command removes the specified section along with all keys it contains.
By default, if the section doesn't exist, the operation completes successfully without error.
Use --strict to return an error when the section doesn't exist. The file is saved using
atomic writes to prevent corruption.

Note: To delete only a specific key within a section, use the 'del' command instead.
Warning: This operation cannot be undone. The entire section will be permanently removed.

Required flags:
  -f, --file       Path to the INI file to modify
  -s, --section    Section name to delete (cannot be empty for this command)

Optional flags:
  --strict    Fail with error if section doesn't exist`,
	Example: `  # Delete an entire section with all its keys
  gini delsection -f config.ini -s deprecated

  # Remove a cache configuration section
  gini delsection -f config.ini -s cache

  # Fail if section doesn't exist (strict mode)
  gini delsection --strict -f config.ini -s test_settings

  # Clean up old test configuration
  gini delsection -f config.ini -s test_settings`,
	RunE: func(_ *cobra.Command, _ []string) error {
		slog.Debug("deleting section", "file", cfg.File, "section", cfg.Section)

		ini, lock, err := inifile.ValidateSectionAndLoad(cfg.File, cfg.Section, tools.ExclusiveLock)
		if err != nil {
			return fmt.Errorf("delsection: %w", err)
		}
		defer func() { _ = lock.Unlock() }()

		if cfg.Strict && !ini.HasSection(cfg.Section) {
			return fmt.Errorf("section '%s': %w", cfg.Section, errSectionNotFound)
		}

		ini.DeleteSection(cfg.Section)
		if err := inifile.SaveConfig(ini, cfg.File); err != nil {
			return fmt.Errorf("delsection: %w", err)
		}
		slog.Debug("section deleted successfully", "file", cfg.File, "section", cfg.Section)
		return nil
	},
}
