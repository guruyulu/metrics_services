package main

import (
	"log"
	"net/http"
	"time"

	"github.com/guruyulu/metrics_services/services"
)

func main() {
	services.Perform()

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			services.Perform()
		}
	}()

	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Printf("Failed to start Prometheus metrics server: %v\n", err)
	}

}
