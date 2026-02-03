package extension

import (
	"context"
	"fmt"

	"github.com/mcpchecker/mcpchecker/pkg/extension/sdk"
)

func (e *Extension) handleAuthCanI(ctx context.Context, req *sdk.OperationRequest) (*sdk.OperationResult, error) {
	if e.client == nil {
		return sdk.Failure(fmt.Errorf("kubernetes client not initialized")), nil
	}

	args, ok := req.Args.(map[string]any)
	if !ok {
		return sdk.Failure(fmt.Errorf("args must be an object")), nil
	}

	verb, _ := args["verb"].(string)
	resource, _ := args["resource"].(string)
	as, _ := args["as"].(string)
	namespace, _ := args["namespace"].(string)
	apiGroup, _ := args["apiGroup"].(string)
	resourceName, _ := args["resourceName"].(string)

	if verb == "" {
		return sdk.Failure(fmt.Errorf("verb is required")), nil
	}
	if resource == "" {
		return sdk.Failure(fmt.Errorf("resource is required")), nil
	}
	if as == "" {
		return sdk.Failure(fmt.Errorf("as is required")), nil
	}

	e.LogInfo(ctx, "Checking permissions", map[string]any{
		"verb":         verb,
		"resource":     resource,
		"as":           as,
		"namespace":    namespace,
		"apiGroup":     apiGroup,
		"resourceName": resourceName,
	})

	allowed, reason, err := e.client.CheckAccess(ctx, as, verb, resource, apiGroup, namespace, resourceName)
	if err != nil {
		e.LogError(ctx, "Failed to check permissions", map[string]any{
			"error": err.Error(),
		})
		return sdk.Failure(fmt.Errorf("failed to check permissions: %w", err)), nil
	}

	e.LogInfo(ctx, "Permission check completed", map[string]any{
		"allowed": allowed,
		"reason":  reason,
	})

	// Handle expect.allowed verification
	if expectArg, hasExpect := args["expect"]; hasExpect {
		expect, ok := expectArg.(map[string]any)
		if !ok {
			return sdk.Failure(fmt.Errorf("expect must be an object")), nil
		}

		if expectedAllowed, hasAllowed := expect["allowed"]; hasAllowed {
			expectedBool, ok := expectedAllowed.(bool)
			if !ok {
				return sdk.Failure(fmt.Errorf("expect.allowed must be a boolean")), nil
			}

			if allowed != expectedBool {
				return sdk.FailureWithMessage(
					fmt.Sprintf("permission check failed: expected allowed=%v but got allowed=%v", expectedBool, allowed),
					fmt.Errorf("permission expectation not met"),
				), nil
			}
		}
	}

	msg := fmt.Sprintf("%s can %s %s", as, verb, resource)
	if namespace != "" {
		msg += fmt.Sprintf(" in namespace %s", namespace)
	} else {
		msg += " cluster-wide"
	}

	if allowed {
		return sdk.SuccessWithOutputs(msg+": allowed", map[string]string{
			"allowed": "true",
			"reason":  reason,
		}), nil
	}

	return sdk.SuccessWithOutputs(msg+": denied", map[string]string{
		"allowed": "false",
		"reason":  reason,
	}), nil
}
