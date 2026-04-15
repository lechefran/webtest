package webtest

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fatih/color"
)

type WebClient struct {
	client    http.Client
	headers   *map[string]string
	transport *Transport
	options   *WebClientOptions
}

func InitWebClient() *WebClient {
	t := InitTransport()
	return &WebClient{
		transport: t,
		headers:   nil,
		client: http.Client{
			Transport: t,
		},
		options: &WebClientOptions{},
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

func (w *WebClient) Headers(m *map[string]string) *WebClient {
	w.headers = m
	return w
}

func (w *WebClient) Options(o *WebClientOptions) *WebClient {
	if o == nil {
		w.options = &WebClientOptions{}
		return w
	}
	w.options = o
	return w
}

func SetHeaders(r *http.Request, m map[string]string) {
	for k, v := range m {
		r.Header.Set(k, v)
	}
}

func Is2xxSuccessful(r *http.Response) bool {
	status := r.StatusCode
	return status >= 200 && status <= 299
}

func Is3xxRedirection(r *http.Response) bool {
	status := r.StatusCode
	return status >= 300 && status <= 399
}

func Is4xxClientError(r *http.Response) bool {
	status := r.StatusCode
	return status >= 400 && status <= 499
}

func Is5xxServerError(r *http.Response) bool {
	status := r.StatusCode
	return status >= 500 && status <= 599
}

func (w *WebClient) execute(req *http.Request, url string) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	if w.headers != nil {
		SetHeaders(req, *w.headers)
	}

	res, err := w.client.Do(req)
	if err != nil {
		return res, err
	}

	var s string
	if res != nil {
		s = req.Method + " " + url + " " + res.Status + " " + fmt.Sprintf("%.3fs", w.transport.Duration().Seconds())
		if Is2xxSuccessful(res) {
			color.Green(s)
		} else if Is3xxRedirection(res) {
			color.Yellow(s)
		} else {
			color.HiRed(s)
		}
	} else {
		s = req.Method + " " + url + " ERROR " + fmt.Sprintf("%.3fs", w.transport.Duration().Seconds())
		color.HiRed(s)
	}

	if w.options != nil && w.options.WriteToFile {
		var fileName string
		if w.options.FilePath != "" {
			fileName = w.options.FilePath
		} else {
			fileName = "./" + string(time.Now().Format(time.RFC3339)) + ".log"
			color.HiBlue("Application logs will be saved to ", fileName)
		}

		if f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return res, err
		} else {
			writeErr := WriteToFile(f, []byte(s))
			closeErr := CloseFile(f)
			if writeErr != nil {
				return res, writeErr
			}
			if closeErr != nil {
				return res, closeErr
			}
		}
	}

	w.client.CloseIdleConnections()
	return res, err
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
