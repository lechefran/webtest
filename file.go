package webtest

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
)

// ReadCsv reads CSV records from path. Callers must not pass paths influenced
// by untrusted input unless they first constrain them to an approved directory.
// filepath.Clean normalizes a path but does not prevent traversal or symlink escape.
func ReadCsv(path string) (records [][]string, err error) {
	cleanPath := filepath.Clean(path)
	f, err := os.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := closeFile(f)
		if closeErr == nil {
			return
		}
		if err == nil {
			err = closeErr
			return
		}
		err = errors.Join(err, closeErr)
	}()

	records, err = csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func writeToFile(f *os.File, b []byte) error {
	if _, err := f.Write(b); err != nil {
		return err
	}
	if _, err := f.Write([]byte{'\n'}); err != nil {
		return err
	}
	return nil
}

func closeFile(file *os.File) error {
	if file == nil {
		return nil
	}

	if err := file.Close(); err != nil {
		return err
	}
	return nil
}
