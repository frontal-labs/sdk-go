// Package frontal provides one context-first client for the Frontal API.
package frontal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/frontal-labs/sdk-go/pkg/resources"
)

// Option configures a Frontal client.
type Option func(*config) error

type config struct {
	apiKey      string
	baseURL     string
	timeout     time.Duration
	maxRetries  int
	retryDelay  time.Duration
	userAgent   string
	environment string
	debug       bool
	httpClient  *http.Client
	headers     http.Header
}

// Client is a shared HTTP client with a field for every service in the
// committed SDK endpoint inventory. It is safe for concurrent use.
type Client struct {
	AI            *Service
	Agents        *Service
	Audit         *Service
	Auth          *Service
	Billing       *Service
	Blob          *Service
	Connectors    *Service
	Data          *Service
	Governance    *Service
	Lineage       *Service
	Observability *Service
	Ontology      *Service
	Pipelines     *Service
	React         *Service
	Sandbox       *Service
	Schedules     *Service
	Webhooks      *Service
	Workflows     *Service

	// Core exposes the shared low-level request client for APIs that are not in
	// the endpoint inventory or need direct HTTP control.
	Core *resources.Client
}

// New creates a client from options and FRONTAL_* environment variables.
// FRONTAL_API_KEY is required unless WithAPIKey is supplied.
func New(options ...Option) (*Client, error) {
	configuration, err := configFromEnvironment()
	if err != nil {
		return nil, err
	}
	for _, option := range options {
		if option == nil {
			return nil, errors.New("frontal: client option is nil")
		}
		if err := option(&configuration); err != nil {
			return nil, err
		}
	}
	resourceOptions := []resources.Option{
		resources.WithBaseURL(configuration.baseURL),
		resources.WithTimeout(configuration.timeout),
		resources.WithMaxRetries(configuration.maxRetries),
		resources.WithRetryDelay(configuration.retryDelay),
		resources.WithUserAgent(configuration.userAgent),
		resources.WithEnvironment(configuration.environment),
		resources.WithDebug(configuration.debug),
		resources.WithHeaders(configuration.headers),
	}
	if configuration.httpClient != nil {
		resourceOptions = append(resourceOptions, resources.WithHTTPClient(configuration.httpClient))
	}
	core, err := resources.NewClient(configuration.apiKey, resourceOptions...)
	if err != nil {
		return nil, err
	}
	client := &Client{Core: core}
	client.AI = newService(core, "ai")
	client.Agents = newService(core, "agents")
	client.Audit = newService(core, "audit")
	client.Auth = newService(core, "auth")
	client.Billing = newService(core, "billing")
	client.Blob = newService(core, "blob")
	client.Connectors = newService(core, "connectors")
	client.Data = newService(core, "data")
	client.Governance = newService(core, "governance")
	client.Lineage = newService(core, "lineage")
	client.Observability = newService(core, "observability")
	client.Ontology = newService(core, "ontology")
	client.Pipelines = newService(core, "pipelines")
	client.React = newService(core, "react")
	client.Sandbox = newService(core, "sandbox")
	client.Schedules = newService(core, "schedules")
	client.Webhooks = newService(core, "webhooks")
	client.Workflows = newService(core, "workflows")
	return client, nil
}

func configFromEnvironment() (config, error) {
	settings := config{
		apiKey:      os.Getenv("FRONTAL_API_KEY"),
		baseURL:     valueOr(os.Getenv("FRONTAL_API_URL"), resources.DefaultBaseURL),
		timeout:     30 * time.Second,
		maxRetries:  3,
		retryDelay:  time.Second,
		userAgent:   "frontal-go/1.0.0",
		environment: valueOr(os.Getenv("FRONTAL_ENV"), "development"),
		headers:     make(http.Header),
	}
	if err := validateEnvironment(settings.environment); err != nil {
		return config{}, err
	}
	if value := strings.TrimSpace(os.Getenv("FRONTAL_TIMEOUT")); value != "" {
		timeout, err := parseTimeout(value)
		if err != nil {
			return config{}, err
		}
		settings.timeout = timeout
	}
	if value := strings.TrimSpace(os.Getenv("FRONTAL_DEBUG")); value != "" {
		switch value {
		case "true", "1":
			settings.debug = true
		case "false", "0":
			settings.debug = false
		default:
			return config{}, fmt.Errorf("frontal: invalid FRONTAL_DEBUG value %q; use true, false, 1, or 0", value)
		}
	}
	return settings, nil
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func validateEnvironment(environment string) error {
	switch environment {
	case "development", "test", "production":
		return nil
	default:
		return fmt.Errorf("frontal: invalid environment %q; use development, test, or production", environment)
	}
}

func parseTimeout(value string) (time.Duration, error) {
	if timeout, err := time.ParseDuration(value); err == nil {
		if timeout > 0 {
			return timeout, nil
		}
		return 0, errors.New("frontal: timeout must be positive")
	}
	milliseconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || milliseconds <= 0 {
		return 0, fmt.Errorf("frontal: invalid FRONTAL_TIMEOUT %q; use a duration or integer milliseconds", value)
	}
	if milliseconds > int64((1<<63-1)/int64(time.Millisecond)) {
		return 0, errors.New("frontal: timeout is too large")
	}
	return time.Duration(milliseconds) * time.Millisecond, nil
}

// WithAPIKey sets the Frontal API key.
func WithAPIKey(apiKey string) Option {
	return func(configuration *config) error {
		configuration.apiKey = apiKey
		return nil
	}
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(baseURL string) Option {
	return func(configuration *config) error {
		configuration.baseURL = baseURL
		return nil
	}
}

// WithHTTPClient uses a caller-provided net/http client.
func WithHTTPClient(client *http.Client) Option {
	return func(configuration *config) error {
		if client == nil {
			return errors.New("frontal: HTTP client cannot be nil")
		}
		configuration.httpClient = client
		return nil
	}
}

// WithTimeout sets the timeout on the default HTTP client.
func WithTimeout(timeout time.Duration) Option {
	return func(configuration *config) error {
		if timeout <= 0 {
			return errors.New("frontal: timeout must be positive")
		}
		configuration.timeout = timeout
		return nil
	}
}

// WithMaxRetries sets the extra attempts for safe GET and HEAD requests.
func WithMaxRetries(retries int) Option {
	return func(configuration *config) error {
		if retries < 0 || retries > 10 {
			return errors.New("frontal: max retries must be between 0 and 10")
		}
		configuration.maxRetries = retries
		return nil
	}
}

// WithRetryDelay sets the initial exponential backoff delay.
func WithRetryDelay(delay time.Duration) Option {
	return func(configuration *config) error {
		if delay < 0 || delay > 30*time.Second {
			return errors.New("frontal: retry delay must be between 0 and 30s")
		}
		configuration.retryDelay = delay
		return nil
	}
}

// WithEnvironment sets the environment sent in X-Frontal-Environment.
func WithEnvironment(environment string) Option {
	return func(configuration *config) error {
		environment = strings.TrimSpace(environment)
		if environment == "" || strings.ContainsAny(environment, "\r\n") {
			return errors.New("frontal: environment must be non-empty and contain no newlines")
		}
		if err := validateEnvironment(environment); err != nil {
			return err
		}
		configuration.environment = environment
		return nil
	}
}

// WithDebug enables request metadata logging. Request bodies and credentials are not logged.
func WithDebug(enabled bool) Option {
	return func(configuration *config) error {
		configuration.debug = enabled
		return nil
	}
}

// WithHeaders adds headers to every request.
func WithHeaders(headers http.Header) Option {
	return func(configuration *config) error {
		for name, values := range headers {
			if strings.EqualFold(name, "Authorization") || strings.EqualFold(name, "Host") {
				return fmt.Errorf("frontal: header %q is managed by the client", name)
			}
			for _, value := range values {
				if strings.ContainsAny(value, "\r\n") {
					return fmt.Errorf("frontal: header %q contains a newline", name)
				}
				configuration.headers.Add(name, value)
			}
		}
		return nil
	}
}

// BaseURL returns the configured API origin and path.
func (client *Client) BaseURL() string {
	if client == nil || client.Core == nil {
		return ""
	}
	return client.Core.BaseURL()
}

// Request sends a JSON request to a path. Use Call for an inventory-backed operation.
func (client *Client) Request(ctx context.Context, method, path string, body, out any) error {
	if client == nil || client.Core == nil {
		return errors.New("frontal: client is not configured")
	}
	return client.Core.Request(ctx, method, path, body, out)
}

// NewRequest creates a request for a relative path, with ctx attached. The
// client adds authentication and default headers when Do sends it.
func (client *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if client == nil || client.Core == nil {
		return nil, errors.New("frontal: client is not configured")
	}
	return client.Core.NewRequest(ctx, method, path, body)
}

// Do sends a request through this client's configured origin and transport.
// The caller owns and must close the response body.
func (client *Client) Do(ctx context.Context, request *http.Request) (*http.Response, error) {
	if client == nil || client.Core == nil {
		return nil, errors.New("frontal: client is not configured")
	}
	if ctx == nil {
		return nil, errors.New("frontal: request context is required")
	}
	if request == nil {
		return nil, errors.New("frontal: request is nil")
	}
	return client.Core.Do(request.WithContext(ctx))
}

// Call dispatches an inventory-backed operation through the matching service.
func (client *Client) Call(ctx context.Context, request resources.Request, out any) error {
	if client == nil || client.Core == nil {
		return errors.New("frontal: client is not configured")
	}
	service := client.service(request.Endpoint.Service)
	if service == nil {
		return fmt.Errorf("frontal: unknown service %q", request.Endpoint.Service)
	}
	return service.Call(ctx, request, out)
}

// CloseIdleConnections closes the client's idle connections.
func (client *Client) CloseIdleConnections() {
	if client != nil && client.Core != nil {
		client.Core.CloseIdleConnections()
	}
}

func (client *Client) service(name string) *Service {
	for _, service := range []*Service{
		client.AI, client.Agents, client.Audit, client.Auth, client.Billing, client.Blob,
		client.Connectors, client.Data, client.Governance, client.Lineage, client.Observability,
		client.Ontology, client.Pipelines, client.React, client.Sandbox, client.Schedules,
		client.Webhooks, client.Workflows,
	} {
		if service != nil && strings.EqualFold(service.name, name) {
			return service
		}
	}
	return nil
}
