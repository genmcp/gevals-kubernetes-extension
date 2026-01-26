package extension

import (
	"context"
	"fmt"

	"github.com/mcpchecker/mcpchecker/pkg/extension/sdk"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *Extension) handleDelete(ctx context.Context, req *sdk.OperationRequest) (*sdk.OperationResult, error) {
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

	ignoreNotFound, _ := args["ignoreNotFound"].(bool)

	gvr, err := ref.gvr()
	if err != nil {
		return sdk.Failure(err), nil
	}

	e.LogInfo(ctx, "Deleting resource", map[string]any{
		"kind":           ref.kind,
		"name":           ref.name,
		"namespace":      ref.namespace,
		"ignoreNotFound": ignoreNotFound,
	})

	propagation := metav1.DeletePropagationForeground
	deleteOpts := metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	}

	err = e.client.Delete(ctx, gvr, ref.name, ref.namespace, deleteOpts)
	if err != nil {
		if ignoreNotFound && apierrors.IsNotFound(err) {
			e.LogInfo(ctx, "Resource not found (ignored)", map[string]any{
				"kind": ref.kind,
				"name": ref.name,
			})
			return sdk.Success(fmt.Sprintf("%s/%s not found (ignored)", ref.kind, ref.name)), nil
		}
		e.LogError(ctx, "Failed to delete resource", map[string]any{
			"kind":  ref.kind,
			"name":  ref.name,
			"error": err.Error(),
		})
		return sdk.Failure(fmt.Errorf("failed to delete resource: %w", err)), nil
	}

	e.LogInfo(ctx, "Resource deleted successfully", map[string]any{
		"kind": ref.kind,
		"name": ref.name,
	})

	return sdk.Success(fmt.Sprintf("Deleted %s/%s", ref.kind, ref.name)), nil
}
