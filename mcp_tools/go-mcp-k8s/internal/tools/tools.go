package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"go-mcp-k8s/internal/k8s"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreateGetAPIResourcesTool creates a tool for getting API resources
func CreateGetAPIResourcesTool() mcp.Tool {
	return mcp.NewTool("get_api_resources",
		mcp.WithDescription("Get all supported API resource types in the cluster, including built-in resources and CRDs"),
		mcp.WithBoolean("includeNamespaceScoped",
			mcp.Description("Include namespace-scoped resources"),
			mcp.DefaultBool(true),
		),
		mcp.WithBoolean("includeClusterScoped",
			mcp.Description("Include cluster-scoped resources"),
			mcp.DefaultBool(true),
		),
	)
}

// CreateGetNamespacesTool
func CreateGetNamespacesTool() mcp.Tool {
	return mcp.NewTool("get_namespaces",
		mcp.WithDescription("Get all namespaces in the cluster"),
	)
}

// CreateGetPodsTool
func CreateGetPodsTool() mcp.Tool {
	return mcp.NewTool("get_pods_by_namespace",
		mcp.WithDescription("Get all pods in a specific namespace"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
	)
}

// CreateGetDeploymentsTool
func CreateGetDeploymentsTool() mcp.Tool {
	return mcp.NewTool("get_deployments_by_namespace",
		mcp.WithDescription("Get all deployments in a specific namespace"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
	)
}

// CreateGetServicesTool
func CreateGetServicesTool() mcp.Tool {
	return mcp.NewTool("get_services_by_namespace",
		mcp.WithDescription("Get all services in a specific namespace"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
	)
}

// CreateGetConfigMapsTool
func CreateGetConfigMapsTool() mcp.Tool {
	return mcp.NewTool("get_configmaps_by_namespace",
		mcp.WithDescription("Get all configmaps in a specific namespace"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
	)
}

// CreateGetResourceTool creates a tool for getting a specific resource
func CreateGetResourceTool() mcp.Tool {
	return mcp.NewTool("get_resource",
		mcp.WithDescription("Get detailed information about a specific resource"),
		mcp.WithString("kind",
			mcp.Required(),
			mcp.Description("Resource type"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Resource name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
	)
}

// CreateListResourcesTool creates a tool for listing resources
func CreateListResourcesTool() mcp.Tool {
	return mcp.NewTool("list_resources",
		mcp.WithDescription("List all instances of a resource type"),
		mcp.WithString("kind",
			mcp.Required(),
			mcp.Description("Resource type"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (only list resources in this namespace)"),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Label selector (format: key1=value1,key2=value2)"),
		),
		mcp.WithString("fieldSelector",
			mcp.Description("Field selector (format: key1=value1,key2=value2)"),
		),
	)
}

// CreateCreateResourceTool creates a tool for creating resources
func CreateCreateResourceTool() mcp.Tool {
	return mcp.NewTool("create_resource",
		mcp.WithDescription("Create a new resource"),
		mcp.WithString("kind",
			mcp.Required(),
			mcp.Description("Resource type"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
		mcp.WithString("manifest",
			mcp.Required(),
			mcp.Description("Resource manifest (JSON/YAML)"),
		),
	)
}

// CreateUpdateResourceTool creates a tool for updating resources
func CreateUpdateResourceTool() mcp.Tool {
	return mcp.NewTool("update_resource",
		mcp.WithDescription("Update an existing resource"),
		mcp.WithString("kind",
			mcp.Required(),
			mcp.Description("Resource type"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Resource name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
		mcp.WithString("manifest",
			mcp.Required(),
			mcp.Description("Resource manifest (JSON/YAML)"),
		),
	)
}

// CreateDeleteResourceTool creates a tool for deleting resources
func CreateDeleteResourceTool() mcp.Tool {
	return mcp.NewTool("delete_resource",
		mcp.WithDescription("Delete a resource"),
		mcp.WithString("kind",
			mcp.Required(),
			mcp.Description("Resource type"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Resource name"),
		),
		mcp.WithString("namespace",
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
	)
}

// CreateDescribePodTool
func CreateDescribePodTool() mcp.Tool {
	return mcp.NewTool("describe_pod",
		mcp.WithDescription("Get all details of a specific pod"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
		mcp.WithString("podName",
			mcp.Required(),
			mcp.Description("Pod name"),
		),
	)
}

// CreateQueryLogsByPodNameTool
func CreateQueryLogsByPodNameTool() mcp.Tool {
	return mcp.NewTool("query_logs_by_pod_name",
		mcp.WithDescription("Get logs of a specific pod"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
		mcp.WithString("podName",
			mcp.Required(),
			mcp.Description("Pod name"),
		),
	)
}

// CreateScaleDeploymentTool
func CreateScaleDeploymentTool() mcp.Tool {
	return mcp.NewTool("scale_deployment",
		mcp.WithDescription("Scale a deployment"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("Namespace (required for namespace-scoped resources)"),
		),
		mcp.WithString("deploymentName",
			mcp.Required(),
			mcp.Description("Deployment name"),
		),
		mcp.WithNumber("replicas",
			mcp.Required(),
			mcp.Description("Number of replicas"),
		),
	)
}

// HandleGetNamespaces handles the get namespaces tool
func HandleGetNamespaces(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespaces, err := client.GetNamespaces(ctx)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(namespaces)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleGetPods handles the get pods tool
func HandleGetPods(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := request.Params.Arguments["namespace"].(string)
		pods, err := client.GetPodList(ctx, namespace)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(pods)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleGetDeployments handles the get deployments tool
func HandleGetDeployments(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := request.Params.Arguments["namespace"].(string)
		deployments, err := client.GetDeploymentList(ctx, namespace)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(deployments)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleGetServices handles the get services tool
func HandleGetServices(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := request.Params.Arguments["namespace"].(string)
		services, err := client.GetServiceList(ctx, namespace)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(services)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleGetConfigMaps handles the get configmaps tool
func HandleGetConfigMaps(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := request.Params.Arguments["namespace"].(string)
		configmaps, err := client.GetConfigMapList(ctx, namespace)

		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(configmaps)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleGetAPIResources handles the get API resources tool
func HandleGetAPIResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		includeNamespaceScoped := true
		includeClusterScoped := true

		if val, ok := request.Params.Arguments["includeNamespaceScoped"].(bool); ok {
			includeNamespaceScoped = val
		}

		if val, ok := request.Params.Arguments["includeClusterScoped"].(bool); ok {
			includeClusterScoped = val
		}

		resources, err := client.GetAPIResources(ctx, includeNamespaceScoped, includeClusterScoped)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleGetResource handles the get resource tool
func HandleGetResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing required parameter: name")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

		resource, err := client.GetResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleListResources handles the list resources tool
func HandleListResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)
		labelSelector, _ := request.Params.Arguments["labelSelector"].(string)
		fieldSelector, _ := request.Params.Arguments["fieldSelector"].(string)

		resources, err := client.ListResources(ctx, kind, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleCreateResource handles the create resource tool
func HandleCreateResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		manifest, ok := request.Params.Arguments["manifest"].(string)
		if !ok || manifest == "" {
			return nil, fmt.Errorf("missing required parameter: manifest")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

		resource, err := client.CreateResource(ctx, kind, namespace, manifest)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleUpdateResource handles the update resource tool
func HandleUpdateResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing required parameter: name")
		}

		manifest, ok := request.Params.Arguments["manifest"].(string)
		if !ok || manifest == "" {
			return nil, fmt.Errorf("missing required parameter: manifest")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

		resource, err := client.UpdateResource(ctx, kind, name, namespace, manifest)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HandleDeleteResource handles the delete resource tool
func HandleDeleteResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing required parameter: name")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

		err := client.DeleteResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted resource %s/%s", kind, name)), nil
	}
}

// handle query logs by pod name
func HandleQueryLogsByPodName(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		podName, _ := request.Params.Arguments["podName"].(string)
		namespace, _ := request.Params.Arguments["namespace"].(string)

		logs, err := client.GetPodLogs(ctx, namespace, podName)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(logs), nil
	}
}

// handle describe pod
func HandleDescribePod(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		podName, _ := request.Params.Arguments["podName"].(string)
		namespace, _ := request.Params.Arguments["namespace"].(string)

		pod, err := client.GetPodDescription(ctx, namespace, podName)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(pod)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// handle scale deployment
func HandleScaleDeployment(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := request.Params.Arguments["namespace"].(string)
		deploymentName, _ := request.Params.Arguments["deploymentName"].(string)
		replicas, _ := request.Params.Arguments["replicas"].(float64)
		// float64 to int32
		replicasInt := int32(replicas)
		err := client.ScaleDeployment(ctx, namespace, deploymentName, replicasInt)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully scaled namespace %s deployment %s to %d replicas", namespace, deploymentName, replicasInt)), nil
	}
}
