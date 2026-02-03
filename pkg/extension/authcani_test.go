package extension

import (
	"context"
	"fmt"
	"testing"

	"github.com/mcpchecker/mcpchecker/pkg/extension/sdk"
)

func TestHandleAuthCanI(t *testing.T) {
	tests := []struct {
		name        string
		args        any
		client      *mockClient
		wantSuccess bool
	}{
		{
			name: "allowed action",
			args: map[string]any{
				"verb":      "get",
				"resource":  "pods",
				"as":        "system:serviceaccount:default:test-sa",
				"namespace": "default",
			},
			client: &mockClient{
				checkAccessFn: func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
					return true, "allowed by RBAC", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "denied action",
			args: map[string]any{
				"verb":      "delete",
				"resource":  "pods",
				"as":        "system:serviceaccount:default:test-sa",
				"namespace": "default",
			},
			client: &mockClient{
				checkAccessFn: func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
					return false, "denied by RBAC", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "cluster-wide check allowed",
			args: map[string]any{
				"verb":     "list",
				"resource": "namespaces",
				"as":       "admin-user",
			},
			client: &mockClient{
				checkAccessFn: func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
					if namespace != "" {
						return false, "", fmt.Errorf("expected cluster-wide check")
					}
					return true, "", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "expect allowed matches",
			args: map[string]any{
				"verb":      "get",
				"resource":  "pods",
				"as":        "alice",
				"namespace": "test-ns",
				"expect": map[string]any{
					"allowed": true,
				},
			},
			client: &mockClient{
				checkAccessFn: func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
					return true, "", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "expect allowed mismatch",
			args: map[string]any{
				"verb":      "delete",
				"resource":  "pods",
				"as":        "alice",
				"namespace": "test-ns",
				"expect": map[string]any{
					"allowed": true,
				},
			},
			client: &mockClient{
				checkAccessFn: func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
					return false, "denied", nil
				},
			},
			wantSuccess: false,
		},
		{
			name: "expect denied matches",
			args: map[string]any{
				"verb":      "delete",
				"resource":  "pods",
				"as":        "readonly-user",
				"namespace": "prod",
				"expect": map[string]any{
					"allowed": false,
				},
			},
			client: &mockClient{
				checkAccessFn: func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
					return false, "denied by policy", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "missing verb",
			args: map[string]any{
				"resource": "pods",
				"as":       "alice",
			},
			client:      &mockClient{},
			wantSuccess: false,
		},
		{
			name: "missing resource",
			args: map[string]any{
				"verb": "get",
				"as":   "alice",
			},
			client:      &mockClient{},
			wantSuccess: false,
		},
		{
			name: "missing as",
			args: map[string]any{
				"verb":     "get",
				"resource": "pods",
			},
			client:      &mockClient{},
			wantSuccess: false,
		},
		{
			name: "with api group",
			args: map[string]any{
				"verb":      "create",
				"resource":  "deployments",
				"as":        "developer",
				"namespace": "apps",
				"apiGroup":  "apps",
			},
			client: &mockClient{
				checkAccessFn: func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
					if apiGroup != "apps" {
						return false, "", fmt.Errorf("expected apiGroup=apps, got %s", apiGroup)
					}
					return true, "", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "with specific resource name",
			args: map[string]any{
				"verb":         "get",
				"resource":     "secrets",
				"as":           "limited-user",
				"namespace":    "default",
				"resourceName": "my-secret",
			},
			client: &mockClient{
				checkAccessFn: func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
					if resourceName != "my-secret" {
						return false, "", fmt.Errorf("expected resourceName=my-secret, got %s", resourceName)
					}
					return true, "", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "service account format",
			args: map[string]any{
				"verb":      "get",
				"resource":  "pods",
				"as":        "system:serviceaccount:create-simple-rbac:reader-sa",
				"namespace": "create-simple-rbac",
			},
			client: &mockClient{
				checkAccessFn: func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
					if user != "system:serviceaccount:create-simple-rbac:reader-sa" {
						return false, "", fmt.Errorf("expected full service account name")
					}
					return true, "", nil
				},
			},
			wantSuccess: true,
		},
		{
			name: "client error",
			args: map[string]any{
				"verb":      "get",
				"resource":  "pods",
				"as":        "alice",
				"namespace": "default",
			},
			client: &mockClient{
				checkAccessFn: func(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
					return false, "", fmt.Errorf("connection refused")
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
			result, err := ext.handleAuthCanI(context.Background(), req)

			if err != nil {
				t.Fatalf("handleAuthCanI() returned error: %v", err)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("handleAuthCanI() success = %v, want %v, message = %s", result.Success, tt.wantSuccess, result.Message)
			}
		})
	}
}
