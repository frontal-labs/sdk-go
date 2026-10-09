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

	"github.com/frontal-labs/sdk-go/v2/internal/core"
	"github.com/frontal-labs/sdk-go/v2/internal/handlers"
)

// ErrUnknownEndpoint reports a request that does not match the committed endpoint inventory.
var ErrUnknownEndpoint = errors.New("frontal: endpoint is not in the contract inventory")

// Endpoint identifies an operation from the committed endpoint inventory.
type Endpoint struct {
	Service string `json:"service"`
	Method  string `json:"method"`
	Path    string `json:"path"`
}

// Request describes a contract-backed request sent through Client.Call.
type Request struct {
	Endpoint   Endpoint
	PathParams []string
	Query      url.Values
	Headers    http.Header
	Body       any
}

// TypedRequest contains a caller-defined body and parameters for a bound operation.
type TypedRequest[Body any] struct {
	PathParams []string
	Query      url.Values
	Headers    http.Header
	Body       *Body
}

// Operation binds endpoint identity to caller-defined request and response types.
type Operation[RequestBody, Response any] struct {
	client   *Client
	endpoint Endpoint
}

// BindOperation binds a route in service to caller-defined types.
func BindOperation[RequestBody, Response any](client *Client, service, method, path string) (*Operation[RequestBody, Response], error) {
	if client == nil || client.core == nil {
		return nil, errors.New("frontal: client is not configured")
	}
	endpoint, ok := core.FindEndpoint(service, method, path)
	if !ok {
		return nil, fmt.Errorf("%w: %s %s for %s", ErrUnknownEndpoint, method, path, service)
	}
	return &Operation[RequestBody, Response]{client: client, endpoint: Endpoint{Service: endpoint.Service, Method: endpoint.Method, Path: endpoint.Path}}, nil
}

// Call sends a bound operation and decodes its response into Response.
func (operation *Operation[RequestBody, Response]) Call(ctx context.Context, request TypedRequest[RequestBody]) (Response, error) {
	var zero Response
	if operation == nil || operation.client == nil {
		return zero, errors.New("frontal: operation is not configured")
	}
	var result Response
	err := operation.client.Call(ctx, Request{
		Endpoint: operation.endpoint, PathParams: request.PathParams,
		Query: request.Query, Headers: request.Headers, Body: request.Body,
	}, &result)
	return result, err
}

// CallJSON sends a request and decodes the response into T.
func CallJSON[T any](ctx context.Context, client *Client, request Request) (T, error) {
	var result T
	if client == nil {
		return result, errors.New("frontal: client is nil")
	}
	err := client.Call(ctx, request, &result)
	return result, err
}

// Page is one decoded page from a cursor-paginated response.
type Page[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"nextCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`
	Total      *int64 `json:"total,omitempty"`
}

// FetchPage decodes one collection page from a response envelope.
func FetchPage[T any](ctx context.Context, client *Client, request Request, collectionKey string) (Page[T], error) {
	var page Page[T]
	if client == nil {
		return page, errors.New("frontal: client is nil")
	}
	var envelope map[string]json.RawMessage
	if err := client.Call(ctx, request, &envelope); err != nil {
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

// StreamItem carries a decoded event or a terminal stream error.
type StreamItem[T any] struct {
	Event *handlers.Event[T]
	Err   error
}

// Watch opens a contract-backed JSON SSE stream and returns decoded events.
func Watch[T any](ctx context.Context, client *Client, request Request) (<-chan StreamItem[T], error) {
	if ctx == nil {
		return nil, errors.New("frontal: watch context is required")
	}
	response, err := client.Stream(ctx, request)
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

// PollUntil fetches at interval until done reports a completed value or ctx is cancelled.
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
	var lastValue T
	hasValue := false
	for {
		if err := ctx.Err(); err != nil {
			if hasValue {
				return lastValue, err
			}
			return zero, err
		}
		value, err := fetch(ctx)
		if err != nil {
			if hasValue && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
				return lastValue, err
			}
			return zero, err
		}
		lastValue, hasValue = value, true
		if err := ctx.Err(); err != nil {
			return value, err
		}
		if done(value) {
			return value, nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return lastValue, ctx.Err()
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
