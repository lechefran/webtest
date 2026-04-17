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
