package contracts

import (
	"context"
	"embed"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	frontal "github.com/frontal-labs/sdk-go"
	"github.com/frontal-labs/sdk-go/pkg/resources"
	"github.com/frontal-labs/sdk-go/pkg/utils"
)

//go:embed openapi/api.openapi.json openapi/ai.openapi.generated.json sdk-endpoints.json
var snapshots embed.FS

type openAPISpec struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

type openAPIOperation struct {
	OperationID string   `json:"operationId"`
	Tags        []string `json:"tags"`
}

func TestOpenAPIOperationsResolveToHTTPRequests(t *testing.T) {
	base, err := utils.ParseBaseURL(resources.DefaultBaseURL)
	if err != nil {
		t.Fatal(err)
	}
	for _, filename := range []string{"openapi/api.openapi.json", "openapi/ai.openapi.generated.json"} {
		filename := filename
		t.Run(filename, func(t *testing.T) {
			contents, err := snapshots.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}
			var spec openAPISpec
			if err := json.Unmarshal(contents, &spec); err != nil {
				t.Fatalf("decode spec: %v", err)
			}
			operationCount := 0
			for path, pathItem := range spec.Paths {
				for method, rawOperation := range pathItem {
					method = strings.ToUpper(method)
					if !isHTTPMethod(method) {
						continue
					}
					operationCount++
					var operation openAPIOperation
					if err := json.Unmarshal(rawOperation, &operation); err != nil {
						t.Fatalf("decode %s %s: %v", method, path, err)
					}
					if operation.OperationID == "" {
						t.Errorf("%s %s has no operationId", method, path)
						continue
					}
					service := serviceForTag(operation.Tags)
					if service == "" {
						t.Errorf("%s %s has no supported SDK service tag", method, path)
						continue
					}
					if !catalogContainsEndpoint(service, method, normalizeOperationPath(path)) {
						t.Errorf("%s %s (%s) has no %s SDK endpoint descriptor", method, path, operation.OperationID, service)
					}
					target, err := utils.ResolveEndpoint(base, path)
					if err != nil {
						t.Errorf("resolve %s %s: %v", method, path, err)
						continue
					}
					if !utils.SameOrigin(base, target) {
						t.Errorf("%s %s resolved outside the configured API origin", method, path)
						continue
					}
					request, err := http.NewRequestWithContext(context.Background(), method, target.String(), nil)
					if err != nil {
						t.Errorf("build request for %s %s (%s): %v", method, path, operation.OperationID, err)
						continue
					}
					if request.Method != method || !utils.SameOrigin(base, request.URL) {
						t.Errorf("%s %s did not map to a request at the configured API origin", method, path)
					}
				}
			}
			if operationCount == 0 {
				t.Errorf("%s contains no HTTP operations", filename)
			}
		})
	}
}

func catalogContainsEndpoint(service, method, path string) bool {
	for _, endpoint := range resources.Endpoints() {
		candidateMethod := endpoint.Method
		switch candidateMethod {
		case "GETRAW", "STREAM":
			candidateMethod = http.MethodGet
		case "POSTRAW", "POSTFORMDATA":
			candidateMethod = http.MethodPost
		}
		if endpoint.Service == service && candidateMethod == method && normalizeOperationPath(endpoint.Path) == path {
			return true
		}
	}
	return false
}

func normalizeOperationPath(path string) string {
	if before, _, found := strings.Cut(path, "?"); found {
		path = before
	}
	if strings.HasPrefix(path, "/v1/") {
		path = strings.TrimPrefix(path, "/v1")
	}
	segments := strings.Split(path, "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			segments[index] = "{param}"
		}
	}
	return strings.Join(segments, "/")
}

func serviceForTag(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	switch tags[0] {
	case "ontology":
		return "ontology"
	case "data":
		return "data"
	case "events", "observability":
		return "observability"
	case "integrations", "action-runs", "connection-tests":
		return "connectors"
	case "auth":
		return "auth"
	case "webhook-endpoints":
		return "webhooks"
	case "workflows":
		return "workflows"
	case "invocations", "providers", "runtime-api", "gateway", "unified-api":
		return "ai"
	default:
		return ""
	}
}

func TestEverySDKServiceHasAClientNamespace(t *testing.T) {
	contents, err := snapshots.ReadFile("sdk-endpoints.json")
	if err != nil {
		t.Fatal(err)
	}
	var inventory map[string]json.RawMessage
	if err := json.Unmarshal(contents, &inventory); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FRONTAL_API_URL", "")
	t.Setenv("FRONTAL_TIMEOUT", "")
	t.Setenv("FRONTAL_DEBUG", "0")
	client, err := frontal.New(frontal.WithAPIKey("frt_contract_test"), frontal.WithBaseURL(resources.DefaultBaseURL), frontal.WithEnvironment("test"))
	if err != nil {
		t.Fatal(err)
	}
	for name := range inventory {
		service := serviceFor(client, name)
		if service == nil {
			t.Errorf("SDK service %q has no unified client namespace", name)
			continue
		}
		if service.Name() != name {
			t.Errorf("namespace for %q reports name %q", name, service.Name())
		}
	}
}

func serviceFor(client *frontal.Client, name string) *frontal.Service {
	switch name {
	case "ai":
		return client.AI
	case "agents":
		return client.Agents
	case "audit":
		return client.Audit
	case "auth":
		return client.Auth
	case "billing":
		return client.Billing
	case "blob":
		return client.Blob
	case "connectors":
		return client.Connectors
	case "data":
		return client.Data
	case "governance":
		return client.Governance
	case "lineage":
		return client.Lineage
	case "observability":
		return client.Observability
	case "ontology":
		return client.Ontology
	case "pipelines":
		return client.Pipelines
	case "react":
		return client.React
	case "sandbox":
		return client.Sandbox
	case "schedules":
		return client.Schedules
	case "webhooks":
		return client.Webhooks
	case "workflows":
		return client.Workflows
	default:
		return nil
	}
}

func isHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
