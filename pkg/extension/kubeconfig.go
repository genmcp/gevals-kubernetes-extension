package extension

import (
	"context"
	"fmt"

	"github.com/mcpchecker/mcpchecker/pkg/extension/sdk"
)

// ContextInfo represents information about a Kubernetes context
type ContextInfo struct {
	Name      string `json:"name"`
	Cluster   string `json:"cluster"`
	User      string `json:"user"`
	Namespace string `json:"namespace,omitempty"`
	IsCurrent bool   `json:"isCurrent"`
}

// handleListContexts lists all contexts from the kubeconfig file.
// Returns the current context name and total count.
func (e *Extension) handleListContexts(ctx context.Context, req *sdk.OperationRequest) (*sdk.OperationResult, error) {
	if e.client == nil {
		return sdk.Failure(fmt.Errorf("kubernetes client not initialized")), nil
	}

	e.LogInfo(ctx, "Listing kubeconfig contexts", nil)

	contexts, err := e.client.ListContexts(ctx)
	if err != nil {
		e.LogError(ctx, "Failed to list contexts", map[string]any{
			"error": err.Error(),
		})
		return sdk.Failure(fmt.Errorf("failed to list contexts: %w", err)), nil
	}

	if len(contexts) == 0 {
		return sdk.Failure(fmt.Errorf("no contexts found in kubeconfig")), nil
	}

	// Find current context
	var currentContext string
	for _, c := range contexts {
		if c.IsCurrent {
			currentContext = c.Name
			break
		}
	}

	e.LogInfo(ctx, "Contexts listed successfully", map[string]any{
		"count":   len(contexts),
		"current": currentContext,
	})

	return sdk.SuccessWithOutputs(
		fmt.Sprintf("Found %d context(s), current: %s", len(contexts), currentContext),
		map[string]string{
			"current": currentContext,
			"count":   fmt.Sprintf("%d", len(contexts)),
		},
	), nil
}

// handleGetCurrentContext returns the current context name from the kubeconfig.
// Fails if no current context is set.
func (e *Extension) handleGetCurrentContext(ctx context.Context, req *sdk.OperationRequest) (*sdk.OperationResult, error) {
	if e.client == nil {
		return sdk.Failure(fmt.Errorf("kubernetes client not initialized")), nil
	}

	e.LogInfo(ctx, "Getting current kubeconfig context", nil)

	currentContext, err := e.client.GetCurrentContext(ctx)
	if err != nil {
		e.LogError(ctx, "Failed to get current context", map[string]any{
			"error": err.Error(),
		})
		return sdk.Failure(fmt.Errorf("failed to get current context: %w", err)), nil
	}

	if currentContext == "" {
		return sdk.Failure(fmt.Errorf("no current context set in kubeconfig")), nil
	}

	e.LogInfo(ctx, "Current context retrieved", map[string]any{
		"context": currentContext,
	})

	return sdk.SuccessWithOutputs(
		fmt.Sprintf("Current context: %s", currentContext),
		map[string]string{
			"context": currentContext,
		},
	), nil
}

// handleViewConfig returns the kubeconfig as YAML.
// When minify is true, returns only the current context and its dependencies.
func (e *Extension) handleViewConfig(ctx context.Context, req *sdk.OperationRequest) (*sdk.OperationResult, error) {
	if e.client == nil {
		return sdk.Failure(fmt.Errorf("kubernetes client not initialized")), nil
	}

	// Parse args
	args, ok := req.Args.(map[string]any)
	if !ok {
		args = make(map[string]any)
	}

	minify := false
	if m, ok := args["minify"].(bool); ok {
		minify = m
	}

	e.LogInfo(ctx, "Viewing kubeconfig", map[string]any{
		"minify": minify,
	})

	configYAML, err := e.client.ViewConfig(ctx, minify)
	if err != nil {
		e.LogError(ctx, "Failed to view config", map[string]any{
			"error": err.Error(),
		})
		return sdk.Failure(fmt.Errorf("failed to view config: %w", err)), nil
	}

	e.LogInfo(ctx, "Kubeconfig retrieved successfully", nil)

	return sdk.SuccessWithOutputs(
		"Kubeconfig retrieved",
		map[string]string{
			"config": configYAML,
		},
	), nil
}
