package main

import (
	"fmt"

	"github.com/guruyulu/metrics_services/services"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// ServiceNamespaceInfo holds information about services and their labels for a namespace
type ServiceNamespaceInfo struct {
	Namespace string                    
	Services  []services.ServiceInfo
}

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
	fmt.Printf("Total aggregated metrics: %.2f\n\n\n", totalMetrics)

	// Fetch namespaces
	namespaces, err := services.FetchNamespacesFromKubernetes()
	if err != nil {
		fmt.Println("Error fetching namespaces:", err)
		return
	}

	// Create a map to hold services and labels for each namespace
	namespaceServices := make(map[string]ServiceNamespaceInfo)


	// Iterate over each namespace
	for _, ns := range namespaces {
		// Fetch services for the namespace
		servicesList, err := services.FetchServicesForNamespace(ns)
		if err != nil {
			fmt.Printf("Error fetching services for namespace %s: %v\n", ns, err)
			continue
		}

		// Create a slice to hold services and their labels
		var serviceInfos []services.ServiceInfo

		// Fetch labels for each service in the namespace
		for _, service := range servicesList {
			labels, err := services.FetchLabelsForService(ns, service.Name)
			if err != nil {
				fmt.Printf("Error fetching labels for service %s in namespace %s: %v\n", service.Name, ns, err)
				continue
			}
			serviceInfo := services.ServiceInfo{Name: service.Name, Labels: labels}
			serviceInfos = append(serviceInfos, serviceInfo)
		}

		// Store services and labels for the namespace
		namespaceInfo := ServiceNamespaceInfo{Namespace: ns, Services: serviceInfos}
		namespaceServices[ns] = namespaceInfo

		// Print services and labels for the namespace
		fmt.Printf("Namespace: %s\n", ns)
		for _, service := range serviceInfos {
			fmt.Printf("Service: %s, Labels: %+v\n", service.Name, service.Labels)
		}
		fmt.Println()
	}

	// Print the map containing services and labels for each namespace
	fmt.Println("Namespace Services and Labels Map:")
	for ns, info := range namespaceServices {
		fmt.Printf("Namespace: %s\n", ns)
		for _, service := range info.Services {
			fmt.Printf("Service: %s, Labels: %+v\n", service.Name, service.Labels)
		}
		fmt.Println()
	}
}
