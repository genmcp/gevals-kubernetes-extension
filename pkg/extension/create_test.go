package extension

import (
	"context"
	"errors"
	"testing"

	"github.com/genmcp/gevals/pkg/extension/sdk"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestHandleCreate(t *testing.T) {
	tests := []struct {
		name        string
		args        any
		client      *mockClient
		wantSuccess bool
	}{
		{
			name: "successful create",
			args: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name":      "test-cm",
					"namespace": "default",
				},
			},
			client: &mockClient{
				createFn: func(ctx context.Context, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, namespace string) (*unstructured.Unstructured, error) {
					result := obj.DeepCopy()
					result.SetUID("test-uid")
					result.SetResourceVersion("1")
					return result, nil
				},
			},
			wantSuccess: true,
		},
		{
			name:        "invalid args type",
			args:        "not a map",
			client:      &mockClient{},
			wantSuccess: false,
		},
		{
			name: "client error",
			args: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name": "test-cm",
				},
			},
			client: &mockClient{
				createFn: func(ctx context.Context, gvr schema.GroupVersionResource, obj *unstructured.Unstructured, namespace string) (*unstructured.Unstructured, error) {
					return nil, errors.New("connection refused")
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
			result, err := ext.handleCreate(context.Background(), req)

			if err != nil {
				t.Fatalf("handleCreate() returned error: %v", err)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("handleCreate() success = %v, want %v", result.Success, tt.wantSuccess)
			}
		})
	}
}
