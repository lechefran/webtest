package webtest

import (
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
)

type closeErrReadCloser struct{}

func (c closeErrReadCloser) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (c closeErrReadCloser) Close() error {
	return errors.New("close failed")
}

func TestReadCsvReturnsErrorForMissingFile(t *testing.T) {
	_, err := ReadCsv("./does-not-exist.csv")
	if err == nil {
		t.Fatal("expected error for missing csv file")
	}
}

func TestWriteToFileReturnsError(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "webtest-*.log")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := WriteToFile(f, []byte("entry")); err == nil {
		t.Fatal("expected write error for closed file")
	}
}

func TestCloseResponseReturnsErrorByDefault(t *testing.T) {
	w := InitWebClient()
	res := &http.Response{
		Body: closeErrReadCloser{},
	}

	if err := w.CloseResponse(res); err == nil {
		t.Fatal("expected close error")
	}
}

func TestAddQueryParamsReturnsErrorForInvalidURL(t *testing.T) {
	w := InitWebClient()

	_, err := w.AddQueryParams("://bad-url", map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected URL parse error")
	}
}
