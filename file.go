package webtest

import (
	"bytes"
	"encoding/csv"
	"log"
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

func ReadCsv(path string) [][]string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer CloseFile(f)
	res, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func WriteToFile(f *os.File, b []byte) {
	if _, err := f.Write(b); err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte{'\n'}); err != nil {
		log.Fatal(err)
	}
}

func CloseFile(file *os.File) {
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}
}
