package main

import (
	"log"
	"net/http"

	email "github.com/gustavolmo/portfolio-v2-api/internal/email"
	health "github.com/gustavolmo/portfolio-v2-api/internal/health"
	observe "github.com/gustavolmo/portfolio-v2-api/internal/observe"
)

var allowedOrigins = map[string]bool{
	"http://localhost:5173":        true,
	"https://gustavolmo.github.io": true,
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("GET /health", health.Handler)
	router.HandleFunc("GET /observe", observe.Handler)
	router.HandleFunc("POST /email", email.Handler)

	log.Println("Server running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", cors(router)); err != nil {
		log.Fatal(err)
	}
}
