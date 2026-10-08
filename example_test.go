package frontal_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	frontal "github.com/frontal-labs/sdk-go"
	"github.com/frontal-labs/sdk-go/pkg/resources"
)

func ExampleNew() {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client, err := frontal.New(
		frontal.WithAPIKey("frt_example_key"),
		frontal.WithBaseURL(server.URL+"/v1"),
		frontal.WithHTTPClient(server.Client()),
	)
	if err != nil {
		panic(err)
	}
	endpoint, _ := client.Agents.Endpoint(http.MethodGet, "/agents/health")
	var health map[string]string
	err = client.Agents.Call(context.Background(), resources.Request{Endpoint: endpoint}, &health)
	if err != nil {
		panic(err)
	}
	fmt.Println(health["status"])

	// Output: ok
}

func ExampleService_Call() {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		payload, _ := io.ReadAll(request.Body)
		var input struct {
			Model string `json:"model"`
		}
		_ = json.Unmarshal(payload, &input)
		_, _ = fmt.Fprintf(writer, `{"model":%q}`, input.Model)
	}))
	defer server.Close()

	client, err := frontal.New(frontal.WithAPIKey("frt_example_key"), frontal.WithBaseURL(server.URL+"/v1"), frontal.WithHTTPClient(server.Client()))
	if err != nil {
		panic(err)
	}
	endpoint, _ := client.AI.Endpoint(http.MethodPost, "/ai/chat/completions")
	var result map[string]string
	err = client.AI.Call(context.Background(), resources.Request{Endpoint: endpoint, Body: map[string]any{"model": "example-model", "messages": []any{}}}, &result)
	if err != nil {
		panic(err)
	}
	fmt.Println(result["model"])

	// Output: example-model
}
