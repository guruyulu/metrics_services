package services

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	cpuUsageMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_cpu_usage",
			Help: "Average CPU Usage per Service",
		},
		[]string{"namespace", "service"},
	)
	memoryUsageMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_memory_usage",
			Help: "Average Memory Usage per Service",
		},
		[]string{"namespace", "service"},
	)
	scaleMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_scaling_needed",
			Help: "Scaling Needed per Service",
		},
		[]string{"namespace", "service"},
	)

	once sync.Once
)

func init() {
	// Register the metrics with Prometheus
	prometheus.MustRegister(cpuUsageMetric)
	prometheus.MustRegister(memoryUsageMetric)
	prometheus.MustRegister(scaleMetric)
}

type ServiceStats struct {
	AverageCPU    float64
	AverageMemory float64
	PodCount      int
	IsScale       int // 1 if scaling needed, 0 otherwise
}

func Compute(podData []PodInfo) map[string]map[string]ServiceStats {
	clusterData := TransformToPodMap(podData)
	serviceStats := make(map[string]map[string]ServiceStats)

	fmt.Println("===== INFO =====")

	for namespace, services := range clusterData {
		serviceStats[namespace] = make(map[string]ServiceStats)

		fmt.Println("Namespace:", namespace)

		for service, pods := range services {
			fmt.Println("\tService:", service)

			totalCPU, totalMemory := 0.0, 0.0
			podCount := len(pods)

			for _, pod := range pods {
				totalCPU += pod.CPUUsage
				totalMemory += pod.MemoryUsage
			}

			avgCPU := totalCPU / float64(podCount)
			avgMemory := totalMemory / float64(podCount)

			// Check if scaling is needed
			var isScale int
			if avgCPU > 60 && avgMemory > 60 {
				isScale = 1
			}

			serviceStats[namespace][service] = ServiceStats{
				AverageCPU:    avgCPU,
				AverageMemory: avgMemory,
				PodCount:      podCount,
				IsScale:       isScale,
			}

			fmt.Println("\t\tAverage CPU Usage:", avgCPU)
			fmt.Println("\t\tAverage Memory Usage:", avgMemory)
		}
	}

	return serviceStats
}

func ComputeScale(podData []PodInfo) {
	// Call the Compute function to get service stats
	serviceStats := Compute(podData)

	// Iterate over serviceStats and update Prometheus metrics
	for namespace, services := range serviceStats {
		for service, stats := range services {
			// Update CPU Usage metric
			cpuUsageMetric.WithLabelValues(namespace, service).Set(stats.AverageCPU)

			// Update Memory Usage metric
			memoryUsageMetric.WithLabelValues(namespace, service).Set(stats.AverageMemory)

			// Update Scaling Needed metric
			scaleMetric.WithLabelValues(namespace, service).Set(float64(stats.IsScale))
		}
	}
}
