// Package tools provides utility functions for file operations.
package tools

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

var (
	// ErrEmptyKey is returned when a key name is empty.
	ErrEmptyKey = errors.New("key name cannot be empty")
	// ErrEmptySection is returned when a section name is empty and required.
	ErrEmptySection = errors.New("section name cannot be empty")
	// ErrKeyWhitespace is returned when a key has leading/trailing whitespace.
	ErrKeyWhitespace = errors.New("key name has leading or trailing whitespace")
	// ErrKeyInvalidChar is returned when a key contains invalid characters.
	ErrKeyInvalidChar = errors.New("key name contains invalid character")
	// ErrSectionWhitespace is returned when a section has leading/trailing whitespace.
	ErrSectionWhitespace = errors.New("section name has leading or trailing whitespace")
	// ErrSectionInvalidChar is returned when a section contains invalid characters.
	ErrSectionInvalidChar = errors.New("section name contains invalid character")
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

// ValidateKey validates a key name for INI files.
// Returns an error if the key is empty or contains invalid characters.
func ValidateKey(key string) error {
	if key == "" {
		return ErrEmptyKey
	}

	// Check for leading/trailing whitespace
	if strings.TrimSpace(key) != key {
		return fmt.Errorf("%w: %q", ErrKeyWhitespace, key)
	}

	// Check for invalid characters that have special meaning in INI format
	invalidChars := []string{"=", "[", "]", ";", "#"}
	for _, char := range invalidChars {
		if strings.Contains(key, char) {
			return fmt.Errorf("%w %q in: %q", ErrKeyInvalidChar, char, key)
		}
	}

	return nil
}

// ValidateSection validates a section name for INI files.
// Returns an error if the section contains invalid characters.
// Empty section is allowed as it represents the default section.
func ValidateSection(section string) error {
	// Empty section is valid (default section)
	if section == "" {
		return nil
	}

	// Check for leading/trailing whitespace
	if strings.TrimSpace(section) != section {
		return fmt.Errorf("%w: %q", ErrSectionWhitespace, section)
	}

	// Check for invalid characters
	invalidChars := []string{"[", "]", "=", ";", "#"}
	for _, char := range invalidChars {
		if strings.Contains(section, char) {
			return fmt.Errorf("%w %q in: %q", ErrSectionInvalidChar, char, section)
		}
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
