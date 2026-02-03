package extension

import (
	"context"
	"fmt"
	"sort"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	authorizationv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
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

	// CheckAccess checks if a user can perform an action on a resource.
	CheckAccess(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error)

	// ListContexts returns all contexts from the kubeconfig sorted by name.
	// Each context includes its name, cluster, user, namespace, and whether it's the current context.
	ListContexts(ctx context.Context) ([]ContextInfo, error)

	// GetCurrentContext returns the current context name from the kubeconfig.
	// Returns an error if the kubeconfig cannot be loaded.
	GetCurrentContext(ctx context.Context) (string, error)

	// ViewConfig returns the kubeconfig as YAML.
	// When minify is true, only the current context and its dependencies are included.
	ViewConfig(ctx context.Context, minify bool) (string, error)
}

// dynamicClientAdapter adapts the Kubernetes dynamic client to the ResourceClient interface.
type dynamicClientAdapter struct {
	client       dynamic.Interface
	authzClient  authorizationv1client.AuthorizationV1Interface
	kubeconfigPath string
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

func (a *dynamicClientAdapter) CheckAccess(ctx context.Context, user, verb, resource, apiGroup, namespace, resourceName string) (bool, string, error) {
	sar := &authorizationv1.SubjectAccessReview{
		Spec: authorizationv1.SubjectAccessReviewSpec{
			User: user,
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb:      verb,
				Resource:  resource,
				Group:     apiGroup,
				Namespace: namespace,
				Name:      resourceName,
			},
		},
	}

	result, err := a.authzClient.SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		return false, "", err
	}

	return result.Status.Allowed, result.Status.Reason, nil
}

func (a *dynamicClientAdapter) ListContexts(ctx context.Context) ([]ContextInfo, error) {
	config, err := clientcmd.LoadFromFile(a.kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	var contexts []ContextInfo
	for name, context := range config.Contexts {
		contexts = append(contexts, ContextInfo{
			Name:      name,
			Cluster:   context.Cluster,
			User:      context.AuthInfo,
			Namespace: context.Namespace,
			IsCurrent: name == config.CurrentContext,
		})
	}

	// Sort contexts by name for deterministic output
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Name < contexts[j].Name
	})

	return contexts, nil
}

func (a *dynamicClientAdapter) GetCurrentContext(ctx context.Context) (string, error) {
	config, err := clientcmd.LoadFromFile(a.kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return config.CurrentContext, nil
}

func (a *dynamicClientAdapter) ViewConfig(ctx context.Context, minify bool) (string, error) {
	// Load the full config
	rawConfig, err := clientcmd.LoadFromFile(a.kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Apply minification if requested
	if minify {
		// Get current context
		currentContext := rawConfig.CurrentContext
		if currentContext == "" {
			return "", fmt.Errorf("no current context set in kubeconfig")
		}

		// Create minified config with only current context and its dependencies
		currentCtx, exists := rawConfig.Contexts[currentContext]
		if !exists {
			return "", fmt.Errorf("current context %q not found in kubeconfig", currentContext)
		}

		minifiedConfig := clientcmdapi.NewConfig()
		minifiedConfig.CurrentContext = currentContext
		minifiedConfig.Contexts = map[string]*clientcmdapi.Context{
			currentContext: currentCtx,
		}

		// Add the cluster referenced by current context
		if currentCtx.Cluster == "" {
			return "", fmt.Errorf("current context %q has no cluster", currentContext)
		}
		if cluster, exists := rawConfig.Clusters[currentCtx.Cluster]; exists {
			minifiedConfig.Clusters = map[string]*clientcmdapi.Cluster{
				currentCtx.Cluster: cluster,
			}
		} else {
			return "", fmt.Errorf("cluster %q not found in kubeconfig", currentCtx.Cluster)
		}

		// Add the user referenced by current context (optional)
		if currentCtx.AuthInfo != "" {
			authInfo, exists := rawConfig.AuthInfos[currentCtx.AuthInfo]
			if !exists {
				return "", fmt.Errorf("user %q not found in kubeconfig", currentCtx.AuthInfo)
			}
			minifiedConfig.AuthInfos = map[string]*clientcmdapi.AuthInfo{
				currentCtx.AuthInfo: authInfo,
			}
		}

		rawConfig = minifiedConfig
	}

	// Convert to YAML
	yamlBytes, err := clientcmd.Write(*rawConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	return string(yamlBytes), nil
}
