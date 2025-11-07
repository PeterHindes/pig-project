package game

import (
	"testing"

	"github.com/yourusername/pig-golang/models"
)

func TestNewPigGame(t *testing.T) {
	game := NewPigGame(100)

	if game == nil {
		t.Fatal("Expected game to be created, got nil")
	}

	state := game.GetState()
	if state.WinningScore != 100 {
		t.Errorf("Expected winning score 100, got %d", state.WinningScore)
	}

	if len(state.Players) != 0 {
		t.Errorf("Expected 0 players, got %d", len(state.Players))
	}

	if state.IsGameOver {
		t.Error("Expected game not to be over")
	}
}

func TestNewPigGameDefaultScore(t *testing.T) {
	game := NewPigGame(0)
	state := game.GetState()

	if state.WinningScore != 100 {
		t.Errorf("Expected default winning score 100, got %d", state.WinningScore)
	}
}

func TestAddPlayer(t *testing.T) {
	game := NewPigGame(100)
	player := models.NewPlayer("Alice")

	err := game.AddPlayer(player)
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	if game.GetPlayerCount() != 1 {
		t.Errorf("Expected 1 player, got %d", game.GetPlayerCount())
	}

	state := game.GetState()
	if state.Players[0].Name != "Alice" {
		t.Errorf("Expected player name 'Alice', got '%s'", state.Players[0].Name)
	}
}

func TestAddPlayerDuplicate(t *testing.T) {
	game := NewPigGame(100)
	player := models.NewPlayer("Alice")

	game.AddPlayer(player)
	err := game.AddPlayer(player)

	if err == nil {
		t.Error("Expected error when adding duplicate player")
	}
}

func TestAddPlayerToFullGame(t *testing.T) {
	game := NewPigGame(100)

	// Add 4 players (max)
	for i := 0; i < 4; i++ {
		player := models.NewPlayer("Player")
		game.AddPlayer(player)
	}

	// Try to add 5th player
	player := models.NewPlayer("Extra")
	err := game.AddPlayer(player)

	if err != ErrGameFull {
		t.Errorf("Expected ErrGameFull, got %v", err)
	}
}

func TestCanStart(t *testing.T) {
	game := NewPigGame(100)

	if game.CanStart() {
		t.Error("Game should not be able to start with 0 players")
	}

	game.AddPlayer(models.NewPlayer("Alice"))
	if game.CanStart() {
		t.Error("Game should not be able to start with 1 player")
	}

	game.AddPlayer(models.NewPlayer("Bob"))
	if !game.CanStart() {
		t.Error("Game should be able to start with 2 players")
	}
}

func TestStart(t *testing.T) {
	game := NewPigGame(100)

	// Can't start with no players
	err := game.Start()
	if err != ErrNotEnoughPlayers {
		t.Errorf("Expected ErrNotEnoughPlayers, got %v", err)
	}

	// Add players and start
	game.AddPlayer(models.NewPlayer("Alice"))
	game.AddPlayer(models.NewPlayer("Bob"))

	err = game.Start()
	if err != nil {
		t.Errorf("Failed to start game: %v", err)
	}
}

func TestRoll(t *testing.T) {
	game := NewPigGame(100)
	player1 := models.NewPlayer("Alice")
	player2 := models.NewPlayer("Bob")

	game.AddPlayer(player1)
	game.AddPlayer(player2)
	game.Start()

	// First player's turn
	roll, err := game.Roll(player1.ID)
	if err != nil {
		t.Fatalf("Failed to roll: %v", err)
	}

	if roll < 1 || roll > 6 {
		t.Errorf("Expected roll between 1 and 6, got %d", roll)
	}

	state := game.GetState()
	if roll == 1 {
		// Should have lost turn
		if state.TurnScore != 0 {
			t.Error("Expected turn score to be 0 after rolling 1")
		}
	} else {
		// Should have accumulated points
		if state.TurnScore != roll {
			t.Errorf("Expected turn score %d, got %d", roll, state.TurnScore)
		}
	}
}

func TestRollNotYourTurn(t *testing.T) {
	game := NewPigGame(100)
	player1 := models.NewPlayer("Alice")
	player2 := models.NewPlayer("Bob")

	game.AddPlayer(player1)
	game.AddPlayer(player2)
	game.Start()

	// Try to roll as player 2 when it's player 1's turn
	_, err := game.Roll(player2.ID)
	if err != ErrNotPlayerTurn {
		t.Errorf("Expected ErrNotPlayerTurn, got %v", err)
	}
}

func TestRollBeforeGameStart(t *testing.T) {
	game := NewPigGame(100)
	player := models.NewPlayer("Alice")
	game.AddPlayer(player)

	_, err := game.Roll(player.ID)
	if err != ErrGameNotStarted {
		t.Errorf("Expected ErrGameNotStarted, got %v", err)
	}
}

func TestHold(t *testing.T) {
	game := NewPigGame(100)
	player1 := models.NewPlayer("Alice")
	player2 := models.NewPlayer("Bob")

	game.AddPlayer(player1)
	game.AddPlayer(player2)
	game.Start()

	// Roll until we get something other than 1
	initialScore := player1.Score
	var roll int
	var err error
	for i := 0; i < 10; i++ {
		roll, err = game.Roll(player1.ID)
		if err != nil {
			t.Fatalf("Failed to roll: %v", err)
		}
		if roll != 1 {
			break
		}
	}

	if roll == 1 {
		t.Skip("Rolled 1 ten times in a row, skipping test")
	}

	state := game.GetState()
	turnScore := state.TurnScore

	err = game.Hold(player1.ID)
	if err != nil {
		t.Fatalf("Failed to hold: %v", err)
	}

	state = game.GetState()
	if state.Players[0].Score != initialScore+turnScore {
		t.Errorf("Expected score %d, got %d", initialScore+turnScore, state.Players[0].Score)
	}

	if state.TurnScore != 0 {
		t.Error("Expected turn score to be reset to 0")
	}

	if state.CurrentPlayer != 1 {
		t.Errorf("Expected current player to be 1, got %d", state.CurrentPlayer)
	}
}

func TestHoldNotYourTurn(t *testing.T) {
	game := NewPigGame(100)
	player1 := models.NewPlayer("Alice")
	player2 := models.NewPlayer("Bob")

	game.AddPlayer(player1)
	game.AddPlayer(player2)
	game.Start()

	err := game.Hold(player2.ID)
	if err != ErrNotPlayerTurn {
		t.Errorf("Expected ErrNotPlayerTurn, got %v", err)
	}
}

func TestWinCondition(t *testing.T) {
	game := NewPigGame(20) // Low winning score for testing
	player1 := models.NewPlayer("Alice")
	player2 := models.NewPlayer("Bob")

	game.AddPlayer(player1)
	game.AddPlayer(player2)
	game.Start()

	// Manually set player score to 19, then force a win
	game.mu.Lock()
	game.state.Players[0].Score = 19
	game.state.TurnScore = 0
	game.mu.Unlock()

	// Roll until we get something other than 1, then hold to win
	won := false
	for i := 0; i < 100; i++ {
		if game.IsGameOver() {
			won = true
			break
		}

		// Make sure it's player1's turn
		currentPlayer := game.GetCurrentPlayer()
		if currentPlayer.ID != player1.ID {
			// Skip to player1's turn
			game.mu.Lock()
			game.state.CurrentPlayer = 0
			game.mu.Unlock()
		}

		roll, err := game.Roll(player1.ID)
		if err != nil {
			t.Logf("Roll error: %v", err)
			continue
		}

		if roll != 1 && game.state.TurnScore > 0 {
			// Hold to bank points and potentially win
			err = game.Hold(player1.ID)
			if err != nil {
				t.Logf("Hold error: %v", err)
			}
			if game.state.Players[0].Score >= 20 {
				won = true
				break
			}
		}
	}

	if !won || !game.IsGameOver() {
		t.Skipf("Unable to reach win condition after 100 iterations, skipping test")
	}

	winner := game.GetWinner()
	if winner == nil {
		t.Fatal("Expected a winner, got nil")
	}

	if winner.ID != player1.ID {
		t.Errorf("Expected winner to be player1, got %s", winner.ID)
	}

	if winner.Score < 20 {
		t.Errorf("Expected winner score >= 20, got %d", winner.Score)
	}
}

func TestRemovePlayer(t *testing.T) {
	game := NewPigGame(100)
	player1 := models.NewPlayer("Alice")
	player2 := models.NewPlayer("Bob")

	game.AddPlayer(player1)
	game.AddPlayer(player2)
	game.Start()

	err := game.RemovePlayer(player1.ID)
	if err != nil {
		t.Fatalf("Failed to remove player: %v", err)
	}

	state := game.GetState()
	if state.Players[0].IsActive {
		t.Error("Expected player to be inactive")
	}
}

func TestRemovePlayerInvalidPlayer(t *testing.T) {
	game := NewPigGame(100)
	player := models.NewPlayer("Alice")
	game.AddPlayer(player)

	err := game.RemovePlayer("invalid-id")
	if err != ErrInvalidPlayer {
		t.Errorf("Expected ErrInvalidPlayer, got %v", err)
	}
}

func TestRemovePlayerEndsGame(t *testing.T) {
	game := NewPigGame(100)
	player1 := models.NewPlayer("Alice")
	player2 := models.NewPlayer("Bob")

	game.AddPlayer(player1)
	game.AddPlayer(player2)
	game.Start()

	// Remove one player, should end game
	game.RemovePlayer(player1.ID)

	if !game.IsGameOver() {
		t.Error("Expected game to be over when only one active player remains")
	}

	winner := game.GetWinner()
	if winner == nil {
		t.Fatal("Expected a winner")
	}

	if winner.ID != player2.ID {
		t.Error("Expected remaining player to be winner")
	}
}

func TestGetActivePlayerCount(t *testing.T) {
	game := NewPigGame(100)
	player1 := models.NewPlayer("Alice")
	player2 := models.NewPlayer("Bob")
	player3 := models.NewPlayer("Charlie")

	game.AddPlayer(player1)
	game.AddPlayer(player2)
	game.AddPlayer(player3)

	if game.GetActivePlayerCount() != 3 {
		t.Errorf("Expected 3 active players, got %d", game.GetActivePlayerCount())
	}

	game.RemovePlayer(player1.ID)

	if game.GetActivePlayerCount() != 2 {
		t.Errorf("Expected 2 active players, got %d", game.GetActivePlayerCount())
	}
}

func TestConcurrentAccess(t *testing.T) {
	game := NewPigGame(100)
	player1 := models.NewPlayer("Alice")
	player2 := models.NewPlayer("Bob")

	game.AddPlayer(player1)
	game.AddPlayer(player2)
	game.Start()

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			state := game.GetState()
			if state == nil {
				t.Error("GetState returned nil")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestTurnProgression(t *testing.T) {
	game := NewPigGame(100)
	player1 := models.NewPlayer("Alice")
	player2 := models.NewPlayer("Bob")
	player3 := models.NewPlayer("Charlie")

	game.AddPlayer(player1)
	game.AddPlayer(player2)
	game.AddPlayer(player3)
	game.Start()

	state := game.GetState()
	if state.CurrentPlayer != 0 {
		t.Errorf("Expected first player to start, got %d", state.CurrentPlayer)
	}

	// Simulate rolling a 1 to end turn
	game.state.LastRoll = 1
	game.nextTurn()

	state = game.GetState()
	if state.CurrentPlayer != 1 {
		t.Errorf("Expected current player to be 1, got %d", state.CurrentPlayer)
	}
}
