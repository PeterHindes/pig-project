package server

import (
	"log"
	"sync"
	"time"

	"github.com/yourusername/pig-golang/game"
	"github.com/yourusername/pig-golang/models"
)

// MatchManager manages all active game matches
type MatchManager struct {
	matches     map[string]*Match
	mu          sync.RWMutex
	waitingRoom *Match // A lobby where players wait for opponents
}

// Match represents a game session
type Match struct {
	ID             string
	Game           *game.PigGame
	Players        map[string]*PlayerConnection
	MinPlayers     int
	MaxPlayers     int
	IsStarted      bool
	Broadcast      chan *models.Message
	Register       chan *PlayerConnection
	Unregister     chan *PlayerConnection
	PlayerActions  chan *models.PlayerAction
	CreatedAt      time.Time
	LastActivityAt time.Time
	mu             sync.RWMutex
}

// PlayerConnection represents a connected player with WebSocket
type PlayerConnection struct {
	PlayerID   string
	PlayerName string
	GameID     string
	Send       chan *models.Message
	conn       interface{} // Will be *websocket.Conn in websocket.go
}

// NewMatchManager creates a new match manager
func NewMatchManager() *MatchManager {
	return &MatchManager{
		matches:     make(map[string]*Match),
		waitingRoom: nil,
	}
}

// NewMatch creates a new match
func NewMatch(gameID string, winningScore int) *Match {
	return &Match{
		ID:             gameID,
		Game:           game.NewPigGame(winningScore),
		Players:        make(map[string]*PlayerConnection),
		MinPlayers:     2,
		MaxPlayers:     4,
		IsStarted:      false,
		Broadcast:      make(chan *models.Message, 256),
		Register:       make(chan *PlayerConnection),
		Unregister:     make(chan *PlayerConnection),
		PlayerActions:  make(chan *models.PlayerAction, 256),
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}
}

// CreateMatch creates a new match and adds it to the manager
func (mm *MatchManager) CreateMatch(winningScore int) *Match {
	state := models.NewGameState(winningScore)
	match := NewMatch(state.GameID, winningScore)

	mm.mu.Lock()
	mm.matches[match.ID] = match
	mm.mu.Unlock()

	// Start the match goroutine
	go match.Run()

	log.Printf("Created new match: %s", match.ID)
	return match
}

// GetMatch retrieves a match by ID
func (mm *MatchManager) GetMatch(gameID string) (*Match, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	match, exists := mm.matches[gameID]
	return match, exists
}

// FindOrCreateMatch finds an available match or creates a new one
func (mm *MatchManager) FindOrCreateMatch(winningScore int) *Match {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Check if there's a waiting room with space
	if mm.waitingRoom != nil {
		mm.waitingRoom.mu.RLock()
		playerCount := len(mm.waitingRoom.Players)
		isStarted := mm.waitingRoom.IsStarted
		mm.waitingRoom.mu.RUnlock()

		if !isStarted && playerCount < mm.waitingRoom.MaxPlayers {
			return mm.waitingRoom
		}
	}

	// Create a new match and set it as waiting room
	state := models.NewGameState(winningScore)
	match := NewMatch(state.GameID, winningScore)
	mm.matches[match.ID] = match
	mm.waitingRoom = match

	// Start the match goroutine
	go match.Run()

	log.Printf("Created new waiting room match: %s", match.ID)
	return match
}

// RemoveMatch removes a match from the manager
func (mm *MatchManager) RemoveMatch(gameID string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.waitingRoom != nil && mm.waitingRoom.ID == gameID {
		mm.waitingRoom = nil
	}

	delete(mm.matches, gameID)
	log.Printf("Removed match: %s", gameID)
}

// CleanupInactiveMatches removes matches that have been inactive for too long
func (mm *MatchManager) CleanupInactiveMatches(timeout time.Duration) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	now := time.Now()
	for gameID, match := range mm.matches {
		match.mu.RLock()
		lastActivity := match.LastActivityAt
		playerCount := len(match.Players)
		isGameOver := match.Game.IsGameOver()
		match.mu.RUnlock()

		// Remove if inactive for too long or if game is over and no players
		if now.Sub(lastActivity) > timeout || (isGameOver && playerCount == 0) {
			if mm.waitingRoom != nil && mm.waitingRoom.ID == gameID {
				mm.waitingRoom = nil
			}
			delete(mm.matches, gameID)
			log.Printf("Cleaned up inactive match: %s", gameID)
		}
	}
}

// Run starts the main loop for a match
func (m *Match) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case conn := <-m.Register:
			m.handleRegister(conn)

		case conn := <-m.Unregister:
			m.handleUnregister(conn)

		case action := <-m.PlayerActions:
			m.handlePlayerAction(action)

		case message := <-m.Broadcast:
			m.broadcastMessage(message)

		case <-ticker.C:
			// Periodic cleanup check
			m.mu.RLock()
			playerCount := len(m.Players)
			m.mu.RUnlock()

			if playerCount == 0 && time.Since(m.LastActivityAt) > 5*time.Minute {
				log.Printf("Match %s has no players, stopping...", m.ID)
				return
			}
		}
	}
}

// handleRegister handles player registration
func (m *Match) handleRegister(conn *PlayerConnection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add player to match
	m.Players[conn.PlayerID] = conn
	m.LastActivityAt = time.Now()

	// Create player model and add to game
	player := models.NewPlayer(conn.PlayerName)
	player.ID = conn.PlayerID // Use the provided player ID

	if err := m.Game.AddPlayer(player); err != nil {
		log.Printf("Error adding player to game: %v", err)
		delete(m.Players, conn.PlayerID)
		return
	}

	log.Printf("Player %s (%s) registered to match %s", player.Name, player.ID, m.ID)

	// Send joined confirmation to the player
	joinedMsg := models.NewMessage(string(models.TypeJoined))
	joinedMsg.PlayerID = conn.PlayerID
	joinedMsg.GameState = m.Game.GetState()
	conn.Send <- joinedMsg

	// Broadcast to all players that someone joined
	updateMsg := models.NewMessage(string(models.TypeGameUpdate))
	updateMsg.GameState = m.Game.GetState()
	updateMsg.Data = map[string]string{
		"event":       "player_joined",
		"player_name": player.Name,
	}
	m.Broadcast <- updateMsg

	// Auto-start if we have enough players
	if m.Game.CanStart() && !m.IsStarted {
		m.IsStarted = true
		if err := m.Game.Start(); err != nil {
			log.Printf("Error starting game: %v", err)
		} else {
			startMsg := models.NewMessage(string(models.TypeGameStart))
			startMsg.GameState = m.Game.GetState()
			m.Broadcast <- startMsg
			log.Printf("Game %s started with %d players", m.ID, m.Game.GetPlayerCount())
		}
	}
}

// handleUnregister handles player disconnection
func (m *Match) handleUnregister(conn *PlayerConnection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.Players[conn.PlayerID]; exists {
		delete(m.Players, conn.PlayerID)
		close(conn.Send)
		m.LastActivityAt = time.Now()

		// Remove player from game
		if err := m.Game.RemovePlayer(conn.PlayerID); err != nil {
			log.Printf("Error removing player from game: %v", err)
		}

		log.Printf("Player %s unregistered from match %s", conn.PlayerID, m.ID)

		// Notify remaining players
		leftMsg := models.NewMessage(string(models.TypePlayerLeft))
		leftMsg.PlayerID = conn.PlayerID
		leftMsg.GameState = m.Game.GetState()
		m.Broadcast <- leftMsg

		// Check if game is over due to not enough players
		if m.Game.IsGameOver() {
			gameOverMsg := models.NewMessage(string(models.TypeGameOver))
			gameOverMsg.GameState = m.Game.GetState()
			m.Broadcast <- gameOverMsg
		}
	}
}

// handlePlayerAction processes player actions (roll, hold)
func (m *Match) handlePlayerAction(action *models.PlayerAction) {
	m.mu.Lock()
	m.LastActivityAt = time.Now()
	m.mu.Unlock()

	var err error
	responseMsg := models.NewMessage(string(models.TypeGameUpdate))
	responseMsg.PlayerID = action.PlayerID
	responseMsg.Action = action.Action

	switch action.Action {
	case models.ActionRoll:
		roll, rollErr := m.Game.Roll(action.PlayerID)
		if rollErr != nil {
			err = rollErr
		} else {
			responseMsg.Data = map[string]interface{}{
				"roll":   roll,
				"action": "roll",
			}
		}

	case models.ActionHold:
		err = m.Game.Hold(action.PlayerID)
		if err == nil {
			responseMsg.Data = map[string]interface{}{
				"action": "hold",
			}
		}

	default:
		log.Printf("Unknown action: %s", action.Action)
		return
	}

	if err != nil {
		errorMsg := models.NewMessage(string(models.TypeError))
		errorMsg.Error = err.Error()
		errorMsg.PlayerID = action.PlayerID

		m.mu.RLock()
		if conn, exists := m.Players[action.PlayerID]; exists {
			conn.Send <- errorMsg
		}
		m.mu.RUnlock()
		return
	}

	// Get updated game state
	responseMsg.GameState = m.Game.GetState()

	// Check if game is over
	if m.Game.IsGameOver() {
		responseMsg.Type = string(models.TypeGameOver)
		log.Printf("Game %s is over. Winner: %v", m.ID, m.Game.GetWinner())
	}

	// Broadcast state to all players
	m.Broadcast <- responseMsg
}

// broadcastMessage sends a message to all connected players
func (m *Match) broadcastMessage(message *models.Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, conn := range m.Players {
		select {
		case conn.Send <- message:
		default:
			// Channel is full or closed, skip
			log.Printf("Failed to send message to player %s", conn.PlayerID)
		}
	}
}

// GetPlayerCount returns the current number of connected players
func (m *Match) GetPlayerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Players)
}

// IsFull returns whether the match is full
func (m *Match) IsFull() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Players) >= m.MaxPlayers
}
