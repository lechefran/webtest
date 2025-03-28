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

func TestTransportDuration(t *testing.T) {
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
	if client.transport.connStart.IsZero() {
		t.Error("Client made a request but connection start time was not updated")
	}
	if client.transport.connEnd.IsZero() {
		t.Error("Client made a request but connection end time was not updated")
	}
	if client.transport.reqStart.IsZero() {
		t.Error("Client made a request but request start time was not updated")
	}
	if client.transport.reqEnd.IsZero() {
		t.Error("Client made a request but request end time was not updated")
	}
	if client.transport.ReqDuration().Seconds() == 0.0 {
		t.Error("Client made a request but client's transport request duration was not updated")
	}
	if client.transport.ConnDuration().Seconds() == 0.0 {
		t.Error("Client made a request but client's transport connection duration was not updated")
	}
	if client.transport.Duration().Seconds() == 0.0 {
		t.Error("Client made a request but client's transport duration was not updated")
	}
}
