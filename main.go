package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

type LogData struct {
	Timestamp time.Time
	Method    string
	Path      string
	IP        string
}

var logChannel = make(chan LogData, 10000)

func logWorker(db *sql.DB, ctx context.Context) {
	stmt, err := db.Prepare("INSERT INTO api_logs (timestamp, method, path, ip) VALUES ($1, $2, $3, $4)")
	if err != nil {
		fmt.Printf("❌ Failed to prepare SQL statement: %v\n", err)
		return
	}
	defer stmt.Close()

	for {
		select {
		case log, ok := <-logChannel:
			if !ok {
				return
			}

			_, err := stmt.Exec(log.Timestamp, log.Method, log.Path, log.IP)
			if err != nil {
				fmt.Printf("❌ Error saving log to DB: %v\n", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	select {
	case logChannel <- LogData{
		Timestamp: time.Now(),
		Method:    r.Method,
		Path:      r.URL.Path,
		IP:        r.RemoteAddr,
	}:
	default:
		fmt.Println("⚠️ Log channel full, dropping log to protect RAM")
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	const response = "Hello, World!"
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(response))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Printf("❌ Database connection error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		fmt.Printf("❌ Database ping failed: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", helloHandler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	go logWorker(db, serverCtx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		const timeLayout = "02-Jan-2006 03:04:05 PM"
		shutdownTime := time.Now().Format(timeLayout)
		fmt.Printf("[%s] ⏳ Shutting down server gracefully...\n", shutdownTime)

		close(logChannel)

		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("[%s] ❌ Server forced to shutdown: %v\n", shutdownTime, err)
		}
		serverStopCtx()
	}()

	const timeLayout = "02-Jan-2006 03:04:05 PM"
	go func() {
		currentTime := time.Now().Format(timeLayout)
		fmt.Printf("[%s] 🚀 Server is running on http://localhost%s\n", currentTime, server.Addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			currentTime = time.Now().Format(timeLayout)
			fmt.Printf("[%s] ❌ Server crashed: %v\n", currentTime, err)
			os.Exit(1)
		}
	}()

	<-serverCtx.Done()

	currentTime := time.Now().Format(timeLayout)
	fmt.Printf("[%s] 🛑 Server stopped cleanly. Goodbye!\n", currentTime)
}
