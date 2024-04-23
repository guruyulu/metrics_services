package controllers

import (
	"net/http"

	"github.com/guruyulu/metrics_services/services"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func GetCPUMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Configure Prometheus API client
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090", // Prometheus server address
	})
	if err != nil {
		http.Error(w, "Failed to create Prometheus API client", http.StatusInternalServerError)
		return
	}

	// Create an instance of PrometheusAPIClient
	apiClient := services.NewPrometheusAPIClient(v1.NewAPI(client))

	// Delegate the logic to services
	result, err := services.FetchCPUMetrics(apiClient)
	if err != nil {
		http.Error(w, "Failed to fetch CPU metrics", http.StatusInternalServerError)
		return
	}
	// Write response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func GetMemoryUsageHandler(w http.ResponseWriter, r *http.Request) {
	// Delegate the logic to services
	result, err := services.FetchMemoryUsage()
	if err != nil {
		http.Error(w, "Failed to fetch memory usage", http.StatusInternalServerError)
		return
	}
	// Write response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func GetDBConnectionsHandler(w http.ResponseWriter, r *http.Request) {
	// Configure Prometheus API client
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090", // Replace with your Prometheus server address
	})
	if err != nil {
		http.Error(w, "Failed to create Prometheus API client", http.StatusInternalServerError)
		return
	}

	// Create an instance of PrometheusAPIClient
	apiClient := services.NewPrometheusAPIClient(v1.NewAPI(client))

	// Delegate the logic to services
	result, err := services.FetchDBConnections(apiClient)
	if err != nil {
		http.Error(w, "Failed to fetch DB connections", http.StatusInternalServerError)
		return
	}
	// Write response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}
