package services


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

