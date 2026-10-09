// Package handlers creates HTTP requests and decodes Frontal API responses.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/frontal-labs/sdk-go/v2/internal/headers"
)

// NewRequest creates an HTTP request with the supplied body reader.
func NewRequest(ctx context.Context, method, target string, body io.Reader) (*http.Request, error) {
	if ctx == nil {
		return nil, fmt.Errorf("frontal: request context is required")
	}
	request, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, fmt.Errorf("frontal: create request: %w", err)
	}
	return request, nil
}

// NewJSONRequest creates an HTTP request with a JSON-encoded body.
func NewJSONRequest(ctx context.Context, method, target string, body any) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("frontal: encode request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}
	request, err := NewRequest(ctx, method, target, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		request.Header = headers.SetJSON(request.Header)
	}
	return request, nil
}
