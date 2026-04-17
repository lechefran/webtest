package webtest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestInitWebClient(t *testing.T) {
	client := InitWebClient()
	if client == nil {
		t.Error("No client was initialized")
	}
}

func assertCloseResponse(t *testing.T, client *WebClient, res *http.Response) {
	t.Helper()
	if err := client.CloseResponse(res); err != nil {
		t.Fatal(err)
	}
}

func TestSuccessfulGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(headers)
	res, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer assertCloseResponse(t, client, res)
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
	client := InitWebClient().Headers(headers)
	res, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer assertCloseResponse(t, client, res)
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
	client := InitWebClient().Headers(headers)
	res, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer assertCloseResponse(t, client, res)
	if !Is4xxClientError(res) {
		t.Fail()
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
	client := InitWebClient().Headers(headers)
	res, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer assertCloseResponse(t, client, res)
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
	client := InitWebClient().Headers(headers)
	res, err := client.Post(server.URL, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	defer assertCloseResponse(t, client, res)
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
	client := InitWebClient().Headers(headers)
	res, err := client.Patch(server.URL, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	defer assertCloseResponse(t, client, res)
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
	client := InitWebClient().Headers(headers)
	res, err := client.Put(server.URL, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	defer assertCloseResponse(t, client, res)
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
	client := InitWebClient().Headers(headers)
	res, err := client.Delete(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer assertCloseResponse(t, client, res)
	if !Is2xxSuccessful(res) {
		t.Fail()
	}
}

func TestCloseIdleConnections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	client := InitWebClient()
	res, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CloseResponse(res); err != nil {
		t.Fatal(err)
	}

	client.CloseIdleConnections()

	res, err = client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CloseResponse(res); err != nil {
		t.Fatal(err)
	}
}

func TestClientHeaders(t *testing.T) {
	headers := map[string]string{}
	headers["Content-Type"] = "application/html"
	client := InitWebClient().Headers(headers)

	if client.headers == nil {
		t.Error("Initialized client with explicit headers have no headers")
	}
	if client.headers["Content-Type"] != "application/html" {
		t.Error("Initialized client headers did not match expected value")
	}
}

func TestOptionsNilNormalizesToDefault(t *testing.T) {
	client := InitWebClient().
		Options(WebClientOptions{WriteToFile: true, FilePath: "./tmp.log"}).
		Options(WebClientOptions{})

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
	client := InitWebClient().Headers(headers)
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
		WriteToFile:   true,
		FilePath:      "./initial.log",
		WriteHeaders:  true,
		WriteRequest:  true,
		WriteResponse: true,
		LogMetadata:   true,
	}
	client := InitWebClient()
	client.Options(opts)

	opts.WriteToFile = false
	opts.FilePath = "./changed.log"
	opts.WriteHeaders = false
	opts.WriteRequest = false
	opts.WriteResponse = false
	opts.LogMetadata = false

	if !client.options.WriteToFile {
		t.Fatal("expected options to be copied on set")
	}
	if client.options.FilePath != "./initial.log" {
		t.Fatalf("expected copied file path ./initial.log, got %q", client.options.FilePath)
	}
	if !client.options.WriteHeaders || !client.options.WriteRequest || !client.options.WriteResponse || !client.options.LogMetadata {
		t.Fatal("expected logging flags to be copied on set")
	}
}

func TestWriteHeadersToFileWhenEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	tmpFile, err := os.CreateTemp(t.TempDir(), "webtest-headers-*.log")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Test":       "abc123",
	}
	opts := WebClientOptions{
		WriteToFile:  true,
		FilePath:     tmpPath,
		WriteHeaders: true,
	}

	client := InitWebClient().Headers(headers).Options(opts)
	res, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	assertCloseResponse(t, client, res)

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 log lines (call + headers), got %d", len(lines))
	}
	if !strings.HasPrefix(lines[1], logPrefixHeaders) {
		t.Fatalf("expected headers line prefix %q, got %q", logPrefixHeaders, lines[1])
	}

	var headersJSON map[string][]string
	headersPayload := strings.TrimPrefix(lines[1], logPrefixHeaders)
	if err := json.Unmarshal([]byte(headersPayload), &headersJSON); err != nil {
		t.Fatalf("expected valid json headers line, got %q: %v", lines[1], err)
	}
	if len(headersJSON["Content-Type"]) == 0 || headersJSON["Content-Type"][0] != "application/json" {
		t.Fatalf("expected Content-Type header in json log output, got: %#v", headersJSON)
	}
	if len(headersJSON["X-Test"]) == 0 || headersJSON["X-Test"][0] != "abc123" {
		t.Fatalf("expected X-Test header in json log output, got: %#v", headersJSON)
	}
}

func TestWriteRequestToFileWhenEnabled(t *testing.T) {
	requestBody := []byte(`{"query":"pizza","limit":5}`)
	bodyCh := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodyCh <- b
		_ = r.Body.Close()
		_, _ = w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	tmpFile, err := os.CreateTemp(t.TempDir(), "webtest-request-*.log")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	opts := WebClientOptions{
		WriteToFile:  true,
		FilePath:     tmpPath,
		WriteRequest: true,
	}

	client := InitWebClient().Options(opts)
	res, err := client.Post(server.URL, requestBody)
	if err != nil {
		t.Fatal(err)
	}
	assertCloseResponse(t, client, res)

	serverBody := <-bodyCh
	if !bytes.Equal(serverBody, requestBody) {
		t.Fatalf("expected server to receive body %q, got %q", string(requestBody), string(serverBody))
	}

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 log lines (call + request body), got %d", len(lines))
	}
	if !strings.HasPrefix(lines[1], logPrefixRequest) {
		t.Fatalf("expected request line prefix %q, got %q", logPrefixRequest, lines[1])
	}

	var requestJSON map[string]interface{}
	requestPayload := strings.TrimPrefix(lines[1], logPrefixRequest)
	if err := json.Unmarshal([]byte(requestPayload), &requestJSON); err != nil {
		t.Fatalf("expected valid json request body line, got %q: %v", lines[1], err)
	}
	if requestJSON["query"] != "pizza" {
		t.Fatalf("expected request query in json log output, got: %#v", requestJSON)
	}
	if requestJSON["limit"] != float64(5) {
		t.Fatalf("expected request limit in json log output, got: %#v", requestJSON)
	}
}

func TestWriteResponseToFileWhenEnabled(t *testing.T) {
	responseBody := []byte(`{"ok":true,"count":3}`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(responseBody)
	}))
	defer server.Close()

	tmpFile, err := os.CreateTemp(t.TempDir(), "webtest-response-*.log")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	opts := WebClientOptions{
		WriteToFile:   true,
		FilePath:      tmpPath,
		WriteResponse: true,
	}

	client := InitWebClient().Options(opts)
	res, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	gotResponseBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotResponseBody, responseBody) {
		t.Fatalf("expected response body %q, got %q", string(responseBody), string(gotResponseBody))
	}
	assertCloseResponse(t, client, res)

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 log lines (call + response body), got %d", len(lines))
	}
	if !strings.HasPrefix(lines[1], logPrefixResponse) {
		t.Fatalf("expected response line prefix %q, got %q", logPrefixResponse, lines[1])
	}

	var responseJSON map[string]interface{}
	responsePayload := strings.TrimPrefix(lines[1], logPrefixResponse)
	if err := json.Unmarshal([]byte(responsePayload), &responseJSON); err != nil {
		t.Fatalf("expected valid json response body line, got %q: %v", lines[1], err)
	}
	if responseJSON["ok"] != true {
		t.Fatalf("expected response ok=true in json log output, got: %#v", responseJSON)
	}
	if responseJSON["count"] != float64(3) {
		t.Fatalf("expected response count in json log output, got: %#v", responseJSON)
	}
}

func TestWriteMetadataToFileWhenEnabled(t *testing.T) {
	requestBody := []byte(`{"query":"pizza","limit":5}`)
	responseBody := []byte(`{"ok":true,"count":3}`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
		_, _ = w.Write(responseBody)
	}))
	defer server.Close()

	tmpFile, err := os.CreateTemp(t.TempDir(), "webtest-metadata-*.log")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	opts := WebClientOptions{
		WriteToFile: true,
		FilePath:    tmpPath,
		LogMetadata: true,
	}

	client := InitWebClient().Options(opts)
	res, err := client.Post(server.URL, requestBody)
	if err != nil {
		t.Fatal(err)
	}
	assertCloseResponse(t, client, res)

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 1 {
		t.Fatalf("expected at least 1 log line (call), got %d", len(lines))
	}

	callLine := lines[0]
	expectedSent := " sent=" + formatByteSize(len(requestBody))
	expectedReceived := " received=" + formatByteSize(len(responseBody))

	connectIdx := strings.Index(callLine, " connect=")
	if connectIdx < 0 {
		connectIdx = strings.Index(callLine, " conn=reused")
		if connectIdx < 0 {
			t.Fatalf("expected call line to include connection timing or reuse, got %q", callLine)
		}
	}
	sentIdx := strings.Index(callLine, expectedSent)
	if sentIdx < 0 {
		t.Fatalf("expected call line to contain %q, got %q", expectedSent, callLine)
	}
	if sentIdx < connectIdx {
		t.Fatalf("expected sent metadata to appear after connection data, got %q", callLine)
	}

	receivedIdx := strings.Index(callLine, expectedReceived)
	if receivedIdx < 0 {
		t.Fatalf("expected call line to contain %q, got %q", expectedReceived, callLine)
	}
	if receivedIdx < sentIdx {
		t.Fatalf("expected received metadata to appear after sent metadata, got %q", callLine)
	}
}

func TestRequestErrorIsLoggedToFileWhenEnabled(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "webtest-request-error-*.log")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	opts := WebClientOptions{
		WriteToFile: true,
		FilePath:    tmpPath,
	}

	client := InitWebClient().Options(opts)
	client.client.Transport = roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("forced transport failure")
	})

	_, err = client.Get("http://example.com")
	if err == nil {
		t.Fatal("expected request error")
	}

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 1 {
		t.Fatalf("expected at least 1 log line (error call line), got %d", len(lines))
	}
	if !strings.Contains(lines[0], " ERROR total=") {
		t.Fatalf("expected error call line in log output, got %q", lines[0])
	}
}

func TestFormatByteSize(t *testing.T) {
	cases := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "bytes",
			input:    512,
			expected: "512b",
		},
		{
			name:     "kilobytes",
			input:    2048,
			expected: "2.00Kb",
		},
		{
			name:     "megabytes",
			input:    1024 * 1024,
			expected: "1.00Mb",
		},
		{
			name:     "gigabytes",
			input:    1024 * 1024 * 1024,
			expected: "1.00Gb",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatByteSize(tc.input)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
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
				client.Headers(headersA)
			} else {
				client.Headers(headersB)
			}
			if i%3 == 0 {
				client.Options(WebClientOptions{})
			} else {
				client.Options(WebClientOptions{WriteToFile: false})
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

func TestDefaultLogFilePathUsesWindowsSafeTimestamp(t *testing.T) {
	ts := time.Date(2026, time.April, 15, 20, 5, 6, 0, time.FixedZone("UTC-5", -5*60*60))
	got := defaultLogFilePath(ts)
	want := "./20260416T010506Z.log"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if strings.Contains(got, ":") {
		t.Fatalf("expected filename without ':', got %q", got)
	}
}

func TestSetHeadersWithNilRequest(t *testing.T) {
	SetHeaders(nil, map[string]string{"key": "value"})
}

func TestSetHeadersWithNilMap(t *testing.T) {
	req := &http.Request{
		Header: map[string][]string{},
	}

	SetHeaders(req, nil)
	if req.Header.Get("key") != "" {
		t.Error("Expected header map to remain unchanged")
	}
}

func TestIs2xxSuccessful(t *testing.T) {
	mockRes := http.Response{
		StatusCode: 200,
	}

	if !Is2xxSuccessful(&mockRes) {
		t.Error("Response did not return a 2xx status code")
	}
	if Is2xxSuccessful(nil) {
		t.Error("Nil response should not return a 2xx status code")
	}
}

func TestIs3xxRedirection(t *testing.T) {
	mockRes := http.Response{
		StatusCode: 301,
	}

	if !Is3xxRedirection(&mockRes) {
		t.Error("Response did not return a 3xx status code")
	}
	if Is3xxRedirection(nil) {
		t.Error("Nil response should not return a 3xx status code")
	}
}

func TestIs4xxClientError(t *testing.T) {
	mockRes := http.Response{
		StatusCode: 400,
	}

	if !Is4xxClientError(&mockRes) {
		t.Error("Response did not return a 4xx status code")
	}
	if Is4xxClientError(nil) {
		t.Error("Nil response should not return a 4xx status code")
	}
}

func TestIs5xxServerError(t *testing.T) {
	mockRes := http.Response{
		StatusCode: 500,
	}

	if !Is5xxServerError(&mockRes) {
		t.Error("Response did not return a 5xx status code")
	}
	if Is5xxServerError(nil) {
		t.Error("Nil response should not return a 5xx status code")
	}
}
