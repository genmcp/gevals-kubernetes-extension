package extension

import (
	"github.com/mcpchecker/mcpchecker/pkg/extension/sdk"
	"github.com/google/jsonschema-go/jsonschema"
)

// registerOperations adds all available Kubernetes operations to the extension.
// Each operation is defined with a JSON schema for input validation and a handler function.
func (e *Extension) registerOperations() {
	e.AddOperation(
		sdk.NewOperation("create",
			sdk.WithDescription("Create a Kubernetes resource"),
			sdk.WithParams(jsonschema.Schema{
				Type:        "object",
				Description: "Kubernetes resource spec (apiVersion, kind, metadata, spec, etc.)",
				Properties: map[string]*jsonschema.Schema{
					"apiVersion": {
						Type:        "string",
						Description: "API version (e.g., v1, apps/v1)",
					},
					"kind": {
						Type:        "string",
						Description: "Resource kind (e.g., Pod, Namespace, Deployment)",
					},
					"metadata": {
						Type:        "object",
						Description: "Resource metadata (name, namespace, labels, annotations)",
					},
					"spec": {
						Type:        "object",
						Description: "Resource spec (optional, depends on resource type)",
					},
				},
				Required: []string{"apiVersion", "kind", "metadata"},
			}),
		),
		e.handleCreate,
	)

	e.AddOperation(
		sdk.NewOperation("wait",
			sdk.WithDescription("Wait for a condition on a Kubernetes resource"),
			sdk.WithParams(jsonschema.Schema{
				Type:        "object",
				Description: "Resource reference with condition to wait for",
				Properties: map[string]*jsonschema.Schema{
					"apiVersion": {
						Type:        "string",
						Description: "API version (e.g., v1, apps/v1)",
					},
					"kind": {
						Type:        "string",
						Description: "Resource kind (e.g., Pod, Deployment)",
					},
					"metadata": {
						Type:        "object",
						Description: "Resource metadata (name, namespace)",
					},
					"condition": {
						Type:        "string",
						Description: "Condition type to wait for (e.g., Ready, Available)",
					},
					"status": {
						Type:        "string",
						Description: "Expected condition status (default: True)",
					},
					"timeout": {
						Type:        "string",
						Description: "Timeout duration (e.g., 60s, 5m, default: 60s)",
					},
				},
				Required: []string{"apiVersion", "kind", "metadata", "condition"},
			}),
		),
		e.handleWait,
	)

	e.AddOperation(
		sdk.NewOperation("delete",
			sdk.WithDescription("Delete a Kubernetes resource"),
			sdk.WithParams(jsonschema.Schema{
				Type:        "object",
				Description: "Resource reference to delete",
				Properties: map[string]*jsonschema.Schema{
					"apiVersion": {
						Type:        "string",
						Description: "API version (e.g., v1, apps/v1)",
					},
					"kind": {
						Type:        "string",
						Description: "Resource kind (e.g., Pod, Namespace)",
					},
					"metadata": {
						Type:        "object",
						Description: "Resource metadata (name, namespace)",
					},
					"ignoreNotFound": {
						Type:        "boolean",
						Description: "If true, do not fail when the resource does not exist",
					},
				},
				Required: []string{"apiVersion", "kind", "metadata"},
			}),
		),
		e.handleDelete,
	)

	e.AddOperation(
		sdk.NewOperation("authCanI",
			sdk.WithDescription("Check if a user or service account can perform an action on a resource"),
			sdk.WithParams(jsonschema.Schema{
				Type:        "object",
				Description: "Permission check parameters",
				Properties: map[string]*jsonschema.Schema{
					"verb": {
						Type:        "string",
						Description: "Action verb (get, list, create, delete, watch, patch, update, etc.)",
					},
					"resource": {
						Type:        "string",
						Description: "Resource name (pods, deployments, configmaps, etc.)",
					},
					"as": {
						Type:        "string",
						Description: "User or service account to impersonate (e.g., alice, system:serviceaccount:ns:sa-name)",
					},
					"namespace": {
						Type:        "string",
						Description: "Namespace scope (optional, empty for cluster-wide check)",
					},
					"apiGroup": {
						Type:        "string",
						Description: "API group (optional, empty for core API, e.g., apps, batch, rbac.authorization.k8s.io)",
					},
					"resourceName": {
						Type:        "string",
						Description: "Specific resource name to check access for (optional)",
					},
					"expect": {
						Type:        "object",
						Description: "Expected result for inline verification",
						Properties: map[string]*jsonschema.Schema{
							"allowed": {
								Type:        "boolean",
								Description: "Expected permission result (true for allowed, false for denied)",
							},
						},
					},
				},
				Required: []string{"verb", "resource", "as"},
			}),
		),
		e.handleAuthCanI,
	)
}
