package utils

import (
	"io/ioutil"
	"os"
)

// CreateDir if not exists
func CreateDir(path string) error {
	exists, err := PathExists(path)
	if err != nil {
		return err
	}
	if !exists {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// PathExists check path exists
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ReadFile return bytes
func ReadFile(filepath string) ([]byte, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// WriteFile use string
func WriteFile(filepath string, data []byte) error {
	if err := ioutil.WriteFile(filepath, data, 0777); err != nil {
		return nil
	}
	return nil
}
