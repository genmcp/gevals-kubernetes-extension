package extension

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type mockClient struct {
	createFn func(ctx context.Context, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, namespace string) (*unstructured.Unstructured, error)
	getFn    func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error)
	deleteFn func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, opts metav1.DeleteOptions) error
}

func (m *mockClient) Create(ctx context.Context, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, namespace string) (*unstructured.Unstructured, error) {
	if m.createFn != nil {
		return m.createFn(ctx, gvr, obj, namespace)
	}
	return obj, nil
}

func (m *mockClient) Get(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error) {
	if m.getFn != nil {
		return m.getFn(ctx, gvr, name, namespace)
	}
	return nil, nil
}

func (m *mockClient) Delete(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, opts metav1.DeleteOptions) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, gvr, name, namespace, opts)
	}
	return nil
}
