// Package cmd implements the CLI commands for gini, a tool for reading and
// writing INI configuration files. It uses the Cobra framework and provides
// get, set, del, and delsection subcommands.
package cmd

import (
	"errors"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	errKeyNotFound     = errors.New("key not found")
	errSectionNotFound = errors.New("section not found")
)

// Config holds all flag values shared across commands.
type Config struct {
	File    string
	Section string
	Key     string
	Value   string
	Create  bool
	Strict  bool
	Debug   bool
	Quiet   bool
}

var cfg Config

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gini",
	Short: "Tool to get/set key from an ini file.",
	Long: `gini is a CLI utility for performing basic operations on INI configuration files.

It provides commands to read, write, update, and delete keys and sections in INI files.
All operations use atomic writes to prevent file corruption during updates.

Common flags:
  -f, --file string      Path to the INI file (required for all commands)
  -s, --section string   Section name (use empty string for default section)
  -k, --key string       Key name within the section
  --debug                Enable debug logging (outputs to stderr)
  --quiet                Suppress non-error output

Examples:
  gini get -f config.ini -s database -k host
  gini set -f config.ini -s database -k port -v 5432
  gini del -f config.ini -s cache -k ttl
  gini delsection -f config.ini -s deprecated`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		level := slog.LevelInfo
		if cfg.Debug {
			level = slog.LevelDebug
		}
		if cfg.Quiet {
			level = slog.LevelError
		}
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
		return nil
	},
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
	rootCmd.PersistentFlags().StringVarP(&cfg.File, "file", "f", "", "INI file to read/update")
	rootCmd.PersistentFlags().StringVarP(&cfg.Key, "key", "k", "", "key to read or update")
	rootCmd.PersistentFlags().StringVarP(&cfg.Section, "section", "s", "", "section of ini file (can be empty)")
	rootCmd.PersistentFlags().BoolVar(&cfg.Strict, "strict", false, "fail with error if key/section doesn't exist")
	rootCmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&cfg.Quiet, "quiet", false, "suppress non-error output")

	// get command - requires: file, key (section can be empty for default section)
	rootCmd.AddCommand(getCmd)
	_ = getCmd.MarkPersistentFlagRequired("file")
	_ = getCmd.MarkPersistentFlagRequired("key")

	// set command - requires: file, key, value (section can be empty for default section)
	setCmd.Flags().StringVarP(&cfg.Value, "value", "v", "", "value to set")
	setCmd.Flags().BoolVarP(&cfg.Create, "create", "c", false, "create file if not present")
	rootCmd.AddCommand(setCmd)
	_ = setCmd.MarkPersistentFlagRequired("file")
	_ = setCmd.MarkPersistentFlagRequired("key")
	_ = setCmd.MarkFlagRequired("value")

	// del command - requires: file, key (section can be empty for default section)
	rootCmd.AddCommand(delCmd)
	_ = delCmd.MarkPersistentFlagRequired("file")
	_ = delCmd.MarkPersistentFlagRequired("key")

	// delsection command - requires: file, section (section cannot be empty for this command)
	rootCmd.AddCommand(delSectionCmd)
	_ = delSectionCmd.MarkPersistentFlagRequired("file")
	_ = delSectionCmd.MarkPersistentFlagRequired("section")
}
