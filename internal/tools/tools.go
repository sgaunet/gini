// Package tools provides utility functions for file operations.
package tools

import (
	"fmt"
	"os"
)

// IsFileExists checks if a file exists and is not a directory.
func IsFileExists(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// TouchFile creates an empty file at the specified path.
func TouchFile(filename string) error {
	f, err := os.Create(filename) // #nosec G304 - file path is intended to be user-provided
	if err == nil {
		if closeErr := f.Close(); closeErr != nil {
			return fmt.Errorf("failed to close file: %w", closeErr)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	return nil
}
