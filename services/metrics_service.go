package services

import (
	"context"
	"fmt"
	"time"
	"os"
	"errors"

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

func FetchCPUMetrics(jobName string, namespace string, apiClient APIClient) (string, error) {
	query := fmt.Sprintf(`my_counter{instance="%s.%s.svc.cluster.local:80", job="%s"}`, jobName, namespace, jobName)

	result, warnings, err := apiClient.Query(context.Background(), query, time.Now())
	if err != nil {
		return "", err
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Printf("Warning: %s\n", warning)
		}
	}

	if result.Type() == model.ValVector {
		vector := result.(model.Vector)
		avgCPUUsage := calculateAvgCPUUsage(vector)
		return fmt.Sprintf("Average CPU Usage: %.2f", avgCPUUsage), nil
	}

	return "", fmt.Errorf("unexpected query result type: %s", result.Type())
}

func calculateAvgCPUUsage(vector model.Vector) float64 {
	sum := 0.0
	count := 0

	for _, sample := range vector {
		if !sample.Value.Equal(0) {
			sum += float64(sample.Value)
			count++
		}
	}

	if count > 0 {
		return sum / float64(count)
	}

	return 0.0
}

func calculateDB_Connections(vector model.Vector) float64 {
	sum := 0.0
	count := 0

	for _, sample := range vector {
		if !sample.Value.Equal(0) {
			sum += float64(sample.Value)
			count++
		}
	}

	if count > 0 {

		return sum / float64(count)
	}

	return 0.0
}

func FetchMemoryUsage() (string, error) {
	return "Memory usage will be fetched here", nil
}

func FetchDBConnections(job_name string, namespace string, apiClient APIClient) (string, error) {

	query := fmt.Sprintf(`database_connections{instance="%s.%s.svc.cluster.local:80", job="%s"}`, job_name, namespace, job_name)

	result, warnings, err := apiClient.Query(context.Background(), query, time.Now())
	if err != nil {
		return "", err
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Printf("Warning: %s\n", warning)
		}
	}

	if result.Type() == model.ValVector {
		vector := result.(model.Vector)
		avgDbUsage := calculateDB_Connections(vector)
		return fmt.Sprintf("Average DB Usage: %.2f", avgDbUsage), nil
	}

	return "", fmt.Errorf("unexpected query result type: %s", result.Type())
}

func FetchNamespacesFromKubernetes() ([]string, error) {
	var config *rest.Config
	var err error

	// Check if running inside Kubernetes cluster
	if _, err := rest.InClusterConfig(); err != nil {
		// Not running inside a Kubernetes cluster, use out-of-cluster configuration
		kubeConfigPath := os.Getenv("KUBECONFIG")
		if kubeConfigPath == "" {
			return nil, errors.New("KUBECONFIG environment variable is not set")
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
