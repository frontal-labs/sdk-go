// Package models defines shared response models used by the Frontal SDK.
package models

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError describes an unsuccessful response from the Frontal API.
type APIError struct {
	StatusCode int             `json:"-"`
	Code       string          `json:"code,omitempty"`
	Type       string          `json:"type,omitempty"`
	Message    string          `json:"message,omitempty"`
	RequestID  string          `json:"request_id,omitempty"`
	Details    json.RawMessage `json:"details,omitempty"`
}

// Error formats the API error without including request credentials.
func (apiError *APIError) Error() string {
	if apiError == nil {
		return "frontal: <nil API error>"
	}
	message := apiError.Message
	if message == "" {
		message = http.StatusText(apiError.StatusCode)
	}
	if message == "" {
		message = "unknown API error"
	}
	if apiError.StatusCode == 0 && apiError.Code == "" {
		return "frontal API error: " + message
	}
	if apiError.Code != "" {
		return fmt.Sprintf("frontal API error (status %d, code %s): %s", apiError.StatusCode, apiError.Code, message)
	}
	return fmt.Sprintf("frontal API error (status %d): %s", apiError.StatusCode, message)
}
