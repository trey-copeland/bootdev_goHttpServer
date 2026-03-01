package main

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

//go:embed app/admin_metrics.html
var adminMetricsTemplateFile string

var adminMetricsTemplate = template.Must(
	template.New("admin_metrics").Parse(adminMetricsTemplateFile),
)

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

var badWordsSet = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func cleanChirp(body string) string {
	words := strings.Split(body, " ")

	for i, word := range words {
		if _, found := badWordsSet[strings.ToLower(word)]; found {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req struct {
		Body string `json:"body"`
	}

	dec := json.NewDecoder(r.Body)
	// dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	if len(req.Body) > 140 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Chirp is too long"})
		return
	}

	cleanedBody := cleanChirp(req.Body)
	writeJSON(w, http.StatusOK, struct {
		CleanedBody string `json:"cleaned_body"`
	}{
		CleanedBody: cleanedBody,
	})

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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := adminMetricsTemplate.Execute(w, struct {
		Hits int32
	}{
		Hits: cfg.fileserverHits.Load(),
	})
	if err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}

}

func (cfg *apiConfig) resetCount(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
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
	mux.Handle("GET /{$}", fileServer)

	mux.HandleFunc("GET /api/healthz", healthzHandler)

	mux.HandleFunc("GET /admin/metrics", apiCfg.requestCount)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetCount)

	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server listening on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
