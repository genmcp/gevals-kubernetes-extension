package extension

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpchecker/mcpchecker/pkg/extension/sdk"
)

func TestHandleListContexts(t *testing.T) {
	tests := []struct {
		name        string
		client      *mockClient
		wantSuccess bool
		wantOutputs bool
	}{
		{
			name: "successful list with multiple contexts",
			client: &mockClient{
				listContextsFn: func(ctx context.Context) ([]ContextInfo, error) {
					return []ContextInfo{
						{Name: "dev", Cluster: "dev-cluster", User: "dev-user", IsCurrent: false},
						{Name: "prod", Cluster: "prod-cluster", User: "prod-user", IsCurrent: true},
					}, nil
				},
			},
			wantSuccess: true,
			wantOutputs: true,
		},
		{
			name: "successful list with single context",
			client: &mockClient{
				listContextsFn: func(ctx context.Context) ([]ContextInfo, error) {
					return []ContextInfo{
						{Name: "kind-kind", Cluster: "kind-kind", User: "kind-kind", IsCurrent: true},
					}, nil
				},
			},
			wantSuccess: true,
			wantOutputs: true,
		},
		{
			name: "no contexts found",
			client: &mockClient{
				listContextsFn: func(ctx context.Context) ([]ContextInfo, error) {
					return []ContextInfo{}, nil
				},
			},
			wantSuccess: false,
		},
		{
			name: "client error",
			client: &mockClient{
				listContextsFn: func(ctx context.Context) ([]ContextInfo, error) {
					return nil, errors.New("failed to load kubeconfig")
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

			req := &sdk.OperationRequest{Args: map[string]any{}}
			result, err := ext.handleListContexts(context.Background(), req)

			if err != nil {
				t.Fatalf("handleListContexts() returned error: %v", err)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("handleListContexts() success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if tt.wantOutputs && result.Outputs == nil {
				t.Errorf("handleListContexts() outputs = nil, want outputs")
			}
		})
	}
}

func TestHandleGetCurrentContext(t *testing.T) {
	tests := []struct {
		name        string
		client      *mockClient
		wantSuccess bool
		wantContext string
	}{
		{
			name: "successful get current context",
			client: &mockClient{
				getCurrentContextFn: func(ctx context.Context) (string, error) {
					return "prod", nil
				},
			},
			wantSuccess: true,
			wantContext: "prod",
		},
		{
			name: "empty current context",
			client: &mockClient{
				getCurrentContextFn: func(ctx context.Context) (string, error) {
					return "", nil
				},
			},
			wantSuccess: false,
		},
		{
			name: "client error",
			client: &mockClient{
				getCurrentContextFn: func(ctx context.Context) (string, error) {
					return "", errors.New("failed to load kubeconfig")
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

			req := &sdk.OperationRequest{Args: map[string]any{}}
			result, err := ext.handleGetCurrentContext(context.Background(), req)

			if err != nil {
				t.Fatalf("handleGetCurrentContext() returned error: %v", err)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("handleGetCurrentContext() success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if tt.wantSuccess && result.Outputs != nil {
				if ctx := result.Outputs["context"]; ctx != tt.wantContext {
					t.Errorf("handleGetCurrentContext() context = %v, want %v", ctx, tt.wantContext)
				}
			}
		})
	}
}

func TestHandleViewConfig(t *testing.T) {
	tests := []struct {
		name        string
		args        any
		client      *mockClient
		wantSuccess bool
	}{
		{
			name: "successful view without minify",
			args: map[string]any{
				"minify": false,
			},
			client: &mockClient{
				viewConfigFn: func(ctx context.Context, minify bool) (string, error) {
					if minify {
						t.Error("expected minify=false")
					}
					return "apiVersion: v1\nkind: Config\nclusters:\n- cluster:\n    server: https://example.com\n", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "successful view with minify",
			args: map[string]any{
				"minify": true,
			},
			client: &mockClient{
				viewConfigFn: func(ctx context.Context, minify bool) (string, error) {
					if !minify {
						t.Error("expected minify=true")
					}
					return "apiVersion: v1\nkind: Config\ncurrent-context: prod\n", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "default minify to false",
			args: map[string]any{},
			client: &mockClient{
				viewConfigFn: func(ctx context.Context, minify bool) (string, error) {
					if minify {
						t.Error("expected minify=false by default")
					}
					return "apiVersion: v1\nkind: Config\n", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "client error",
			args: map[string]any{},
			client: &mockClient{
				viewConfigFn: func(ctx context.Context, minify bool) (string, error) {
					return "", errors.New("failed to read kubeconfig")
				},
			},
			wantSuccess: false,
		},
		{
			name: "minify with missing cluster reference",
			args: map[string]any{
				"minify": true,
			},
			client: &mockClient{
				viewConfigFn: func(ctx context.Context, minify bool) (string, error) {
					return "", errors.New("cluster \"missing-cluster\" not found in kubeconfig")
				},
			},
			wantSuccess: false,
		},
		{
			name: "minify with missing authInfo reference",
			args: map[string]any{
				"minify": true,
			},
			client: &mockClient{
				viewConfigFn: func(ctx context.Context, minify bool) (string, error) {
					return "", errors.New("user \"missing-user\" not found in kubeconfig")
				},
			},
			wantSuccess: false,
		},
		{
			name: "minify with empty cluster name",
			args: map[string]any{
				"minify": true,
			},
			client: &mockClient{
				viewConfigFn: func(ctx context.Context, minify bool) (string, error) {
					return "", errors.New("current context \"prod\" has no cluster")
				},
			},
			wantSuccess: false,
		},
		{
			name: "minify with empty authInfo (should succeed)",
			args: map[string]any{
				"minify": true,
			},
			client: &mockClient{
				viewConfigFn: func(ctx context.Context, minify bool) (string, error) {
					// AuthInfo is optional, so empty authInfo should succeed
					return "apiVersion: v1\nkind: Config\ncurrent-context: prod\n", nil
				},
			},
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := &Extension{
				Extension: sdk.NewExtension(sdk.ExtensionInfo{Name: "test"}),
				client:    tt.client,
			}

			req := &sdk.OperationRequest{Args: tt.args}
			result, err := ext.handleViewConfig(context.Background(), req)

			if err != nil {
				t.Fatalf("handleViewConfig() returned error: %v", err)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("handleViewConfig() success = %v, want %v", result.Success, tt.wantSuccess)
			}
		})
	}
}
