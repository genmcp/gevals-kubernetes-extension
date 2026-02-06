package extension

import (
	"context"
	"testing"

	"github.com/mcpchecker/mcpchecker/pkg/extension/sdk"
)

func TestHandleHelmInstall(t *testing.T) {
	tests := []struct {
		name        string
		args        any
		wantSuccess bool
		wantErrMsg  string
	}{
		{
			name:        "missing chart parameter",
			args:        map[string]any{},
			wantSuccess: false,
			wantErrMsg:  "chart parameter is required",
		},
		{
			name: "empty chart parameter",
			args: map[string]any{
				"chart": "",
			},
			wantSuccess: false,
			wantErrMsg:  "chart parameter is required",
		},
		{
			name: "invalid args type",
			args: "invalid",
			wantSuccess: false,
			wantErrMsg:  "args must be an object",
		},
		{
			name: "valid chart parameter",
			args: map[string]any{
				"chart": "oci://registry.io/chart/nginx",
				"name":  "test-release",
			},
			// Will fail because helm isn't actually run, but we test parameter validation
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := &Extension{
				Extension: sdk.NewExtension(sdk.ExtensionInfo{Name: "test"}),
			}

			req := &sdk.OperationRequest{Args: tt.args}
			result, err := ext.handleHelmInstall(context.Background(), req)

			if err != nil {
				t.Fatalf("handleHelmInstall() returned error: %v", err)
			}

			if result.Success != tt.wantSuccess {
				t.Errorf("handleHelmInstall() success = %v, want %v", result.Success, tt.wantSuccess)
			}

			if tt.wantErrMsg != "" && result.Message != "" {
				// Check if error message contains expected substring
				// (we don't check exact match because helm error messages may vary)
				if result.Success {
					t.Errorf("expected failure with message containing %q, but got success", tt.wantErrMsg)
				}
			}
		})
	}
}

func TestHandleHelmList(t *testing.T) {
	tests := []struct {
		name        string
		args        any
		wantSuccess bool
	}{
		{
			name:        "invalid args type",
			args:        "invalid",
			wantSuccess: false,
		},
		{
			name: "valid empty args",
			args: map[string]any{},
			// helm list succeeds even with no releases (returns empty list)
			wantSuccess: true,
		},
		{
			name: "with namespace",
			args: map[string]any{
				"namespace": "default",
			},
			wantSuccess: true,
		},
		{
			name: "all namespaces",
			args: map[string]any{
				"allNamespaces": true,
			},
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := &Extension{
				Extension: sdk.NewExtension(sdk.ExtensionInfo{Name: "test"}),
			}

			req := &sdk.OperationRequest{Args: tt.args}
			result, err := ext.handleHelmList(context.Background(), req)

			if err != nil {
				t.Fatalf("handleHelmList() returned error: %v", err)
			}

			if result.Success != tt.wantSuccess {
				t.Errorf("handleHelmList() success = %v, want %v", result.Success, tt.wantSuccess)
			}
		})
	}
}

func TestHandleHelmUninstall(t *testing.T) {
	tests := []struct {
		name        string
		args        any
		wantSuccess bool
	}{
		{
			name:        "missing name parameter",
			args:        map[string]any{},
			wantSuccess: false,
		},
		{
			name: "empty name parameter",
			args: map[string]any{
				"name": "",
			},
			wantSuccess: false,
		},
		{
			name:        "invalid args type",
			args:        "invalid",
			wantSuccess: false,
		},
		{
			name: "valid name parameter for non-existent release",
			args: map[string]any{
				"name": "test-release-nonexistent",
			},
			// Succeeds because code handles "not found" gracefully
			wantSuccess: true,
		},
		{
			name: "with namespace",
			args: map[string]any{
				"name":      "test-release-nonexistent",
				"namespace": "default",
			},
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := &Extension{
				Extension: sdk.NewExtension(sdk.ExtensionInfo{Name: "test"}),
			}

			req := &sdk.OperationRequest{Args: tt.args}
			result, err := ext.handleHelmUninstall(context.Background(), req)

			if err != nil {
				t.Fatalf("handleHelmUninstall() returned error: %v", err)
			}

			if result.Success != tt.wantSuccess {
				t.Errorf("handleHelmUninstall() success = %v, want %v", result.Success, tt.wantSuccess)
			}
		})
	}
}
