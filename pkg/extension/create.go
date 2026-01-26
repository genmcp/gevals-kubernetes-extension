package extension

import (
	"context"
	"fmt"

	"github.com/mcpchecker/mcpchecker/pkg/extension/sdk"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (e *Extension) handleCreate(ctx context.Context, req *sdk.OperationRequest) (*sdk.OperationResult, error) {
	if e.client == nil {
		return sdk.Failure(fmt.Errorf("kubernetes client not initialized")), nil
	}

	// Args is the resource spec as a map
	resourceSpec, ok := req.Args.(map[string]any)
	if !ok {
		return sdk.Failure(fmt.Errorf("args must be a resource spec object")), nil
	}

	obj := &unstructured.Unstructured{Object: resourceSpec}

	gvk := obj.GroupVersionKind()
	if gvk.Kind == "" {
		return sdk.Failure(fmt.Errorf("kind is required")), nil
	}

	gvr := gvkToGVR(gvk)
	namespace := obj.GetNamespace()

	e.LogInfo(ctx, "Creating resource", map[string]any{
		"kind":      gvk.Kind,
		"name":      obj.GetName(),
		"namespace": namespace,
	})

	result, err := e.client.Create(ctx, gvr, obj, namespace)
	if err != nil {
		e.LogError(ctx, "Failed to create resource", map[string]any{
			"kind":  gvk.Kind,
			"name":  obj.GetName(),
			"error": err.Error(),
		})
		return sdk.Failure(fmt.Errorf("failed to create resource: %w", err)), nil
	}

	e.LogInfo(ctx, "Resource created successfully", map[string]any{
		"kind": gvk.Kind,
		"name": result.GetName(),
		"uid":  string(result.GetUID()),
	})

	return sdk.SuccessWithOutputs(
		fmt.Sprintf("Created %s/%s", gvk.Kind, result.GetName()),
		map[string]string{
			"name":            result.GetName(),
			"namespace":       result.GetNamespace(),
			"uid":             string(result.GetUID()),
			"resourceVersion": result.GetResourceVersion(),
		},
	), nil
}
