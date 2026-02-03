package extension

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type mockClient struct {
	createFn            func(ctx context.Context, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, namespace string) (*unstructured.Unstructured, error)
	getFn               func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string) (*unstructured.Unstructured, error)
	deleteFn            func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, opts metav1.DeleteOptions) error
	checkAccessFn       func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error)
	listContextsFn      func(ctx context.Context) ([]ContextInfo, error)
	getCurrentContextFn func(ctx context.Context) (string, error)
	viewConfigFn        func(ctx context.Context, minify bool) (string, error)
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

func (m *mockClient) CheckAccess(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
	if m.checkAccessFn != nil {
		return m.checkAccessFn(ctx, user, verb, resource, apiGroup, namespace, resourceName)
	}
	return true, "", nil
}

func (m *mockClient) ListContexts(ctx context.Context) ([]ContextInfo, error) {
	if m.listContextsFn != nil {
		return m.listContextsFn(ctx)
	}
	return []ContextInfo{
		{Name: "default", Cluster: "default-cluster", User: "default-user", IsCurrent: true},
	}, nil
}

func (m *mockClient) GetCurrentContext(ctx context.Context) (string, error) {
	if m.getCurrentContextFn != nil {
		return m.getCurrentContextFn(ctx)
	}
	return "default", nil
}

func (m *mockClient) ViewConfig(ctx context.Context, minify bool) (string, error) {
	if m.viewConfigFn != nil {
		return m.viewConfigFn(ctx, minify)
	}
	return "apiVersion: v1\nkind: Config\n", nil
}
