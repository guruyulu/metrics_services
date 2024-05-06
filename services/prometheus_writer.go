package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

func Compute(podData []PodInfo, clientset *kubernetes.Clientset) map[string]map[string]ServiceStats {
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

			var actions []string

			fmt.Println("No. of pods in %s : %s", namespace, podCount)
			if avgCPU > 60 && avgMemory > 60 {
				actions = append(actions, "scaleUp")
				err := scaleDeployment(clientset, model.LabelValue(namespace), model.LabelValue(service), int32(podCount)+1)
				if err != nil {
					log.Printf("Error scaling up deployment %s in namespace %s: %v\n", service, namespace, err)
				} else {
					fmt.Printf("Scaled up deployment %s in namespace %s\n", service, namespace)
				}
			} else {
				actions = append(actions, "scaleUp")
				err := scaleDeployment(clientset, model.LabelValue(namespace), model.LabelValue(service), int32(podCount)+1)
				if err != nil {
					log.Printf("Error scaling up deployment %s in namespace %s: %v\n", service, namespace, err)
				} else {
					fmt.Printf("Scaled up deployment %s in namespace %s\n", service, namespace)
				}
			}

			serviceStats[namespace][service] = ServiceStats{
				AverageCPU:    avgCPU,
				AverageMemory: avgMemory,
				PodCount:      podCount,
				IsScale:       1,
			}
			fmt.Println("\t\tScale needed:", serviceStats)
			fmt.Println("\t\tAverage CPU Usage:", avgCPU)
			fmt.Println("\t\tAverage Memory Usage:", avgMemory)
		}
	}

	return serviceStats
}

func scaleDeployment(clientset *kubernetes.Clientset, namespace model.LabelValue, deploymentName model.LabelValue, replicas int32) error {
	fmt.Println("replica inside scale : ", replicas)
	deploymentsClient := clientset.AppsV1().Deployments(string(namespace))

	var deployment *appsv1.Deployment
	var err error

	for attempts := 0; attempts < 3; attempts++ {
		deployment, err = deploymentsClient.Get(context.TODO(), string(deploymentName), metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return fmt.Errorf("Deployment %s not found in namespace %s", deploymentName, namespace)
			}
			return err
		}

		deployment.Spec.Replicas = &replicas
		_, err = deploymentsClient.Update(context.TODO(), deployment, metav1.UpdateOptions{})
		if err == nil {
			return nil
		}

		// Sleep for a short duration before retrying
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("Failed to update deployment %s in namespace %s after 3 attempts", deploymentName, namespace)
}

func ComputeScale(podData []PodInfo) {
	kubeconfig := "/Users/raushan/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	// Call the Compute function to get service stats
	serviceStats := Compute(podData, clientset)

	// Iterate over serviceStats and update Prometheus metrics
	for namespace, services := range serviceStats {
		for service, stats := range services {
			// Update CPU Usage metric
			cpuUsageMetric.WithLabelValues(namespace, service).Set(stats.AverageCPU)

			// Update Memory Usage metric
			memoryUsageMetric.WithLabelValues(namespace, service).Set(stats.AverageMemory)

			// Update Scaling Needed metric
			scaleMetric.WithLabelValues(namespace, service).Set(float64(stats.IsScale))

			// Convert namespace and service to model.LabelValue
			ns := model.LabelValue(namespace)
			srv := model.LabelValue(service)

			// Scale deployment
			if err := scaleDeployment(clientset, ns, srv, 2); err != nil {
				log.Printf("Error scaling deployment: %v", err)
			}
		}
	}
}
