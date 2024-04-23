package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/guruyulu/metrics_services/controllers"
)

func main() {
	// Create a new HTTP router
	router := mux.NewRouter()

	// Setup routes using the handlers package
	SetupRoutes(router)

	// Start the HTTP server
	fmt.Println("Starting the Metrics service...")
	http.ListenAndServe(":8080", router)
}

func SetupRoutes(router *mux.Router) {
	// Map GET CPU metrics endpoint to controller handler
	router.HandleFunc("/cpu-metrics", controllers.GetCPUMetricsHandler).Methods("GET")

	// Map GET Memory usage endpoint to controller handler
	router.HandleFunc("/memory-usage", controllers.GetMemoryUsageHandler).Methods("GET")

	// Map GET DB connections endpoint to controller handler
	router.HandleFunc("/db-connections", controllers.GetDBConnectionsHandler).Methods("GET")
}
