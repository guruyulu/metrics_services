package controllers

import (
	"net/http"

	"github.com/guruyulu/metrics_services/services"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func GetCPUMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse URL query parameters
	query := r.URL.Query()
	namespace := query.Get("namespace")
	jobName := query.Get("job_name")

	// Create a Prometheus API client
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090", // Prometheus server address
	})
	if err != nil {
		http.Error(w, "Failed to create Prometheus API client", http.StatusInternalServerError)
		return
	}

	// Create a Prometheus API client service
	apiClient := services.NewPrometheusAPIClient(v1.NewAPI(client))

	// Fetch CPU metrics for the specified job and namespace
	result, err := services.FetchCPUMetrics(jobName, namespace, apiClient)
	if err != nil {
		http.Error(w, "Failed to fetch CPU metrics", http.StatusInternalServerError)
		return
	}

	// Respond with the fetched data
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func GetMemoryUsageHandler(w http.ResponseWriter, r *http.Request) {

	result, err := services.FetchMemoryUsage()
	if err != nil {
		http.Error(w, "Failed to fetch memory usage", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func GetDBConnectionsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	namespace := query.Get("namespace")
	jobName := query.Get("job_name")

	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090",
	})
	if err != nil {
		http.Error(w, "Failed to create Prometheus API client", http.StatusInternalServerError)
		return
	}

	apiClient := services.NewPrometheusAPIClient(v1.NewAPI(client))

	result, err := services.FetchDBConnections(jobName, namespace, apiClient)
	if err != nil {
		http.Error(w, "Failed to fetch DB connections", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}
