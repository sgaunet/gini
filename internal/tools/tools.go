// Package tools provides utility functions for file operations.
package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
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

// AtomicSave atomically saves an ini.File to the specified path.
// It writes to a temporary file first, then renames it to prevent corruption.
func AtomicSave(cfg *ini.File, targetPath string) error {
	// Create temp file in the same directory as target
	dir := filepath.Dir(targetPath)
	tempFile, err := os.CreateTemp(dir, ".gini-*.tmp") // #nosec G304 - file path is intended to be user-provided
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Clean up temp file on error
	defer func() {
		if err != nil {
			_ = os.Remove(tempPath)
		}
	}()

	// Close the temp file so ini.SaveTo can write to it
	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Save to temp file
	if err = cfg.SaveTo(tempPath); err != nil {
		return fmt.Errorf("failed to save to temp file: %w", err)
	}

	// Atomic rename (on UNIX systems)
	if err = os.Rename(tempPath, targetPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
