package extension

import (
	"context"
	"fmt"
	"time"

	"github.com/mcpchecker/mcpchecker/pkg/extension/sdk"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (e *Extension) handleWait(ctx context.Context, req *sdk.OperationRequest) (*sdk.OperationResult, error) {
	if e.client == nil {
		return sdk.Failure(fmt.Errorf("kubernetes client not initialized")), nil
	}

	args, ok := req.Args.(map[string]any)
	if !ok {
		return sdk.Failure(fmt.Errorf("args must be an object")), nil
	}

	ref, err := parseResourceRef(args)
	if err != nil {
		return sdk.Failure(err), nil
	}

	condition, _ := args["condition"].(string)
	if condition == "" {
		return sdk.Failure(fmt.Errorf("condition is required")), nil
	}

	status, _ := args["status"].(string)
	if status == "" {
		status = "True"
	}

	timeoutStr, _ := args["timeout"].(string)
	if timeoutStr == "" {
		timeoutStr = "60s"
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return sdk.Failure(fmt.Errorf("invalid timeout format: %w", err)), nil
	}

	gvr, err := ref.gvr()
	if err != nil {
		return sdk.Failure(err), nil
	}

	e.LogInfo(ctx, "Waiting for condition", map[string]any{
		"kind":      ref.kind,
		"name":      ref.name,
		"namespace": ref.namespace,
		"condition": condition,
		"status":    status,
		"timeout":   timeoutStr,
	})

	var lastStatus string
	err = wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		obj, getErr := e.client.Get(ctx, gvr, ref.name, ref.namespace)
		if getErr != nil {
			return false, nil // Keep polling on transient errors
		}

		conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
		if err != nil || !found {
			lastStatus = "NoConditions"
			return false, nil
		}

		for _, c := range conditions {
			cond, ok := c.(map[string]any)
			if !ok {
				continue
			}

			condType, _, _ := unstructured.NestedString(cond, "type")
			condStatus, _, _ := unstructured.NestedString(cond, "status")

			if condType == condition {
				lastStatus = condStatus
				return condStatus == status, nil
			}
		}

		lastStatus = "ConditionNotFound"
		return false, nil
	})

	if err != nil {
		e.LogError(ctx, "Condition wait timed out", map[string]any{
			"kind":       ref.kind,
			"name":       ref.name,
			"condition":  condition,
			"lastStatus": lastStatus,
		})
		return sdk.FailureWithMessage(
			fmt.Sprintf("Condition %s=%s not met", condition, status),
			fmt.Errorf("timed out waiting for %s/%s: last status was %s", ref.kind, ref.name, lastStatus),
		), nil
	}

	e.LogInfo(ctx, "Condition met", map[string]any{
		"kind":      ref.kind,
		"name":      ref.name,
		"condition": condition,
		"status":    status,
	})

	return sdk.Success(fmt.Sprintf("%s/%s condition %s=%s", ref.kind, ref.name, condition, status)), nil
}
