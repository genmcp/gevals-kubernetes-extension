package extension

import (
	"testing"
)

func TestParseResourceRef(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		want    *resourceRef
		wantErr bool
	}{
		{
			name: "valid namespaced resource",
			args: map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]any{
					"name":      "nginx",
					"namespace": "default",
				},
			},
			want: &resourceRef{
				apiVersion: "apps/v1",
				kind:       "Deployment",
				name:       "nginx",
				namespace:  "default",
			},
		},
		{
			name: "missing apiVersion",
			args: map[string]any{
				"kind": "Pod",
				"metadata": map[string]any{
					"name": "test",
				},
			},
			wantErr: true,
		},
		{
			name: "missing metadata.name",
			args: map[string]any{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata":   map[string]any{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResourceRef(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResourceRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				if got.apiVersion != tt.want.apiVersion ||
					got.kind != tt.want.kind ||
					got.name != tt.want.name ||
					got.namespace != tt.want.namespace {
					t.Errorf("parseResourceRef() = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestResourceRefGVR(t *testing.T) {
	tests := []struct {
		name         string
		ref          *resourceRef
		wantGroup    string
		wantVersion  string
		wantResource string
		wantErr      bool
	}{
		{
			name:         "core v1 pod",
			ref:          &resourceRef{apiVersion: "v1", kind: "Pod"},
			wantGroup:    "",
			wantVersion:  "v1",
			wantResource: "pods",
		},
		{
			name:         "apps/v1 deployment",
			ref:          &resourceRef{apiVersion: "apps/v1", kind: "Deployment"},
			wantGroup:    "apps",
			wantVersion:  "v1",
			wantResource: "deployments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gvr, err := tt.ref.gvr()
			if (err != nil) != tt.wantErr {
				t.Errorf("gvr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gvr.Group != tt.wantGroup || gvr.Version != tt.wantVersion || gvr.Resource != tt.wantResource {
				t.Errorf("gvr() = %v, want %s/%s/%s", gvr, tt.wantGroup, tt.wantVersion, tt.wantResource)
			}
		})
	}
}
