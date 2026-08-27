package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

func main() {
	address := "127.0.0.1:9090"
	if value := os.Getenv("MOCK_API_ADDR"); value != "" {
		address = value
	}
	server := &http.Server{Addr: address, Handler: routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	slog.Info("mock API listening", "address", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("mock API stopped", "error", err)
		os.Exit(1)
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()
	var flaky atomic.Int64
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":10,"status":"ok"}`)
	})
	mux.HandleFunc("/unauthorized", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="demo"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
	mux.HandleFunc("/forbidden", func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "forbidden", http.StatusForbidden) })
	mux.HandleFunc("/server-error", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "synthetic failure", http.StatusInternalServerError)
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		fmt.Fprint(w, "slow")
	})
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/ok", http.StatusFound) })
	mux.HandleFunc("/redirect-loop", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/redirect-loop", http.StatusFound) })
	mux.HandleFunc("/rate-limit", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "30")
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	})
	mux.HandleFunc("/bad-json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "{broken")
	})
	mux.HandleFunc("/large-response", func(w http.ResponseWriter, _ *http.Request) {
		for i := 0; i < 12000; i++ {
			fmt.Fprint(w, "0123456789")
		}
	})
	mux.HandleFunc("/gzip", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		writer := gzip.NewWriter(w)
		defer writer.Close()
		fmt.Fprint(writer, "compressed")
	})
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"method": r.Method, "path": r.URL.Path, "query": r.URL.Query(), "headers": r.Header})
	})
	mux.HandleFunc("/idempotency", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Idempotency-Key") == "" {
			http.Error(w, "missing Idempotency-Key", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("/flaky", func(w http.ResponseWriter, _ *http.Request) {
		if flaky.Add(1)%3 != 0 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprint(w, "recovered")
	})
	return mux
}
