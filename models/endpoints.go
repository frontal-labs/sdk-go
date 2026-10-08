//go:generate python3 ../scripts/generate_endpoints.py

package models

import "strings"

// Endpoints returns a copy of all generated endpoint descriptors.
func Endpoints() []Endpoint {
	return append([]Endpoint(nil), generatedEndpointCatalog[:]...)
}

// EndpointsFor returns a copy of the descriptors belonging to service.
func EndpointsFor(service string) []Endpoint {
	var result []Endpoint
	for _, endpoint := range generatedEndpointCatalog {
		if strings.EqualFold(endpoint.Service, service) {
			result = append(result, endpoint)
		}
	}
	return result
}

// FindEndpoint returns the matching service, method, and path descriptor.
func FindEndpoint(service, method, endpointPath string) (Endpoint, bool) {
	for _, endpoint := range generatedEndpointCatalog {
		if strings.EqualFold(endpoint.Service, service) && strings.EqualFold(endpoint.Method, method) && endpoint.Path == endpointPath {
			return endpoint, true
		}
	}
	return Endpoint{}, false
}
