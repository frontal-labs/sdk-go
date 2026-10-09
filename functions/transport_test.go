package functions_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	frontal "github.com/frontal-labs/sdk-go/v2"
	"github.com/frontal-labs/sdk-go/v2/functions"
)

func TestSharedTransportConfigurationAndJSON(t *testing.T) {
	const functionResponse = `{"id":"fn-1","name":"hello","runtime":"nodejs22","entrypoint":"index.handler","status":"active","version":1,"createdAt":"2025-01-01T00:00:00Z","updatedAt":"2025-01-01T00:00:00Z"}`
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Authorization"); got != "Bearer frt_test_key" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer frt_test_key")
		}
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/v1/functions":
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Errorf("decode function definition: %v", err)
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			schema, ok := body["inputSchema"].(map[string]any)
			if !ok || schema["customKey"] == nil {
				t.Errorf("inputSchema keys were changed: %#v", body["inputSchema"])
			}
			if nested, ok := schema["customKey"].(map[string]any); !ok || nested["inner_key"] != true {
				t.Errorf("nested schema keys were changed: %#v", schema["customKey"])
			}
			writeJSON(t, writer, json.RawMessage(functionResponse))
		case request.Method == http.MethodGet && request.URL.Path == "/v1/functions":
			if got := request.URL.Query().Get("cursor"); got != "cursor +/?&" {
				t.Errorf("cursor query = %q, want %q", got, "cursor +/?&")
			}
			if got := request.URL.Query().Get("limit"); got != "7" {
				t.Errorf("limit query = %q, want %q", got, "7")
			}
			writeJSON(t, writer, json.RawMessage(`{"functions":[],"pagination":{"cursor":"next","hasMore":true}}`))
		case request.Method == http.MethodGet && request.URL.Path == "/v1/functions/executions":
			query := request.URL.Query()
			if query.Get("cursor") != "execution-next" || query.Get("limit") != "3" || query.Get("functionId") != "fn-1" || query.Get("status") != "failed" {
				t.Errorf("execution query = %v, want cursor, limit, functionId, and status filters", query)
			}
			writeJSON(t, writer, json.RawMessage(`{"executions":[],"pagination":{"hasMore":false}}`))
		case request.Method == http.MethodPost && request.URL.Path == "/v1/functions/invoke":
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Errorf("decode invocation input: %v", err)
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			input, ok := body["input"].(map[string]any)
			if !ok || input["camelCase"] != "value" || input["snake_key"] != "also-preserved" {
				t.Errorf("invocation input keys were changed: %#v", body["input"])
			}
			writeJSON(t, writer, json.RawMessage(`{"executionId":"exec-1","result":{"outerKey":{"inner_key":[1,true,{"leafValue":"ok"}]}}}`))
		case request.Method == http.MethodGet && request.URL.EscapedPath() == "/v1/functions/org%2Ffunction%20one":
			writeJSON(t, writer, json.RawMessage(functionResponse))
		default:
			t.Errorf("unexpected request: %s %s", request.Method, request.URL.RequestURI())
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := frontal.New(
		frontal.WithAPIKey("frt_test_key"),
		frontal.WithBaseURL(server.URL+"/v1"),
		frontal.WithTimeout(time.Second),
		frontal.WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("frontal.New() returned an error: %v", err)
	}
	created, err := client.Functions.Create(context.Background(), functions.FunctionDefinition{
		Name:       "hello",
		Runtime:    functions.RuntimeNodeJS22,
		Entrypoint: "index.handler",
		InputSchema: map[string]any{
			"type":      "object",
			"customKey": map[string]any{"inner_key": true},
		},
	})
	if err != nil {
		t.Fatalf("Functions.Create() returned an error: %v", err)
	}
	if created.ID != "fn-1" || created.Runtime != functions.RuntimeNodeJS22 {
		t.Errorf("Functions.Create() = %#v, want function fn-1 with nodejs22 runtime", created)
	}

	page, err := client.Functions.List(context.Background(), functions.ListOptions{Cursor: "cursor +/?&", Limit: 7})
	if err != nil {
		t.Fatalf("Functions.List() returned an error: %v", err)
	}
	if page.Pagination.Cursor != "next" || !page.Pagination.HasMore {
		t.Errorf("Functions.List() pagination = %#v, want cursor next and hasMore", page.Pagination)
	}
	if _, err := client.Functions.ListExecutions(context.Background(), functions.ExecutionListOptions{
		Cursor: "execution-next", Limit: 3, FunctionID: "fn-1", Status: functions.StatusFailed,
	}); err != nil {
		t.Fatalf("Functions.ListExecutions() returned an error: %v", err)
	}

	result, err := client.Functions.Invoke(context.Background(), functions.FunctionInvocationInput{
		FunctionID: "fn-1",
		Input: map[string]any{
			"camelCase": "value",
			"snake_key": "also-preserved",
		},
	})
	if err != nil {
		t.Fatalf("Functions.Invoke() returned an error: %v", err)
	}
	if result.ExecutionID != "exec-1" {
		t.Errorf("Functions.Invoke() execution ID = %q, want %q", result.ExecutionID, "exec-1")
	}
	outer, ok := result.Result["outerKey"].(map[string]any)
	if !ok {
		t.Fatalf("Functions.Invoke() result = %#v, want outerKey object", result.Result)
	}
	inner, ok := outer["inner_key"].([]any)
	if !ok || len(inner) != 3 {
		t.Fatalf("nested result = %#v, want three elements under inner_key", outer["inner_key"])
	}
	leaf, ok := inner[2].(map[string]any)
	if !ok || leaf["leafValue"] != "ok" {
		t.Errorf("nested result leaf = %#v, want leafValue preserved", inner[2])
	}

	if _, err := client.Functions.Get(context.Background(), "org/function one"); err != nil {
		t.Fatalf("Functions.Get() with a reserved path ID returned an error: %v", err)
	}
}

func TestAPIErrorMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		if _, err := io.WriteString(writer, `{"error":{"code":"invalid_function","message":"definition rejected"}}`); err != nil {
			t.Errorf("write API error response: %v", err)
		}
	}))
	defer server.Close()
	client := newTestClient(t, server.URL, frontal.WithMaxRetries(0))

	_, err := client.Functions.Get(context.Background(), "bad-id")
	if err == nil {
		t.Fatal("Functions.Get() succeeded for an API error response")
	}
	var apiError *frontal.APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("Functions.Get() error = %T, want *frontal.APIError", err)
	}
	if apiError.StatusCode != http.StatusUnprocessableEntity || apiError.Code != "invalid_function" {
		t.Errorf("API error = %#v, want status 422 and code invalid_function", apiError)
	}
	if !frontal.IsValidationError(err) {
		t.Errorf("IsValidationError(%v) = false, want true", err)
	}
}

func TestDefaultSafeGetRetriesThreeTimes(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		attempt := requests.Add(1)
		if attempt <= 3 {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writeJSON(t, writer, json.RawMessage(`{"functions":[],"pagination":{"hasMore":false}}`))
	}))
	defer server.Close()
	client, err := frontal.New(
		frontal.WithAPIKey("frt_test_key"),
		frontal.WithBaseURL(server.URL),
		frontal.WithRetryDelay(0),
	)
	if err != nil {
		t.Fatalf("frontal.New() returned an error: %v", err)
	}
	if _, err := client.Functions.List(context.Background(), functions.ListOptions{}); err != nil {
		t.Fatalf("Functions.List() returned an error after retries: %v", err)
	}
	if got := requests.Load(); got != 4 {
		t.Errorf("GET attempts = %d, want 4 (one initial attempt and three retries)", got)
	}
}

func TestPostIsNotRetried(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client, err := frontal.New(
		frontal.WithAPIKey("frt_test_key"),
		frontal.WithBaseURL(server.URL),
		frontal.WithMaxRetries(3),
		frontal.WithRetryDelay(0),
	)
	if err != nil {
		t.Fatalf("frontal.New() returned an error: %v", err)
	}
	_, err = client.Functions.Create(context.Background(), functions.FunctionDefinition{
		Name: "hello", Runtime: functions.RuntimeNodeJS20, Entrypoint: "index.handler",
	})
	if err == nil {
		t.Fatal("Functions.Create() succeeded for a 503 response")
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("POST attempts = %d, want 1", got)
	}
}

func TestConfiguredTimeout(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		close(started)
		<-request.Context().Done()
	}))
	defer server.Close()
	client, err := frontal.New(
		frontal.WithAPIKey("frt_test_key"),
		frontal.WithBaseURL(server.URL),
		frontal.WithTimeout(25*time.Millisecond),
		frontal.WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("frontal.New() returned an error: %v", err)
	}
	_, err = client.Functions.List(context.Background(), functions.ListOptions{})
	if err == nil {
		t.Fatal("Functions.List() succeeded after the configured timeout")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Functions.List() error = %v, want context deadline exceeded", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("timed-out request did not reach the test server")
	}
}

func newTestClient(t *testing.T, baseURL string, options ...frontal.Option) *frontal.Client {
	t.Helper()
	options = append([]frontal.Option{frontal.WithAPIKey("frt_test_key"), frontal.WithBaseURL(baseURL)}, options...)
	client, err := frontal.New(options...)
	if err != nil {
		t.Fatalf("frontal.New() returned an error: %v", err)
	}
	return client
}

func writeJSON(t *testing.T, writer http.ResponseWriter, value any) {
	t.Helper()
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		t.Errorf("encode response JSON: %v", err)
	}
}
