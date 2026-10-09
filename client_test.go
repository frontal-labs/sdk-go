package frontal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	frontal "github.com/frontal-labs/sdk-go/v2"
	"github.com/frontal-labs/sdk-go/v2/agents"
)

func newTestClient(t *testing.T, server *httptest.Server, options ...frontal.Option) *frontal.Client {
	t.Helper()
	settings := []frontal.Option{
		frontal.WithAPIKey("frt_test_key"),
		frontal.WithBaseURL(server.URL + "/v1"),
		frontal.WithHTTPClient(server.Client()),
		frontal.WithMaxRetries(0),
	}
	settings = append(settings, options...)
	client, err := frontal.New(settings...)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	return client
}

func TestUnifiedClientDispatchesServiceCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/v1/agents/health" {
			t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer frt_test_key" {
			t.Errorf("authorization = %q", got)
		}
		if got := request.Header.Get("X-Frontal-Environment"); got != "test" {
			t.Errorf("environment = %q", got)
		}
		if got := request.Header.Get("X-Frontal-Core"); got != "go@2.0.0" {
			t.Errorf("SDK header = %q", got)
		}
		if got := request.Header.Get("X-Request-ID"); len(got) != 36 {
			t.Errorf("request ID = %q, want UUID format", got)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server, frontal.WithEnvironment("test"))
	var health struct {
		Status string `json:"status"`
	}
	endpoint, ok := client.Agents.Endpoint(http.MethodGet, "/agents/health")
	if !ok {
		t.Fatal("agents health endpoint missing")
	}
	err := client.Agents.Call(context.Background(), agents.Request{Endpoint: endpoint}, &health)
	if err != nil {
		t.Fatalf("call agents health: %v", err)
	}
	if health.Status != "ok" {
		t.Fatalf("health status = %q", health.Status)
	}
	if len(client.AI.Endpoints()) == 0 || len(client.Workflows.Endpoints()) == 0 || len(client.Webhooks.Endpoints()) == 0 {
		t.Fatal("unified client service fields are not initialized")
	}
}

func TestServiceCallRejectsAnEndpointOutsideItsContract(t *testing.T) {
	client, err := frontal.New(frontal.WithAPIKey("frt_test_key"))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	err = client.Call(context.Background(), frontal.Request{
		Endpoint: frontal.Endpoint{Service: "workflows", Method: http.MethodGet, Path: "/workflows"},
	}, nil)
	if !errors.Is(err, frontal.ErrUnknownEndpoint) {
		t.Fatalf("expected unknown endpoint error, got %v", err)
	}
}

func TestRetryUsesRetryAfterAndSucceeds(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			writer.Header().Set("Retry-After", "0")
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = writer.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()
	client := newTestClient(t, server, frontal.WithMaxRetries(1), frontal.WithRetryDelay(0))
	var result map[string]string
	if err := client.Call(context.Background(), frontal.Request{
		Endpoint: frontal.Endpoint{Service: "agents", Method: http.MethodGet, Path: "/agents/health"},
	}, &result); err != nil {
		t.Fatalf("call after retry: %v", err)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("request count = %d, want 2", got)
	}
}

func TestRawGETRetriesBeforeWritingSuccessfulBody(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			writer.Header().Set("Retry-After", "0")
			writer.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = writer.Write([]byte("invoice-pdf"))
	}))
	defer server.Close()
	client := newTestClient(t, server, frontal.WithMaxRetries(1), frontal.WithRetryDelay(0))
	var output bytes.Buffer
	err := client.Call(context.Background(), frontal.Request{
		Endpoint:   frontal.Endpoint{Service: "billing", Method: "GETRAW", Path: "/billing/invoices/{param}/pdf"},
		PathParams: []string{"inv_1"},
	}, &output)
	if err != nil {
		t.Fatalf("download raw endpoint: %v", err)
	}
	if output.String() != "invoice-pdf" || calls.Load() != 2 {
		t.Fatalf("body %q after %d requests", output.String(), calls.Load())
	}
}

func TestAPIErrorCarriesCategoryAndRequestID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("X-Request-ID", "req_test_123")
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusTooManyRequests)
		_, _ = writer.Write([]byte(`{"error":{"code":"RATE_LIMITED","message":"slow down"}}`))
	}))
	defer server.Close()
	client := newTestClient(t, server)
	err := client.Call(context.Background(), frontal.Request{
		Endpoint: frontal.Endpoint{Service: "agents", Method: http.MethodGet, Path: "/agents/health"},
	}, nil)
	var apiError *frontal.APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiError.RequestID != "req_test_123" || apiError.Code != "RATE_LIMITED" || !apiError.Retryable {
		t.Fatalf("unexpected API error: %#v", apiError)
	}
	if !frontal.IsRateLimitError(err) || frontal.IsServerError(err) || frontal.IsAuthError(err) {
		t.Fatalf("API error category helpers did not classify the error")
	}
}

func TestAPIErrorReadsCamelCaseRequestIDFromBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(`{"code":"VALIDATION_ERROR","message":"invalid input","requestId":"req_body_456","fields":[{"field":"name"}]}`))
	}))
	defer server.Close()
	client := newTestClient(t, server)
	err := client.Call(context.Background(), frontal.Request{
		Endpoint: frontal.Endpoint{Service: "agents", Method: http.MethodGet, Path: "/agents/health"},
	}, nil)
	var apiError *frontal.APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiError.RequestID != "req_body_456" || !frontal.IsValidationError(err) || len(apiError.Fields) == 0 {
		t.Fatalf("unexpected API error: %#v", apiError)
	}
}

func TestFetchPageReadsCollectionAndCursor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("status") != "active" {
			t.Errorf("status query = %q", request.URL.Query().Get("status"))
		}
		_, _ = writer.Write([]byte(`{"workflows":[{"id":"wf_1"}],"nextPageToken":"cursor_2"}`))
	}))
	defer server.Close()
	client := newTestClient(t, server)
	endpoint, _ := client.Workflows.Endpoint(http.MethodGet, "/workflows")
	page, err := frontal.FetchPage[map[string]string](context.Background(), client, frontal.Request{
		Endpoint: frontal.Endpoint{Service: "workflows", Method: endpoint.Method, Path: endpoint.Path},
		Query:    map[string][]string{"status": {"active"}},
	}, "workflows")
	if err != nil {
		t.Fatalf("fetch page: %v", err)
	}
	if len(page.Data) != 1 || page.Data[0]["id"] != "wf_1" || !page.HasMore || page.NextCursor != "cursor_2" {
		t.Fatalf("unexpected page: %#v", page)
	}
}

func TestWatchStreamsEventsAndUsesRequestContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/agents/runs/run-1/stream" {
			t.Errorf("stream path = %q", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer frt_test_key" {
			t.Errorf("stream authorization missing")
		}
		if request.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("stream accept = %q", request.Header.Get("Accept"))
		}
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(writer, "id: ev_1\nevent: update\ndata: {\"value\":\"ready\"}\n\n")
	}))
	defer server.Close()
	client := newTestClient(t, server)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	items, err := frontal.Watch[struct {
		Value string `json:"value"`
	}](ctx, client, frontal.Request{
		Endpoint:   frontal.Endpoint{Service: "agents", Method: "STREAM", Path: "/agents/runs/{param}/stream"},
		PathParams: []string{"run-1"},
	})
	if err != nil {
		t.Fatalf("open watch: %v", err)
	}
	item, ok := <-items
	if !ok || item.Err != nil || item.Event == nil {
		t.Fatalf("unexpected stream item: %#v, open=%v", item, ok)
	}
	if item.Event.ID != "ev_1" || item.Event.Name != "update" || item.Event.Data.Value != "ready" {
		t.Fatalf("unexpected event: %#v", item.Event)
	}
	if _, ok := <-items; ok {
		t.Fatal("stream channel should close after response ends")
	}
}

func TestNewReadsEnvironmentConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Frontal-Environment") != "test" {
			t.Errorf("environment = %q", request.Header.Get("X-Frontal-Environment"))
		}
		_, _ = writer.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()
	t.Setenv("FRONTAL_API_KEY", "frt_env_key")
	t.Setenv("FRONTAL_API_URL", server.URL+"/v1")
	t.Setenv("FRONTAL_ENV", "test")
	t.Setenv("FRONTAL_DEBUG", "0")
	t.Setenv("FRONTAL_TIMEOUT", "2500")
	client, err := frontal.New(frontal.WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("create client from env: %v", err)
	}
	var result map[string]string
	if err := client.Call(context.Background(), frontal.Request{
		Endpoint: frontal.Endpoint{Service: "agents", Method: http.MethodGet, Path: "/agents/health"},
	}, &result); err != nil {
		t.Fatalf("call with env client: %v", err)
	}
}

func TestNewRequiresAPIKey(t *testing.T) {
	t.Setenv("FRONTAL_API_KEY", "")
	_, err := frontal.New()
	if err == nil {
		t.Fatal("New should require an API key")
	}
}

func TestIsNetworkError(t *testing.T) {
	if !frontal.IsNetworkError(&url.Error{Op: "Get", URL: "https://api.frontal.dev", Err: errors.New("connection refused")}) {
		t.Fatal("transport failure should be classified as a network error")
	}
	if frontal.IsNetworkError(context.Canceled) || frontal.IsNetworkError(errors.New("bad request")) {
		t.Fatal("cancellation and local errors are not network failures")
	}
}

func TestPageJSONRemainsStandardJSON(t *testing.T) {
	encoded, err := json.Marshal(frontal.Page[string]{Data: []string{"a"}, NextCursor: "cursor", HasMore: true})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" {
		t.Fatal("page did not encode")
	}
}

func TestPollUntilStopsOnContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := frontal.PollUntil(ctx, time.Hour, func(context.Context) (int, error) { return 0, nil }, func(int) bool { return false })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("poll error = %v", err)
	}
}
