package frontal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/frontal-labs/sdk-go/pkg/handlers"
	"github.com/frontal-labs/sdk-go/pkg/resources"
)

// ErrUnknownEndpoint is returned when a service call uses an unlisted operation.
var ErrUnknownEndpoint = errors.New("frontal: endpoint is not in this service's contract")

// Service is one namespace on a shared Frontal client. Use Endpoints to
// inspect its contract and Call to send an operation with Go request values.
type Service struct {
	name string
	core *resources.Client
}

func newService(core *resources.Client, name string) *Service {
	return &Service{name: name, core: core}
}

// Name returns the service's contract name.
func (service *Service) Name() string {
	if service == nil {
		return ""
	}
	return service.name
}

// Endpoints returns a copy of this service's contract-backed operations.
func (service *Service) Endpoints() []resources.Endpoint {
	if service == nil {
		return nil
	}
	return resources.EndpointsFor(service.name)
}

// Endpoint finds a method/path operation in this service's contract.
func (service *Service) Endpoint(method, path string) (resources.Endpoint, bool) {
	if service == nil {
		return resources.Endpoint{}, false
	}
	return resources.FindEndpoint(service.name, method, path)
}

// Call validates that request belongs to this service, then sends it. Path
// parameters are supplied in contract order through Request.PathParams.
func (service *Service) Call(ctx context.Context, request resources.Request, out any) error {
	if service == nil || service.core == nil {
		return errors.New("frontal: service is not configured")
	}
	endpoint, ok := resources.FindEndpoint(service.name, request.Endpoint.Method, request.Endpoint.Path)
	if !ok {
		return fmt.Errorf("%w: %s %s for %s", ErrUnknownEndpoint, request.Endpoint.Method, request.Endpoint.Path, service.name)
	}
	request.Endpoint = endpoint
	return service.core.Call(ctx, request, out)
}

// CallJSON sends a contract operation and decodes the response into T.
func CallJSON[T any](ctx context.Context, service *Service, request resources.Request) (T, error) {
	var result T
	if service == nil {
		return result, errors.New("frontal: service is nil")
	}
	err := service.Call(ctx, request, &result)
	return result, err
}

// Page is one decoded page from a paginated endpoint. Pass NextCursor back in
// the request query field using the endpoint's cursor parameter to fetch the next page.
type Page[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"nextCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`
	Total      *int64 `json:"total,omitempty"`
}

// FetchPage fetches a page and extracts a collection from a JSON envelope.
// collectionKey may be "data", "workflows", or another top-level array key.
func FetchPage[T any](ctx context.Context, service *Service, request resources.Request, collectionKey string) (Page[T], error) {
	var page Page[T]
	if service == nil {
		return page, errors.New("frontal: service is nil")
	}
	var envelope map[string]json.RawMessage
	if err := service.Call(ctx, request, &envelope); err != nil {
		return page, err
	}
	if collectionKey == "" {
		collectionKey = "data"
	}
	rawCollection := envelope[collectionKey]
	if len(rawCollection) == 0 {
		return Page[T]{}, fmt.Errorf("frontal: page response has no %q collection", collectionKey)
	}
	if err := json.Unmarshal(rawCollection, &page.Data); err != nil {
		return Page[T]{}, fmt.Errorf("frontal: decode page collection %q: %w", collectionKey, err)
	}
	var pagination struct {
		Cursor         string `json:"cursor"`
		NextCursor     string `json:"next_cursor"`
		HasMore        bool   `json:"hasMore"`
		HasMoreSnake   bool   `json:"has_more"`
		NextPageToken  string `json:"nextPageToken"`
		NextPageToken2 string `json:"next_page_token"`
		Total          *int64 `json:"total"`
	}
	if raw := envelope["pagination"]; len(raw) > 0 {
		if err := json.Unmarshal(raw, &pagination); err != nil {
			return Page[T]{}, fmt.Errorf("frontal: decode page metadata: %w", err)
		}
	}
	page.NextCursor = firstNonEmpty(pagination.Cursor, pagination.NextCursor, pagination.NextPageToken, pagination.NextPageToken2, rawString(envelope, "nextPageToken"), rawString(envelope, "next_page_token"))
	page.HasMore = pagination.HasMore || pagination.HasMoreSnake || page.NextCursor != ""
	page.Total = pagination.Total
	return page, nil
}

func rawString(envelope map[string]json.RawMessage, key string) string {
	var value string
	_ = json.Unmarshal(envelope[key], &value)
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// Stream sends an SSE request for an endpoint in this service. The caller owns
// and must close the response body; the request context cancels the stream.
func (service *Service) Stream(ctx context.Context, request resources.Request) (*http.Response, error) {
	if service == nil || service.core == nil {
		return nil, errors.New("frontal: service is not configured")
	}
	endpoint, ok := resources.FindEndpoint(service.name, request.Endpoint.Method, request.Endpoint.Path)
	if !ok {
		return nil, fmt.Errorf("%w: %s %s for %s", ErrUnknownEndpoint, request.Endpoint.Method, request.Endpoint.Path, service.name)
	}
	request.Endpoint = endpoint
	return service.core.Stream(ctx, request)
}

// StreamItem carries a decoded event or a terminal stream error.
type StreamItem[T any] struct {
	Event *handlers.Event[T]
	Err   error
}

// Watch opens a contract-backed SSE endpoint and returns decoded events on a
// receive-only channel. Cancel the context to stop the request and goroutine.
func Watch[T any](ctx context.Context, service *Service, request resources.Request) (<-chan StreamItem[T], error) {
	if ctx == nil {
		return nil, errors.New("frontal: watch context is required")
	}
	response, err := service.Stream(ctx, request)
	if err != nil {
		return nil, err
	}
	if response.Body == nil {
		return nil, errors.New("frontal: stream response body is nil")
	}
	items := make(chan StreamItem[T], 1)
	go func() {
		defer close(items)
		defer func() { _ = response.Body.Close() }()
		err := handlers.DecodeEventStream(ctx, response.Body, func(event handlers.Event[T]) error {
			select {
			case items <- StreamItem[T]{Event: &event}:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			select {
			case items <- StreamItem[T]{Err: err}:
			case <-ctx.Done():
			}
		}
	}()
	return items, nil
}

// PollUntil fetches at interval until done reports a completed value or ctx
// is cancelled. The first fetch happens immediately.
func PollUntil[T any](ctx context.Context, interval time.Duration, fetch func(context.Context) (T, error), done func(T) bool) (T, error) {
	var zero T
	if ctx == nil {
		return zero, errors.New("frontal: polling context is required")
	}
	if interval <= 0 {
		return zero, errors.New("frontal: polling interval must be positive")
	}
	if fetch == nil || done == nil {
		return zero, errors.New("frontal: poll fetch and completion functions are required")
	}
	for {
		value, err := fetch(ctx)
		if err != nil {
			return zero, err
		}
		if done(value) {
			return value, nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}
}

// APIError is the typed error returned for non-success API responses.
type APIError = handlers.APIError

// IsAPIError reports whether err wraps an APIError.
func IsAPIError(err error) bool {
	var apiError *handlers.APIError
	return errors.As(err, &apiError) && apiError != nil
}

// IsAuthError reports whether err is an API authentication or authorization error.
func IsAuthError(err error) bool { return handlers.IsAuthError(err) }

// IsRateLimitError reports whether err is an API rate limit error.
func IsRateLimitError(err error) bool { return handlers.IsRateLimitError(err) }

// IsValidationError reports whether err is an API request validation error.
func IsValidationError(err error) bool { return handlers.IsValidationError(err) }

// IsServerError reports whether err is an API 5xx error.
func IsServerError(err error) bool { return handlers.IsServerError(err) }

// IsNetworkError reports whether err is a transport-level network failure.
func IsNetworkError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var networkError net.Error
	if errors.As(err, &networkError) {
		return true
	}
	var urlError *url.Error
	return errors.As(err, &urlError)
}
