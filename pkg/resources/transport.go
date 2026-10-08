package resources

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

// httpTransport owns the HTTP clients used by the public SDK client.
type httpTransport struct {
	client       *http.Client
	streamClient *http.Client
}

// newHTTPTransport creates clients with a bounded request timeout and an unbounded stream timeout.
// When client is supplied, its redirect and transport settings are preserved.
func newHTTPTransport(client *http.Client, timeout time.Duration) *httpTransport {
	if client == nil {
		redirectPolicy := sameOriginRedirect
		return &httpTransport{
			client: &http.Client{
				Timeout:       timeout,
				CheckRedirect: redirectPolicy,
			},
			streamClient: &http.Client{
				CheckRedirect: redirectPolicy,
			},
		}
	}

	streamClient := *client
	streamClient.Timeout = 0
	return &httpTransport{client: client, streamClient: &streamClient}
}

// Do sends a request using the configured request timeout.
func (transport *httpTransport) Do(request *http.Request) (*http.Response, error) {
	if transport == nil || transport.client == nil {
		return nil, errors.New("frontal: HTTP transport is not configured")
	}
	if request == nil {
		return nil, errors.New("frontal: request is nil")
	}
	return transport.client.Do(request)
}

// DoStream sends a request without applying a client-wide timeout. The request context still applies.
func (transport *httpTransport) DoStream(request *http.Request) (*http.Response, error) {
	if transport == nil || transport.streamClient == nil {
		return nil, errors.New("frontal: HTTP stream transport is not configured")
	}
	if request == nil {
		return nil, errors.New("frontal: request is nil")
	}
	return transport.streamClient.Do(request)
}

// CloseIdleConnections closes idle connections used by both clients.
func (transport *httpTransport) CloseIdleConnections() {
	if transport == nil {
		return
	}
	if transport.client != nil {
		transport.client.CloseIdleConnections()
	}
	if transport.streamClient != nil && transport.streamClient != transport.client {
		transport.streamClient.CloseIdleConnections()
	}
}

func sameOriginRedirect(request *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return nil
	}
	previous := via[len(via)-1].URL
	if previous == nil || request.URL == nil ||
		!strings.EqualFold(previous.Scheme, request.URL.Scheme) ||
		!strings.EqualFold(previous.Host, request.URL.Host) {
		return http.ErrUseLastResponse
	}
	return nil
}
