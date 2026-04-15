package webtest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitWebClient(t *testing.T) {
	client := InitWebClient()
	if client == nil {
		t.Error("No client was initialized")
	}
}

func TestSuccessfulGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(&headers)
	res, err := client.Get(server.URL)
	if err != nil {
		t.Error(err)
	}
	if !Is2xxSuccessful(res) {
		t.Fail()
	}
}

func TestRedirectedGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(300)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(&headers)
	res, err := client.Get(server.URL)
	if err != nil {
		t.Error(err)
	}
	if !Is3xxRedirection(res) {
		t.Fail()
	}
}

func TestClientErrorGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(&headers)
	res, err := client.Get(server.URL)
	if err != nil {
		t.Error(err)
	}
	if !Is4xxClientError(res) {
		t.Error(err)
	}
}

func TestServerErrorGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(&headers)
	res, err := client.Get(server.URL)
	if err != nil {
		t.Error(err)
	}
	if !Is5xxServerError(res) {
		t.Fail()
	}
}

func TestSuccessfulPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(&headers)
	res, err := client.Post(server.URL, []byte{})
	if err != nil {
		t.Error(err)
	}
	if !Is2xxSuccessful(res) {
		t.Fail()
	}
}

func TestSuccessfulPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(&headers)
	res, err := client.Patch(server.URL, []byte{})
	if err != nil {
		t.Error(err)
	}
	if !Is2xxSuccessful(res) {
		t.Fail()
	}
}

func TestSuccessfulPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(&headers)
	res, err := client.Put(server.URL, []byte{})
	if err != nil {
		t.Error(err)
	}
	if !Is2xxSuccessful(res) {
		t.Fail()
	}
}

func TestSuccessfulDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(&headers)
	res, err := client.Delete(server.URL)
	if err != nil {
		t.Error(err)
	}
	if !Is2xxSuccessful(res) {
		t.Fail()
	}
}

func TestClientHeaders(t *testing.T) {
	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(&headers)

	if client.headers == nil {
		t.Error("Initialized client with explicit headers have no headers")
	}
}

func TestOptionsNilNormalizesToDefault(t *testing.T) {
	client := InitWebClient().Options(nil)

	if client.options == nil {
		t.Fatal("expected default options after passing nil options")
	}
	if client.options.WriteToFile {
		t.Error("expected WriteToFile to default to false")
	}
	if client.options.FilePath != "" {
		t.Error("expected FilePath to default to empty string")
	}
}

func TestExecuteNormalizesNilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	client := InitWebClient()
	client.options = nil

	res, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CloseResponse(res); err != nil {
		t.Fatal(err)
	}
	if client.options == nil {
		t.Fatal("expected options to be normalized during request execution")
	}
}

func TestClientSetHeaders(t *testing.T) {
	headers := map[string][]string{}
	headers["Content-Type"] = []string{"application/html"}
	req := http.Request{
		Header: headers,
	}

	SetHeaders(&req, map[string]string{"key": "value"})
	if req.Header.Get("key") != "value" {
		t.Error("Request has no headers attached")
	}
}

func TestIs2xxSuccessful(t *testing.T) {
	mockRes := http.Response{
		StatusCode: 200,
	}

	if !Is2xxSuccessful(&mockRes) {
		t.Error("Response did not return a 2xx status code")
	}
}

func TestIs3xxRedirection(t *testing.T) {
	mockRes := http.Response{
		StatusCode: 301,
	}

	if !Is3xxRedirection(&mockRes) {
		t.Error("Response did not return a 3xx status code")
	}
}

func TestIs4xxClientError(t *testing.T) {
	mockRes := http.Response{
		StatusCode: 400,
	}

	if !Is4xxClientError(&mockRes) {
		t.Error("Response did not return a 4xx status code")
	}
}

func TestIs5xxServerError(t *testing.T) {
	mockRes := http.Response{
		StatusCode: 500,
	}

	if !Is5xxServerError(&mockRes) {
		t.Error("Response did not return a 5xx status code")
	}
}
