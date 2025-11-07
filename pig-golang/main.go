package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourusername/pig-golang/server"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8080", "Port to run the server on")
	wsURL := flag.String("wsurl", "ws://localhost:8080", "WebSocket URL for client connections")
	flag.Parse()

	log.Printf("Starting Pig Game Server...")

	// Create match manager
	matchManager := server.NewMatchManager()

	// Start cleanup routine for inactive matches
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			matchManager.CleanupInactiveMatches(30 * time.Minute)
		}
	}()

	// Create servers
	restServer := server.NewRESTServer(matchManager, *wsURL)
	wsServer := server.NewWebSocketServer(matchManager)

	// Setup router
	router := mux.NewRouter()

	// REST API endpoints
	router.HandleFunc("/api/health", restServer.HandleHealthCheck).Methods("GET")
	router.HandleFunc("/api/matches", restServer.HandleListMatches).Methods("GET")
	router.HandleFunc("/api/match/create", restServer.HandleCreateMatch).Methods("POST")
	router.HandleFunc("/api/match/join", restServer.HandleJoinMatch).Methods("POST")
	router.HandleFunc("/api/match/{gameId}", restServer.HandleGetMatch).Methods("GET")

	// WebSocket endpoint
	router.HandleFunc("/ws/game/{gameId}", wsServer.HandleWebSocket)

	// Add CORS middleware
	router.Use(corsMiddleware)

	// Add logging middleware
	router.Use(loggingMiddleware)

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Server listening on %s", addr)
	log.Printf("WebSocket URL: %s", *wsURL)
	log.Printf("REST API available at http://localhost:%s/api", *port)
	log.Printf("Health check: http://localhost:%s/api/health", *port)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf(
			"%s %s %s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}
