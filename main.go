package main

import (
	"fmt"
	"net/http"

	"github.com/guruyulu/metrics_services/services"
)

func main() {
	// Register the Perform function as a handler for the /cluster_data route
	http.HandleFunc("/cluster_data", services.MetricsHandler)

	// Start the HTTP server to expose the /cluster_data route
	fmt.Println("Server listening on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
