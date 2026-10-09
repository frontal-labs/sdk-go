// Package agents provides types and functions for working with Frontal Agents.
package agents

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/url"
	"strconv"
	"time"
)

// AgentStatus is the lifecycle state of an agent.
type AgentStatus string

const (
	// AgentStatusDraft represents the draft status of an agent.
	AgentStatusDraft AgentStatus = "draft"
	// AgentStatusActive represents the active status of an agent.
	AgentStatusActive AgentStatus = "active"
	// AgentStatusPaused represents the paused status of an agent.
	AgentStatusPaused AgentStatus = "paused"
	// AgentStatusDeprecated represents the deprecated status of an agent.
	AgentStatusDeprecated AgentStatus = "deprecated"
)

// Trigger describes an event that can activate an agent.
type Trigger struct {
	Event    string         `json:"event"`
	Filter   map[string]any `json:"filter,omitempty"`
	Debounce string         `json:"debounce,omitempty"`
}

// Scope contains the permissions granted to an agent.
type Scope struct {
	Read            []string `json:"read"`
	Write           []string `json:"write"`
	Actions         []string `json:"actions"`
	Escalate        []string `json:"escalate"`
	InvokeAgents    []string `json:"invokeAgents"`
	InvokeFunctions []string `json:"invokeFunctions"`
}

// Confidence configures automatic execution and escalation thresholds.
type Confidence struct {
	AutoExecuteAbove     *float64 `json:"autoExecuteAbove,omitempty"`
	EscalateBelow        *float64 `json:"escalateBelow,omitempty"`
	RequireReviewBetween *bool    `json:"requireReviewBetween,omitempty"`
}

// Memory configures the agent's memory strategy.
type Memory struct {
	Type      string `json:"type,omitempty"`
	TTL       string `json:"ttl,omitempty"`
	MaxTokens int    `json:"maxTokens,omitempty"`
}

// Retry configures retries after failed executions.
type Retry struct {
	MaxRetries *int    `json:"maxRetries,omitempty"`
	RetryDelay *int    `json:"retryDelay,omitempty"` // milliseconds
	Backoff    *string `json:"backoff,omitempty"`
	RetryOn    *[]int  `json:"retryOn,omitempty"`
}

// RateLimit caps an agent's execution rate.
type RateLimit struct {
	MaxExecutionsPerMinute int `json:"maxExecutionsPerMinute,omitempty"`
	MaxConcurrent          int `json:"maxConcurrent,omitempty"`
}

// Definition is the configuration used to create or update an agent.
type Definition struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Triggers    []Trigger  `json:"triggers"`
	Scope       Scope      `json:"scope,omitempty"`
	Confidence  Confidence `json:"confidence,omitempty"`
	Memory      Memory     `json:"memory,omitempty"`
	Retry       Retry      `json:"retry,omitempty"`
	Timeout     string     `json:"timeout,omitempty"`
	RateLimit   *RateLimit `json:"rateLimit,omitempty"`
	Tags        []string   `json:"tags"`
}

// UpdateDefinition contains only the fields to change on an existing agent.
type UpdateDefinition struct {
	Name        *string     `json:"name,omitempty"`
	Description *string     `json:"description,omitempty"`
	Triggers    *[]Trigger  `json:"triggers,omitempty"`
	Scope       *Scope      `json:"scope,omitempty"`
	Confidence  *Confidence `json:"confidence,omitempty"`
	Memory      *Memory     `json:"memory,omitempty"`
	Retry       *Retry      `json:"retry,omitempty"`
	Timeout     *string     `json:"timeout,omitempty"`
	RateLimit   *RateLimit  `json:"rateLimit,omitempty"`
	Tags        *[]string   `json:"tags,omitempty"`
}

// Metrics summarizes an agent's recent execution performance.
type Metrics struct {
	ExecutionsToday int     `json:"executionsToday"`
	EscalationRate  float64 `json:"escalationRate"`
	AvgExecutionMS  int     `json:"avgExecutionMs"`
	SuccessRate     float64 `json:"successRate"`
}

// Agent is a configured agent and its current lifecycle metadata.
type Agent struct {
	Definition
	ID          string      `json:"id"`
	Version     int         `json:"version"`
	Status      AgentStatus `json:"status"`
	Environment string      `json:"environment"`
	Metrics     *Metrics    `json:"metrics,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// ListOptions filters and paginates the agent collection.
type ListOptions struct {
	Status  AgentStatus
	Trigger string
	Limit   int
	Cursor  string
}

// Page is a page of agent resources.
type Page struct {
	Data       []Agent `json:"data"`
	Pagination struct {
		Cursor  string `json:"cursor"`
		HasMore bool   `json:"hasMore"`
		Total   *int64 `json:"total,omitempty"`
	} `json:"pagination"`
}

// ExecutionStatus is the state of an agent execution.
type ExecutionStatus string

const (
	// ExecutionStatusRunning represents the running status of an agent execution.
	ExecutionStatusRunning ExecutionStatus = "running"
	// ExecutionStatusCompleted represents the completed status of an agent execution.
	ExecutionStatusCompleted ExecutionStatus = "completed"
	// ExecutionStatusFailed represents the failed status of an agent execution.
	ExecutionStatusFailed ExecutionStatus = "failed"
	// ExecutionStatusEscalated represents the escalated status of an agent execution.
	ExecutionStatusEscalated ExecutionStatus = "escalated"
)

// Execution is an agent run and its decision trace.
type Execution struct {
	ID             string          `json:"id"`
	AgentID        string          `json:"agentId"`
	TriggerEvent   string          `json:"triggerEvent"`
	TriggerPayload map[string]any  `json:"triggerPayload"`
	Status         ExecutionStatus `json:"status"`
	Outcome        string          `json:"outcome,omitempty"`
	Confidence     *float64        `json:"confidence,omitempty"`
	DecisionTrace  []DecisionStep  `json:"decisionTrace,omitempty"`
	ActionsTaken   []any           `json:"actionsTaken,omitempty"`
	EscalationID   string          `json:"escalationId,omitempty"`
	StartedAt      time.Time       `json:"startedAt"`
	CompletedAt    *time.Time      `json:"completedAt,omitempty"`
	DurationMS     *int            `json:"durationMs,omitempty"`
	Error          string          `json:"error,omitempty"`
}

// DecisionStep is one step in an agent's execution trace.
type DecisionStep struct {
	Step        int          `json:"step"`
	Type        string       `json:"type"`
	Description string       `json:"description"`
	DataRead    []DataRead   `json:"dataRead,omitempty"`
	Reasoning   string       `json:"reasoning,omitempty"`
	Confidence  *float64     `json:"confidence,omitempty"`
	ActionTaken *ActionTaken `json:"actionTaken,omitempty"`
	DurationMS  int          `json:"durationMs"`
}

// DataRead identifies the fields read from an entity during a run.
type DataRead struct {
	Entity string   `json:"entity"`
	ID     string   `json:"id"`
	Fields []string `json:"fields"`
}

// ActionTaken describes an action selected by the agent.
type ActionTaken struct {
	Type       string         `json:"type"`
	Parameters map[string]any `json:"parameters"`
}

// Health is the Agents service health response.
type Health struct {
	Status string `json:"status"`
}

// RunListOptions filters and paginates agent executions.
type RunListOptions struct {
	Status string
	From   string
	To     string
	Limit  int
	Cursor string
}

// ExecutionPage is one page of agent runs.
type ExecutionPage struct {
	Data       []Execution `json:"data"`
	Pagination struct {
		Cursor  string `json:"cursor"`
		HasMore bool   `json:"hasMore"`
		Total   *int64 `json:"total,omitempty"`
	} `json:"pagination"`
}

// VersionPage contains version records. The public schema currently treats
// each version entry as an open object, so entries remain RawMessages.
type VersionPage struct {
	Data       []json.RawMessage `json:"data"`
	Pagination struct {
		Cursor  string `json:"cursor"`
		HasMore bool   `json:"hasMore"`
		Total   *int64 `json:"total,omitempty"`
	} `json:"pagination"`
}

// Conversation contains the messages associated with an execution.
type Conversation struct {
	Messages []any `json:"messages"`
}

// Builder incrementally defines an agent before creating it.
type Builder struct {
	definition Definition
	client     *Client
	err        error
}

// Define starts a fluent definition for a new agent.
func Define(name string) *Builder {
	return &Builder{definition: defaultDefinition(Definition{Name: name})}
}

// Define starts a fluent definition attached to client.
func (client *Client) Define(name string) *Builder {
	builder := Define(name)
	builder.client = client
	return builder
}

// Description sets the agent's description.
func (builder *Builder) Description(description string) *Builder {
	if builder != nil {
		builder.definition.Description = description
	}
	return builder
}

// Trigger adds an event trigger to the definition.
func (builder *Builder) Trigger(event string, filter map[string]any) *Builder {
	if builder != nil {
		builder.definition.Triggers = append(builder.definition.Triggers, Trigger{Event: event, Filter: filter})
	}
	return builder
}

// Permissions sets the agent's access scope.
func (builder *Builder) Permissions(scope Scope) *Builder {
	if builder != nil {
		builder.definition.Scope = scope
	}
	return builder
}

// WithConfidence sets the agent's confidence thresholds.
func (builder *Builder) WithConfidence(confidence Confidence) *Builder {
	if builder != nil {
		builder.definition.Confidence = confidence
	}
	return builder
}

// WithMemory sets the agent's memory strategy.
func (builder *Builder) WithMemory(memory Memory) *Builder {
	if builder != nil {
		builder.definition.Memory = memory
	}
	return builder
}

// WithRetry sets the agent's retry policy.
func (builder *Builder) WithRetry(retry Retry) *Builder {
	if builder != nil {
		builder.definition.Retry = retry
	}
	return builder
}

// WithTimeout sets the agent's execution timeout.
func (builder *Builder) WithTimeout(timeout time.Duration) *Builder {
	if builder != nil {
		if timeout <= 0 {
			builder.err = errors.New("agents: timeout must be positive")
		} else {
			builder.err = nil
			builder.definition.Timeout = timeout.String()
		}
	}
	return builder
}

// WithRateLimit sets the agent's execution limits.
func (builder *Builder) WithRateLimit(rateLimit RateLimit) *Builder {
	if builder != nil {
		builder.definition.RateLimit = &rateLimit
	}
	return builder
}

// WithTags sets the agent's tags.
func (builder *Builder) WithTags(tags ...string) *Builder {
	if builder != nil {
		builder.definition.Tags = append([]string(nil), tags...)
	}
	return builder
}

// Build returns the configured definition after validating required fields.
func (builder *Builder) Build() (Definition, error) {
	if builder == nil {
		return Definition{}, errors.New("agents: builder is nil")
	}
	if builder.err != nil {
		return Definition{}, builder.err
	}
	definition := defaultDefinition(builder.definition)
	if err := validateDefinition(definition); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func validateDefinition(definition Definition) error {
	if definition.Name == "" {
		return errors.New("agents: agent name is required")
	}
	if len(definition.Triggers) == 0 {
		return errors.New("agents: at least one trigger is required")
	}
	for _, trigger := range definition.Triggers {
		if trigger.Event == "" {
			return errors.New("agents: trigger event is required")
		}
	}
	for _, threshold := range []*float64{definition.Confidence.AutoExecuteAbove, definition.Confidence.EscalateBelow} {
		if threshold != nil && (math.IsNaN(*threshold) || math.IsInf(*threshold, 0) || *threshold < 0 || *threshold > 1) {
			return errors.New("agents: confidence thresholds must be between 0 and 1")
		}
	}
	switch definition.Memory.Type {
	case "working", "persistent", "episodic":
	default:
		return errors.New("agents: memory type must be working, persistent, or episodic")
	}
	if definition.Memory.MaxTokens < 0 {
		return errors.New("agents: memory max tokens must be positive")
	}
	if definition.Retry.MaxRetries != nil && *definition.Retry.MaxRetries < 0 {
		return errors.New("agents: retry max retries must not be negative")
	}
	if definition.Retry.RetryDelay != nil && *definition.Retry.RetryDelay <= 0 {
		return errors.New("agents: retry delay must be positive")
	}
	if definition.Retry.Backoff != nil {
		switch *definition.Retry.Backoff {
		case "constant", "linear", "exponential":
		default:
			return errors.New("agents: retry backoff must be constant, linear, or exponential")
		}
	}
	if definition.RateLimit != nil && (definition.RateLimit.MaxExecutionsPerMinute < 0 || definition.RateLimit.MaxConcurrent < 0) {
		return errors.New("agents: rate limits must be positive")
	}
	return nil
}

func defaultDefinition(definition Definition) Definition {
	if definition.Triggers == nil {
		definition.Triggers = []Trigger{}
	}
	if definition.Scope.Read == nil {
		definition.Scope.Read = []string{}
	}
	if definition.Scope.Write == nil {
		definition.Scope.Write = []string{}
	}
	if definition.Scope.Actions == nil {
		definition.Scope.Actions = []string{}
	}
	if definition.Scope.Escalate == nil {
		definition.Scope.Escalate = []string{}
	}
	if definition.Scope.InvokeAgents == nil {
		definition.Scope.InvokeAgents = []string{}
	}
	if definition.Scope.InvokeFunctions == nil {
		definition.Scope.InvokeFunctions = []string{}
	}
	if definition.Confidence.AutoExecuteAbove == nil {
		value := 0.85
		definition.Confidence.AutoExecuteAbove = &value
	}
	if definition.Confidence.EscalateBelow == nil {
		value := 0.6
		definition.Confidence.EscalateBelow = &value
	}
	if definition.Confidence.RequireReviewBetween == nil {
		value := true
		definition.Confidence.RequireReviewBetween = &value
	}
	if definition.Memory.Type == "" {
		definition.Memory.Type = "working"
	}
	if definition.Retry.MaxRetries == nil {
		value := 3
		definition.Retry.MaxRetries = &value
	}
	if definition.Retry.RetryDelay == nil {
		value := 1000
		definition.Retry.RetryDelay = &value
	}
	if definition.Retry.Backoff == nil {
		value := "exponential"
		definition.Retry.Backoff = &value
	}
	if definition.Retry.RetryOn == nil {
		value := []int{408, 409, 425, 429, 500, 502, 503, 504}
		definition.Retry.RetryOn = &value
	}
	if definition.Timeout == "" {
		definition.Timeout = "30s"
	}
	if definition.Tags == nil {
		definition.Tags = []string{}
	}
	return definition
}

// Create builds the definition and creates the agent through its attached client.
func (builder *Builder) Create(ctx context.Context) (Agent, error) {
	definition, err := builder.Build()
	if err != nil {
		return Agent{}, err
	}
	if builder.client == nil {
		return Agent{}, errors.New("agents: client is nil")
	}
	return builder.client.Create(ctx, definition)
}

// Create creates an agent from a typed definition.
func (client *Client) Create(ctx context.Context, definition Definition) (Agent, error) {
	var agent Agent
	definition = defaultDefinition(definition)
	if err := validateDefinition(definition); err != nil {
		return agent, err
	}
	endpoint, _ := client.Endpoint("POST", "/agents")
	err := client.Call(ctx, Request{Endpoint: endpoint, Body: definition}, &agent)
	return agent, err
}

// List returns one page of agents.
func (client *Client) List(ctx context.Context, options ListOptions) (Page, error) {
	var page Page
	query := make(url.Values)
	if options.Status != "" {
		query.Set("status", string(options.Status))
	}
	if options.Trigger != "" {
		query.Set("trigger", options.Trigger)
	}
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.Cursor != "" {
		query.Set("cursor", options.Cursor)
	}
	endpoint, _ := client.Endpoint("GET", "/agents")
	err := client.Call(ctx, Request{Endpoint: endpoint, Query: query}, &page)
	return page, err
}

// Get fetches an agent by ID.
func (client *Client) Get(ctx context.Context, id string) (Agent, error) {
	var agent Agent
	if id == "" {
		return agent, errors.New("agents: id is required")
	}
	endpoint, _ := client.Endpoint("GET", "/agents/{param}")
	err := client.Call(ctx, Request{Endpoint: endpoint, PathParams: []string{id}}, &agent)
	return agent, err
}

// Update changes selected fields on an agent by ID.
func (client *Client) Update(ctx context.Context, id string, definition UpdateDefinition) (Agent, error) {
	var agent Agent
	if id == "" {
		return agent, errors.New("agents: id is required")
	}
	endpoint, _ := client.Endpoint("PUT", "/agents/{param}")
	err := client.Call(ctx, Request{Endpoint: endpoint, PathParams: []string{id}, Body: definition}, &agent)
	return agent, err
}

// Delete removes an agent by ID.
func (client *Client) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("agents: id is required")
	}
	endpoint, _ := client.Endpoint("DELETE", "/agents/{param}")
	return client.Call(ctx, Request{Endpoint: endpoint, PathParams: []string{id}}, nil)
}

// Rollback returns an agent to a previous version.
func (client *Client) Rollback(ctx context.Context, id string, toVersion int) (Agent, error) {
	var agent Agent
	if id == "" {
		return agent, errors.New("agents: id is required")
	}
	body := struct {
		ToVersion int `json:"toVersion,omitempty"`
	}{ToVersion: toVersion}
	endpoint, _ := client.Endpoint("POST", "/agents/{param}/rollback")
	err := client.Call(ctx, Request{Endpoint: endpoint, PathParams: []string{id}, Body: body}, &agent)
	return agent, err
}

// Health returns the Agents service health status.
func (client *Client) Health(ctx context.Context) (Health, error) {
	var health Health
	endpoint, _ := client.Endpoint("GET", "/agents/health")
	err := client.Call(ctx, Request{Endpoint: endpoint}, &health)
	return health, err
}

// ListRuns returns one page of an agent's executions.
func (client *Client) ListRuns(ctx context.Context, agentID string, options RunListOptions) (ExecutionPage, error) {
	var page ExecutionPage
	if agentID == "" {
		return page, errors.New("agents: id is required")
	}
	query := make(url.Values)
	if options.Status != "" {
		query.Set("status", options.Status)
	}
	if options.From != "" {
		query.Set("from", options.From)
	}
	if options.To != "" {
		query.Set("to", options.To)
	}
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.Cursor != "" {
		query.Set("cursor", options.Cursor)
	}
	endpoint, _ := client.Endpoint("GET", "/agents/{param}/runs")
	err := client.Call(ctx, Request{Endpoint: endpoint, PathParams: []string{agentID}, Query: query}, &page)
	return page, err
}

// StartRun sends an event and payload to an agent.
func (client *Client) StartRun(ctx context.Context, agentID, event string, payload map[string]any) (Execution, error) {
	var execution Execution
	if agentID == "" || event == "" {
		return execution, errors.New("agents: id and event are required")
	}
	body := struct {
		Event   string         `json:"event"`
		Payload map[string]any `json:"payload"`
	}{Event: event, Payload: payload}
	endpoint, _ := client.Endpoint("POST", "/agents/{param}/runs")
	err := client.Call(ctx, Request{Endpoint: endpoint, PathParams: []string{agentID}, Body: body}, &execution)
	return execution, err
}

// GetRun fetches an execution by ID.
func (client *Client) GetRun(ctx context.Context, runID string) (Execution, error) {
	var execution Execution
	if runID == "" {
		return execution, errors.New("agents: run id is required")
	}
	endpoint, _ := client.Endpoint("GET", "/agents/runs/{param}")
	err := client.Call(ctx, Request{Endpoint: endpoint, PathParams: []string{runID}}, &execution)
	return execution, err
}

// GetConversation fetches the conversation transcript for an execution.
func (client *Client) GetConversation(ctx context.Context, runID string) (Conversation, error) {
	var conversation Conversation
	if runID == "" {
		return conversation, errors.New("agents: run id is required")
	}
	endpoint, _ := client.Endpoint("GET", "/agents/runs/{param}/conversation")
	err := client.Call(ctx, Request{Endpoint: endpoint, PathParams: []string{runID}}, &conversation)
	return conversation, err
}

// ListVersions returns version records for an agent.
func (client *Client) ListVersions(ctx context.Context, agentID string, options ListOptions) (VersionPage, error) {
	var page VersionPage
	if agentID == "" {
		return page, errors.New("agents: id is required")
	}
	query := make(url.Values)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.Cursor != "" {
		query.Set("cursor", options.Cursor)
	}
	endpoint, _ := client.Endpoint("GET", "/agents/{param}/versions")
	err := client.Call(ctx, Request{Endpoint: endpoint, PathParams: []string{agentID}, Query: query}, &page)
	return page, err
}
