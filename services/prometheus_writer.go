package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/common/model"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func decisionHandler(podData []PodInfo, clientset *kubernetes.Clientset) {
	clusterData := TransformToPodMap(podData)

	fmt.Println("===== INFO =====")

	for namespace, services := range clusterData {

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

			if avgCPU > 60 && avgMemory > 60 {
				if podCount == 10 {
					continue
				}
				err := scaleDeployment(clientset, model.LabelValue(namespace), model.LabelValue(service), int32(podCount)+1)
				if err != nil {
					log.Printf("Error scaling up deployment %s in namespace %s: %v\n", service, namespace, err)
				} else {
					fmt.Printf("Scaled up deployment %s in namespace %s\n", service, namespace)
				}
			} else {
				// if podCount == 1 {
				// 	continue
				// }
				err := scaleDeployment(clientset, model.LabelValue(namespace), model.LabelValue(service), int32(podCount)+1)
				if err != nil {
					log.Printf("Error scaling up deployment %s in namespace %s: %v\n", service, namespace, err)
				} else {
					fmt.Printf("Scaled up deployment %s in namespace %s\n", service, namespace)
				}
			}
		}
	}
}

func scaleDeployment(clientset *kubernetes.Clientset, namespace model.LabelValue, deploymentName model.LabelValue, replicas int32) error {
	fmt.Println("replica should be scale to: ", replicas)
	deploymentsClient := clientset.AppsV1().Deployments(string(namespace))

	var deployment *appsv1.Deployment
	var err error

	for attempts := 0; attempts < 3; attempts++ {
		deployment, err = deploymentsClient.Get(context.TODO(), string(deploymentName), metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return fmt.Errorf("deployment %s not found in namespace %s", deploymentName, namespace)
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

	return fmt.Errorf("failed to update deployment %s in namespace %s after 3 attempts", deploymentName, namespace)
}

func Scale(podData []PodInfo) {
	kubeconfig := "/Users/guru/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	// Call the Compute function to get service stats
	decisionHandler(podData, clientset)
}
