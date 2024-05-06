package services

import (
	"fmt"
)

func Perform() {
	namespaces, err := FetchNamespacesFromKubernetes()
	if err != nil {
		fmt.Println("Error fetching namespaces:", err)
		return
	}

	// Iterate over each namespace
	for _, ns := range namespaces {
		if ns == "metrics-service-namespace" || ns == "default" || ns == "keda" || ns == "kube-node-lease" || ns == "kube-public" || ns == "kube-system" {
			continue
		}
		podData, err := PrintReplicasAndDuration(ns)
		ComputeScale(podData)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}
