package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) requestCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "Hits: %d\n", cfg.fileserverHits.Load())

}

func (cfg *apiConfig) resetCount(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	apiCfg := &apiConfig{}

	mux := http.NewServeMux()

	fileServer := apiCfg.middlewareMetricsInc(
		http.FileServer(http.Dir("./app")),
	)
	mux.Handle("GET /app/", http.StripPrefix("/app/", fileServer))

	mux.HandleFunc("GET /healthz", healthzHandler)

	mux.HandleFunc("GET /metrics", apiCfg.requestCount)

	mux.HandleFunc("POST /reset", apiCfg.resetCount)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server listening on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
