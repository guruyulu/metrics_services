// metrics_aggregator.go
package services

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
