package models

import (
	"time"

	"github.com/google/uuid"
)

// Player represents a player in the game
type Player struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Score    int    `json:"score"`
	IsActive bool   `json:"is_active"`
}

// GameState represents the current state of a Pig game
type GameState struct {
	GameID         string    `json:"game_id"`
	Players        []*Player `json:"players"`
	CurrentPlayer  int       `json:"current_player"`
	TurnScore      int       `json:"turn_score"`
	LastRoll       int       `json:"last_roll"`
	WinningScore   int       `json:"winning_score"`
	Winner         *Player   `json:"winner,omitempty"`
	IsGameOver     bool      `json:"is_game_over"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

// GameAction represents actions players can take
type GameAction string

const (
	ActionRoll GameAction = "roll"
	ActionHold GameAction = "hold"
	ActionJoin GameAction = "join"
	ActionQuit GameAction = "quit"
)

// Message represents WebSocket messages
type Message struct {
	Type      string      `json:"type"`
	Action    GameAction  `json:"action,omitempty"`
	PlayerID  string      `json:"player_id,omitempty"`
	GameState *GameState  `json:"game_state,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// MessageType defines the type of WebSocket message
type MessageType string

const (
	TypeGameUpdate MessageType = "game_update"
	TypeError      MessageType = "error"
	TypeJoined     MessageType = "joined"
	TypePlayerLeft MessageType = "player_left"
	TypeGameStart  MessageType = "game_start"
	TypeGameOver   MessageType = "game_over"
)

// MatchRequest represents a request to join or create a match
type MatchRequest struct {
	PlayerName string `json:"player_name"`
}

// MatchResponse represents the response when a match is allocated
type MatchResponse struct {
	GameID    string    `json:"game_id"`
	PlayerID  string    `json:"player_id"`
	WSURL     string    `json:"ws_url"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// Match represents a game session
type Match struct {
	ID             string
	GameState      *GameState
	Players        map[string]*PlayerConnection
	MinPlayers     int
	MaxPlayers     int
	IsStarted      bool
	Broadcast      chan *Message
	Register       chan *PlayerConnection
	Unregister     chan *PlayerConnection
	PlayerActions  chan *PlayerAction
	CreatedAt      time.Time
	LastActivityAt time.Time
}

// PlayerConnection represents a connected player
type PlayerConnection struct {
	PlayerID string
	GameID   string
	Send     chan *Message
}

// PlayerAction represents an action taken by a player
type PlayerAction struct {
	PlayerID string
	Action   GameAction
}

// NewPlayer creates a new player
func NewPlayer(name string) *Player {
	return &Player{
		ID:       uuid.New().String(),
		Name:     name,
		Score:    0,
		IsActive: true,
	}
}

// NewGameState creates a new game state
func NewGameState(winningScore int) *GameState {
	return &GameState{
		GameID:         uuid.New().String(),
		Players:        make([]*Player, 0),
		CurrentPlayer:  0,
		TurnScore:      0,
		LastRoll:       0,
		WinningScore:   winningScore,
		Winner:         nil,
		IsGameOver:     false,
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}
}

// NewMessage creates a new message with timestamp
func NewMessage(msgType string) *Message {
	return &Message{
		Type:      msgType,
		Timestamp: time.Now(),
	}
}
