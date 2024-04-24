package services

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

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

func FetchCPUMetrics(apiClient APIClient) (string, error) {
	query := `my_counter{instance="hello-app.hello-app-namespace.svc.cluster.local:80", job="hello-app"}`

	
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
