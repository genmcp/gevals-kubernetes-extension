package extension

import (
	"context"
	"testing"

	"github.com/genmcp/gevals/pkg/extension/sdk"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestHandleDelete(t *testing.T) {
	tests := []struct {
		name        string
		args        any
		client      *mockClient
		wantSuccess bool
	}{
		{
			name: "successful delete",
			args: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name":      "test-cm",
					"namespace": "default",
				},
			},
			client:      &mockClient{},
			wantSuccess: true,
		},
		{
			name: "not found with ignoreNotFound",
			args: map[string]any{
				"apiVersion":     "v1",
				"kind":           "ConfigMap",
				"metadata":       map[string]any{"name": "missing"},
				"ignoreNotFound": true,
			},
			client: &mockClient{
				deleteFn: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, opts metav1.DeleteOptions) error {
					return apierrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, "missing")
				},
			},
			wantSuccess: true,
		},
		{
			name: "not found without ignoreNotFound",
			args: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]any{"name": "missing"},
			},
			client: &mockClient{
				deleteFn: func(ctx context.Context, gvr schema.GroupVersionResource, name, namespace string, opts metav1.DeleteOptions) error {
					return apierrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, "missing")
				},
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := &Extension{
				Extension: sdk.NewExtension(sdk.ExtensionInfo{Name: "test"}),
				client:    tt.client,
			}

			req := &sdk.OperationRequest{Args: tt.args}
			result, err := ext.handleDelete(context.Background(), req)

			if err != nil {
				t.Fatalf("handleDelete() returned error: %v", err)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("handleDelete() success = %v, want %v", result.Success, tt.wantSuccess)
			}
		})
	}
}
