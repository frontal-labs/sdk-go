package functions

import (
	"context"
	"errors"
	"fmt"
)

// Builder incrementally defines a function before creating it.
type Builder struct {
	client     *Client
	definition FunctionDefinition
}

// Define starts building a function definition attached to client.
func (client *Client) Define(name string) *Builder {
	return &Builder{client: client, definition: FunctionDefinition{Name: name}}
}

// Description sets the function description.
func (builder *Builder) Description(description string) *Builder {
	if builder != nil {
		builder.definition.Description = description
	}
	return builder
}

// Runtime sets the function runtime.
func (builder *Builder) Runtime(runtime Runtime) *Builder {
	if builder != nil {
		builder.definition.Runtime = runtime
	}
	return builder
}

// Entrypoint sets the function entrypoint.
func (builder *Builder) Entrypoint(entrypoint string) *Builder {
	if builder != nil {
		builder.definition.Entrypoint = entrypoint
	}
	return builder
}

// Source sets the function source reference.
func (builder *Builder) Source(source string) *Builder {
	if builder != nil {
		builder.definition.Source = source
	}
	return builder
}

// InputSchema sets the function input JSON Schema.
func (builder *Builder) InputSchema(schema map[string]any) *Builder {
	if builder != nil {
		builder.definition.InputSchema = schema
	}
	return builder
}

// OutputSchema sets the function output JSON Schema.
func (builder *Builder) OutputSchema(schema map[string]any) *Builder {
	if builder != nil {
		builder.definition.OutputSchema = schema
	}
	return builder
}

// Dependencies sets the function's dependencies.
func (builder *Builder) Dependencies(dependencies []string) *Builder {
	if builder != nil {
		builder.definition.Dependencies = dependencies
	}
	return builder
}

// EnvVars sets the function's environment variables.
func (builder *Builder) EnvVars(envVars map[string]string) *Builder {
	if builder != nil {
		builder.definition.EnvVars = envVars
	}
	return builder
}

// Secrets sets the secret names available to the function.
func (builder *Builder) Secrets(secrets []string) *Builder {
	if builder != nil {
		builder.definition.Secrets = secrets
	}
	return builder
}

// Memory sets the function memory limit in megabytes.
func (builder *Builder) Memory(memory int) *Builder {
	if builder != nil {
		builder.definition.Memory = memory
	}
	return builder
}

// Timeout sets the function execution timeout in seconds.
func (builder *Builder) Timeout(timeout int) *Builder {
	if builder != nil {
		builder.definition.Timeout = timeout
	}
	return builder
}

// Permissions sets the function's ontology and action permissions.
func (builder *Builder) Permissions(permissions Permission) *Builder {
	if builder != nil {
		builder.definition.Permissions = &permissions
	}
	return builder
}

// Build validates and returns the function definition.
func (builder *Builder) Build() (FunctionDefinition, error) {
	if builder == nil {
		return FunctionDefinition{}, errors.New("functions: builder is nil")
	}
	if err := validateDefinition(builder.definition); err != nil {
		return FunctionDefinition{}, err
	}
	return builder.definition, nil
}

// Create validates and creates the function through the attached client.
func (builder *Builder) Create(ctx context.Context) (FunctionResource, error) {
	definition, err := builder.Build()
	if err != nil {
		return FunctionResource{}, fmt.Errorf("functions: build function definition: %w", err)
	}
	if builder.client == nil {
		return FunctionResource{}, errors.New("functions: client is nil")
	}
	created, err := builder.client.Create(ctx, definition)
	if err != nil {
		return FunctionResource{}, fmt.Errorf("functions: create from builder: %w", err)
	}
	return created, nil
}
