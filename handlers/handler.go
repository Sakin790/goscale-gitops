package handlers

import (
	"encoding/json"
	"gitops/database"
	"gitops/utils"
	"net/http"
	"time"
)

func Root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to the API Server!"))
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
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
	response := utils.StatusResponse{
		Status:    "success",
		Message:   "API SERVER IS WORKING (BUFFERED)",
		Timestamp: now.Format("2006-01-02 15:04:05"),
	}
	json.NewEncoder(w).Encode(response)
}
