// Package utils contains URL and configuration helpers shared by the SDK.
package utils

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.frontal.dev/v1"

var ErrInvalidBaseURL = errors.New("frontal: base URL must be an absolute HTTP or HTTPS URL without credentials, query, or fragment")

// ParseBaseURL validates a base URL and removes trailing slashes from its path.
func ParseBaseURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil || parsed == nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Opaque != "" {
		return nil, ErrInvalidBaseURL
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return nil, ErrInvalidBaseURL
	}
	if strings.ContainsAny(parsed.Host, "\r\n\t ") {
		return nil, ErrInvalidBaseURL
	}
	for _, segment := range strings.Split(parsed.Path, "/") {
		if segment == "." || segment == ".." {
			return nil, ErrInvalidBaseURL
		}
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = ""
	return parsed, nil
}

// ExpandPath substitutes positional path parameters in a contract path.
// Path parameters are escaped as individual URL path segments.
func ExpandPath(endpoint string, parameters []string) (string, error) {
	reference, err := url.Parse(endpoint)
	if err != nil || reference.IsAbs() || reference.Host != "" || reference.User != nil || reference.Fragment != "" || reference.Opaque != "" {
		return "", errors.New("frontal: endpoint must be a relative URL path")
	}
	escapedPath := strings.Split(reference.EscapedPath(), "/")
	parameterIndex := 0
	for index, escapedSegment := range escapedPath {
		segment, err := url.PathUnescape(escapedSegment)
		if err != nil {
			return "", errors.New("frontal: endpoint contains an invalid escape")
		}
		if len(segment) < 2 || segment[0] != '{' || segment[len(segment)-1] != '}' {
			continue
		}
		if parameterIndex >= len(parameters) {
			return "", fmt.Errorf("frontal: endpoint requires a value for path parameter %s", segment)
		}
		parameter := parameters[parameterIndex]
		if parameter == "" {
			return "", fmt.Errorf("frontal: path parameter %s cannot be empty", segment)
		}
		escapedPath[index] = url.PathEscape(parameter)
		parameterIndex++
	}
	if parameterIndex != len(parameters) {
		return "", fmt.Errorf("frontal: endpoint has %d path parameters but %d values were supplied", parameterIndex, len(parameters))
	}
	joinedPath := strings.Join(escapedPath, "/")
	decodedPath, err := url.PathUnescape(joinedPath)
	if err != nil {
		return "", errors.New("frontal: endpoint contains an invalid path parameter")
	}
	reference.Path = decodedPath
	reference.RawPath = joinedPath
	return reference.String(), nil
}

// ResolveEndpoint joins a relative API endpoint to a base URL while preserving its base path.
// It accepts both inventory paths such as "/agents" and OpenAPI paths such as "/v1/agents".
func ResolveEndpoint(base *url.URL, endpoint string) (*url.URL, error) {
	if base == nil {
		return nil, ErrInvalidBaseURL
	}
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil, errors.New("frontal: endpoint path is required")
	}
	reference, err := url.Parse(endpoint)
	if err != nil || reference.IsAbs() || reference.Host != "" || reference.User != nil || reference.Opaque != "" || reference.Fragment != "" {
		return nil, errors.New("frontal: endpoint must be a relative URL path")
	}
	endpointPath := reference.Path
	for _, segment := range strings.Split(endpointPath, "/") {
		if segment == "." || segment == ".." {
			return nil, errors.New("frontal: endpoint path cannot contain dot segments")
		}
	}
	basePath := strings.TrimRight(base.EscapedPath(), "/")
	endpointEscapedPath := strings.TrimLeft(reference.EscapedPath(), "/")
	if strings.HasSuffix(base.Path, "/v1") && (endpointPath == "v1" || strings.HasPrefix(endpointPath, "v1/")) {
		endpointPath = strings.TrimPrefix(endpointPath, "v1")
		endpointPath = strings.TrimLeft(endpointPath, "/")
		endpointEscapedPath = strings.TrimPrefix(endpointEscapedPath, "v1")
		endpointEscapedPath = strings.TrimLeft(endpointEscapedPath, "/")
	}
	joinedEscapedPath := path.Join(basePath, endpointEscapedPath)
	if joinedEscapedPath == "." || joinedEscapedPath == "" {
		joinedEscapedPath = "/"
	}
	if !strings.HasPrefix(joinedEscapedPath, "/") {
		joinedEscapedPath = "/" + joinedEscapedPath
	}
	if strings.HasSuffix(endpointPath, "/") && !strings.HasSuffix(joinedEscapedPath, "/") {
		joinedEscapedPath += "/"
	}
	joinedPath, err := url.PathUnescape(joinedEscapedPath)
	if err != nil {
		return nil, errors.New("frontal: endpoint contains an invalid escape")
	}
	resolved := *base
	resolved.Path = joinedPath
	resolved.RawPath = joinedEscapedPath
	resolved.RawQuery = reference.RawQuery
	resolved.ForceQuery = reference.ForceQuery
	resolved.Fragment = ""
	return &resolved, nil
}

// SameOrigin reports whether target uses the same scheme and authority as base.
func SameOrigin(base, target *url.URL) bool {
	return base != nil && target != nil && target.User == nil &&
		strings.EqualFold(base.Scheme, target.Scheme) && strings.EqualFold(base.Host, target.Host)
}

// ParseTimeout accepts Go duration strings or integer milliseconds.
func ParseTimeout(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errors.New("frontal: timeout is required")
	}
	if duration, err := time.ParseDuration(raw); err == nil {
		if duration <= 0 {
			return 0, errors.New("frontal: timeout must be positive")
		}
		return duration, nil
	}
	milliseconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || milliseconds <= 0 {
		return 0, fmt.Errorf("frontal: invalid timeout %q; use a duration or milliseconds", raw)
	}
	const maxMilliseconds = int64((1<<63 - 1) / int64(time.Millisecond))
	if milliseconds > maxMilliseconds {
		return 0, errors.New("frontal: timeout is too large")
	}
	return time.Duration(milliseconds) * time.Millisecond, nil
}
