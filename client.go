// Package frontal provides one context-first client for the Frontal API.
package frontal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/frontal-labs/sdk-go/v2/agents"
	"github.com/frontal-labs/sdk-go/v2/ai"
	"github.com/frontal-labs/sdk-go/v2/audit"
	"github.com/frontal-labs/sdk-go/v2/auth"
	"github.com/frontal-labs/sdk-go/v2/billing"
	"github.com/frontal-labs/sdk-go/v2/blob"
	"github.com/frontal-labs/sdk-go/v2/connectors"
	"github.com/frontal-labs/sdk-go/v2/data"
	"github.com/frontal-labs/sdk-go/v2/governance"
	"github.com/frontal-labs/sdk-go/v2/internal/core"
	"github.com/frontal-labs/sdk-go/v2/lineage"
	"github.com/frontal-labs/sdk-go/v2/observability"
	"github.com/frontal-labs/sdk-go/v2/ontology"
	"github.com/frontal-labs/sdk-go/v2/pipelines"
	"github.com/frontal-labs/sdk-go/v2/sandbox"
	"github.com/frontal-labs/sdk-go/v2/schedules"
	"github.com/frontal-labs/sdk-go/v2/webhooks"
	"github.com/frontal-labs/sdk-go/v2/workflows"
)

// Option configures a Frontal client.
type Option func(*config) error

type config struct {
	apiKey           string
	baseURL          string
	timeout          time.Duration
	maxRetries       int
	retryDelay       time.Duration
	userAgent        string
	environment      string
	debug            bool
	logger           *slog.Logger
	maxResponseBytes int64
	httpClient       *http.Client
	headers          http.Header
}

// Client is a shared HTTP client with a field for every service in the
// committed SDK endpoint inventory. It is safe for concurrent use.
type Client struct {
	AI            *ai.Client
	Agents        *agents.Client
	Audit         *audit.Client
	Auth          *auth.Client
	Billing       *billing.Client
	Blob          *blob.Client
	Connectors    *connectors.Client
	Data          *data.Client
	Governance    *governance.Client
	Lineage       *lineage.Client
	Observability *observability.Client
	Ontology      *ontology.Client
	Pipelines     *pipelines.Client
	Sandbox       *sandbox.Client
	Schedules     *schedules.Client
	Webhooks      *webhooks.Client
	Workflows     *workflows.Client

	core *core.Client
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
	resourceOptions := []core.Option{
		core.WithBaseURL(configuration.baseURL),
		core.WithTimeout(configuration.timeout),
		core.WithMaxRetries(configuration.maxRetries),
		core.WithRetryDelay(configuration.retryDelay),
		core.WithUserAgent(configuration.userAgent),
		core.WithEnvironment(configuration.environment),
		core.WithDebug(configuration.debug),
		core.WithLogger(configuration.logger),
		core.WithMaxResponseBytes(configuration.maxResponseBytes),
		core.WithHeaders(configuration.headers),
	}
	if configuration.httpClient != nil {
		resourceOptions = append(resourceOptions, core.WithHTTPClient(configuration.httpClient))
	}
	core, err := core.NewClient(configuration.apiKey, resourceOptions...)
	if err != nil {
		return nil, err
	}
	client := &Client{core: core}
	client.AI = ai.NewClient(serviceCall(core, "ai"), serviceStream(core, "ai"))
	client.Agents = agents.NewClient(serviceCall(core, "agents"), serviceStream(core, "agents"))
	client.Audit = audit.NewClient(serviceCall(core, "audit"), serviceStream(core, "audit"))
	client.Auth = auth.NewClient(serviceCall(core, "auth"), serviceStream(core, "auth"))
	client.Billing = billing.NewClient(serviceCall(core, "billing"), serviceStream(core, "billing"))
	client.Blob = blob.NewClient(serviceCall(core, "blob"), serviceStream(core, "blob"))
	client.Connectors = connectors.NewClient(serviceCall(core, "connectors"), serviceStream(core, "connectors"))
	client.Data = data.NewClient(serviceCall(core, "data"), serviceStream(core, "data"))
	client.Governance = governance.NewClient(serviceCall(core, "governance"), serviceStream(core, "governance"))
	client.Lineage = lineage.NewClient(serviceCall(core, "lineage"), serviceStream(core, "lineage"))
	client.Observability = observability.NewClient(serviceCall(core, "observability"), serviceStream(core, "observability"))
	client.Ontology = ontology.NewClient(serviceCall(core, "ontology"), serviceStream(core, "ontology"))
	client.Pipelines = pipelines.NewClient(serviceCall(core, "pipelines"), serviceStream(core, "pipelines"))
	client.Sandbox = sandbox.NewClient(serviceCall(core, "sandbox"), serviceStream(core, "sandbox"))
	client.Schedules = schedules.NewClient(serviceCall(core, "schedules"), serviceStream(core, "schedules"))
	client.Webhooks = webhooks.NewClient(serviceCall(core, "webhooks"), serviceStream(core, "webhooks"))
	client.Workflows = workflows.NewClient(serviceCall(core, "workflows"), serviceStream(core, "workflows"))
	return client, nil
}

func configFromEnvironment() (config, error) {
	settings := config{
		apiKey:           os.Getenv("FRONTAL_API_KEY"),
		baseURL:          valueOr(os.Getenv("FRONTAL_API_URL"), core.DefaultBaseURL),
		timeout:          30 * time.Second,
		maxRetries:       3,
		retryDelay:       time.Second,
		userAgent:        "frontal-go/2.0.0",
		environment:      valueOr(os.Getenv("FRONTAL_ENV"), "development"),
		logger:           slog.Default(),
		maxResponseBytes: core.DefaultMaxResponseBytes,
		headers:          make(http.Header),
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

// WithHTTPClient uses a caller-provided net/http client. The SDK preserves its
// transport and redirect policy but blocks redirects outside the API origin.
func WithHTTPClient(client *http.Client) Option {
	return func(configuration *config) error {
		if client == nil {
			return errors.New("frontal: HTTP client cannot be nil")
		}
		configuration.httpClient = client
		return nil
	}
}

// WithLogger sets the structured logger used for debug request metadata.
func WithLogger(logger *slog.Logger) Option {
	return func(configuration *config) error {
		if logger == nil {
			return errors.New("frontal: logger cannot be nil")
		}
		configuration.logger = logger
		return nil
	}
}

// WithMaxResponseBytes sets the maximum JSON response size decoded by the client.
func WithMaxResponseBytes(maxBytes int64) Option {
	return func(configuration *config) error {
		if maxBytes <= 0 || maxBytes == int64(^uint64(0)>>1) {
			return errors.New("frontal: maximum response size must be positive and below the maximum int64 value")
		}
		configuration.maxResponseBytes = maxBytes
		return nil
	}
}

// WithUserAgent sets the User-Agent header sent by the client.
func WithUserAgent(userAgent string) Option {
	return func(configuration *config) error {
		userAgent = strings.TrimSpace(userAgent)
		if userAgent == "" || strings.ContainsAny(userAgent, "\r\n") {
			return errors.New("frontal: user agent must be non-empty and contain no newlines")
		}
		configuration.userAgent = userAgent
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
	if client == nil || client.core == nil {
		return ""
	}
	return client.core.BaseURL()
}

// Request sends a JSON request to a path. Use Call for an inventory-backed operation.
func (client *Client) Request(ctx context.Context, method, path string, body, out any) error {
	if client == nil || client.core == nil {
		return errors.New("frontal: client is not configured")
	}
	return client.core.Request(ctx, method, path, body, out)
}

// NewRequest creates a request for a relative path, with ctx attached. The
// client adds authentication and default headers when Do sends it.
func (client *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if client == nil || client.core == nil {
		return nil, errors.New("frontal: client is not configured")
	}
	return client.core.NewRequest(ctx, method, path, body)
}

// Do sends a request through this client's configured origin and transport.
// The caller owns and must close the response body.
func (client *Client) Do(ctx context.Context, request *http.Request) (*http.Response, error) {
	if client == nil || client.core == nil {
		return nil, errors.New("frontal: client is not configured")
	}
	if ctx == nil {
		return nil, errors.New("frontal: request context is required")
	}
	if request == nil {
		return nil, errors.New("frontal: request is nil")
	}
	return client.core.Do(request.WithContext(ctx))
}

// Call dispatches an inventory-backed operation through the matching service.
func (client *Client) Call(ctx context.Context, request Request, out any) error {
	if client == nil || client.core == nil {
		return errors.New("frontal: client is not configured")
	}
	endpoint, ok := core.FindEndpoint(request.Endpoint.Service, request.Endpoint.Method, request.Endpoint.Path)
	if !ok {
		return fmt.Errorf("%w: %s %s for %s", ErrUnknownEndpoint, request.Endpoint.Method, request.Endpoint.Path, request.Endpoint.Service)
	}
	return client.core.Call(ctx, core.Request{
		Endpoint:   endpoint,
		PathParams: request.PathParams,
		Query:      request.Query,
		Headers:    request.Headers,
		Body:       request.Body,
	}, out)
}

// Stream opens a contract-backed streaming request. The caller owns and must
// close the returned response body.
func (client *Client) Stream(ctx context.Context, request Request) (*http.Response, error) {
	if client == nil || client.core == nil {
		return nil, errors.New("frontal: client is not configured")
	}
	endpoint, ok := core.FindEndpoint(request.Endpoint.Service, request.Endpoint.Method, request.Endpoint.Path)
	if !ok {
		return nil, fmt.Errorf("%w: %s %s for %s", ErrUnknownEndpoint, request.Endpoint.Method, request.Endpoint.Path, request.Endpoint.Service)
	}
	return client.core.Stream(ctx, core.Request{
		Endpoint:   endpoint,
		PathParams: request.PathParams,
		Query:      request.Query,
		Headers:    request.Headers,
		Body:       request.Body,
	})
}

// CloseIdleConnections closes the client's idle connections.
func (client *Client) CloseIdleConnections() {
	if client != nil && client.core != nil {
		client.core.CloseIdleConnections()
	}
}

func serviceCall(client *core.Client, service string) func(context.Context, string, string, []string, url.Values, http.Header, any, any) error {
	return func(ctx context.Context, method, path string, pathParams []string, query url.Values, customHeaders http.Header, body, out any) error {
		return client.Call(ctx, core.Request{
			Endpoint:   core.Endpoint{Service: service, Method: method, Path: path},
			PathParams: pathParams,
			Query:      query,
			Headers:    customHeaders,
			Body:       body,
		}, out)
	}
}

func serviceStream(client *core.Client, service string) func(context.Context, string, string, []string, url.Values, http.Header, any) (*http.Response, error) {
	return func(ctx context.Context, method, path string, pathParams []string, query url.Values, customHeaders http.Header, body any) (*http.Response, error) {
		return client.Stream(ctx, core.Request{
			Endpoint:   core.Endpoint{Service: service, Method: method, Path: path},
			PathParams: pathParams,
			Query:      query,
			Headers:    customHeaders,
			Body:       body,
		})
	}
}
