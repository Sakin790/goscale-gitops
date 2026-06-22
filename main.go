package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gitops/database"
	"gitops/handlers"
)

func main() {
	database.InitDB()
	defer database.DB.Close()

	database.StartBufferLogWorker()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := ":" + port

	http.HandleFunc("/status", handlers.StatusHandler)
	http.HandleFunc("/", handlers.Root)

	fmt.Printf("🚀🌐 High-Performance Buffer Engine initiated! Server running on http://localhost%s\n", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
