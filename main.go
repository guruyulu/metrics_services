package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/guruyulu/metrics_services/services"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	services.Perform()

	// Expose Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Prometheus metrics server started at :8082/metrics")

	// Start HTTP server

	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Printf("Failed to start Prometheus metrics server: %v\n", err)
	}

}
