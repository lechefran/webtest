package webtest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitTransport(t *testing.T) {
	transport := InitTransport()
	if transport == nil {
		t.Error("No transport was initialized")
	}
}

func TestTransportRoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Hello, World!"))
		if err != nil {
			return
		}
	}))
	defer server.Close()

	client := InitWebClient()
	res, err := client.Get(server.URL)
	if err != nil {
		t.Error(err)
	}
	if !Is2xxSuccessful(res) {
		t.Error(err)
	}
	if err := client.CloseResponse(res); err != nil {
		t.Fatal(err)
	}
}

func TestInitTransportUsesDialContext(t *testing.T) {
	transport := InitTransport()
	httpTransport, ok := transport.rtp.(*http.Transport)
	if !ok {
		t.Fatal("expected transport.rtp to be *http.Transport")
	}
	if httpTransport.DialContext == nil {
		t.Fatal("expected DialContext to be configured")
	}
}

func TestInitTransportPreservesDefaultTransportSettings(t *testing.T) {
	transport := InitTransport()
	httpTransport, ok := transport.rtp.(*http.Transport)
	if !ok {
		t.Fatal("expected transport.rtp to be *http.Transport")
	}

	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		t.Fatal("expected http.DefaultTransport to be *http.Transport")
	}

	if httpTransport == defaultTransport {
		t.Fatal("expected InitTransport to clone the default transport")
	}
	if httpTransport.ForceAttemptHTTP2 != defaultTransport.ForceAttemptHTTP2 {
		t.Fatal("expected ForceAttemptHTTP2 to match the default transport")
	}
	if httpTransport.MaxIdleConns != defaultTransport.MaxIdleConns {
		t.Fatal("expected MaxIdleConns to match the default transport")
	}
	if httpTransport.IdleConnTimeout != defaultTransport.IdleConnTimeout {
		t.Fatal("expected IdleConnTimeout to match the default transport")
	}
}
