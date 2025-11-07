package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourusername/pig-golang/models"
)

// RESTServer handles REST API requests
type RESTServer struct {
	matchManager *MatchManager
	wsURL        string
}

// NewRESTServer creates a new REST API server
func NewRESTServer(matchManager *MatchManager, wsURL string) *RESTServer {
	return &RESTServer{
		matchManager: matchManager,
		wsURL:        wsURL,
	}
}

// HandleCreateMatch creates a new match
func (rs *RESTServer) HandleCreateMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PlayerName == "" {
		req.PlayerName = "Player"
	}

	// Create a new player
	player := models.NewPlayer(req.PlayerName)

	// Create a new match
	match := rs.matchManager.CreateMatch(100) // Default winning score of 100

	// Create response
	response := models.MatchResponse{
		GameID:    match.ID,
		PlayerID:  player.ID,
		WSURL:     fmt.Sprintf("%s/ws/game/%s?playerId=%s&playerName=%s", rs.wsURL, match.ID, player.ID, req.PlayerName),
		Message:   "Match created successfully. Connect via WebSocket to join.",
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	log.Printf("Created match %s for player %s (%s)", match.ID, req.PlayerName, player.ID)
}

// HandleJoinMatch joins an existing match or finds an available one
func (rs *RESTServer) HandleJoinMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PlayerName == "" {
		req.PlayerName = "Player"
	}

	// Create a new player
	player := models.NewPlayer(req.PlayerName)

	// Find or create a match
	match := rs.matchManager.FindOrCreateMatch(100) // Default winning score of 100

	// Create response
	response := models.MatchResponse{
		GameID:    match.ID,
		PlayerID:  player.ID,
		WSURL:     fmt.Sprintf("%s/ws/game/%s?playerId=%s&playerName=%s", rs.wsURL, match.ID, player.ID, req.PlayerName),
		Message:   "Match found. Connect via WebSocket to join.",
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("Player %s (%s) allocated to match %s", req.PlayerName, player.ID, match.ID)
}

// HandleGetMatch retrieves match information
func (rs *RESTServer) HandleGetMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	gameID := vars["gameId"]

	if gameID == "" {
		http.Error(w, "Missing gameId", http.StatusBadRequest)
		return
	}

	match, exists := rs.matchManager.GetMatch(gameID)
	if !exists {
		http.Error(w, "Match not found", http.StatusNotFound)
		return
	}

	// Get game state
	gameState := match.Game.GetState()

	// Return game state
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameState)
}

// HandleListMatches lists all active matches
func (rs *RESTServer) HandleListMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rs.matchManager.mu.RLock()
	defer rs.matchManager.mu.RUnlock()

	type MatchInfo struct {
		GameID       string    `json:"game_id"`
		PlayerCount  int       `json:"player_count"`
		MaxPlayers   int       `json:"max_players"`
		IsStarted    bool      `json:"is_started"`
		IsGameOver   bool      `json:"is_game_over"`
		CreatedAt    time.Time `json:"created_at"`
		WinningScore int       `json:"winning_score"`
	}

	matches := make([]MatchInfo, 0)
	for _, match := range rs.matchManager.matches {
		match.mu.RLock()
		gameState := match.Game.GetState()
		matchInfo := MatchInfo{
			GameID:       match.ID,
			PlayerCount:  len(match.Players),
			MaxPlayers:   match.MaxPlayers,
			IsStarted:    match.IsStarted,
			IsGameOver:   gameState.IsGameOver,
			CreatedAt:    match.CreatedAt,
			WinningScore: gameState.WinningScore,
		}
		match.mu.RUnlock()
		matches = append(matches, matchInfo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

// HandleHealthCheck returns server health status
func (rs *RESTServer) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"matches":   len(rs.matchManager.matches),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
