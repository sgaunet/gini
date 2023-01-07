package tools

import "os"

func IsFileExists(file string) bool {
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		return false
	}
	defer f.Close()
	i, _ := os.Stat(file)
	return !i.IsDir()
}

func TouchFile(filename string) error {
	f, err := os.Create(filename)
	if err == nil {
		f.Close()
	}
	return err
}
