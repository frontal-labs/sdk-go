// Package authentication builds requests authenticated with a Frontal API key.
package authentication

import (
	"errors"
	"net/http"
	"strings"

	"github.com/frontal-labs/sdk-go/v2/internal/headers"
)

// ErrInvalidAPIKey is returned when an API key is not in a supported format.
var ErrInvalidAPIKey = errors.New("frontal: API key is not in a supported Frontal key format")

// APIKey stores an opaque Frontal API key without exposing its value in logs.
type APIKey struct {
	value string
}

// NewAPIKey validates and stores an opaque Frontal API key.
func NewAPIKey(value string) (APIKey, error) {
	if len(value) < 9 || len(value) > 128 || strings.ContainsAny(value, " \t\r\n") {
		return APIKey{}, ErrInvalidAPIKey
	}
	var suffix string
	allowHyphen := false
	switch {
	case strings.HasPrefix(value, "frt_"):
		suffix = strings.TrimPrefix(value, "frt_")
		allowHyphen = true
	case strings.HasPrefix(value, "fr_typed"):
		suffix = strings.TrimPrefix(value, "fr_typed")
	default:
		return APIKey{}, ErrInvalidAPIKey
	}
	if suffix == "" {
		return APIKey{}, ErrInvalidAPIKey
	}
	for _, char := range suffix {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_' || allowHyphen && char == '-' {
			continue
		}
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
