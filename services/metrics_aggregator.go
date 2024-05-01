package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
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

// func (a *MetricsAggregator) AggregateMetrics(jobName, namespace string, apiClient APIClient) (float64, error) {
// 	metrics, err := FetchMetrics(jobName, namespace, apiClient)
// 	if err != nil {
// 		return 0.0, err
// 	}
// 	// Perform aggregation here, for example, adding CPU metrics and DB connections
// 	totalMetrics := metrics.CPUMetrics + metrics.DBConnections

// 	return totalMetrics, nil
// }

func Perform() {
	// Create a Prometheus API client
	// client, err := api.NewClient(api.Config{
	// 	// Address: "http://172.17.0.2:30841", // Prometheus server address if inside cluster
	// 	Address: "http://localhost:9090", // Prometheus server address if outside cluster
	// })
	// if err != nil {
	// 	fmt.Printf("Failed to create Prometheus API client: %v\n", err)
	// 	return
	// }

	// Initialize the API client for Prometheus
	// apiClient := NewPrometheusAPIClient(v1.NewAPI(client))

	// Initialize the metrics aggregator
	// aggregator := NewMetricsAggregator()

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

		if ns == "metrics-service-namespace" || ns == "default" || ns == "keda" || ns == "kube-node-lease" || ns == "kube-public" || ns == "kube-system" {
			continue
		}

		// Fetch services for the namespace
		servicesList, err := FetchServicesForNamespace(ns)
		if err != nil {
			fmt.Printf("Error fetching services for namespace %s: %v\n", ns, err)
			return
		}

		// Fetch labels for each service in the namespace
		for _, service := range servicesList {

			fmt.Println(service.Name, ns)

			// totalMetrics, err := aggregator.AggregateMetrics(service.Name, ns, apiClient)
			// if err != nil {
			// 	fmt.Printf("Error aggregating metrics: %v\n", err)
			// 	return
			// }

			// // Print the total aggregated metrics
			// fmt.Printf("Total aggregated metrics: %.2f\n\n\n", totalMetrics)

			labels, err := FetchLabelsForService(ns, service.Name)
			if err != nil {
				fmt.Printf("Error fetching labels for service %s in namespace %s: %v\n", service.Name, ns, err)
				return
			}

			var labelInfos []LabelInfo
			for key, value := range labels {
				labelInfo := LabelInfo{
					Key:   key,
					Value: value,
				}
				labelInfos = append(labelInfos, labelInfo)
			}
			scaleNeeded := false

			// if totalMetrics > 100 {
			// 	scaleNeeded = true
			// }

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
	jsonData, err := json.Marshal(serviceInfos)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	if err != nil {
		fmt.Printf("Error writing response: %v\n", err)
		return
	}

	// Send JSON data to Prometheus
	SendJSONToPrometheus(jsonData)

	// Print the JSON data
	fmt.Println(string(jsonData))
}

var jsonDataMetric prometheus.Gauge

func SendJSONToPrometheus(jsonData []byte) {
	// Check if the metric collector is already registered
	if jsonDataMetric == nil {
		// Create a gauge metric to store the presence of JSON data
		jsonDataMetric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "cluster_data",
				Help: "Cluster data for services",
				ConstLabels: prometheus.Labels{
					"json_data": string(jsonData),
				},
			},
		)
	}

	// Register the metric collector if it's not already registered
	if err := prometheus.Register(jsonDataMetric); err != nil {
		// Metric is already registered, no need to re-register
	}

	// Set the value to 1 indicating the presence of data
	jsonDataMetric.Set(1)
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	Perform()
}
