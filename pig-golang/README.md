# Pig Game Backend Server

A real-time multiplayer backend server for the dice game **Pig**, built with Go, WebSockets, and REST API.

## Overview

This server provides:
- **REST API** for player matchmaking and game allocation
- **WebSocket** connections for real-time game play
- **Thread-safe** match management with concurrent game sessions
- **Automatic matchmaking** to pair players together

## Game Rules: Pig

Pig is a simple dice game with the following rules:

1. Players take turns rolling a single die
2. On each turn, a player may:
   - **Roll**: Add the die value to their turn score (unless they roll a 1)
   - **Hold**: Add their turn score to their total score and end their turn
3. If a player rolls a **1**, they lose all points accumulated in that turn and their turn ends
4. The first player to reach the winning score (default: 100) wins the game

## Installation

### Prerequisites
- Go 1.21 or higher

### Setup

1. Clone the repository:
```bash
cd pig-golang
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run main.go
```

Or with custom settings:
```bash
go run main.go -port=8080 -wsurl=ws://localhost:8080
```

### Command Line Flags

- `-port`: Port to run the server on (default: `8080`)
- `-wsurl`: WebSocket URL for client connections (default: `ws://localhost:8080`)

## API Documentation

### REST API Endpoints

#### Health Check
```
GET /api/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "matches": 3
}
```

#### List Active Matches
```
GET /api/matches
```

Response:
```json
[
  {
    "game_id": "550e8400-e29b-41d4-a716-446655440000",
    "player_count": 2,
    "max_players": 4,
    "is_started": true,
    "is_game_over": false,
    "created_at": "2024-01-15T10:25:00Z",
    "winning_score": 100
  }
]
```

#### Create New Match
```
POST /api/match/create
Content-Type: application/json

{
  "player_name": "Alice"
}
```

Response:
```json
{
  "game_id": "550e8400-e29b-41d4-a716-446655440000",
  "player_id": "660e8400-e29b-41d4-a716-446655440001",
  "ws_url": "ws://localhost:8080/ws/game/550e8400-e29b-41d4-a716-446655440000?playerId=660e8400-e29b-41d4-a716-446655440001&playerName=Alice",
  "message": "Match created successfully. Connect via WebSocket to join.",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Join Available Match
```
POST /api/match/join
Content-Type: application/json

{
  "player_name": "Bob"
}
```

Response: Same as Create Match. This endpoint will find an existing match with available slots or create a new one.

#### Get Match Info
```
GET /api/match/{gameId}
```

Response:
```json
{
  "game_id": "550e8400-e29b-41d4-a716-446655440000",
  "players": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "Alice",
      "score": 45,
      "is_active": true
    },
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "name": "Bob",
      "score": 32,
      "is_active": true
    }
  ],
  "current_player": 0,
  "turn_score": 12,
  "last_roll": 6,
  "winning_score": 100,
  "is_game_over": false,
  "created_at": "2024-01-15T10:25:00Z",
  "last_activity_at": "2024-01-15T10:30:00Z"
}
```

### WebSocket Protocol

#### Connecting

Connect to: `ws://localhost:8080/ws/game/{gameId}?playerId={playerId}&playerName={playerName}`

#### Message Types

All messages are JSON with the following structure:

```json
{
  "type": "message_type",
  "action": "action_type",
  "player_id": "player-uuid",
  "game_state": { /* game state object */ },
  "data": { /* additional data */ },
  "error": "error message",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Message Types (Server → Client)

1. **joined** - Player successfully joined the game
2. **game_start** - Game has started with enough players
3. **game_update** - Game state updated (after roll/hold)
4. **game_over** - Game has ended
5. **player_left** - A player disconnected
6. **error** - An error occurred

#### Actions (Client → Server)

Send JSON messages with an `action` field:

**Roll the die:**
```json
{
  "action": "roll"
}
```

**Hold and bank points:**
```json
{
  "action": "hold"
}
```

#### Example Game Flow

1. **Connect**: Player connects via WebSocket
```
→ Server sends: { "type": "joined", "game_state": {...} }
```

2. **Game Starts**: When 2+ players join
```
→ Server broadcasts: { "type": "game_start", "game_state": {...} }
```

3. **Player Rolls**:
```
← Client sends: { "action": "roll" }
→ Server broadcasts: { 
    "type": "game_update",
    "data": { "roll": 5, "action": "roll" },
    "game_state": {...}
  }
```

4. **Player Holds**:
```
← Client sends: { "action": "hold" }
→ Server broadcasts: {
    "type": "game_update",
    "data": { "action": "hold" },
    "game_state": {...}
  }
```

5. **Game Ends**:
```
→ Server broadcasts: {
    "type": "game_over",
    "game_state": {
      "winner": { "id": "...", "name": "Alice", "score": 100 },
      "is_game_over": true,
      ...
    }
  }
```

## Example Usage

### Using cURL

**Create a match:**
```bash
curl -X POST http://localhost:8080/api/match/create \
  -H "Content-Type: application/json" \
  -d '{"player_name": "Alice"}'
```

**Join a match:**
```bash
curl -X POST http://localhost:8080/api/match/join \
  -H "Content-Type: application/json" \
  -d '{"player_name": "Bob"}'
```

**Check server health:**
```bash
curl http://localhost:8080/api/health
```

**List active matches:**
```bash
curl http://localhost:8080/api/matches
```

### Using JavaScript (Browser/Node.js)

```javascript
// 1. Create or join a match
const response = await fetch('http://localhost:8080/api/match/join', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ player_name: 'Alice' })
});

const { ws_url, player_id, game_id } = await response.json();

// 2. Connect via WebSocket
const ws = new WebSocket(ws_url);

ws.onopen = () => {
  console.log('Connected to game!');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
  
  switch (message.type) {
    case 'joined':
      console.log('Joined game:', message.game_state);
      break;
    case 'game_start':
      console.log('Game started!');
      break;
    case 'game_update':
      console.log('Game state:', message.game_state);
      break;
    case 'game_over':
      console.log('Winner:', message.game_state.winner);
      break;
  }
};

// 3. Take actions
function roll() {
  ws.send(JSON.stringify({ action: 'roll' }));
}

function hold() {
  ws.send(JSON.stringify({ action: 'hold' }));
}

// Example: Roll when it's your turn
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'game_update' || msg.type === 'game_start') {
    const state = msg.game_state;
    const currentPlayer = state.players[state.current_player];
    if (currentPlayer.id === player_id) {
      console.log('Your turn!');
      // Roll or hold based on your strategy
      if (state.turn_score < 20) {
        roll();
      } else {
        hold();
      }
    }
  }
};
```

## Project Structure

```
pig-golang/
├── main.go              # Entry point and server setup
├── go.mod               # Go module dependencies
├── models/
│   └── types.go         # Shared data structures and types
├── game/
│   └── pig.go          # Core game logic and rules
└── server/
    ├── match.go        # Match/session management
    ├── rest.go         # REST API handlers
    └── websocket.go    # WebSocket handlers
```

## Architecture

### Components

1. **Match Manager**: Central coordinator for all game sessions
   - Manages active matches
   - Handles matchmaking (finding or creating matches)
   - Cleans up inactive matches
   - Thread-safe with mutex protection

2. **Match**: Individual game session
   - Runs in its own goroutine
   - Manages player connections
   - Processes player actions (roll, hold)
   - Broadcasts updates to all players
   - Handles player registration/disconnection

3. **Game Logic (PigGame)**: Core game rules
   - Thread-safe game state management
   - Dice rolling and scoring
   - Turn management
   - Win condition checking

4. **REST Server**: HTTP API for matchmaking
   - Create new matches
   - Join existing matches
   - Query match status
   - Health checks

5. **WebSocket Server**: Real-time communication
   - Upgrades HTTP to WebSocket
   - Bidirectional message handling
   - Connection management with ping/pong
   - Message broadcasting

### Concurrency Model

- Each match runs in a dedicated goroutine
- Channels used for thread-safe communication:
  - `Register`: Add new player connections
  - `Unregister`: Remove disconnected players
  - `PlayerActions`: Process game actions (roll, hold)
  - `Broadcast`: Send messages to all players
- Mutexes protect shared state

### Scalability

- Multiple concurrent matches supported
- Each match is independent
- Automatic cleanup of inactive matches
- WebSocket connections with timeouts and health checks

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o pig-server
./pig-server -port=8080
```

### Building for Production
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o pig-server-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o pig-server.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o pig-server-mac
```

## Configuration

### Environment Variables

You can also use environment variables:
```bash
export PORT=8080
export WS_URL=ws://localhost:8080
go run main.go
```

### Customizing Game Rules

Edit `game/pig.go` to modify:
- Winning score (default: 100)
- Max players per match (default: 4)
- Min players to start (default: 2)

## Error Handling

The server handles various error conditions:
- Invalid player turns
- Disconnected players
- Full matches
- Invalid game states
- WebSocket connection errors

All errors are logged and appropriate error messages are sent to clients.

## Security Considerations

**For Production Use:**

1. **Origin Checking**: Update `websocket.go` to validate WebSocket origins
2. **Authentication**: Add player authentication and authorization
3. **Rate Limiting**: Implement rate limiting for API endpoints
4. **Input Validation**: Add more robust input validation
5. **TLS/SSL**: Use HTTPS and WSS in production
6. **CORS**: Configure appropriate CORS policies

## License

MIT License

## Contributing

Contributions welcome! Please feel free to submit a Pull Request.