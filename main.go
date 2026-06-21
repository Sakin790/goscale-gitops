package main

import (
	"encoding/json"
	"fmt"
	"gitops/database"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Layer7Log struct {
	Timestamp     string              `yaml:"timestamp"`
	Method        string              `yaml:"method"`
	Proto         string              `yaml:"protocol"`
	Path          string              `yaml:"path"`
	RemoteAddress string              `yaml:"remote_address"`
	Headers       map[string][]string `yaml:"headers"`
	Body          string              `yaml:"body,omitempty"`
}

type StatusResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func main() {

	database.InitDB()
	defer database.DB.Close()

	
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/", root)

	fmt.Println("🚀 Advanced YAML logger initiated! Server running on http://localhost:8080")
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

	var bodyBytes []byte
	if r.Body != nil {
		// [FIX] রিকোয়েস্ট বডি রিসোর্স লিক এড়াতে ক্লোজ করা নিশ্চিত করুন
		defer r.Body.Close()
		bodyBytes, _ = io.ReadAll(r.Body)
	}

	now := time.Now()

	logItem := Layer7Log{
		Timestamp:     now.Format("2006-01-02 15:04:05.000"),
		Method:        r.Method,
		Proto:         r.Proto,
		Path:          r.URL.Path,
		RemoteAddress: r.RemoteAddr,
		Headers:       r.Header,
		Body:          string(bodyBytes),
	}

	saveLogToUniqueYAML(logItem, now)

	w.Header().Set("Content-Type", "application/json")
	response := StatusResponse{
		Status:    "success",
		Message:   "API SERVER IS WORKING",
		Timestamp: now.Format("2006-01-02 15:04:05"),
	}
	json.NewEncoder(w).Encode(response)
}

func saveLogToUniqueYAML(logItem Layer7Log, now time.Time) {
	folderPath := fmt.Sprintf("logs/%s", now.Format("2006-01-02"))
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		log.Printf("[ERROR] Failed to create directory: %v\n", err)
		return
	}

	fileName := fmt.Sprintf("%s/req_%s_%d.yml",
		folderPath,
		now.Format("150405"),
		now.Nanosecond(),
	)

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("[ERROR] Failed to create log file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	defer encoder.Close()

	if err := encoder.Encode(logItem); err != nil {
		log.Printf("[ERROR] Failed to encode YAML: %v\n", err)
		return
	}

	fmt.Printf("[📝 LOG SAVED] %s\n", fileName)
}
