package webtest

import (
	"net/http"
	"net/http/httptest"
	"sync"
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
	if client.headers["Content-Type"] != "application/html" {
		t.Error("Initialized client headers did not match expected value")
	}
}

func TestOptionsNilNormalizesToDefault(t *testing.T) {
	client := InitWebClient().
		Options(&WebClientOptions{WriteToFile: true, FilePath: "./tmp.log"}).
		Options(nil)

	if client.options.WriteToFile {
		t.Error("expected WriteToFile to default to false")
	}
	if client.options.FilePath != "" {
		t.Error("expected FilePath to default to empty string")
	}
}

func TestHeadersCopiesInputMap(t *testing.T) {
	received := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Get("X-Test-Header")
		_, _ = w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	headers := map[string]string{"X-Test-Header": "v1"}
	client := InitWebClient().Headers(&headers)
	headers["X-Test-Header"] = "v2"

	res, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CloseResponse(res); err != nil {
		t.Fatal(err)
	}

	got := <-received
	if got != "v1" {
		t.Fatalf("expected copied header value v1, got %q", got)
	}
}

func TestOptionsCopiesInputStruct(t *testing.T) {
	opts := WebClientOptions{
		WriteToFile: true,
		FilePath:    "./initial.log",
	}
	client := InitWebClient()
	client.Options(&opts)

	opts.WriteToFile = false
	opts.FilePath = "./changed.log"

	if !client.options.WriteToFile {
		t.Fatal("expected options to be copied on set")
	}
	if client.options.FilePath != "./initial.log" {
		t.Fatalf("expected copied file path ./initial.log, got %q", client.options.FilePath)
	}
}

func TestConcurrentRequestsWithConfigUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	client := InitWebClient()
	headersA := map[string]string{"X-Test": "A"}
	headersB := map[string]string{"X-Test": "B"}

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	sendErr := func(err error) {
		select {
		case errCh <- err:
		default:
		}
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 40; j++ {
				res, err := client.Get(server.URL)
				if err != nil {
					sendErr(err)
					return
				}
				if err := client.CloseResponse(res); err != nil {
					sendErr(err)
					return
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 400; i++ {
			if i%2 == 0 {
				client.Headers(&headersA)
			} else {
				client.Headers(&headersB)
			}
			if i%3 == 0 {
				client.Options(nil)
			} else {
				client.Options(&WebClientOptions{WriteToFile: false})
			}
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
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
