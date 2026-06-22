package main

import (
	"fmt"
	"log"
	"net/http"

	"gitops/database"
	"gitops/handlers"
)

func main() {
	database.InitDB()
	defer database.DB.Close()

	database.StartBufferLogWorker()

	http.HandleFunc("/status", handlers.StatusHandler)
	http.HandleFunc("/", handlers.Root)

	fmt.Println("🚀 High-Performance Buffer Engine initiated! Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
