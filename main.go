package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	const response = "Hello, World!"
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(response))
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", helloHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Shutting down server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Server forced to shutdown", "error", err)
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
