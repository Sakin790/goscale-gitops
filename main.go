package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gitops/database" 
)

type StatusResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func main() {
	database.InitDB()
	defer database.DB.Close()

	database.StartBufferLogWorker()

	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/", root)

	fmt.Println("🚀 High-Performance Buffer Engine initiated! Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to the API Server!"))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	
	if r.Body != nil {
		defer r.Body.Close()
	}

	now := time.Now()


	logItem := database.DbLog{
		Timestamp:     now,
		Method:        r.Method,
		Proto:         r.Proto,
		Path:          r.URL.Path,
		RemoteAddress: r.RemoteAddr,
	}

	database.LogQueue <- logItem

	w.Header().Set("Content-Type", "application/json")
	response := StatusResponse{
		Status:    "success",
		Message:   "API SERVER IS WORKING (BUFFERED)",
		Timestamp: now.Format("2006-01-02 15:04:05"),
	}
	json.NewEncoder(w).Encode(response)
}
