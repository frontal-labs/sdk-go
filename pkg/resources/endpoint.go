// Package resources provides endpoint descriptors and request data for Frontal API resources.
package resources

import (
	"net/http"
	"net/url"
)

// Endpoint identifies an operation from the committed endpoint inventory.
type Endpoint struct {
	Service string `json:"service"`
	Method  string `json:"method"`
	Path    string `json:"path"`
}

// Request describes a contract-backed request sent by Client.Call.
type Request struct {
	Endpoint   Endpoint
	PathParams []string
	Query      url.Values
	Headers    http.Header
	Body       any
}
