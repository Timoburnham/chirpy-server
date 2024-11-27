package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	
	"regexp"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)

	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	hits := cfg.fileserverHits.Load()
	fmt.Fprintf(w, `<html>
		<body>
		  <h1>Welcome, Chirpy Admin</h1>
		  <p>Chirpy has been visited %d times!</p>
		</body>
	  </html>`, hits)
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(405)
		data, err := json.Marshal(map[string]string{"error": "Something went wrong"})
		if err != nil {
			w.WriteHeader(500)
			return
		}
		w.Write(data)
		return
	}

	type Chirp struct {
		Body string `json:"body"`
	}

	var chirp Chirp
	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		w.WriteHeader(400)
		data, _ := json.Marshal(map[string]string{"error": "Invalid JSON"})
		w.Write(data)
		return
	}
	if len(chirp.Body) > 140 {
		w.WriteHeader(400)
		data, _ := json.Marshal(map[string]string{"error": "Chirp is too long"})
		w.Write(data)
		return
	}
	
	
	
	
	cleanedBody := cleanChirp(chirp.Body)
	
	response, _ := json.Marshal(map[string]string{"cleaned_body": cleanedBody})
	w.WriteHeader(200)
	w.Write(response)
}

func cleanChirp(body string) string {
    
    profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
    
    
    for _, word := range profaneWords {
        
        re := regexp.MustCompile(`(?i)\b` + word + `\b`)
        
        
        body = re.ReplaceAllStringFunc(body, func(s string) string {
            return "****"
        })
    }
    
    return body
}

func main() {

	apiCfg := &apiConfig{}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("/api/validate_chirp", validateChirpHandler)

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
