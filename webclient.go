package webtest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

type WebClient struct {
	mu        sync.RWMutex
	client    http.Client
	headers   map[string]string
	transport *Transport
	options   WebClientOptions
}

const (
	logPrefixHeaders = "[HEADERS] "
	logPrefixRequest = "[REQUEST] "
	logPrefixResponse = "[RESPONSE] "
)

func InitWebClient() *WebClient {
	t := InitTransport()
	return &WebClient{
		transport: t,
		headers:   nil,
		client: http.Client{
			Transport: t,
		},
		options: WebClientOptions{},
	}
}

func (w *WebClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return w.execute(req, url)
}

func (w *WebClient) Post(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return w.execute(req, url)
}

func (w *WebClient) Patch(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return w.execute(req, url)
}

func (w *WebClient) Put(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return w.execute(req, url)
}

func (w *WebClient) Delete(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	return w.execute(req, url)
}

func (w *WebClient) Headers(m map[string]string) *WebClient {
	w.mu.Lock()
	defer w.mu.Unlock()

	if m == nil {
		w.headers = nil
		return w
	}
	w.headers = cloneStringMap(m)
	return w
}

func (w *WebClient) Options(o WebClientOptions) *WebClient {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.options = o
	return w
}

func SetHeaders(r *http.Request, m map[string]string) {
	if r == nil || m == nil {
		return
	}
	for k, v := range m {
		r.Header.Set(k, v)
	}
}

func cloneStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	cloned := make(map[string]string, len(m))
	maps.Copy(cloned, m)
	return cloned
}

func defaultLogFilePath(now time.Time) string {
	return "./" + now.UTC().Format("20060102T150405Z") + ".log"
}

func formatRequestHeadersForLog(headers http.Header) (string, error) {
	if len(headers) == 0 {
		return "{}", nil
	}
	body, err := json.Marshal(headers)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func formatRequestBodyForLog(body []byte) (string, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return "{}", nil
	}
	if json.Valid(trimmed) {
		var compact bytes.Buffer
		if err := json.Compact(&compact, trimmed); err != nil {
			return "", err
		}
		return compact.String(), nil
	}

	serialized, err := json.Marshal(map[string]string{"raw": string(body)})
	if err != nil {
		return "", err
	}
	return string(serialized), nil
}

func prepareRequestBodyForLog(req *http.Request) (string, bool, error) {
	if req == nil || req.Body == nil || req.Body == http.NoBody {
		return "", false, nil
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return "", false, err
	}
	if err := req.Body.Close(); err != nil {
		return "", false, err
	}

	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}

	if len(bytes.TrimSpace(bodyBytes)) == 0 {
		return "", false, nil
	}

	bodyLine, err := formatRequestBodyForLog(bodyBytes)
	if err != nil {
		return "", false, err
	}
	return bodyLine, true, nil
}

func prepareResponseBodyForLog(res *http.Response) (string, bool, error) {
	if res == nil || res.Body == nil || res.Body == http.NoBody {
		return "", false, nil
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", false, err
	}
	if err := res.Body.Close(); err != nil {
		return "", false, err
	}

	// Restore the response body so callers can still read it.
	res.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if len(bytes.TrimSpace(bodyBytes)) == 0 {
		return "", false, nil
	}

	bodyLine, err := formatRequestBodyForLog(bodyBytes)
	if err != nil {
		return "", false, err
	}
	return bodyLine, true, nil
}

func (w *WebClient) snapshotConfig() (map[string]string, WebClientOptions) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return cloneStringMap(w.headers), w.options
}

func Is2xxSuccessful(r *http.Response) bool {
	if r == nil {
		return false
	}
	status := r.StatusCode
	return status >= 200 && status <= 299
}

func Is3xxRedirection(r *http.Response) bool {
	if r == nil {
		return false
	}
	status := r.StatusCode
	return status >= 300 && status <= 399
}

func Is4xxClientError(r *http.Response) bool {
	if r == nil {
		return false
	}
	status := r.StatusCode
	return status >= 400 && status <= 499
}

func Is5xxServerError(r *http.Response) bool {
	if r == nil {
		return false
	}
	status := r.StatusCode
	return status >= 500 && status <= 599
}

func (w *WebClient) execute(req *http.Request, url string) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	headers, opts := w.snapshotConfig()

	if headers != nil {
		SetHeaders(req, headers)
	}

	var requestBodyLine string
	hasRequestBodyLine := false
	if opts.WriteToFile && opts.WriteSettings.writeRequest {
		var prepareErr error
		requestBodyLine, hasRequestBodyLine, prepareErr = prepareRequestBodyForLog(req)
		if prepareErr != nil {
			return nil, prepareErr
		}
	}

	start := time.Now()
	var connStart time.Time
	var connDuration time.Duration
	var connReused bool
	trace := &httptrace.ClientTrace{
		ConnectStart: func(_, _ string) {
			connStart = time.Now()
		},
		ConnectDone: func(_, _ string, err error) {
			if err != nil || connStart.IsZero() {
				return
			}
			connDuration += time.Since(connStart)
			connStart = time.Time{}
		},
		GotConn: func(info httptrace.GotConnInfo) {
			connReused = info.Reused
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	res, err := w.client.Do(req)
	totalDuration := time.Since(start)
	if err != nil {
		return res, err
	}

	var responseBodyLine string
	hasResponseBodyLine := false
	if opts.WriteToFile && opts.WriteSettings.writeResponse {
		var prepareErr error
		responseBodyLine, hasResponseBodyLine, prepareErr = prepareResponseBodyForLog(res)
		if prepareErr != nil {
			return res, prepareErr
		}
	}

	var s string
	if res != nil {
		s = req.Method + " " + url + " " + res.Status + " total=" + fmt.Sprintf("%.3fs", totalDuration.Seconds())
		if connReused {
			s += " conn=reused"
		} else {
			s += " connect=" + fmt.Sprintf("%.3fs", connDuration.Seconds())
		}
		if Is2xxSuccessful(res) {
			color.Green(s)
		} else if Is3xxRedirection(res) {
			color.Yellow(s)
		} else {
			color.HiRed(s)
		}
	} else {
		s = req.Method + " " + url + " ERROR total=" + fmt.Sprintf("%.3fs", totalDuration.Seconds())
		if connReused {
			s += " conn=reused"
		} else {
			s += " connect=" + fmt.Sprintf("%.3fs", connDuration.Seconds())
		}
		color.HiRed(s)
	}

	if opts.WriteToFile {
		var fileName string
		if opts.FilePath != "" {
			fileName = opts.FilePath
		} else {
			fileName = defaultLogFilePath(time.Now())
			color.HiBlue("Application logs will be saved to ", fileName)
		}

		if f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return res, err
		} else {
			writeErr := WriteToFile(f, []byte(s))
			if writeErr != nil {
				_ = CloseFile(f)
				return res, writeErr
			}
			if opts.WriteSettings.writeHeaders {
				headerLine, headerFormatErr := formatRequestHeadersForLog(req.Header)
				if headerFormatErr != nil {
					_ = CloseFile(f)
					return res, headerFormatErr
				}
				headerErr := WriteToFile(f, []byte(logPrefixHeaders+headerLine))
				if headerErr != nil {
					_ = CloseFile(f)
					return res, headerErr
				}
			}
			if opts.WriteSettings.writeRequest && hasRequestBodyLine {
				bodyErr := WriteToFile(f, []byte(logPrefixRequest+requestBodyLine))
				if bodyErr != nil {
					_ = CloseFile(f)
					return res, bodyErr
				}
			}
			if opts.WriteSettings.writeResponse && hasResponseBodyLine {
				bodyErr := WriteToFile(f, []byte(logPrefixResponse+responseBodyLine))
				if bodyErr != nil {
					_ = CloseFile(f)
					return res, bodyErr
				}
			}
			closeErr := CloseFile(f)
			if closeErr != nil {
				return res, closeErr
			}
		}
	}

	return res, err
}

func (w *WebClient) CloseIdleConnections() {
	w.client.CloseIdleConnections()
}

func (w *WebClient) CloseResponse(res *http.Response) error {
	if res == nil || res.Body == nil {
		return nil
	}
	if err := res.Body.Close(); err != nil {
		return err
	}
	return nil
}

func (w *WebClient) AddQueryParams(s string, m map[string]string) (string, error) {
	parsed, err := url.Parse(s)
	if err != nil {
		return "", err
	}

	q := parsed.Query()
	for k, v := range m {
		q.Add(k, v)
	}
	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}
