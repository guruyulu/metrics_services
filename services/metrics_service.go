package services

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

)

// APIClient defines the methods required for interacting with the Prometheus API.
type APIClient interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, v1.Warnings, error)
}

// PrometheusAPIClient is an implementation of APIClient using the Prometheus client library.
type PrometheusAPIClient struct {
	api v1.API
}

// NewPrometheusAPIClient creates a new instance of PrometheusAPIClient.
func NewPrometheusAPIClient(apiClient v1.API) *PrometheusAPIClient {
	return &PrometheusAPIClient{api: apiClient}
}

// Query executes a Prometheus query.
func (c *PrometheusAPIClient) Query(ctx context.Context, query string, ts time.Time) (model.Value, v1.Warnings, error) {
	return c.api.Query(ctx, query, ts)
}

// FetchCPUMetrics fetches CPU metrics from Prometheus.
func FetchCPUMetrics(apiClient APIClient) (string, error) {
	query := `my_counter{instance="hello-app.hello-app-namespace.svc.cluster.local:80", job="hello-app"}`

	// Execute the query
	result, warnings, err := apiClient.Query(context.Background(), query, time.Now())
	if err != nil {
		return "", err
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Printf("Warning: %s\n", warning)
		}
	}

	// Process the query result
	if result.Type() == model.ValVector {
		vector := result.(model.Vector)
		// Process the vector and return the processed data as a string
		avgCPUUsage := calculateAvgCPUUsage(vector)
		return fmt.Sprintf("Average CPU Usage: %.2f", avgCPUUsage), nil
	}

	return "", fmt.Errorf("unexpected query result type: %s", result.Type())
}

func calculateAvgCPUUsage(vector model.Vector) float64 {
	sum := 0.0
	count := 0

	for _, sample := range vector {
		// Check if the sample value is a model.SampleValue
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

// FetchMemoryUsage fetches memory usage metrics.
func FetchMemoryUsage() (string, error) {
	// Implement logic to fetch memory usage from Prometheus
	return "Memory usage will be fetched here", nil
}

// FetchDBConnections fetches DB connection metrics.
func FetchDBConnections(job_name string, namespace string, apiClient APIClient) (string, error) {

	query := fmt.Sprintf(`database_connections{instance="%s.%s.svc.cluster.local:80", job="%s"}`, job_name, namespace, job_name)

	// Execute the query
	result, warnings, err := apiClient.Query(context.Background(), query, time.Now())
	if err != nil {
		return "", err
	}

	if len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Printf("Warning: %s\n", warning)
		}
	}

	// Process the query result
	if result.Type() == model.ValVector {
		vector := result.(model.Vector)
		// Process the vector and return the processed data as a string
		avgDbUsage := calculateAvgCPUUsage(vector)
		return fmt.Sprintf("Average DB Usage: %.2f", avgDbUsage), nil
	}

	return "", fmt.Errorf("unexpected query result type: %s", result.Type())
}
