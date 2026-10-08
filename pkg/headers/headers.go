// Package headers defines Frontal HTTP header names and request defaults.
package headers

import "net/http"

const (
	Authorization = "Authorization"
	Accept        = "Accept"
	ContentType   = "Content-Type"
	RequestID     = "X-Request-ID"
	UserAgent     = "User-Agent"
	JSON          = "application/json"
)

// ApplyDefaults adds JSON response and user-agent headers when callers have not set them.
func ApplyDefaults(header http.Header, userAgent string) http.Header {
	if header == nil {
		header = make(http.Header)
	}
	if header.Get(Accept) == "" {
		header.Set(Accept, JSON)
	}
	if userAgent != "" && header.Get(UserAgent) == "" {
		header.Set(UserAgent, userAgent)
	}
	return header
}

// SetJSON marks a request body as JSON.
func SetJSON(header http.Header) http.Header {
	if header == nil {
		header = make(http.Header)
	}
	if header.Get(ContentType) == "" {
		header.Set(ContentType, JSON)
	}
	return header
}
