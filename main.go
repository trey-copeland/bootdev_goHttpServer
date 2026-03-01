package main

import (
	"fmt"
	"log"
	"net/http"
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
	_, _ = fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits.Load())

}

func (cfg *apiConfig) resetCount(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}

func main() {
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	fileServer := apiCfg.middlewareMetricsInc(
		http.FileServer(http.Dir("./app")),
	)
	mux.Handle("/app/", http.StripPrefix("/app/", fileServer))

	mux.HandleFunc("/healthz", healthzHandler)

	mux.HandleFunc("/metrics", apiCfg.requestCount)

	mux.HandleFunc("/reset", apiCfg.resetCount)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Printf("Server lisenting on port %s\n", port)
	log.Fatal(server.ListenAndServe())
}
