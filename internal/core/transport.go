// Package core owns low-level HTTP execution for the Frontal SDK.
package core

import (
	"errors"
	"net/http"
	"time"
)

// Transport owns the HTTP client used by the public SDK client.
type Transport struct {
	client *http.Client
}

// NewTransport creates a transport, using a default client when client is nil.
func NewTransport(client *http.Client, timeout time.Duration) *Transport {
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	return &Transport{client: client}
}

// Do sends an HTTP request.
func (transport *Transport) Do(request *http.Request) (*http.Response, error) {
	if transport == nil || transport.client == nil {
		return nil, errors.New("frontal: HTTP transport is not configured")
	}
	if request == nil {
		return nil, errors.New("frontal: request is nil")
	}
	return transport.client.Do(request)
}

// CloseIdleConnections closes idle connections managed by the HTTP client.
func (transport *Transport) CloseIdleConnections() {
	if transport != nil && transport.client != nil {
		transport.client.CloseIdleConnections()
	}
}
