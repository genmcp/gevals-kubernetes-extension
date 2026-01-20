package extension

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// resourceRef holds parsed resource reference information from operation arguments.
type resourceRef struct {
	apiVersion string
	kind       string
	name       string
	namespace  string
}

// parseResourceRef extracts resource reference info from operation arguments.
// It validates that required fields (apiVersion, kind, metadata.name) are present.
func parseResourceRef(args map[string]any) (*resourceRef, error) {
	apiVersion, _ := args["apiVersion"].(string)
	kind, _ := args["kind"].(string)

	if apiVersion == "" {
		return nil, fmt.Errorf("apiVersion is required")
	}
	if kind == "" {
		return nil, fmt.Errorf("kind is required")
	}

	ref := &resourceRef{
		apiVersion: apiVersion,
		kind:       kind,
	}

	if metadata, ok := args["metadata"].(map[string]any); ok {
		ref.name, _ = metadata["name"].(string)
		ref.namespace, _ = metadata["namespace"].(string)
	}

	if ref.name == "" {
		return nil, fmt.Errorf("metadata.name is required")
	}

	return ref, nil
}

// gvr converts the resource reference to a GroupVersionResource.
func (r *resourceRef) gvr() (schema.GroupVersionResource, error) {
	gv, err := schema.ParseGroupVersion(r.apiVersion)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("invalid apiVersion: %w", err)
	}
	gvk := gv.WithKind(r.kind)
	return gvkToGVR(gvk), nil
}

// gvkToGVR converts a GroupVersionKind to GroupVersionResource using Kubernetes
// built-in pluralization logic which correctly handles compound CamelCase kinds
// and common irregular plurals (e.g., Ingress → ingresses, NetworkPolicy → networkpolicies).
func gvkToGVR(gvk schema.GroupVersionKind) schema.GroupVersionResource {
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	return gvr
}
