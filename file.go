package webtest

import (
	"bytes"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
)

func FormatStringArray(buf *bytes.Buffer, arr []string, delimiter string) *bytes.Buffer {
	buf.WriteString("[")
	for i := range arr {
		if i != len(arr)-1 {
			buf.WriteByte('"')
			buf.WriteString(arr[i])
			buf.WriteByte('"')
			buf.WriteString(delimiter)
			buf.WriteByte(' ')
		} else {
			buf.WriteByte('"')
			buf.WriteString(arr[i])
			buf.WriteByte('"')
		}
	}
	buf.WriteString("]")
	return buf
}

func ReadCsv(path string) (records [][]string, err error) {
	cleanPath := filepath.Clean(path)
	f, err := os.Open(cleanPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := CloseFile(f)
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

func WriteToFile(f *os.File, b []byte) error {
	if _, err := f.Write(b); err != nil {
		return err
	}
	if _, err := f.Write([]byte{'\n'}); err != nil {
		return err
	}
	return nil
}

func CloseFile(file *os.File) error {
	if file == nil {
		return nil
	}

	if err := file.Close(); err != nil {
		return err
	}
	return nil
}
