package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "development"

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version of gini",
	Long: `Print the version information of the gini CLI tool.

The version command displays the current version of gini. During development,
this shows "development". In official releases, this displays the semantic
version number (e.g., v1.2.3).

Examples:
  # Display the current version
  gini version

  # Use in scripts to check version
  if [ "$(gini version)" = "v1.0.0" ]; then
    echo "Version matches"
  fi`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
