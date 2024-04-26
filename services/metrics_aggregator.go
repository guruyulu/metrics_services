package services

import (
	"context"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


type ServiceInfo struct {
	Name   string            // Service name
	Labels map[string]string // Labels associated with the service
}

// NewServiceInfo creates a new instance of ServiceInfo
func NewServiceInfo(name string, labels map[string]string) *ServiceInfo {
	return &ServiceInfo{
		Name:   name,
		Labels: labels,
	}
}

type MetricsAggregator struct {
}

func NewMetricsAggregator() *MetricsAggregator {
	return &MetricsAggregator{}
}

func (a *MetricsAggregator) AggregateMetrics(jobName, namespace string, apiClient APIClient) (float64, error) {
	metrics, err := FetchMetrics(jobName, namespace, apiClient)
	if err != nil {
		return 0.0, err
	}
	// Perform aggregation here, for example, adding CPU metrics and DB connections
	totalMetrics := metrics.CPUMetrics + metrics.DBConnections

	return totalMetrics, nil
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
