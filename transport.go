package webtest

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Transport struct {
	rtp    http.RoundTripper
	dialer *net.Dialer
}

func InitTransport() *Transport {
	t := &Transport{
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}

	t.rtp = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         t.dialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return t
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.rtp.RoundTrip(r)
}

func (t *Transport) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return t.dialer.DialContext(ctx, network, addr)
}
