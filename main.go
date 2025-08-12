package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

func main() {
	apiCfg := &apiConfig{}

	const port = "8080"
	const filepathRoot = "."

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.HandleFunc("GET /api/healthz", handleReadiness)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handleRequests)
	mux.HandleFunc("POST /admin/reset", apiCfg.handleReset)
	mux.HandleFunc("POST /api/validate_chirp", handleJson)

	s.ListenAndServe()
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func handleJson(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}

	type cleanResponse struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	req := requestBody{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, 400, "Something went wrong")
		return
	}

	if len(req.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	cleanedBody := cleanResponse{Cleaned_body: replaceProfanity(req.Body)}
	respondWithJSON(w, 200, cleanedBody)
}

func replaceProfanity(text string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	splitText := strings.Split(text, " ")
	for i, word := range splitText {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				splitText[i] = "****"
			}
		}
	}
	joinText := strings.Join(splitText, " ")
	return joinText
}
func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	respondWithJSON(w, code, errorResponse{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	dat, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Write(dat)
}

func handleReadiness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handleRequests(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Type", "text/html")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>`,
		cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})

}

