package cmd

import (
	"fmt"

	"github.com/sgaunet/gini/internal/tools"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
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
  -i    Path to the INI file to modify
  -s    Section name to delete (cannot be empty for this command)

Optional flags:
  --strict    Fail with error if section doesn't exist

Examples:
  # Delete an entire section with all its keys
  gini delsection -i config.ini -s deprecated

  # Remove a cache configuration section
  gini delsection -i config.ini -s cache

  # Fail if section doesn't exist (strict mode)
  gini delsection --strict -i config.ini -s test_settings

  # Clean up old test configuration
  gini delsection -i config.ini -s test_settings`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if iniFile == "" {
			return errNoIniFile
		}
		if err := tools.ValidateSection(section); err != nil {
			return fmt.Errorf("invalid section: %w", err)
		}

		cfg, err := ini.Load(iniFile)
		if err != nil {
			return fmt.Errorf("fail to load file: %w", err)
		}

		if strict && !cfg.HasSection(section) {
			return fmt.Errorf("section '%s' not found", section)
		}

		cfg.DeleteSection(section)
		err = tools.AtomicSave(cfg, iniFile)
		if err != nil {
			return fmt.Errorf("fail to save file: %w", err)
		}
		return nil
	},
}
