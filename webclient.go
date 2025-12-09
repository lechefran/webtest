package webtest

import (
	"bytes"
	"errors"
	"fmt"
	"log"
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
		options: &WebClientOptions{
			HandleError: DEFAULT,
		},
	}
}

func (w *WebClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if w.options.HandleError != DEFAULT && err != nil {
		if w.options.HandleError == SOFT {
			color.HiRed("Error creating GET request: ", err)
		} else if w.options.HandleError == STRICT {
			return nil, err
		}
	}

	res, err := w.execute(req, url)
	if w.options.HandleError == SOFT && err != nil {
		color.HiRed("Error executing GET request: ", err)
		return res, nil
	}
	return res, err
}

func (w *WebClient) Post(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if w.options.HandleError != DEFAULT && err != nil {
		if w.options.HandleError == SOFT {
			color.HiRed("Error creating POST request: ", err)
		} else if w.options.HandleError == STRICT {
			return nil, err
		}
	}

	res, err := w.execute(req, url)
	if w.options.HandleError == SOFT && err != nil {
		color.HiRed("Error executing POST request: ", err)
		return res, nil
	}
	return res, err
}

func (w *WebClient) Patch(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
	if w.options.HandleError != DEFAULT && err != nil {
		if w.options.HandleError == SOFT {
			color.HiRed("Error creating PATCH request: ", err)
		} else if w.options.HandleError == STRICT {
			return nil, err
		}
	}

	res, err := w.execute(req, url)
	if w.options.HandleError == SOFT && err != nil {
		color.HiRed("Error executing PATCH request: ", err)
		return res, nil
	}
	return res, err
}

func (w *WebClient) Put(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if w.options.HandleError != DEFAULT && err != nil {
		if w.options.HandleError == SOFT {
			color.HiRed("Error creating PUT request: ", err)
		} else if w.options.HandleError == STRICT {
			return nil, err
		}
	}

	res, err := w.execute(req, url)
	if w.options.HandleError == SOFT && err != nil {
		color.HiRed("Error executing PUT request: ", err)
		return res, nil
	}
	return res, err
}

func (w *WebClient) Delete(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if w.options.HandleError != DEFAULT && err != nil {
		if w.options.HandleError == SOFT {
			color.HiRed("Error creating DELETE request: ", err)
		} else if w.options.HandleError == STRICT {
			return nil, err
		}
	}

	res, err := w.execute(req, url)
	if w.options.HandleError == SOFT && err != nil {
		color.HiRed("Error executing DELETE request: ", err)
		return res, nil
	}
	return res, err
}

func (w *WebClient) Headers(m *map[string]string) *WebClient {
	w.headers = m
	return w
}

func (w *WebClient) Options(o *WebClientOptions) *WebClient {
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
	if w.options != nil && w.options.HandleError != DEFAULT && err != nil {
		if w.options.HandleError == SOFT {
			color.HiRed("Error executing request: ", err)
		} else if w.options.HandleError == STRICT {
			return nil, err
		}
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
			if w.options.HandleError != DEFAULT {
				if w.options.HandleError == SOFT {
					color.HiRed("Error creating or opening file: ", err)
				} else if w.options.HandleError == STRICT {
					return nil, err
				}
			}
		} else {
			WriteToFile(f, []byte(s))
			CloseFile(f)
		}
	}

	w.client.CloseIdleConnections()
	if w.options.HandleError == SOFT && err != nil {
		return res, nil
	}
	return res, err
}

func (w *WebClient) CloseResponse(res *http.Response) {
	if res == nil || res.Body == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			switch {
			case w.options != nil && w.options.HandleError == STRICT:
				log.Fatal("panic while closing response body: ", r)
			case w.options != nil && w.options.HandleError == SOFT:
				color.HiRed("Recovered from panic while closing response body: ", r)
			default:
				log.Printf("Recovered from panic while closing response body: %v", r)
			}
		}
	}()

	if cerr := res.Body.Close(); cerr != nil {
		panic(cerr)
	}
}

func (w *WebClient) AddQueryParams(s string, m map[string]string) string {
	parsed, err := url.Parse(s)
	if err != nil {
		log.Fatal(err)
	}

	q := parsed.Query()
	for k, v := range m {
		q.Add(k, v)
	}
	parsed.RawQuery = q.Encode()
	return parsed.String()
}
