package services

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type APIClient interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, v1.Warnings, error)
}

type PrometheusAPIClient struct {
	api v1.API
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

func FetchMetrics(jobName, namespace string, apiClient APIClient) (Metrics, error) {
	var metrics Metrics
	var err error

	metrics.CPUMetrics, err = FetchCPUMetrics(jobName, namespace, apiClient)
	if err != nil {
		return metrics, err
	}

	metrics.MemoryUsage, err = FetchMemoryUsage(jobName, namespace, apiClient)
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

func FetchCPUMetrics(jobName, namespace string, apiClient APIClient) (float64, error) {
	cpuQuery := fmt.Sprintf(`avg_over_time(my_counter{instance="%s.%s.svc.cluster.local:80", job="%s"}[5m])`, jobName, namespace, jobName)
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

func FetchMemoryUsage(jobName, namespace string, apiClient APIClient) (float64, error) {
	memoryQuery := fmt.Sprintf(`database_connections{instance="%s.%s.svc.cluster.local:80", job="%s"}`, jobName, namespace, jobName)

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
