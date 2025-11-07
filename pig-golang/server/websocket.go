package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/yourusername/pig-golang/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// WebSocketServer handles WebSocket connections
type WebSocketServer struct {
	matchManager *MatchManager
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(matchManager *MatchManager) *WebSocketServer {
	return &WebSocketServer{
		matchManager: matchManager,
	}
}

// HandleWebSocket handles WebSocket connection requests
func (ws *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["gameId"]
	playerID := r.URL.Query().Get("playerId")

	if gameID == "" || playerID == "" {
		http.Error(w, "Missing gameId or playerId", http.StatusBadRequest)
		return
	}

	// Get the match
	match, exists := ws.matchManager.GetMatch(gameID)
	if !exists {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Get player name from query params (if reconnecting) or generate default
	playerName := r.URL.Query().Get("playerName")
	if playerName == "" {
		playerName = "Player"
	}

	// Create player connection
	playerConn := &PlayerConnection{
		PlayerID:   playerID,
		PlayerName: playerName,
		GameID:     gameID,
		Send:       make(chan *models.Message, 256),
		conn:       conn,
	}

	// Register player with the match
	match.Register <- playerConn

	// Start goroutines for reading and writing
	go ws.writePump(conn, playerConn)
	go ws.readPump(conn, playerConn, match)
}

// readPump reads messages from the WebSocket connection
func (ws *WebSocketServer) readPump(conn *websocket.Conn, playerConn *PlayerConnection, match *Match) {
	defer func() {
		match.Unregister <- playerConn
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetReadLimit(maxMessageSize)
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse the message
		var msg models.Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle the action
		if msg.Action != "" {
			action := &models.PlayerAction{
				PlayerID: playerConn.PlayerID,
				Action:   msg.Action,
			}
			match.PlayerActions <- action
		}
	}
}

// writePump writes messages to the WebSocket connection
func (ws *WebSocketServer) writePump(conn *websocket.Conn, playerConn *PlayerConnection) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-playerConn.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// Write the message as JSON
			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				w.Close()
				continue
			}

			w.Write(messageBytes)

			// Add queued messages to the current websocket message
			n := len(playerConn.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				nextMsg := <-playerConn.Send
				nextMsgBytes, err := json.Marshal(nextMsg)
				if err != nil {
					log.Printf("Error marshaling queued message: %v", err)
					continue
				}
				w.Write(nextMsgBytes)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
