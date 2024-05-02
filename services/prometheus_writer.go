package services

import (
	"fmt"
)

func Compute(podData []PodInfo) {
	clusterData := TransformToPodMap(podData)
	fmt.Println("===== INFO =====")
	// Assume clusterData is a map[string]map[string][]PodInfo
	for namespace, services := range clusterData {
		fmt.Println("Namespace:", namespace)
		for service, pods := range services {
			fmt.Println("\tService:", service)
			for _, pod := range pods {
				fmt.Println("\t\tPod:", pod.Name)
				fmt.Println("\t\t\tCreation Time:", pod.CreationTime)
				fmt.Println("\t\t\tAge:", pod.Age)
				fmt.Println("\t\t\tCPU Usage:", pod.CPUUsage)
				fmt.Println("\t\t\tMemory Usage:", pod.MemoryUsage)
				fmt.Println("\t\t\tDatabase Connections:", pod.DatabaseConnections)
				fmt.Println("\t\t\tStatus:", pod.Status)
				fmt.Println("\t\t\tLabels:")
				for _, label := range pod.Labels {
					fmt.Println("\t\t\t\tKey:", label.Key)
					fmt.Println("\t\t\t\tValue:", label.Value)
				}
			}
		}
	}

}

func ComputeScale(podData []PodInfo) {
	Compute(podData)

}
