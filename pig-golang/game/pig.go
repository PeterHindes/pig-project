package game

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/yourusername/pig-golang/models"
)

var (
	ErrGameNotStarted   = errors.New("game has not started yet")
	ErrGameOver         = errors.New("game is already over")
	ErrNotPlayerTurn    = errors.New("it is not this player's turn")
	ErrInvalidPlayer    = errors.New("invalid player")
	ErrGameFull         = errors.New("game is full")
	ErrNotEnoughPlayers = errors.New("not enough players to start")
)

// PigGame manages the game logic for Pig
type PigGame struct {
	state *models.GameState
	mu    sync.RWMutex
	rng   *rand.Rand
}

// NewPigGame creates a new Pig game instance
func NewPigGame(winningScore int) *PigGame {
	if winningScore <= 0 {
		winningScore = 100 // Default winning score
	}

	return &PigGame{
		state: models.NewGameState(winningScore),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AddPlayer adds a player to the game
func (g *PigGame) AddPlayer(player *models.Player) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.state.IsGameOver {
		return ErrGameOver
	}

	// Check if game is full (max 4 players for Pig)
	if len(g.state.Players) >= 4 {
		return ErrGameFull
	}

	// Check if player already exists
	for _, p := range g.state.Players {
		if p.ID == player.ID {
			return errors.New("player already in game")
		}
	}

	g.state.Players = append(g.state.Players, player)
	g.state.LastActivityAt = time.Now()

	return nil
}

// RemovePlayer removes a player from the game
func (g *PigGame) RemovePlayer(playerID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	for i, player := range g.state.Players {
		if player.ID == playerID {
			player.IsActive = false

			// Check if all remaining players are inactive
			activeCount := 0
			for _, p := range g.state.Players {
				if p.IsActive {
					activeCount++
				}
			}

			// If only one or zero active players remain, end the game
			if activeCount <= 1 && len(g.state.Players) > 1 {
				g.state.IsGameOver = true
				// Find the last active player as winner
				for _, p := range g.state.Players {
					if p.IsActive {
						g.state.Winner = p
						break
					}
				}
			}

			// If current player left, move to next player
			if i == g.state.CurrentPlayer {
				g.nextTurn()
			}

			g.state.LastActivityAt = time.Now()
			return nil
		}
	}

	return ErrInvalidPlayer
}

// CanStart checks if the game can start
func (g *PigGame) CanStart() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.state.Players) >= 2
}

// Start starts the game
func (g *PigGame) Start() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.state.Players) < 2 {
		return ErrNotEnoughPlayers
	}

	if g.state.IsGameOver {
		return ErrGameOver
	}

	// Game is implicitly started when there are enough players
	g.state.LastActivityAt = time.Now()
	return nil
}

// Roll performs a dice roll for the current player
func (g *PigGame) Roll(playerID string) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.state.IsGameOver {
		return 0, ErrGameOver
	}

	if len(g.state.Players) < 2 {
		return 0, ErrGameNotStarted
	}

	// Verify it's this player's turn
	currentPlayer := g.state.Players[g.state.CurrentPlayer]
	if currentPlayer.ID != playerID {
		return 0, ErrNotPlayerTurn
	}

	if !currentPlayer.IsActive {
		return 0, ErrInvalidPlayer
	}

	// Roll the die (1-6)
	roll := g.rng.Intn(6) + 1
	g.state.LastRoll = roll
	g.state.LastActivityAt = time.Now()

	if roll == 1 {
		// Rolled a 1 - lose turn score and move to next player
		g.state.TurnScore = 0
		g.nextTurn()
	} else {
		// Add to turn score
		g.state.TurnScore += roll
	}

	return roll, nil
}

// Hold banks the current turn score and moves to the next player
func (g *PigGame) Hold(playerID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.state.IsGameOver {
		return ErrGameOver
	}

	if len(g.state.Players) < 2 {
		return ErrGameNotStarted
	}

	// Verify it's this player's turn
	currentPlayer := g.state.Players[g.state.CurrentPlayer]
	if currentPlayer.ID != playerID {
		return ErrNotPlayerTurn
	}

	if !currentPlayer.IsActive {
		return ErrInvalidPlayer
	}

	// Add turn score to player's total score
	currentPlayer.Score += g.state.TurnScore
	g.state.LastActivityAt = time.Now()

	// Check for winner
	if currentPlayer.Score >= g.state.WinningScore {
		g.state.IsGameOver = true
		g.state.Winner = currentPlayer
		g.state.TurnScore = 0
		return nil
	}

	// Move to next player
	g.state.TurnScore = 0
	g.nextTurn()

	return nil
}

// nextTurn moves to the next active player
func (g *PigGame) nextTurn() {
	if len(g.state.Players) == 0 {
		return
	}

	// Find next active player
	startIdx := g.state.CurrentPlayer
	for i := 0; i < len(g.state.Players); i++ {
		g.state.CurrentPlayer = (g.state.CurrentPlayer + 1) % len(g.state.Players)
		if g.state.Players[g.state.CurrentPlayer].IsActive {
			break
		}
		// If we've looped back to start and no active players found
		if g.state.CurrentPlayer == startIdx && i == len(g.state.Players)-1 {
			g.state.IsGameOver = true
			return
		}
	}

	g.state.LastRoll = 0
}

// GetState returns a copy of the current game state
func (g *PigGame) GetState() *models.GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Create a deep copy to prevent race conditions
	stateCopy := &models.GameState{
		GameID:         g.state.GameID,
		Players:        make([]*models.Player, len(g.state.Players)),
		CurrentPlayer:  g.state.CurrentPlayer,
		TurnScore:      g.state.TurnScore,
		LastRoll:       g.state.LastRoll,
		WinningScore:   g.state.WinningScore,
		Winner:         g.state.Winner,
		IsGameOver:     g.state.IsGameOver,
		CreatedAt:      g.state.CreatedAt,
		LastActivityAt: g.state.LastActivityAt,
	}

	// Copy players
	for i, player := range g.state.Players {
		stateCopy.Players[i] = &models.Player{
			ID:       player.ID,
			Name:     player.Name,
			Score:    player.Score,
			IsActive: player.IsActive,
		}
	}

	return stateCopy
}

// GetCurrentPlayer returns the current player
func (g *PigGame) GetCurrentPlayer() *models.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.state.Players) == 0 {
		return nil
	}

	return g.state.Players[g.state.CurrentPlayer]
}

// IsGameOver returns whether the game is over
func (g *PigGame) IsGameOver() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.state.IsGameOver
}

// GetWinner returns the winner if game is over
func (g *PigGame) GetWinner() *models.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.state.Winner
}

// GetPlayerCount returns the number of players
func (g *PigGame) GetPlayerCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.state.Players)
}

// GetActivePlayerCount returns the number of active players
func (g *PigGame) GetActivePlayerCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for _, player := range g.state.Players {
		if player.IsActive {
			count++
		}
	}
	return count
}
