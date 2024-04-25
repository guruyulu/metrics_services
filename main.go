// main.go or another file
package main

import (
	"fmt"

	"github.com/guruyulu/metrics_services/services"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func main() {
	// Create a Prometheus API client
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090", // Prometheus server address
	})
	if err != nil {
		fmt.Printf("Failed to create Prometheus API client: %v\n", err)
		return
	}

	// Initialize the API client for Prometheus
	apiClient := services.NewPrometheusAPIClient(v1.NewAPI(client))

	// Initialize the metrics aggregator
	aggregator := services.NewMetricsAggregator()

	// Fetch and aggregate metrics
	totalMetrics, err := aggregator.AggregateMetrics("hello-app", "hello-app-namespace", apiClient)
	if err != nil {
		fmt.Printf("Error aggregating metrics: %v\n", err)
		return
	}

	// Print the total aggregated metrics
	fmt.Printf("Total aggregated metrics: %.2f\n", totalMetrics)
}
