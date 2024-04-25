package services

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
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

type Metrics struct {
	CPUMetrics    float64 `json:"cpu_metrics"`
	MemoryUsage   string  `json:"memory_usage"`
	DBConnections float64 `json:"db_connections"`
}

func FetchMetrics(jobName, namespace string, apiClient APIClient) (Metrics, error) {
	var metrics Metrics

	cpuQuery := fmt.Sprintf(`my_counter{instance="%s.%s.svc.cluster.local:80", job="%s"}`, jobName, namespace, jobName)
	cpuResult, _, err := apiClient.Query(context.Background(), cpuQuery, time.Now())
	fmt.Println(cpuQuery)
	if err != nil {
		return metrics, err
	}
	metrics.CPUMetrics = calculateAvgCPUUsage(cpuResult.(model.Vector))

	metrics.MemoryUsage, err = FetchMemoryUsage()
	if err != nil {
		return metrics, err
	}

	dbQuery := fmt.Sprintf(`database_connections{instance="%s.%s.svc.cluster.local:80", job="%s"}`, jobName, namespace, jobName)
	dbResult, _, err := apiClient.Query(context.Background(), dbQuery, time.Now())
	if err != nil {
		return metrics, err
	}
	metrics.DBConnections = calculateDBConnections(dbResult.(model.Vector))

	return metrics, nil
}

func calculateAvgCPUUsage(vector model.Vector) float64 {
	var sum float64
	var count int

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

func calculateDBConnections(vector model.Vector) float64 {
	var sum float64
	var count int

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
