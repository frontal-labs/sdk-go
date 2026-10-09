package agents

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestDefineBuild(t *testing.T) {
	definition, err := Define("ticket-triager").
		Description("Triage support tickets").
		Trigger("support.ticket.created", nil).
		Permissions(Scope{Read: []string{"tickets"}, Actions: []string{"assign"}}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if definition.Name != "ticket-triager" || len(definition.Triggers) != 1 {
		t.Fatalf("unexpected definition: %#v", definition)
	}
	if definition.Timeout != "30s" || definition.Memory.Type != "working" || definition.Retry.MaxRetries == nil || *definition.Retry.MaxRetries != 3 {
		t.Fatalf("schema defaults were not applied: %#v", definition)
	}
	if definition.Confidence.AutoExecuteAbove == nil || *definition.Confidence.AutoExecuteAbove != 0.85 {
		t.Fatalf("confidence default was not applied: %#v", definition.Confidence)
	}
}

func TestRetryDefaultsPreserveExplicitZero(t *testing.T) {
	zero := 0
	definition, err := Define("retry-policy").
		Trigger("event", nil).
		WithRetry(Retry{MaxRetries: &zero}).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if definition.Retry.MaxRetries == nil || *definition.Retry.MaxRetries != 0 {
		t.Fatalf("max retries = %v, want explicit zero", definition.Retry.MaxRetries)
	}
	if definition.Retry.RetryDelay == nil || *definition.Retry.RetryDelay != 1000 {
		t.Fatalf("retry delay = %v, want schema default 1000ms", definition.Retry.RetryDelay)
	}
	if definition.Retry.Backoff == nil || *definition.Retry.Backoff != "exponential" {
		t.Fatalf("backoff = %v, want schema default exponential", definition.Retry.Backoff)
	}
	if definition.Retry.RetryOn == nil || len(*definition.Retry.RetryOn) != 8 {
		t.Fatalf("retry statuses = %v, want schema default status list", definition.Retry.RetryOn)
	}
}

func TestConversationDecodesArbitraryMessages(t *testing.T) {
	var conversation Conversation
	if err := json.Unmarshal([]byte(`{"messages":["hello",{"role":"assistant"},[1,2],null]}`), &conversation); err != nil {
		t.Fatalf("decode conversation: %v", err)
	}
	if len(conversation.Messages) != 4 {
		t.Fatalf("message count = %d, want 4", len(conversation.Messages))
	}
}

func TestDefineBuildRequiresNameAndTrigger(t *testing.T) {
	for _, builder := range []*Builder{Define(""), Define("agent")} {
		if _, err := builder.Build(); err == nil {
			t.Fatal("Build() succeeded without required definition fields")
		}
	}
}

func TestBuildRejectsInvalidSchemaValues(t *testing.T) {
	aboveOne := 1.1
	negative := -1
	zero := 0
	invalidCases := []struct {
		name   string
		modify func(*Builder) *Builder
	}{
		{"confidence threshold", func(builder *Builder) *Builder {
			return builder.WithConfidence(Confidence{AutoExecuteAbove: &aboveOne})
		}},
		{"memory type", func(builder *Builder) *Builder {
			return builder.WithMemory(Memory{Type: "volatile"})
		}},
		{"memory token limit", func(builder *Builder) *Builder {
			return builder.WithMemory(Memory{MaxTokens: negative})
		}},
		{"retry count", func(builder *Builder) *Builder {
			return builder.WithRetry(Retry{MaxRetries: &negative})
		}},
		{"retry delay", func(builder *Builder) *Builder {
			return builder.WithRetry(Retry{RetryDelay: &zero})
		}},
		{"retry backoff", func(builder *Builder) *Builder {
			backoff := "random"
			return builder.WithRetry(Retry{Backoff: &backoff})
		}},
		{"rate limit", func(builder *Builder) *Builder {
			return builder.WithRateLimit(RateLimit{MaxConcurrent: negative})
		}},
	}
	for _, test := range invalidCases {
		t.Run(test.name, func(t *testing.T) {
			builder := test.modify(Define("agent").Trigger("event", nil))
			if _, err := builder.Build(); err == nil {
				t.Fatal("Build() accepted a value outside the agent schema")
			}
		})
	}
}

func TestValidTimeoutClearsPriorTimeoutError(t *testing.T) {
	definition, err := Define("agent").
		Trigger("event", nil).
		WithTimeout(-time.Second).
		WithTimeout(30 * time.Second).
		Build()
	if err != nil {
		t.Fatalf("Build() after correcting timeout: %v", err)
	}
	if definition.Timeout != "30s" {
		t.Fatalf("timeout = %q, want 30s", definition.Timeout)
	}
}

func TestClientBuilderCreatesAgent(t *testing.T) {
	var gotMethod, gotPath string
	client := NewClient(func(_ context.Context, method, path string, _ []string, _ url.Values, _ http.Header, body, out any) error {
		gotMethod, gotPath = method, path
		if definition, ok := body.(Definition); !ok || definition.Timeout != "30s" {
			t.Fatalf("request body = %#v, want defaulted Definition", body)
		}
		if agent, ok := out.(*Agent); ok {
			agent.ID = "agent-1"
		}
		return nil
	}, nil)
	created, err := client.Define("triager").Trigger("ticket.created", nil).Create(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "agent-1" || gotMethod != http.MethodPost || gotPath != "/agents" {
		t.Fatalf("created=%#v request=%s %s", created, gotMethod, gotPath)
	}
}

func TestTypedAgentMethods(t *testing.T) {
	tests := []struct {
		name       string
		wantMethod string
		wantPath   string
		invoke     func(*Client) error
	}{
		{"create", http.MethodPost, "/agents", func(client *Client) error {
			_, err := client.Create(context.Background(), Definition{Name: "triager", Triggers: []Trigger{{Event: "ticket.created"}}})
			return err
		}},
		{"get", http.MethodGet, "/agents/{param}", func(client *Client) error {
			_, err := client.Get(context.Background(), "agent-1")
			return err
		}},
		{"delete", http.MethodDelete, "/agents/{param}", func(client *Client) error {
			return client.Delete(context.Background(), "agent-1")
		}},
		{"start run", http.MethodPost, "/agents/{param}/runs", func(client *Client) error {
			_, err := client.StartRun(context.Background(), "agent-1", "ticket.created", map[string]any{"id": "ticket-1"})
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var gotMethod, gotPath string
			client := NewClient(func(_ context.Context, method, path string, _ []string, _ url.Values, _ http.Header, body, out any) error {
				gotMethod, gotPath = method, path
				if output, ok := out.(*Agent); ok {
					output.ID = "agent-1"
				}
				if method == http.MethodPost {
					encoded, err := json.Marshal(body)
					if err != nil {
						return err
					}
					if len(encoded) == 0 {
						t.Fatal("expected request body")
					}
				}
				return nil
			}, nil)
			if err := test.invoke(client); err != nil {
				t.Fatal(err)
			}
			if gotMethod != test.wantMethod || gotPath != test.wantPath {
				t.Fatalf("called %s %s; want %s %s", gotMethod, gotPath, test.wantMethod, test.wantPath)
			}
		})
	}
}

func TestListBuildsQuery(t *testing.T) {
	want := url.Values{"status": {"active"}, "trigger": {"ticket.created"}, "limit": {"25"}, "cursor": {"next"}}
	client := NewClient(func(_ context.Context, _, _ string, _ []string, query url.Values, _ http.Header, _, _ any) error {
		if !reflect.DeepEqual(query, want) {
			t.Fatalf("query = %v, want %v", query, want)
		}
		return nil
	}, nil)
	_, err := client.List(context.Background(), ListOptions{Status: AgentStatusActive, Trigger: "ticket.created", Limit: 25, Cursor: "next"})
	if err != nil {
		t.Fatal(err)
	}
}
