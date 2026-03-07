// Package inifile provides shared INI file validation, loading, and saving
// operations used across CLI commands.
package inifile

import (
	"errors"
	"fmt"

	"github.com/sgaunet/gini/internal/tools"
	"gopkg.in/ini.v1"
)

var errNoIniFile = errors.New("specify inifile")

// ValidateAndLoad validates the file path, section, and key, acquires the
// specified lock, and loads the INI file. Use this for commands that operate
// on a specific key (get, set, del).
func ValidateAndLoad(path, section, key string, mode tools.LockMode) (*ini.File, *tools.FileLock, error) {
	if err := tools.ValidateKey(key); err != nil {
		return nil, nil, fmt.Errorf("invalid key: %w", err)
	}

	return ValidateSectionAndLoad(path, section, mode)
}

// ValidateSectionAndLoad validates the file path and section, acquires the
// specified lock, and loads the INI file. Use this for commands that operate
// on a whole section (delsection).
func ValidateSectionAndLoad(path, section string, mode tools.LockMode) (*ini.File, *tools.FileLock, error) {
	if path == "" {
		return nil, nil, errNoIniFile
	}

	if err := tools.ValidateSection(section); err != nil {
		return nil, nil, fmt.Errorf("invalid section: %w", err)
	}

	lock, err := tools.LockFile(path, mode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lock file: %w", err)
	}

	cfg, err := ini.Load(path)
	if err != nil {
		_ = lock.Unlock()
		return nil, nil, fmt.Errorf("fail to load file: %w", err)
	}

	return cfg, lock, nil
}

// SaveConfig atomically saves the INI configuration to the specified path.
func SaveConfig(cfg *ini.File, path string) error {
	if err := tools.AtomicSave(cfg, path); err != nil {
		return fmt.Errorf("fail to save file: %w", err)
	}
	return nil
}
