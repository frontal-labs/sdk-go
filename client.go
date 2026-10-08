// Package frontal provides the Frontal API client.
package frontal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/frontal-labs/sdk-go/authentication"
	"github.com/frontal-labs/sdk-go/handlers"
	"github.com/frontal-labs/sdk-go/headers"
	"github.com/frontal-labs/sdk-go/internal/core"
	"github.com/frontal-labs/sdk-go/models"
	"github.com/frontal-labs/sdk-go/utils"
)

const (
	// DefaultBaseURL is the default Frontal API URL.
	DefaultBaseURL = utils.DefaultBaseURL
	// DefaultTimeout is applied when the caller does not provide an HTTP client or timeout.
	DefaultTimeout  = 30 * time.Second
	defaultUserAgent = "frontal-go/0.1.0"
	defaultMaxRetries = 2
	maxRetriesLimit   = 8
)

var ErrForeignRequest = errors.New("frontal: request URL must use the configured API origin")

// Client sends authenticated requests to the Frontal API. It is safe for concurrent use.
type Client struct {
	baseURL    *url.URL
	apiKey     authentication.APIKey
	transport  *core.Transport
	userAgent  string
	maxRetries int
}

type clientConfig struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	userAgent  string
	maxRetries int
}

// Option configures a Client.
type Option func(*clientConfig) error

// NewClient creates an API client using apiKey and the default API URL.
func NewClient(apiKey string, options ...Option) (*Client, error) {
	key, err := authentication.NewAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("frontal: configure authentication: %w", err)
	}
	config := clientConfig{
		baseURL:    DefaultBaseURL,
		timeout:    DefaultTimeout,
		userAgent:  defaultUserAgent,
		maxRetries: defaultMaxRetries,
	}
	for _, option := range options {
		if option == nil {
			return nil, errors.New("frontal: client option is nil")
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}
	baseURL, err := utils.ParseBaseURL(config.baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL:    baseURL,
		apiKey:     key,
		transport:  core.NewTransport(config.httpClient, config.timeout),
		userAgent:  config.userAgent,
		maxRetries: config.maxRetries,
	}, nil
}

// NewClientFromEnv creates a client from FRONTAL_API_KEY and optional URL and timeout settings.
// FRONTAL_TIMEOUT accepts a Go duration or an integer number of milliseconds.
func NewClientFromEnv(options ...Option) (*Client, error) {
	apiKey := os.Getenv("FRONTAL_API_KEY")
	envOptions := make([]Option, 0, 2+len(options))
	if baseURL := strings.TrimSpace(os.Getenv("FRONTAL_API_URL")); baseURL != "" {
		envOptions = append(envOptions, WithBaseURL(baseURL))
	}
	if timeout := strings.TrimSpace(os.Getenv("FRONTAL_TIMEOUT")); timeout != "" {
		duration, err := utils.ParseTimeout(timeout)
		if err != nil {
			return nil, err
		}
		envOptions = append(envOptions, WithTimeout(duration))
	}
	envOptions = append(envOptions, options...)
	return NewClient(apiKey, envOptions...)
}

// NewFromEnvironment is an alias for NewClientFromEnv.
func NewFromEnvironment(options ...Option) (*Client, error) {
	return NewClientFromEnv(options...)
}

// WithBaseURL overrides the API base URL. The URL must use HTTP or HTTPS and must not include credentials, a query, or a fragment.
func WithBaseURL(baseURL string) Option {
	return func(config *clientConfig) error {
		config.baseURL = baseURL
		return nil
	}
}

// WithHTTPClient uses a caller-provided HTTP client. Its timeout and redirect policy remain under caller control.
func WithHTTPClient(client *http.Client) Option {
	return func(config *clientConfig) error {
		if client == nil {
			return errors.New("frontal: HTTP client cannot be nil")
		}
		config.httpClient = client
		return nil
	}
}

// WithTimeout sets the timeout used by the default HTTP client.
func WithTimeout(timeout time.Duration) Option {
	return func(config *clientConfig) error {
		if timeout <= 0 {
			return errors.New("frontal: timeout must be positive")
		}
		config.timeout = timeout
		return nil
	}
}

// WithUserAgent sets the User-Agent header sent by the client.
func WithUserAgent(userAgent string) Option {
	return func(config *clientConfig) error {
		userAgent = strings.TrimSpace(userAgent)
		if userAgent == "" || strings.ContainsAny(userAgent, "\r\n") {
			return errors.New("frontal: user agent must be non-empty and contain no newlines")
		}
		config.userAgent = userAgent
		return nil
	}
}

// WithMaxRetries sets the number of additional attempts for safe GET and HEAD requests.
func WithMaxRetries(retries int) Option {
	return func(config *clientConfig) error {
		if retries < 0 || retries > maxRetriesLimit {
			return fmt.Errorf("frontal: max retries must be between 0 and %d", maxRetriesLimit)
		}
		config.maxRetries = retries
		return nil
	}
}

// BaseURL returns the configured API base URL.
func (client *Client) BaseURL() string {
	if client == nil || client.baseURL == nil {
		return ""
	}
	return client.baseURL.String()
}

// NewRequest creates a request for a relative API path. The API key is added when Do sends it.
func (client *Client) NewRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	if client == nil || client.baseURL == nil {
		return nil, errors.New("frontal: client is not configured")
	}
	target, err := utils.ResolveEndpoint(client.baseURL, endpoint)
	if err != nil {
		return nil, err
	}
	request, err := handlers.NewRequest(ctx, method, target.String(), body)
	if err != nil {
		return nil, err
	}
	request.Header = headers.ApplyDefaults(request.Header, client.userAgent)
	return request, nil
}

// Do sends a request created for this client's API origin. The caller owns and must close the response body.
func (client *Client) Do(request *http.Request) (*http.Response, error) {
	if client == nil || client.baseURL == nil || client.transport == nil {
		return nil, errors.New("frontal: client is not configured")
	}
	if request == nil || request.URL == nil {
		return nil, errors.New("frontal: request and request URL are required")
	}
	if request.URL.Fragment != "" || !utils.SameOrigin(client.baseURL, request.URL) {
		return nil, ErrForeignRequest
	}
	request = request.Clone(request.Context())
	request.Header = headers.ApplyDefaults(request.Header.Clone(), client.userAgent)
	request.Host = ""
	if err := client.apiKey.Apply(request); err != nil {
		return nil, err
	}
	response, err := client.transport.Do(request)
	if err != nil {
		return nil, fmt.Errorf("frontal: %s request failed: %w", request.Method, err)
	}
	if response == nil {
		return nil, errors.New("frontal: HTTP transport returned a nil response")
	}
	return response, nil
}

// Request sends a JSON request to endpoint and decodes a successful response into out.
// GET and HEAD requests are retried on transient failures. If out implements io.Writer,
// the response body is copied to it without JSON decoding.
func (client *Client) Request(ctx context.Context, method, endpoint string, body, out any) error {
	return client.request(ctx, method, endpoint, nil, nil, nil, body, out)
}

// Call sends a request described by the generated endpoint inventory.
func (client *Client) Call(ctx context.Context, request models.Request, out any) error {
	method := strings.ToUpper(strings.TrimSpace(request.Endpoint.Method))
	if method == "STREAM" {
		return errors.New("frontal: use Stream or StreamEvents for streaming endpoints")
	}
	if request.Endpoint.Path == "" || method == "" {
		return errors.New("frontal: request endpoint method and path are required")
	}
	return client.request(ctx, method, request.Endpoint.Path, request.PathParams, request.Query, request.Headers, request.Body, out)
}

func (client *Client) request(ctx context.Context, method, endpoint string, pathParams []string, query url.Values, customHeaders http.Header, body, out any) error {
	if client == nil || client.baseURL == nil || client.transport == nil {
		return errors.New("frontal: client is not configured")
	}
	expandedEndpoint, err := utils.ExpandPath(endpoint, pathParams)
	if err != nil {
		return err
	}
	target, err := utils.ResolveEndpoint(client.baseURL, expandedEndpoint)
	if err != nil {
		return err
	}
	mergeQuery(target, query)

	buildRequest := func() (*http.Request, error) {
		request, err := handlers.NewJSONRequest(ctx, method, target.String(), body)
		if err != nil {
			return nil, err
		}
		applyCustomHeaders(request, customHeaders)
		request.Header = headers.ApplyDefaults(request.Header, client.userAgent)
		return request, nil
	}
	for attempt := 0; ; attempt++ {
		request, err := buildRequest()
		if err != nil {
			return err
		}
		response, err := client.Do(request)
		if err != nil {
			if attempt >= client.maxRetries || !isRetryableMethod(method) || !isRetryableNetworkError(err) {
				return err
			}
			if err := waitForRetry(ctx, retryDelay(nil, attempt)); err != nil {
				return err
			}
			continue
		}
		if attempt >= client.maxRetries || !isRetryableMethod(method) || !isRetryableStatus(response.StatusCode) {
			return handlers.HandleResponse(response, out)
		}
		delay := retryDelay(response, attempt)
		drainAndClose(response.Body)
		if err := waitForRetry(ctx, delay); err != nil {
			return err
		}
	}
}

// Stream sends a request for a streaming endpoint and returns the open response body.
// The caller must close the returned response body.
func (client *Client) Stream(ctx context.Context, request models.Request) (*http.Response, error) {
	if client == nil || client.baseURL == nil || client.transport == nil {
		return nil, errors.New("frontal: client is not configured")
	}
	method := strings.ToUpper(strings.TrimSpace(request.Endpoint.Method))
	if method == "STREAM" {
		method = http.MethodGet
	}
	if method != http.MethodGet || request.Endpoint.Path == "" {
		return nil, errors.New("frontal: streaming requires a GET or STREAM endpoint with a path")
	}
	endpoint, err := utils.ExpandPath(request.Endpoint.Path, request.PathParams)
	if err != nil {
		return nil, err
	}
	target, err := utils.ResolveEndpoint(client.baseURL, endpoint)
	if err != nil {
		return nil, err
	}
	mergeQuery(target, request.Query)
	httpRequest, err := handlers.NewJSONRequest(ctx, method, target.String(), request.Body)
	if err != nil {
		return nil, err
	}
	applyCustomHeaders(httpRequest, request.Headers)
	httpRequest.Header = headers.ApplyDefaults(httpRequest.Header, client.userAgent)
	if request.Headers.Get(headers.Accept) == "" {
		httpRequest.Header.Set(headers.Accept, "text/event-stream")
	}
	response, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, handlers.HandleResponse(response, nil)
	}
	return response, nil
}

// StreamEvents decodes a JSON Server-Sent Event stream from a generated endpoint.
func StreamEvents[T any](ctx context.Context, client *Client, request models.Request, handle func(models.Event[T]) error) error {
	response, err := client.Stream(ctx, request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return handlers.DecodeEventStream(ctx, response.Body, handle)
}

// DoJSON sends a contract-backed request and returns its decoded JSON response.
func DoJSON[T any](ctx context.Context, client *Client, request models.Request) (T, error) {
	var result T
	if client == nil {
		return result, errors.New("frontal: client is nil")
	}
	if err := client.Call(ctx, request, &result); err != nil {
		return result, err
	}
	return result, nil
}

// Get sends a GET request and decodes its response.
func (client *Client) Get(ctx context.Context, endpoint string, out any) error {
	return client.Request(ctx, http.MethodGet, endpoint, nil, out)
}

// Post sends a JSON POST request and decodes its response.
func (client *Client) Post(ctx context.Context, endpoint string, body, out any) error {
	return client.Request(ctx, http.MethodPost, endpoint, body, out)
}

// Put sends a JSON PUT request and decodes its response.
func (client *Client) Put(ctx context.Context, endpoint string, body, out any) error {
	return client.Request(ctx, http.MethodPut, endpoint, body, out)
}

// Patch sends a JSON PATCH request and decodes its response.
func (client *Client) Patch(ctx context.Context, endpoint string, body, out any) error {
	return client.Request(ctx, http.MethodPatch, endpoint, body, out)
}

// Delete sends a DELETE request and decodes its response.
func (client *Client) Delete(ctx context.Context, endpoint string, out any) error {
	return client.Request(ctx, http.MethodDelete, endpoint, nil, out)
}

// CloseIdleConnections closes idle network connections managed by the client.
func (client *Client) CloseIdleConnections() {
	if client != nil && client.transport != nil {
		client.transport.CloseIdleConnections()
	}
}

func mergeQuery(target *url.URL, query url.Values) {
	if target == nil || len(query) == 0 {
		return
	}
	values := target.Query()
	for key, entries := range query {
		for _, value := range entries {
			values.Add(key, value)
		}
	}
	target.RawQuery = values.Encode()
}

func applyCustomHeaders(request *http.Request, custom http.Header) {
	for name, values := range custom {
		request.Header.Del(name)
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}
}

func isRetryableMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusInternalServerError:
		return true
	default:
		return false
	}
}

func isRetryableNetworkError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var networkError net.Error
	return errors.As(err, &networkError) && (networkError.Timeout() || networkError.Temporary())
}

func retryDelay(response *http.Response, attempt int) time.Duration {
	if response != nil {
		if delay, ok := utils.RetryAfter(response.Header.Get("Retry-After"), time.Now()); ok {
			if delay > 30*time.Second {
				return 30 * time.Second
			}
			return delay
		}
	}
	delay := 100 * time.Millisecond * time.Duration(1<<attempt)
	if delay > 2*time.Second {
		return 2 * time.Second
	}
	return delay
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 64<<10))
	_ = body.Close()
}
