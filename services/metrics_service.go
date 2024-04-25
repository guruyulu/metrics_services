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
