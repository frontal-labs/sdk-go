package functions

import "time"

// Runtime identifies a supported Frontal Functions runtime.
type Runtime string

const (
	// RuntimeNodeJS20 is the Node.js 20 runtime.
	RuntimeNodeJS20 Runtime = "nodejs20"
	// RuntimeNodeJS22 is the Node.js 22 runtime.
	RuntimeNodeJS22 Runtime = "nodejs22"
	// RuntimePython311 is the Python 3.11 runtime.
	RuntimePython311 Runtime = "python311"
)

// Status identifies a function or execution status.
type Status string

const (
	// StatusDraft identifies a draft function.
	StatusDraft Status = "draft"
	// StatusActive identifies an active function.
	StatusActive Status = "active"
	// StatusDeprecated identifies a deprecated function.
	StatusDeprecated Status = "deprecated"
	// StatusFailed identifies a failed function or execution.
	StatusFailed Status = "failed"
)

// Permission describes the Frontal resources and actions a function may use.
type Permission struct {
	Ontology []string `json:"ontology,omitempty"`
	Actions  []string `json:"actions,omitempty"`
}

// FunctionDefinition describes a function that can be created or updated.
type FunctionDefinition struct {
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Runtime      Runtime           `json:"runtime"`
	Entrypoint   string            `json:"entrypoint"`
	Source       string            `json:"source,omitempty"`
	InputSchema  map[string]any    `json:"inputSchema,omitempty"`
	OutputSchema map[string]any    `json:"outputSchema,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	EnvVars      map[string]string `json:"envVars,omitempty"`
	Secrets      []string          `json:"secrets,omitempty"`
	Memory       int               `json:"memory,omitempty"`
	Timeout      int               `json:"timeout,omitempty"`
	Permissions  *Permission       `json:"permissions,omitempty"`
}

// FunctionResource is a function returned by the API.
type FunctionResource struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Description   string      `json:"description,omitempty"`
	Runtime       Runtime     `json:"runtime"`
	Entrypoint    string      `json:"entrypoint"`
	Status        Status      `json:"status"`
	Version       int         `json:"version"`
	LatestVersion *int        `json:"latestVersion,omitempty"`
	Memory        *int        `json:"memory,omitempty"`
	Timeout       *int        `json:"timeout,omitempty"`
	Permissions   *Permission `json:"permissions,omitempty"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

// FunctionVersion describes a version of a function.
type FunctionVersion struct {
	Version     int       `json:"version"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"createdAt"`
	PublishedBy string    `json:"publishedBy,omitempty"`
}

// FunctionExecution describes the state and output of a function execution.
type FunctionExecution struct {
	ID          string         `json:"id"`
	FunctionID  string         `json:"functionId"`
	Version     int            `json:"version"`
	Status      Status         `json:"status"`
	Input       map[string]any `json:"input,omitempty"`
	Output      map[string]any `json:"output,omitempty"`
	Error       string         `json:"error,omitempty"`
	StartedAt   *time.Time     `json:"startedAt,omitempty"`
	CompletedAt *time.Time     `json:"completedAt,omitempty"`
	DurationMS  *int64         `json:"durationMs,omitempty"`
}

// FunctionInvocationInput contains a function ID, optional version, and JSON input.
type FunctionInvocationInput struct {
	FunctionID string         `json:"functionId"`
	Version    int            `json:"version,omitempty"`
	Input      map[string]any `json:"input,omitempty"`
}

// FunctionInvocationResult contains the result or error from an invocation.
type FunctionInvocationResult struct {
	ExecutionID string         `json:"executionId"`
	Result      map[string]any `json:"result,omitempty"`
	Error       string         `json:"error,omitempty"`
	Status      *Status        `json:"status,omitempty"`
}

// AsyncInvocation identifies a function execution started asynchronously.
type AsyncInvocation struct {
	ExecutionID string `json:"executionId"`
}

// DeploymentStatus describes a version deployment and any additional details.
type DeploymentStatus struct {
	Status  string `json:"status"`
	Details any    `json:"details,omitempty"`
}

// Pagination contains cursor metadata returned by a list operation.
type Pagination struct {
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"hasMore"`
}

// FunctionListResponse contains one page of functions.
type FunctionListResponse struct {
	Functions  []FunctionResource `json:"functions"`
	Pagination Pagination         `json:"pagination"`
}

// FunctionVersionListResponse contains one page of function versions.
type FunctionVersionListResponse struct {
	Versions   []FunctionVersion `json:"versions"`
	Pagination Pagination        `json:"pagination"`
}

// FunctionExecutionListResponse contains one page of function executions.
type FunctionExecutionListResponse struct {
	Executions []FunctionExecution `json:"executions"`
	Pagination Pagination          `json:"pagination"`
}

// ListOptions contains cursor pagination parameters.
type ListOptions struct {
	Cursor string
	Limit  int
}

// ExecutionListOptions contains pagination and execution filters.
type ExecutionListOptions struct {
	Cursor     string
	Limit      int
	FunctionID string
	Status     Status
}
