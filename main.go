
package main

go get github.com/lib/pq

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	type errorResponse struct {
		Error string `json:"error"`
	}

	type cleanResponse struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	req := requestBody{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		err := errorResponse{Error: "Something went wrong"}
		dat, _ := json.Marshal(err)
		w.Write(dat)
		return
	}

	if len(req.Body) > 140 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		err := errorResponse{Error: "Chirp is too long"}
		dat, _ := json.Marshal(err)
		w.Write(dat)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	valid := cleanResponse{Valid: true}
	dat, _ := json.Marshal(valid)
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

