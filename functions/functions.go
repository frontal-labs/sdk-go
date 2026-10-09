// Package functions provides typed operations for the Frontal Functions API.
package functions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	functionsPath        = "/functions"
	functionPath         = "/functions/{param}"
	functionVersionsPath = "/functions/{param}/versions"
	functionVersionPath  = "/functions/{param}/versions/{param}"
	executionPath        = "/functions/executions/{param}"
	executionsPath       = "/functions/executions"
)

func (client *Client) callOperation(
	ctx context.Context,
	method string,
	path string,
	pathParams []string,
	query url.Values,
	body any,
	out any,
) error {
	endpoint, ok := client.Endpoint(method, path)
	if !ok {
		return fmt.Errorf("functions: endpoint is missing from the generated inventory: %s %s", method, path)
	}
	if err := client.Call(ctx, Request{
		Endpoint: endpoint, PathParams: pathParams, Query: query, Body: body,
	}, out); err != nil {
		return fmt.Errorf("functions: %s %s: %w", method, path, err)
	}
	return nil
}

// Create creates a function from a complete definition.
func (client *Client) Create(ctx context.Context, definition FunctionDefinition) (FunctionResource, error) {
	var function FunctionResource
	if err := validateDefinition(definition); err != nil {
		return function, err
	}
	if err := client.callOperation(ctx, http.MethodPost, functionsPath, nil, nil, definition, &function); err != nil {
		return FunctionResource{}, err
	}
	return function, nil
}

// List returns one page of functions.
func (client *Client) List(ctx context.Context, options ListOptions) (FunctionListResponse, error) {
	var response FunctionListResponse
	if err := client.callOperation(ctx, http.MethodGet, functionsPath, nil, paginationQuery(options), nil, &response); err != nil {
		return FunctionListResponse{}, err
	}
	return response, nil
}

// Get fetches a function by ID.
func (client *Client) Get(ctx context.Context, functionID string) (FunctionResource, error) {
	var function FunctionResource
	if err := requireID("function", functionID); err != nil {
		return function, err
	}
	if err := client.callOperation(ctx, http.MethodGet, functionPath, []string{functionID}, nil, nil, &function); err != nil {
		return FunctionResource{}, err
	}
	return function, nil
}

// Update applies a complete function definition as a new version.
func (client *Client) Update(ctx context.Context, functionID string, definition FunctionDefinition) (FunctionResource, error) {
	var function FunctionResource
	if err := requireID("function", functionID); err != nil {
		return function, err
	}
	if err := validateDefinition(definition); err != nil {
		return function, err
	}
	if err := client.callOperation(ctx, http.MethodPatch, functionPath, []string{functionID}, nil, definition, &function); err != nil {
		return FunctionResource{}, err
	}
	return function, nil
}

// Delete removes a function by ID.
func (client *Client) Delete(ctx context.Context, functionID string) error {
	if err := requireID("function", functionID); err != nil {
		return err
	}
	return client.callOperation(ctx, http.MethodDelete, functionPath, []string{functionID}, nil, nil, nil)
}

// ListVersions returns one page of versions for a function.
func (client *Client) ListVersions(ctx context.Context, functionID string, options ListOptions) (FunctionVersionListResponse, error) {
	var response FunctionVersionListResponse
	if err := requireID("function", functionID); err != nil {
		return response, err
	}
	if err := client.callOperation(ctx, http.MethodGet, functionVersionsPath, []string{functionID}, paginationQuery(options), nil, &response); err != nil {
		return FunctionVersionListResponse{}, err
	}
	return response, nil
}

// GetVersion fetches one version of a function.
func (client *Client) GetVersion(ctx context.Context, functionID string, version int) (FunctionVersion, error) {
	var result FunctionVersion
	if err := validateFunctionVersion(functionID, version); err != nil {
		return result, err
	}
	if err := client.callOperation(ctx, http.MethodGet, functionVersionPath, versionPathParams(functionID, version), nil, nil, &result); err != nil {
		return FunctionVersion{}, err
	}
	return result, nil
}

// PublishVersion publishes one version of a function.
func (client *Client) PublishVersion(ctx context.Context, functionID string, version int) (FunctionVersion, error) {
	var result FunctionVersion
	if err := validateFunctionVersion(functionID, version); err != nil {
		return result, err
	}
	if err := client.callOperation(ctx, http.MethodPost, functionVersionPath+"/publish", versionPathParams(functionID, version), nil, struct{}{}, &result); err != nil {
		return FunctionVersion{}, err
	}
	return result, nil
}

// DeployVersion deploys one version of a function.
func (client *Client) DeployVersion(ctx context.Context, functionID string, version int) error {
	if err := validateFunctionVersion(functionID, version); err != nil {
		return err
	}
	return client.callOperation(ctx, http.MethodPost, functionVersionPath+"/deploy", versionPathParams(functionID, version), nil, struct{}{}, nil)
}

// GetDeploymentStatus fetches deployment status for one function version.
func (client *Client) GetDeploymentStatus(ctx context.Context, functionID string, version int) (DeploymentStatus, error) {
	var result DeploymentStatus
	if err := validateFunctionVersion(functionID, version); err != nil {
		return result, err
	}
	path := functionVersionPath + "/deployment/status"
	if err := client.callOperation(ctx, http.MethodGet, path, versionPathParams(functionID, version), nil, nil, &result); err != nil {
		return DeploymentStatus{}, err
	}
	return result, nil
}

// Invoke invokes a function synchronously.
func (client *Client) Invoke(ctx context.Context, input FunctionInvocationInput) (FunctionInvocationResult, error) {
	var result FunctionInvocationResult
	if err := requireID("function", input.FunctionID); err != nil {
		return result, err
	}
	if input.Version < 0 {
		return result, errors.New("functions: version cannot be negative")
	}
	if err := client.callOperation(ctx, http.MethodPost, functionsPath+"/invoke", nil, nil, input, &result); err != nil {
		return FunctionInvocationResult{}, err
	}
	return result, nil
}

// InvokeAsync starts a function execution and returns its execution ID.
func (client *Client) InvokeAsync(ctx context.Context, input FunctionInvocationInput) (AsyncInvocation, error) {
	var result AsyncInvocation
	if err := requireID("function", input.FunctionID); err != nil {
		return result, err
	}
	if input.Version < 0 {
		return result, errors.New("functions: version cannot be negative")
	}
	if err := client.callOperation(ctx, http.MethodPost, functionsPath+"/invoke-async", nil, nil, input, &result); err != nil {
		return AsyncInvocation{}, err
	}
	return result, nil
}

// GetExecution fetches the status and details of an execution.
func (client *Client) GetExecution(ctx context.Context, executionID string) (FunctionExecution, error) {
	var result FunctionExecution
	if err := requireID("execution", executionID); err != nil {
		return result, err
	}
	if err := client.callOperation(ctx, http.MethodGet, executionPath, []string{executionID}, nil, nil, &result); err != nil {
		return FunctionExecution{}, err
	}
	return result, nil
}

// GetExecutionResult fetches the result of a completed execution.
func (client *Client) GetExecutionResult(ctx context.Context, executionID string) (FunctionInvocationResult, error) {
	var result FunctionInvocationResult
	if err := requireID("execution", executionID); err != nil {
		return result, err
	}
	if err := client.callOperation(ctx, http.MethodGet, executionPath+"/result", []string{executionID}, nil, nil, &result); err != nil {
		return FunctionInvocationResult{}, err
	}
	return result, nil
}

// ListExecutions returns one page of executions, optionally filtered by function and status.
func (client *Client) ListExecutions(ctx context.Context, options ExecutionListOptions) (FunctionExecutionListResponse, error) {
	var response FunctionExecutionListResponse
	query := make(url.Values)
	if options.Cursor != "" {
		query.Set("cursor", options.Cursor)
	}
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.FunctionID != "" {
		query.Set("functionId", options.FunctionID)
	}
	if options.Status != "" {
		query.Set("status", string(options.Status))
	}
	if err := client.callOperation(ctx, http.MethodGet, executionsPath, nil, query, nil, &response); err != nil {
		return FunctionExecutionListResponse{}, err
	}
	return response, nil
}

// CancelExecution cancels an execution by ID.
func (client *Client) CancelExecution(ctx context.Context, executionID string) error {
	if err := requireID("execution", executionID); err != nil {
		return err
	}
	return client.callOperation(ctx, http.MethodPost, executionPath+"/cancel", []string{executionID}, nil, struct{}{}, nil)
}

func paginationQuery(options ListOptions) url.Values {
	query := make(url.Values)
	if options.Cursor != "" {
		query.Set("cursor", options.Cursor)
	}
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	return query
}

func versionPathParams(functionID string, version int) []string {
	return []string{functionID, strconv.Itoa(version)}
}

func requireID(kind, value string) error {
	if value == "" {
		return fmt.Errorf("functions: %s ID is required", kind)
	}
	return nil
}

func validateFunctionVersion(functionID string, version int) error {
	if err := requireID("function", functionID); err != nil {
		return err
	}
	if version <= 0 {
		return errors.New("functions: version must be positive")
	}
	return nil
}

func validateDefinition(definition FunctionDefinition) error {
	if definition.Name == "" {
		return errors.New("functions: function name is required")
	}
	if definition.Runtime != RuntimeNodeJS20 && definition.Runtime != RuntimeNodeJS22 && definition.Runtime != RuntimePython311 {
		return fmt.Errorf("functions: unsupported runtime %q", definition.Runtime)
	}
	if definition.Entrypoint == "" {
		return errors.New("functions: entrypoint is required")
	}
	if definition.Memory < 0 {
		return errors.New("functions: memory cannot be negative")
	}
	if definition.Timeout < 0 {
		return errors.New("functions: timeout cannot be negative")
	}
	return nil
}
