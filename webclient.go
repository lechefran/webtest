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
	logPrefixHeaders  = "[HEADERS] "
	logPrefixRequest  = "[REQUEST] "
	logPrefixResponse = "[RESPONSE] "
	defaultRequestTimeout = 30 * time.Second
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

func formatByteSize(n int) string {
	if n < 1024 {
		return fmt.Sprintf("%db", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.2fKb", float64(n)/1024)
	}
	if n < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMb", float64(n)/(1024*1024))
	}
	return fmt.Sprintf("%.2fGb", float64(n)/(1024*1024*1024))
}

func prepareRequestBodyForLog(req *http.Request) (string, bool, int, error) {
	if req == nil || req.Body == nil || req.Body == http.NoBody {
		return "", false, 0, nil
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return "", false, 0, err
	}
	if err := req.Body.Close(); err != nil {
		return "", false, 0, err
	}

	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}

	if len(bytes.TrimSpace(bodyBytes)) == 0 {
		return "", false, len(bodyBytes), nil
	}

	bodyLine, err := formatRequestBodyForLog(bodyBytes)
	if err != nil {
		return "", false, len(bodyBytes), err
	}
	return bodyLine, true, len(bodyBytes), nil
}

func prepareResponseBodyForLog(res *http.Response) (string, bool, int, error) {
	if res == nil || res.Body == nil || res.Body == http.NoBody {
		return "", false, 0, nil
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", false, 0, err
	}
	if err := res.Body.Close(); err != nil {
		return "", false, 0, err
	}

	// Restore the response body so callers can still read it.
	res.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if len(bytes.TrimSpace(bodyBytes)) == 0 {
		return "", false, len(bodyBytes), nil
	}

	bodyLine, err := formatRequestBodyForLog(bodyBytes)
	if err != nil {
		return "", false, len(bodyBytes), err
	}
	return bodyLine, true, len(bodyBytes), nil
}

type bodyLogPayload struct {
	line    string
	hasLine bool
	bytes   int
}

type requestTiming struct {
	totalDuration time.Duration
	connDuration  time.Duration
	connReused    bool
}

func shouldPrepareRequestBodyForLog(opts WebClientOptions) bool {
	return opts.WriteToFile && (opts.WriteRequest || opts.LogMetadata)
}

func shouldPrepareResponseBodyForLog(opts WebClientOptions) bool {
	return opts.WriteToFile && (opts.WriteResponse || opts.LogMetadata)
}

func effectiveRequestTimeout(opts WebClientOptions) time.Duration {
	if opts.RequestTimeout > 0 {
		return opts.RequestTimeout
	}
	return defaultRequestTimeout
}

func prepareRequestPayloadForLog(req *http.Request, opts WebClientOptions) (bodyLogPayload, error) {
	if !shouldPrepareRequestBodyForLog(opts) {
		return bodyLogPayload{}, nil
	}

	line, hasLine, bodyBytes, err := prepareRequestBodyForLog(req)
	if err != nil {
		return bodyLogPayload{}, err
	}

	return bodyLogPayload{
		line:    line,
		hasLine: hasLine,
		bytes:   bodyBytes,
	}, nil
}

func prepareResponsePayloadForLog(res *http.Response, opts WebClientOptions) (bodyLogPayload, error) {
	if !shouldPrepareResponseBodyForLog(opts) {
		return bodyLogPayload{}, nil
	}

	line, hasLine, bodyBytes, err := prepareResponseBodyForLog(res)
	if err != nil {
		return bodyLogPayload{}, err
	}

	return bodyLogPayload{
		line:    line,
		hasLine: hasLine,
		bytes:   bodyBytes,
	}, nil
}

func buildCallSummary(
	req *http.Request,
	requestURL string,
	res *http.Response,
	timing requestTiming,
	opts WebClientOptions,
	requestBytes int,
	responseBytes int,
	requestErr error,
) string {
	var summary string
	if requestErr != nil {
		summary = req.Method + " " + requestURL + " ERROR total=" + fmt.Sprintf("%.3fs", timing.totalDuration.Seconds())
	} else if res != nil {
		summary = req.Method + " " + requestURL + " " + res.Status + " total=" + fmt.Sprintf("%.3fs", timing.totalDuration.Seconds())
	} else {
		summary = req.Method + " " + requestURL + " ERROR total=" + fmt.Sprintf("%.3fs", timing.totalDuration.Seconds())
	}

	if timing.connReused {
		summary += " conn=reused"
	} else {
		summary += " connect=" + fmt.Sprintf("%.3fs", timing.connDuration.Seconds())
	}

	if opts.LogMetadata {
		summary += " sent=" + formatByteSize(requestBytes) + " received=" + formatByteSize(responseBytes)
	}

	return summary
}

func printCallSummary(summary string, res *http.Response, requestErr error) {
	if requestErr != nil || res == nil {
		color.HiRed(summary)
		return
	}

	if Is2xxSuccessful(res) {
		color.Green(summary)
		return
	}
	if Is3xxRedirection(res) {
		color.Yellow(summary)
		return
	}
	color.HiRed(summary)
}

func resolveLogFilePath(opts WebClientOptions) string {
	if opts.FilePath != "" {
		return opts.FilePath
	}

	fileName := defaultLogFilePath(time.Now())
	color.HiBlue("Application logs will be saved to ", fileName)
	return fileName
}

func writeRequestLogs(
	opts WebClientOptions,
	req *http.Request,
	callSummary string,
	requestPayload bodyLogPayload,
	responsePayload bodyLogPayload,
) error {
	if !opts.WriteToFile {
		return nil
	}

	fileName := resolveLogFilePath(opts)
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if err := WriteToFile(f, []byte(callSummary)); err != nil {
		_ = CloseFile(f)
		return err
	}

	if opts.WriteHeaders {
		headerLine, err := formatRequestHeadersForLog(req.Header)
		if err != nil {
			_ = CloseFile(f)
			return err
		}
		if err := WriteToFile(f, []byte(logPrefixHeaders+headerLine)); err != nil {
			_ = CloseFile(f)
			return err
		}
	}

	if opts.WriteRequest && requestPayload.hasLine {
		if err := WriteToFile(f, []byte(logPrefixRequest+requestPayload.line)); err != nil {
			_ = CloseFile(f)
			return err
		}
	}

	if opts.WriteResponse && responsePayload.hasLine {
		if err := WriteToFile(f, []byte(logPrefixResponse+responsePayload.line)); err != nil {
			_ = CloseFile(f)
			return err
		}
	}

	return CloseFile(f)
}

func (w *WebClient) doRequestWithTiming(req *http.Request, timeout time.Duration) (*http.Response, requestTiming, error) {
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

	tracedReq := req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	client := w.client
	client.Timeout = timeout
	res, err := client.Do(tracedReq)
	timing := requestTiming{
		totalDuration: time.Since(start),
		connDuration:  connDuration,
		connReused:    connReused,
	}

	if err != nil {
		return res, timing, err
	}
	return res, timing, nil
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
	SetHeaders(req, headers)

	requestPayload, err := prepareRequestPayloadForLog(req, opts)
	if err != nil {
		return nil, err
	}

	res, timing, err := w.doRequestWithTiming(req, effectiveRequestTimeout(opts))
	if err != nil {
		summary := buildCallSummary(req, url, res, timing, opts, requestPayload.bytes, 0, err)
		printCallSummary(summary, res, err)
		logErr := writeRequestLogs(opts, req, summary, requestPayload, bodyLogPayload{})
		if logErr != nil {
			return res, errors.Join(err, logErr)
		}
		return res, err
	}

	responsePayload, err := prepareResponsePayloadForLog(res, opts)
	if err != nil {
		return res, err
	}

	summary := buildCallSummary(req, url, res, timing, opts, requestPayload.bytes, responsePayload.bytes, nil)
	printCallSummary(summary, res, nil)

	if err := writeRequestLogs(opts, req, summary, requestPayload, responsePayload); err != nil {
		return res, err
	}

	return res, nil
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
