package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/guruyulu/metrics_services/controllers"
)

func main() {
	router := mux.NewRouter()

	SetupRoutes(router)

	fmt.Println("Starting the Metrics service...")
	http.ListenAndServe(":8080", router)
}

func SetupRoutes(router *mux.Router) {
	
	router.HandleFunc("/cpu-metrics", controllers.GetCPUMetricsHandler).Methods("GET")

	router.HandleFunc("/memory-usage", controllers.GetMemoryUsageHandler).Methods("GET")

	router.HandleFunc("/db-connections", controllers.GetDBConnectionsHandler).Methods("GET")
}
