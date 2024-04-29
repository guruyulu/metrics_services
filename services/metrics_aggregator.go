package services

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// FormatedServiceInfo holds information about a service and its labels
type FormatedServiceInfo struct {
	Namespace   string      `json:"namespace"`   // Namespace name
	Service     string      `json:"service"`     // Service name
	Labels      []LabelInfo `json:"labels"`      // Slice of LabelInfo structs
	ScaleNeeded bool        `json:"scaleNeeded"` // Indicates if scale is needed
}

// LabelInfo holds information about a label
type LabelInfo struct {
	Key   string `json:"key"`   // Label key
	Value string `json:"value"` // Label value
}

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

func Perform() {
	// Create a Prometheus API client
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090", // Prometheus server address
	})
	if err != nil {
		fmt.Printf("Failed to create Prometheus API client: %v\n", err)
		return
	}

	// Initialize the API client for Prometheus
	apiClient := NewPrometheusAPIClient(v1.NewAPI(client))

	// Initialize the metrics aggregator
	aggregator := NewMetricsAggregator()

	// Fetch and aggregate metrics
	totalMetrics, err := aggregator.AggregateMetrics("hello-app", "hello-app-namespace", apiClient)
	if err != nil {
		fmt.Printf("Error aggregating metrics: %v\n", err)
		return
	}

	// Print the total aggregated metrics
	fmt.Printf("Total aggregated metrics: %.2f\n\n\n", totalMetrics)

	// Fetch namespaces
	namespaces, err := FetchNamespacesFromKubernetes()
	if err != nil {
		fmt.Println("Error fetching namespaces:", err)
		return
	}

	// Create a slice to hold service info
	var serviceInfos []FormatedServiceInfo

	// Iterate over each namespace
	for _, ns := range namespaces {
		// Fetch services for the namespace
		servicesList, err := FetchServicesForNamespace(ns)
		if err != nil {
			fmt.Printf("Error fetching services for namespace %s: %v\n", ns, err)
			continue
		}

		// Fetch labels for each service in the namespace
		for _, service := range servicesList {
			labels, err := FetchLabelsForService(ns, service.Name)
			if err != nil {
				fmt.Printf("Error fetching labels for service %s in namespace %s: %v\n", service.Name, ns, err)
				continue
			}

			var labelInfos []LabelInfo
			for key, value := range labels {
				labelInfo := LabelInfo{
					Key:   key,
					Value: value,
				}
				labelInfos = append(labelInfos, labelInfo)
			}

			// Set the scaleNeeded field based on some condition
			scaleNeeded := false // Example condition, modify as per your requirement

			serviceInfo := FormatedServiceInfo{
				Namespace:   ns,
				Service:     service.Name,
				Labels:      labelInfos,
				ScaleNeeded: scaleNeeded,
			}
			serviceInfos = append(serviceInfos, serviceInfo)
		}
	}

	// Convert the service info slice to JSON
	jsonData, err := json.MarshalIndent(serviceInfos, "", "    ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	// Print the JSON data
	fmt.Println(string(jsonData))
}
