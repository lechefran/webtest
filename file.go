package webtest

import (
	"bytes"
	"encoding/csv"
	"os"
)

type FileSettings struct {
	writeCall, writeHeader, writeRequest, writeResponse bool
}

func InitDefaultFileSettings() *FileSettings {
	return &FileSettings{
		writeCall:     true,
		writeHeader:   false,
		writeRequest:  false,
		writeResponse: false,
	}
}

func FormatStringArray(buf *bytes.Buffer, arr []string, delimiter string) *bytes.Buffer {
	buf.WriteString("[")
	for i := range arr {
		if i != len(arr)-1 {
			buf.WriteString("\"" + arr[i] + "\"" + delimiter + " ")
		} else {
			buf.WriteString("\"" + arr[i] + "\"")
		}
	}
	buf.WriteString("]")
	return buf
}

func ReadCsv(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = CloseFile(f)
	}()

	res, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	return res, nil
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
