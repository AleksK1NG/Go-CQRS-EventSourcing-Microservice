package http_client

import (
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	clientTimeout             = 5 * time.Second
	dialContextTimeout        = 5 * time.Second
	clientTLSHandshakeTimeout = 5 * time.Second
	clientRetryWaitTime       = 300 * time.Millisecond
	retryCount                = 3
	xaxIdleConns              = 20
	maxConnsPerHost           = 40
	idleConnTimeout           = 120 * time.Second
	responseHeaderTimeout     = 5 * time.Second
)

func NewHttpClient(debugMode bool) *resty.Client {
	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: dialContextTimeout,
		}).DialContext,
		TLSHandshakeTimeout:   clientTLSHandshakeTimeout,
		MaxIdleConns:          xaxIdleConns,
		MaxConnsPerHost:       maxConnsPerHost,
		IdleConnTimeout:       idleConnTimeout,
		ResponseHeaderTimeout: responseHeaderTimeout,
	}

	client := resty.New().
		SetDebug(debugMode).
		SetTimeout(clientTimeout).
		SetRetryCount(retryCount).
		SetRetryWaitTime(clientRetryWaitTime).
		SetTransport(t)

	return client
}
