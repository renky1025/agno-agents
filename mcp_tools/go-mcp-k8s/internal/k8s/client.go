package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client 封装了 Kubernetes 客户端功能
type Client struct {
	// 标准 clientset
	clientset *kubernetes.Clientset
	// 动态客户端
	dynamicClient dynamic.Interface
	// 发现客户端
	discoveryClient *discovery.DiscoveryClient
	// REST 配置
	restConfig *rest.Config
}

// NewClient 创建一个新的 Kubernetes 客户端
func NewClient(kubeconfigPath string) (*Client, error) {
	var kubeconfig string
	var config *rest.Config
	var err error

	// 如果提供了 kubeconfig 路径，使用它
	if kubeconfigPath != "" {
		kubeconfig = kubeconfigPath
	} else if home := homedir.HomeDir(); home != "" {
		// 否则尝试使用默认路径
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// 使用提供的 kubeconfig 或尝试集群内配置
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes 配置失败: %w", err)
	}

	// 创建 clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes 客户端失败: %w", err)
	}

	// 创建动态客户端
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建动态客户端失败: %w", err)
	}

	// 创建发现客户端
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建发现客户端失败: %w", err)
	}

	return &Client{
		clientset:       clientset,
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		restConfig:      config,
	}, nil
}

// 获取namespaces 列表
func (c *Client) GetNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 namespaces 失败: %w", err)
	}
	namespaceNames := make([]string, len(namespaces.Items))
	for i, namespace := range namespaces.Items {
		namespaceNames[i] = namespace.Name
	}
	return namespaceNames, nil
}

// 获取所有资源类型
func (c *Client) GetAllResourceTypes(ctx context.Context) ([]string, error) {
	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("获取资源类型失败: %w", err)
	}
	resourceTypes := make([]string, 0)
	for _, resourceList := range resourceLists {
		resourceTypes = append(resourceTypes, resourceList.GroupVersion)
	}
	return resourceTypes, nil
}

// 获取namespace下的pod列表
func (c *Client) GetPodList(ctx context.Context, namespace string) ([]string, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 pod 列表失败: %w", err)
	}
	podNames := make([]string, len(pods.Items))
	for i, pod := range pods.Items {
		podNames[i] = pod.Name
	}
	return podNames, nil
}

// 获取namespace下的deployment列表
func (c *Client) GetDeploymentList(ctx context.Context, namespace string) ([]string, error) {
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 deployment 列表失败: %w", err)
	}
	deploymentNames := make([]string, len(deployments.Items))
	for i, deployment := range deployments.Items {
		deploymentNames[i] = deployment.Name
	}
	return deploymentNames, nil
}

// 获取namespace下的service列表
func (c *Client) GetServiceList(ctx context.Context, namespace string) ([]string, error) {
	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 service 列表失败: %w", err)
	}
	serviceNames := make([]string, len(services.Items))
	for i, service := range services.Items {
		serviceNames[i] = service.Name
	}
	return serviceNames, nil
}

// 获取namespace下的configmap列表
func (c *Client) GetConfigMapList(ctx context.Context, namespace string) ([]string, error) {
	configmaps, err := c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 configmap 列表失败: %w", err)
	}
	configmapNames := make([]string, len(configmaps.Items))
	for i, configmap := range configmaps.Items {
		configmapNames[i] = configmap.Name
	}
	return configmapNames, nil
}

// GetPodLogs 获取指定 pod 的日志
// 默认获取5行日志
func (c *Client) GetPodLogs(ctx context.Context, namespace, podName string) (string, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	count := int64(5)
	podLogOptions := v1.PodLogOptions{
		TailLines: &count,
		// 移除 Follow 选项，避免无限等待
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOptions)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		if err == context.DeadlineExceeded {
			return "", fmt.Errorf("获取日志超时")
		}
		return "", fmt.Errorf("打开日志流失败: %w", err)
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	// 添加读取超时
	errChan := make(chan error, 1)
	go func() {
		_, err := io.Copy(buf, podLogs)
		errChan <- err
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return "", fmt.Errorf("读取日志失败: %w", err)
		}
	case <-ctx.Done():
		return "", fmt.Errorf("读取日志超时")
	}

	return buf.String(), nil
}

// 根据namespace 和podname获取pod的描述
func (c *Client) GetPodDescription(ctx context.Context, namespace, podName string) (string, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("获取 pod 描述失败: %w", err)
	}
	if errors.IsNotFound(err) {
		return "", fmt.Errorf("Pod %s in namespace %s not found\n", podName, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		return "", fmt.Errorf("Error getting pod %s in namespace %s: %v\n",
			podName, namespace, statusError.ErrStatus.Message)
	} else if err != nil {
		return "", fmt.Errorf("获取 pod 描述失败: %w", err)
	}
	bytes, _ := json.Marshal(pod)
	return string(bytes), nil
}

// scale up /down deployment
func (c *Client) ScaleDeployment(ctx context.Context, namespace, deploymentName string, replicas int32) error {
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 deployment 失败: %w", err)
	}
	deployment.Spec.Replicas = &replicas
	_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 deployment 失败: %w", err)
	}
	return nil
}

// GetAPIResources 获取集群中的所有 API 资源类型
func (c *Client) GetAPIResources(ctx context.Context, includeNamespaceScoped, includeClusterScoped bool) ([]map[string]interface{}, error) {
	// 获取集群中所有 API Groups 和 Resources
	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil {
		// 处理部分错误，有些资源可能无法访问
		if !discovery.IsGroupDiscoveryFailedError(err) {
			return nil, fmt.Errorf("获取 API 资源失败: %w", err)
		}
	}

	var resources []map[string]interface{}

	// 处理每个API组中的资源
	for _, resourceList := range resourceLists {
		groupVersion := resourceList.GroupVersion
		for _, resource := range resourceList.APIResources {
			// 忽略子资源
			if len(resource.Group) == 0 {
				resource.Group = resourceList.GroupVersion
			}
			if len(resource.Version) == 0 {
				gv, err := schema.ParseGroupVersion(groupVersion)
				if err != nil {
					continue
				}
				resource.Version = gv.Version
			}

			// 根据命名空间范围过滤
			if (resource.Namespaced && !includeNamespaceScoped) || (!resource.Namespaced && !includeClusterScoped) {
				continue
			}

			resources = append(resources, map[string]interface{}{
				"name":         resource.Name,
				"singularName": resource.SingularName,
				"namespaced":   resource.Namespaced,
				"kind":         resource.Kind,
				"group":        resource.Group,
				"version":      resource.Version,
				"verbs":        resource.Verbs,
			})
		}
	}

	return resources, nil
}

// GetResource 获取特定资源的详细信息
func (c *Client) GetResource(ctx context.Context, kind, name, namespace string) (map[string]interface{}, error) {
	// 获取资源的 GVR
	gvr, err := c.findGroupVersionResource(kind)
	if err != nil {
		return nil, err
	}

	var obj *unstructured.Unstructured
	if namespace != "" {
		obj, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = c.dynamicClient.Resource(*gvr).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("获取资源失败: %w", err)
	}

	return obj.UnstructuredContent(), nil
}

// ListResources 列出某类资源的所有实例
func (c *Client) ListResources(ctx context.Context, kind, namespace string, labelSelector, fieldSelector string) ([]map[string]interface{}, error) {
	// 获取资源的 GVR
	gvr, err := c.findGroupVersionResource(kind)
	if err != nil {
		return nil, err
	}

	options := metav1.ListOptions{}
	if labelSelector != "" {
		options.LabelSelector = labelSelector
	}
	if fieldSelector != "" {
		options.FieldSelector = fieldSelector
	}

	var list *unstructured.UnstructuredList
	if namespace != "" {
		list, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).List(ctx, options)
	} else {
		list, err = c.dynamicClient.Resource(*gvr).List(ctx, options)
	}

	if err != nil {
		return nil, fmt.Errorf("列出资源失败: %w", err)
	}

	var resources []map[string]interface{}
	for _, item := range list.Items {
		resources = append(resources, item.UnstructuredContent())
	}

	return resources, nil
}

// CreateResource 创建一个新的资源
func (c *Client) CreateResource(ctx context.Context, kind, namespace string, manifest string) (map[string]interface{}, error) {
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal([]byte(manifest), &obj.Object); err != nil {
		return nil, fmt.Errorf("解析资源清单失败: %w", err)
	}

	// 获取资源的 GVR
	gvr, err := c.findGroupVersionResource(kind)
	if err != nil {
		return nil, err
	}

	var result *unstructured.Unstructured
	if namespace != "" || obj.GetNamespace() != "" {
		targetNamespace := namespace
		if targetNamespace == "" {
			targetNamespace = obj.GetNamespace()
		}
		result, err = c.dynamicClient.Resource(*gvr).Namespace(targetNamespace).Create(ctx, obj, metav1.CreateOptions{})
	} else {
		result, err = c.dynamicClient.Resource(*gvr).Create(ctx, obj, metav1.CreateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}

	return result.UnstructuredContent(), nil
}

// UpdateResource 更新现有资源
func (c *Client) UpdateResource(ctx context.Context, kind, name, namespace string, manifest string) (map[string]interface{}, error) {
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal([]byte(manifest), &obj.Object); err != nil {
		return nil, fmt.Errorf("解析资源清单失败: %w", err)
	}

	// 检查名称是否匹配
	if obj.GetName() != name {
		return nil, fmt.Errorf("资源清单中的名称 (%s) 与请求的名称 (%s) 不匹配", obj.GetName(), name)
	}

	// 获取资源的 GVR
	gvr, err := c.findGroupVersionResource(kind)
	if err != nil {
		return nil, err
	}

	var result *unstructured.Unstructured
	if namespace != "" {
		result, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	} else {
		result, err = c.dynamicClient.Resource(*gvr).Update(ctx, obj, metav1.UpdateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("更新资源失败: %w", err)
	}

	return result.UnstructuredContent(), nil
}

// DeleteResource 删除资源
func (c *Client) DeleteResource(ctx context.Context, kind, name, namespace string) error {
	// 获取资源的 GVR
	gvr, err := c.findGroupVersionResource(kind)
	if err != nil {
		return err
	}

	var deleteErr error
	if namespace != "" {
		deleteErr = c.dynamicClient.Resource(*gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		deleteErr = c.dynamicClient.Resource(*gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}

	if deleteErr != nil {
		return fmt.Errorf("删除资源失败: %w", deleteErr)
	}

	return nil
}

// findGroupVersionResource 根据 Kind 查找对应的 GroupVersionResource
func (c *Client) findGroupVersionResource(kind string) (*schema.GroupVersionResource, error) {
	// 获取集群中所有 API Groups 和 Resources
	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil {
		// 处理部分错误，有些资源可能无法访问
		if !discovery.IsGroupDiscoveryFailedError(err) {
			return nil, fmt.Errorf("获取 API 资源失败: %w", err)
		}
	}

	// 遍历所有 API 组和资源，查找指定的 Kind
	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range resourceList.APIResources {
			if resource.Kind == kind {
				return &schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: resource.Name,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("找不到资源类型 %s", kind)
}
