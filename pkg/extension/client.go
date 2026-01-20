package extension

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// ResourceClient abstracts Kubernetes resource operations for testability.
// Implementations can use the real dynamic client or a mock for testing.
type ResourceClient interface {
	// Create creates a Kubernetes resource and returns the created object.
	Create(ctx context.Context, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, namespace string) (*unstructured.Unstructured, error)

	// Get retrieves a Kubernetes resource by name.
	Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error)

	// Delete removes a Kubernetes resource.
	Delete(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, opts metav1.DeleteOptions) error
}

// dynamicClientAdapter adapts the Kubernetes dynamic client to the ResourceClient interface.
type dynamicClientAdapter struct {
	client dynamic.Interface
}

func (a *dynamicClientAdapter) Create(ctx context.Context, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, namespace string) (*unstructured.Unstructured, error) {
	if namespace != "" {
		return a.client.Resource(gvr).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
	}
	return a.client.Resource(gvr).Create(ctx, obj, metav1.CreateOptions{})
}

func (a *dynamicClientAdapter) Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error) {
	if namespace != "" {
		return a.client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	}
	return a.client.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
}

func (a *dynamicClientAdapter) Delete(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, opts metav1.DeleteOptions) error {
	if namespace != "" {
		return a.client.Resource(gvr).Namespace(namespace).Delete(ctx, name, opts)
	}
	return a.client.Resource(gvr).Delete(ctx, name, opts)
}
