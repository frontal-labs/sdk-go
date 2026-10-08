package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.frontal.dev/v1"

// Config controls authentication, request deadlines, safe retries, and response limits.
type Config struct {
	APIKey          string
	BaseURL         string
	HTTPClient      *http.Client
	MaxRetries      int
	MaxResponseSize int64
	Headers         http.Header
}

// Endpoint describes one route from contracts/sdk-endpoints.json.
type Endpoint struct {
	Service string
	Method  string
	Path    string
}

// Request carries route parameters, query values, and either a JSON or raw body.
type Request struct {
	PathParams  []string
	Query       url.Values
	JSON        any
	Body        io.Reader
	ContentType string
	Headers     http.Header
}

// Response contains the bounded response body and response metadata.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// APIError contains the structured error details returned by Frontal.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("frontal API %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("frontal API returned HTTP %d", e.StatusCode)
}

// Client owns the shared HTTP transport.
type Client struct {
	apiKey          string
	baseURL         *url.URL
	httpClient      *http.Client
	streamClient    *http.Client
	maxRetries      int
	maxResponseSize int64
	headers         http.Header
}

// NewClient creates a production HTTP client from validated settings.
func NewClient(config Config) (*Client, error) {
	if strings.TrimSpace(config.APIKey) == "" {
		return nil, errors.New("frontal API key must not be empty")
	}
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return nil, fmt.Errorf("invalid Frontal API base URL %q", baseURL)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("frontal API base URL cannot include a query or fragment")
	}
	if config.MaxRetries < 0 || config.MaxRetries > 10 {
		return nil, errors.New("MaxRetries must be between 0 and 10")
	}
	if config.MaxResponseSize == 0 {
		config.MaxResponseSize = 64 << 20
	}
	if config.MaxResponseSize < 1 {
		return nil, errors.New("MaxResponseSize must be positive")
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	streamClient := *httpClient
	streamClient.Timeout = 0
	return &Client{
		apiKey:          config.APIKey,
		baseURL:         parsed,
		httpClient:      httpClient,
		streamClient:    &streamClient,
		maxRetries:      config.MaxRetries,
		maxResponseSize: config.MaxResponseSize,
		headers:         config.Headers.Clone(),
	}, nil
}

// NewClientFromEnvironment reads FRONTAL_API_KEY and optional FRONTAL_API_URL.
func NewClientFromEnvironment() (*Client, error) {
	return NewClient(Config{APIKey: strings.TrimSpace(os.Getenv("FRONTAL_API_KEY")), BaseURL: os.Getenv("FRONTAL_API_URL"), MaxRetries: 2})
}

// Do sends a non-streaming request and returns a bounded response body.
func (c *Client) Do(ctx context.Context, endpoint Endpoint, request Request) (*Response, error) {
	if endpoint.Method == "STREAM" {
		return nil, errors.New("STREAM endpoint requires Stream")
	}
	uri, err := c.endpointURL(endpoint, request.PathParams, request.Query)
	if err != nil {
		return nil, err
	}
	body, contentType, err := encodeRequestBody(endpoint.Method, request)
	if err != nil {
		return nil, err
	}
	attempts := 1
	if endpoint.Method == "GET" || endpoint.Method == "GETRAW" {
		attempts += c.maxRetries
	}
	for attempt := 0; attempt < attempts; attempt++ {
		httpRequest, requestErr := c.newRequest(ctx, endpoint.Method, uri, body, contentType, request.Headers)
		if requestErr != nil {
			return nil, requestErr
		}
		response, sendErr := c.httpClient.Do(httpRequest)
		if sendErr != nil {
			if attempt+1 == attempts {
				return nil, fmt.Errorf("frontal request failed: %w", sendErr)
			}
			if err := waitRetry(ctx, attempt, ""); err != nil {
				return nil, err
			}
			continue
		}
		responseBody, readErr := readBounded(response.Body, c.maxResponseSize)
		if readErr != nil {
			_ = response.Body.Close()
			return nil, readErr
		}
		if closeErr := response.Body.Close(); closeErr != nil {
			return nil, fmt.Errorf("close frontal response body: %w", closeErr)
		}
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			return &Response{StatusCode: response.StatusCode, Header: response.Header.Clone(), Body: responseBody}, nil
		}
		if (endpoint.Method == "GET" || endpoint.Method == "GETRAW") && attempt+1 < attempts && (response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500) {
			if err := waitRetry(ctx, attempt, response.Header.Get("Retry-After")); err != nil {
				return nil, err
			}
			continue
		}
		return nil, apiError(response, responseBody)
	}
	return nil, errors.New("frontal request exhausted retries")
}

// Stream opens an authenticated event stream; the caller must close the returned body.
func (c *Client) Stream(ctx context.Context, endpoint Endpoint, request Request) (*http.Response, error) {
	if endpoint.Method != "STREAM" {
		return nil, errors.New("Stream requires a STREAM endpoint")
	}
	uri, err := c.endpointURL(endpoint, request.PathParams, request.Query)
	if err != nil {
		return nil, err
	}
	httpRequest, err := c.newRequest(ctx, "GET", uri, nil, "", request.Headers)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Accept", "text/event-stream")
	response, err := c.streamClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("frontal stream request failed: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, readErr := readBounded(response.Body, c.maxResponseSize)
		if readErr != nil {
			_ = response.Body.Close()
			return nil, readErr
		}
		if closeErr := response.Body.Close(); closeErr != nil {
			return nil, fmt.Errorf("close frontal error response body: %w", closeErr)
		}
		return nil, apiError(response, body)
	}
	return response, nil
}

func (c *Client) endpointURL(endpoint Endpoint, pathParams []string, query url.Values) (*url.URL, error) {
	if endpoint.Service == "" || endpoint.Method == "" || !strings.HasPrefix(endpoint.Path, "/") {
		return nil, errors.New("invalid endpoint")
	}
	path := endpoint.Path
	count := strings.Count(path, "{param}")
	if count != len(pathParams) {
		return nil, fmt.Errorf("endpoint requires %d path parameters, got %d", count, len(pathParams))
	}
	for _, param := range pathParams {
		if param == "" {
			return nil, errors.New("path parameters must not be empty")
		}
		path = strings.Replace(path, "{param}", url.PathEscape(param), 1)
	}
	uri := *c.baseURL
	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path parameter encoding: %w", err)
	}
	uri.Path = strings.TrimRight(c.baseURL.Path, "/") + decodedPath
	uri.RawPath = strings.TrimRight(c.baseURL.EscapedPath(), "/") + path
	if len(query) > 0 {
		uri.RawQuery = query.Encode()
	}
	return &uri, nil
}

func (c *Client) newRequest(ctx context.Context, method string, uri *url.URL, body io.Reader, contentType string, headers http.Header) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, uri.String(), body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Accept", "application/json, application/octet-stream")
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	for key, values := range c.headers {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	for key, values := range headers {
		request.Header.Del(key)
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	return request, nil
}

func encodeRequestBody(method string, request Request) (io.Reader, string, error) {
	if request.Body != nil && request.JSON != nil {
		return nil, "", errors.New("provide either a raw Body or JSON request value, not both")
	}
	if method == "GET" || method == "GETRAW" || method == "DELETE" {
		if request.Body != nil || request.JSON != nil {
			return nil, "", fmt.Errorf("%s endpoint does not accept a body", method)
		}
		return nil, "", nil
	}
	if request.Body != nil {
		contentType := request.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		return request.Body, contentType, nil
	}
	if request.JSON == nil {
		return nil, "", nil
	}
	if method == "POSTRAW" || method == "POSTFORMDATA" {
		return nil, "", fmt.Errorf("%s requires a raw request Body", method)
	}
	body, err := json.Marshal(request.JSON)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(body), "application/json", nil
}

func readBounded(body io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read frontal response: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("frontal response exceeds %d byte limit", limit)
	}
	return data, nil
}

func apiError(response *http.Response, body []byte) *APIError {
	errorResponse := &APIError{StatusCode: response.StatusCode, RequestID: response.Header.Get("X-Request-ID"), Body: string(body)}
	var envelope struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &envelope) == nil {
		errorResponse.Code = envelope.Code
		errorResponse.Message = envelope.Message
		if envelope.Error != nil {
			errorResponse.Code = envelope.Error.Code
			errorResponse.Message = envelope.Error.Message
		}
	}
	return errorResponse
}

func waitRetry(ctx context.Context, attempt int, retryAfter string) error {
	delay := time.Duration(100*(1<<min(attempt, 5))) * time.Millisecond
	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && seconds >= 0 {
		delay = min(time.Duration(seconds)*time.Second, 5*time.Second)
	}
	timer := time.NewTimer(min(delay, 5*time.Second))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
