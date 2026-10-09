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
	type chatRequest struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	type chatResponse struct {
		Model string `json:"model"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		payload, _ := io.ReadAll(request.Body)
		var input chatRequest
		if err := json.Unmarshal(payload, &input); err != nil {
			http.Error(writer, "invalid request", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprintf(writer, `{"model":%q}`, input.Model)
	}))
	defer server.Close()

	client, err := frontal.New(frontal.WithAPIKey("frt_example_key"), frontal.WithBaseURL(server.URL+"/v1"), frontal.WithHTTPClient(server.Client()))
	if err != nil {
		panic(err)
	}
	operation, err := frontal.BindOperation[chatRequest, chatResponse](client.AI, http.MethodPost, "/ai/chat/completions")
	if err != nil {
		panic(err)
	}
	result, err := operation.Call(context.Background(), frontal.TypedRequest[chatRequest]{
		Body: &chatRequest{Model: "example-model"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Model)

	// Output: example-model
}
