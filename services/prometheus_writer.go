package services

import (
	"fmt"
)

func Compute(podData []PodInfo) {
	fmt.Println("===== INFO =====")
	for _, pod := range podData {
		fmt.Println("Name:", pod.Name)
		fmt.Println("CreationTime:", pod.CreationTime)
		fmt.Println("Age:", pod.Age)
		fmt.Println("CPUUsage:", pod.CPUUsage)
		fmt.Println("MemoryUsage:", pod.MemoryUsage)
		fmt.Println("Namespace:", pod.Namespace)
		fmt.Println("Service:", pod.Service)
		fmt.Println("Labels:")
		for key, value := range pod.Labels {
			fmt.Printf("\t%s: %s\n", key, value)
		}
		fmt.Println("==================================")
	}
}

func ComputeScale(podData []PodInfo) {
	Compute(podData)

}
