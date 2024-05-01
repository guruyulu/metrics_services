package services

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type APIClient interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, v1.Warnings, error)
}

type PrometheusAPIClient struct {
	api v1.API
}

type PodInfo struct {
	Name         string
	CreationTime time.Time
	Age          time.Duration
	CPUUsage     float64
	MemoryUsage  float64
	Namespace    string
	Service      string
}

func NewPrometheusAPIClient(apiClient v1.API) *PrometheusAPIClient {
	return &PrometheusAPIClient{api: apiClient}
}

func (c *PrometheusAPIClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, v1.Warnings, error) {
	return c.api.Query(ctx, query, ts)
}

type Metrics struct {
	CPUMetrics    float64 `json:"cpu_metrics"`
	MemoryUsage   float64 `json:"memory_usage"`
	DBConnections float64 `json:"db_connections"`
}

func FetchMetrics(jobName, podName, namespace string, apiClient APIClient) (Metrics, error) {
	var metrics Metrics
	var err error

	metrics.CPUMetrics, err = FetchCPUMetrics(namespace, podName, apiClient)
	if err != nil {
		return metrics, err
	}

	metrics.MemoryUsage, err = FetchMemoryUsage(namespace, podName, apiClient)
	if err != nil {
		return metrics, err
	}

	metrics.DBConnections, err = FetchDBConnections(jobName, namespace, apiClient)
	if err != nil {
		return metrics, err
	}
	fmt.Println(metrics, "******")
	return metrics, nil
}

func FetchCPUMetrics(namespace, podName string, apiClient APIClient) (float64, error) {
	cpuQuery := fmt.Sprintf(`container_cpu_usage_seconds_total{namespace="%s", pod="%s"}`, namespace, podName)
	cpuResult, _, err := apiClient.Query(context.Background(), cpuQuery, time.Now())
	if err != nil {
		return 0.0, err
	}

	if cpuResult.Type() == model.ValVector {
		vector := cpuResult.(model.Vector)
		if len(vector) > 0 && !vector[0].Value.Equal(0) {
			return float64(vector[0].Value), nil
		}
	}

	return 0.0, fmt.Errorf("unexpected query result type")
}

func FetchMemoryUsage(namespace, podName string, apiClient APIClient) (float64, error) {
	memoryQuery := fmt.Sprintf(`container_memory_max_usage_bytes{namespace="%s", pod="%s"}`, namespace, podName)
	memoryResult, _, err := apiClient.Query(context.Background(), memoryQuery, time.Now())
	if err != nil {
		return 0.0, err
	}

	if memoryResult.Type() == model.ValVector {
		vector := memoryResult.(model.Vector)
		if len(vector) > 0 && !vector[0].Value.Equal(0) {
			return float64(vector[0].Value), nil
		}
	}

	return 0.0, fmt.Errorf("unexpected query result type")
}

func FetchDBConnections(jobName, namespace string, apiClient APIClient) (float64, error) {
	dbQuery := fmt.Sprintf(`database_connections{instance="%s.%s.svc.cluster.local:80", job="%s"}`, jobName, namespace, jobName)
	dbResult, _, err := apiClient.Query(context.Background(), dbQuery, time.Now())
	if err != nil {
		return 0.0, err
	}

	if dbResult.Type() == model.ValVector {
		vector := dbResult.(model.Vector)
		if len(vector) > 0 && !vector[0].Value.Equal(0) {
			return float64(vector[0].Value), nil
		}
	}

	return 0.0, fmt.Errorf("unexpected query result type")
}

// PrintReplicasAndDuration fetches pod information and returns a slice of PodInfo structs
func PrintReplicasAndDuration(namespace string) ([]PodInfo, error) {

	client, err := api.NewClient(api.Config{
		// Address: "http://172.17.0.2:30841", // Prometheus server address if inside cluster
		Address: "http://localhost:9090", // Prometheus server address if outside cluster
	})
	if err != nil {
		fmt.Printf("Failed to create Prometheus API client: %v\n", err)
	}

	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	fmt.Printf("Number of replicas: %d\n", len(deployments.Items))

	var podInfoList []PodInfo

	for _, deployment := range deployments.Items {
		fmt.Printf("Deployment: %s\n", deployment.Name)
		fmt.Printf("Replicas: %d\n", *deployment.Spec.Replicas)
		fmt.Printf("Creation time: %s\n", deployment.CreationTimestamp)

		// List pods of the deployment
		pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", deployment.Name)})
		if err != nil {
			fmt.Printf("Error fetching pods for deployment %s: %v\n", deployment.Name, err)
			continue
		}

		// Print pod names
		fmt.Println("Pods:")

		for _, pod := range pods.Items {
			// Calculate pod age
			creationTime := pod.CreationTimestamp.Time
			age := time.Since(creationTime)

			// Create Prometheus API client
			apiClient := NewPrometheusAPIClient(v1.NewAPI(client))

			// Fetch CPU and memory metrics for the pod
			metrics, err := FetchMetrics(deployment.Name, pod.Name, namespace, apiClient)
			if err != nil {
				fmt.Printf("Error fetching metrics for pod %s: %v\n", pod.Name, err)
				continue
			}

			// Add pod information to the list
			podInfo := PodInfo{
				Name:         pod.Name,
				CreationTime: creationTime,
				Age:          age,
				CPUUsage:     metrics.CPUMetrics,
				MemoryUsage:  metrics.MemoryUsage,
				Namespace:    namespace,
				Service:      deployment.Name,
			}
			podInfoList = append(podInfoList, podInfo)
		}
	}

	fmt.Println("Pods Details:")
	for _, podInfo := range podInfoList {
		fmt.Printf("Name: %s\n", podInfo.Name)
		fmt.Printf("Creation Time: %s\n", podInfo.CreationTime)
		fmt.Printf("Age: %s\n", podInfo.Age)
		fmt.Printf("CPU Usage: %.2f\n", podInfo.CPUUsage)
		fmt.Printf("Memory Usage: %.2f\n", podInfo.MemoryUsage)
		fmt.Printf("Namespace: %s\n", podInfo.Namespace)
		fmt.Printf("Service: %s\n", podInfo.Service)
		fmt.Println("--------------------------------------\n\n")
	}
	return podInfoList, nil
}

// Function to get the cluster name from the Kubernetes API server
// Function to get the cluster name from the Kubernetes configuration
func getClusterName(config *rest.Config) (string, error) {
	if config == nil {
		return "", fmt.Errorf("Kubernetes configuration is nil")
	}

	if config.Host == "" {
		return "", fmt.Errorf("Kubernetes configuration host is empty")
	}

	// Extract the cluster name from the Kubernetes configuration host URL
	// For example, if the URL is "https://kubernetes.default.svc", the cluster name would be "kubernetes"
	clusterURL, err := url.Parse(config.Host)
	if err != nil {
		return "", fmt.Errorf("Failed to parse Kubernetes host URL: %v", err)
	}

	clusterName := strings.TrimSuffix(clusterURL.Hostname(), ".")

	return clusterName, nil
}

// getConfig returns Kubernetes config
func getConfig() (*rest.Config, error) {
	// Check if running inside Kubernetes cluster
	if _, err := rest.InClusterConfig(); err != nil {
		// Not running inside a Kubernetes cluster, use out-of-cluster configuration
		kubeConfigPath := os.Getenv("KUBECONFIG")
		if kubeConfigPath == "" {
			return nil, fmt.Errorf("KUBECONFIG environment variable is not set")
		}
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}
	// Running inside a Kubernetes cluster, use in-cluster configuration
	return rest.InClusterConfig()
}

// ==================================== if running outside cluser ==================================

// FetchNamespacesFromKubernetes fetches namespaces from Kubernetes
func FetchNamespacesFromKubernetes() ([]string, error) {
	var config *rest.Config
	var err error

	// Check if running inside Kubernetes cluster
	if _, err := rest.InClusterConfig(); err != nil {
		// Not running inside a Kubernetes cluster, use out-of-cluster configuration
		kubeConfigPath := os.Getenv("KUBECONFIG")
		if kubeConfigPath == "" {
			return nil, fmt.Errorf("KUBECONFIG environment variable is not set")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		// Running inside a Kubernetes cluster, use in-cluster configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	// fmt.Println("Namespace list : "fetchNamespaces(clientset))
	return fetchNamespaces(clientset)
}

// fetchNamespaces fetches namespaces using the provided clientset
func fetchNamespaces(clientset *kubernetes.Clientset) ([]string, error) {
	// Fetch namespaces
	nsList, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Extract namespace names
	namespaces := make([]string, len(nsList.Items))
	for i, ns := range nsList.Items {
		_, err := PrintReplicasAndDuration(ns.Name)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		namespaces[i] = ns.Name
	}

	return namespaces, nil
}

// Function to fetch labels for a given namespace from the Kubernetes API server
func FetchLabelsForNamespace(namespace string) (map[string]string, error) {
	// Check if the KUBECONFIG environment variable is set
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		return nil, fmt.Errorf("KUBECONFIG environment variable is not set")
	}

	// Create a Kubernetes client using the specified kubeconfig path
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Fetch namespace
	ns, err := clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Extract labels
	labels := ns.Labels

	return labels, nil
}

// FetchServicesForNamespace fetches services along with their respective labels for a given namespace
func FetchServicesForNamespace(namespace string) ([]*ServiceInfo, error) {
	// Check if the KUBECONFIG environment variable is set
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		return nil, fmt.Errorf("KUBECONFIG environment variable is not set")
	}

	// Create a Kubernetes client using the specified kubeconfig path
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Fetch services in the namespace
	serviceList, err := clientset.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Construct slice with service names and labels
	var services []*ServiceInfo
	for _, service := range serviceList.Items {
		// Collect service name and labels
		labels, err := FetchLabelsForService(namespace, service.Name)
		if err != nil {
			return nil, err
		}
		services = append(services, NewServiceInfo(service.Name, labels))
	}

	return services, nil
}

// FetchLabelsForService fetches labels for a given service in the specified namespace
func FetchLabelsForService(namespace, serviceName string) (map[string]string, error) {
	// Check if the KUBECONFIG environment variable is set
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		return nil, fmt.Errorf("KUBECONFIG environment variable is not set")
	}

	// Create a Kubernetes client using the specified kubeconfig path
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Fetch service in the namespace
	service, err := clientset.CoreV1().Services(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Extract labels
	labels := service.Labels

	return labels, nil
}

// ===================================== if running inside cluser =======================================
// FetchNamespacesFromKubernetes fetches namespaces from Kubernetes
// func FetchNamespacesFromKubernetes() ([]string, error) {
// 	// Get the in-cluster configuration
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create the Kubernetes client
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Fetch namespaces
// 	nsList, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Extract namespace names
// 	namespaces := make([]string, len(nsList.Items))
// 	for i, ns := range nsList.Items {
// 		namespaces[i] = ns.Name
// 	}

// 	return namespaces, nil
// }

// // Function to fetch labels for a given namespace from the Kubernetes API server
// func FetchLabelsForNamespace(namespace string) (map[string]string, error) {
// 	// Get the in-cluster configuration
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create the Kubernetes client
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Fetch namespace
// 	ns, err := clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Extract labels
// 	labels := ns.Labels

// 	return labels, nil
// }

// // FetchServicesForNamespace fetches services along with their respective labels for a given namespace
// func FetchServicesForNamespace(namespace string) ([]*ServiceInfo, error) {
// 	// Get the in-cluster configuration
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create the Kubernetes client
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Fetch services in the namespace
// 	serviceList, err := clientset.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Construct slice with service names and labels
// 	var services []*ServiceInfo
// 	for _, service := range serviceList.Items {
// 		// Collect service name and labels
// 		labels, err := FetchLabelsForService(namespace, service.Name)
// 		if err != nil {
// 			return nil, err
// 		}
// 		services = append(services, NewServiceInfo(service.Name, labels))
// 	}

// 	return services, nil
// }

// // FetchLabelsForService fetches labels for a given service in the specified namespace
// func FetchLabelsForService(namespace, serviceName string) (map[string]string, error) {
// 	// Get the in-cluster configuration
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create the Kubernetes client
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Fetch service in the namespace
// 	service, err := clientset.CoreV1().Services(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Extract labels
// 	labels := service.Labels

// 	return labels, nil
// }
