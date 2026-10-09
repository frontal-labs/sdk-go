package functions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestOperations(t *testing.T) {
	definition := FunctionDefinition{
		Name:        "example",
		Description: "Example function",
		Runtime:     RuntimeNodeJS22,
		Entrypoint:  "index.handler",
		Source:      "blob://functions/example",
		InputSchema: map[string]any{"type": "object", "customKey": map[string]any{"inner_key": true}},
		OutputSchema: map[string]any{
			"type":        "object",
			"outputField": map[string]any{"type": "string"},
		},
		Dependencies: []string{"zod"},
		EnvVars:      map[string]string{"MY_VAR": "value"},
		Secrets:      []string{"API_TOKEN"},
		Memory:       256,
		Timeout:      30,
		Permissions:  &Permission{Ontology: []string{"tickets"}, Actions: []string{"assign"}},
	}
	functionResponse := `{"id":"fn-1","name":"example","runtime":"nodejs22","entrypoint":"index.handler","status":"active","version":1,"createdAt":"2025-01-01T00:00:00Z","updatedAt":"2025-01-01T00:00:00Z"}`
	versionResponse := `{"version":2,"source":"blob://functions/example/v2","createdAt":"2025-01-01T00:00:00Z"}`
	executionResponse := `{"id":"exec-1","functionId":"fn-1","version":2,"status":"active","input":{"customerId":"c-1"},"output":{"nested":{"resultKey":true}}}`
	invocationResponse := `{"executionId":"exec-1","result":{"nested":{"resultKey":true}}}`
	query := func(values map[string]string) url.Values {
		result := make(url.Values, len(values))
		for key, value := range values {
			result.Set(key, value)
		}
		return result
	}
	tests := []struct {
		name       string
		method     string
		path       string
		pathParams []string
		query      url.Values
		body       any
		response   string
		outType    reflect.Type
		invoke     func(*Client) error
	}{
		{
			name: "create", method: http.MethodPost, path: functionsPath, body: definition,
			response: functionResponse, outType: reflect.TypeOf((*FunctionResource)(nil)),
			invoke: func(client *Client) error { _, err := client.Create(context.Background(), definition); return err },
		},
		{
			name: "list", method: http.MethodGet, path: functionsPath,
			query:    query(map[string]string{"cursor": "next +/", "limit": "25"}),
			response: `{"functions":[` + functionResponse + `],"pagination":{"cursor":"next","hasMore":true}}`,
			outType:  reflect.TypeOf((*FunctionListResponse)(nil)),
			invoke: func(client *Client) error {
				_, err := client.List(context.Background(), ListOptions{Cursor: "next +/", Limit: 25})
				return err
			},
		},
		{
			name: "get", method: http.MethodGet, path: functionPath, pathParams: []string{"fn-1"},
			response: functionResponse, outType: reflect.TypeOf((*FunctionResource)(nil)),
			invoke: func(client *Client) error { _, err := client.Get(context.Background(), "fn-1"); return err },
		},
		{
			name: "update", method: http.MethodPatch, path: functionPath, pathParams: []string{"fn-1"}, body: definition,
			response: functionResponse, outType: reflect.TypeOf((*FunctionResource)(nil)),
			invoke: func(client *Client) error {
				_, err := client.Update(context.Background(), "fn-1", definition)
				return err
			},
		},
		{
			name: "delete", method: http.MethodDelete, path: functionPath, pathParams: []string{"fn-1"},
			invoke: func(client *Client) error { return client.Delete(context.Background(), "fn-1") },
		},
		{
			name: "list versions", method: http.MethodGet, path: functionVersionsPath, pathParams: []string{"fn-1"},
			query:    query(map[string]string{"cursor": "v2", "limit": "10"}),
			response: `{"versions":[` + versionResponse + `],"pagination":{"cursor":"v3","hasMore":true}}`,
			outType:  reflect.TypeOf((*FunctionVersionListResponse)(nil)),
			invoke: func(client *Client) error {
				_, err := client.ListVersions(context.Background(), "fn-1", ListOptions{Cursor: "v2", Limit: 10})
				return err
			},
		},
		{
			name: "get version", method: http.MethodGet, path: functionVersionPath,
			pathParams: []string{"fn-1", "2"}, response: versionResponse,
			outType: reflect.TypeOf((*FunctionVersion)(nil)),
			invoke:  func(client *Client) error { _, err := client.GetVersion(context.Background(), "fn-1", 2); return err },
		},
		{
			name: "publish version", method: http.MethodPost, path: functionVersionPath + "/publish",
			pathParams: []string{"fn-1", "2"}, body: struct{}{}, response: versionResponse,
			outType: reflect.TypeOf((*FunctionVersion)(nil)),
			invoke: func(client *Client) error {
				_, err := client.PublishVersion(context.Background(), "fn-1", 2)
				return err
			},
		},
		{
			name: "deploy version", method: http.MethodPost, path: functionVersionPath + "/deploy",
			pathParams: []string{"fn-1", "2"}, body: struct{}{},
			invoke: func(client *Client) error { return client.DeployVersion(context.Background(), "fn-1", 2) },
		},
		{
			name: "deployment status", method: http.MethodGet,
			path: functionVersionPath + "/deployment/status", pathParams: []string{"fn-1", "2"},
			response: `{"status":"deployed","details":{"worker":{"ready":true}}}`,
			outType:  reflect.TypeOf((*DeploymentStatus)(nil)),
			invoke: func(client *Client) error {
				_, err := client.GetDeploymentStatus(context.Background(), "fn-1", 2)
				return err
			},
		},
		{
			name: "invoke", method: http.MethodPost, path: functionsPath + "/invoke",
			body:     FunctionInvocationInput{FunctionID: "fn-1", Version: 2, Input: map[string]any{"userKey": "u-1"}},
			response: invocationResponse, outType: reflect.TypeOf((*FunctionInvocationResult)(nil)),
			invoke: func(client *Client) error {
				_, err := client.Invoke(context.Background(), FunctionInvocationInput{FunctionID: "fn-1", Version: 2, Input: map[string]any{"userKey": "u-1"}})
				return err
			},
		},
		{
			name: "invoke async", method: http.MethodPost, path: functionsPath + "/invoke-async",
			body:     FunctionInvocationInput{FunctionID: "fn-1", Input: map[string]any{"userKey": "u-1"}},
			response: `{"executionId":"exec-1"}`, outType: reflect.TypeOf((*AsyncInvocation)(nil)),
			invoke: func(client *Client) error {
				_, err := client.InvokeAsync(context.Background(), FunctionInvocationInput{FunctionID: "fn-1", Input: map[string]any{"userKey": "u-1"}})
				return err
			},
		},
		{
			name: "get execution", method: http.MethodGet, path: executionPath, pathParams: []string{"exec-1"},
			response: executionResponse, outType: reflect.TypeOf((*FunctionExecution)(nil)),
			invoke: func(client *Client) error { _, err := client.GetExecution(context.Background(), "exec-1"); return err },
		},
		{
			name: "get execution result", method: http.MethodGet, path: executionPath + "/result", pathParams: []string{"exec-1"},
			response: invocationResponse, outType: reflect.TypeOf((*FunctionInvocationResult)(nil)),
			invoke: func(client *Client) error {
				_, err := client.GetExecutionResult(context.Background(), "exec-1")
				return err
			},
		},
		{
			name: "list executions", method: http.MethodGet, path: executionsPath,
			query:    query(map[string]string{"cursor": "next", "limit": "40", "functionId": "fn-1", "status": "active"}),
			response: `{"executions":[` + executionResponse + `],"pagination":{"hasMore":false}}`,
			outType:  reflect.TypeOf((*FunctionExecutionListResponse)(nil)),
			invoke: func(client *Client) error {
				_, err := client.ListExecutions(context.Background(), ExecutionListOptions{
					Cursor: "next", Limit: 40, FunctionID: "fn-1", Status: StatusActive,
				})
				return err
			},
		},
		{
			name: "cancel execution", method: http.MethodPost, path: executionPath + "/cancel",
			pathParams: []string{"exec-1"}, body: struct{}{},
			invoke: func(client *Client) error { return client.CancelExecution(context.Background(), "exec-1") },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotPathParams []string
			var gotQuery url.Values
			var gotBody []byte
			var gotOutType reflect.Type
			client := NewClient(func(_ context.Context, method, path string, pathParams []string, got url.Values, _ http.Header, body, out any) error {
				gotMethod, gotPath = method, path
				gotPathParams = append([]string(nil), pathParams...)
				gotQuery = got
				if body != nil {
					encoded, err := json.Marshal(body)
					if err != nil {
						return err
					}
					gotBody = encoded
				}
				if out != nil {
					gotOutType = reflect.TypeOf(out)
					return json.Unmarshal([]byte(test.response), out)
				}
				return nil
			}, nil)

			if err := test.invoke(client); err != nil {
				t.Fatalf("operation returned an error: %v", err)
			}
			if gotMethod != test.method || gotPath != test.path {
				t.Errorf("request = %s %s, want %s %s", gotMethod, gotPath, test.method, test.path)
			}
			if !reflect.DeepEqual(gotPathParams, test.pathParams) {
				t.Errorf("path params = %#v, want %#v", gotPathParams, test.pathParams)
			}
			if !reflect.DeepEqual(gotQuery, test.query) {
				t.Errorf("query = %#v, want %#v", gotQuery, test.query)
			}
			var wantBody []byte
			if test.body != nil {
				encodedBody, err := json.Marshal(test.body)
				if err != nil {
					t.Fatalf("marshal expected request body: %v", err)
				}
				wantBody = encodedBody
			}
			if !reflect.DeepEqual(gotBody, wantBody) {
				t.Errorf("body = %s, want %s", gotBody, wantBody)
			}
			if gotOutType != test.outType {
				t.Errorf("response target type = %v, want %v", gotOutType, test.outType)
			}
		})
	}
}

func TestCreateRequiresDefinitionFields(t *testing.T) {
	client := NewClient(func(context.Context, string, string, []string, url.Values, http.Header, any, any) error {
		t.Fatal("invalid definition reached the transport")
		return nil
	}, nil)
	for _, definition := range []FunctionDefinition{
		{Runtime: RuntimeNodeJS20, Entrypoint: "index.handler"},
		{Name: "example", Entrypoint: "index.handler"},
		{Name: "example", Runtime: RuntimeNodeJS20},
		{Name: "example", Runtime: "unknown", Entrypoint: "index.handler"},
		{Name: "example", Runtime: RuntimeNodeJS20, Entrypoint: "index.handler", Memory: -1},
	} {
		if _, err := client.Create(context.Background(), definition); err == nil {
			t.Errorf("Create(%#v) succeeded without a valid required definition", definition)
		}
	}
}

func TestBuilderCreatesFunction(t *testing.T) {
	client := NewClient(func(_ context.Context, method, path string, _ []string, _ url.Values, _ http.Header, body, out any) error {
		if method != http.MethodPost || path != functionsPath {
			t.Errorf("request = %s %s, want POST %s", method, path, functionsPath)
		}
		definition, ok := body.(FunctionDefinition)
		if !ok || definition.Name != "built" || definition.Runtime != RuntimeNodeJS20 {
			t.Errorf("body = %#v, want builder definition", body)
		}
		return json.Unmarshal([]byte(`{"id":"fn-built","name":"built","runtime":"nodejs20","entrypoint":"index.handler","status":"draft","version":1,"createdAt":"2025-01-01T00:00:00Z","updatedAt":"2025-01-01T00:00:00Z"}`), out)
	}, nil)
	created, err := client.Define("built").Runtime(RuntimeNodeJS20).Entrypoint("index.handler").Create(context.Background())
	if err != nil {
		t.Fatalf("Builder.Create() returned an error: %v", err)
	}
	if created.ID != "fn-built" {
		t.Errorf("Builder.Create() ID = %q, want %q", created.ID, "fn-built")
	}
}
