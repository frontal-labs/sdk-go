// Package authentication builds requests authenticated with a Frontal API key.
package authentication

import (
	"errors"
	"net/http"
	"strings"

	"github.com/frontal-labs/sdk-go/pkg/headers"
)

var ErrInvalidAPIKey = errors.New("frontal: API key is empty or contains whitespace")

// APIKey stores an opaque Frontal API key without exposing its value in logs.
type APIKey struct {
	value string
}

// NewAPIKey validates and stores an opaque Frontal API key.
func NewAPIKey(value string) (APIKey, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, " \t\r\n") {
		return APIKey{}, ErrInvalidAPIKey
	}
	return APIKey{value: value}, nil
}

// Apply adds the API key as an HTTP Bearer token.
func (key APIKey) Apply(request *http.Request) error {
	if request == nil {
		return errors.New("frontal: cannot authenticate a nil request")
	}
	if key.value == "" {
		return ErrInvalidAPIKey
	}
	if request.Header == nil {
		request.Header = make(http.Header)
	}
	request.Header.Set(headers.Authorization, "Bearer "+key.value)
	return nil
}

// String redacts the key when it is formatted with %v.
func (APIKey) String() string { return "[REDACTED]" }

// GoString redacts the key when it is formatted with %#v.
func (APIKey) GoString() string { return "authentication.APIKey([REDACTED])" }
