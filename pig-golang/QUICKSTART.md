# Quick Start Guide - Pig Game Server

Get up and running with the Pig Game Server in 5 minutes!

## Prerequisites

- Go 1.21 or higher installed
- A terminal/command prompt
- A web browser

## Step 1: Install Dependencies

```bash
cd pig-golang
go mod download
```

## Step 2: Start the Server

```bash
go run main.go
```

You should see:
```
Starting Pig Game Server...
Server listening on :8080
WebSocket URL: ws://localhost:8080
REST API available at http://localhost:8080/api
Health check: http://localhost:8080/api/health
```

## Step 3: Test the Server

Open a new terminal and test the health endpoint:

```bash
curl http://localhost:8080/api/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "matches": 0
}
```

## Step 4: Play the Game!

### Option A: Use the Web Client (Easiest)

1. Open `example-client.html` in your web browser
2. Enter your name
3. Click "Join Available Game"
4. Open the same file in another browser window/tab with a different name
5. Start playing!

### Option B: Use cURL and a WebSocket Client

**Create a match:**
```bash
curl -X POST http://localhost:8080/api/match/create \
  -H "Content-Type: application/json" \
  -d '{"player_name": "Alice"}'
```

**Response:**
```json
{
  "game_id": "550e8400-e29b-41d4-a716-446655440000",
  "player_id": "660e8400-e29b-41d4-a716-446655440001",
  "ws_url": "ws://localhost:8080/ws/game/550e8400-e29b-41d4-a716-446655440000?playerId=660e8400-e29b-41d4-a716-446655440001&playerName=Alice",
  "message": "Match created successfully. Connect via WebSocket to join.",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Connect via WebSocket:**

Use the `ws_url` from the response to connect with any WebSocket client (e.g., websocat, wscat, or browser JavaScript).

## Step 5: Play Multiple Games

You can have multiple games running simultaneously:

```bash
# Terminal 1
curl -X POST http://localhost:8080/api/match/create \
  -H "Content-Type: application/json" \
  -d '{"player_name": "Player1"}'

# Terminal 2
curl -X POST http://localhost:8080/api/match/create \
  -H "Content-Type: application/json" \
  -d '{"player_name": "Player2"}'
```

## Game Rules Reminder

1. **Roll**: Roll the die to accumulate points for your turn
2. **Hold**: Bank your turn points and end your turn
3. If you roll a **1**, you lose all points for that turn!
4. First player to reach **100 points** wins

## WebSocket Message Examples

### Roll the dice:
```json
{
  "action": "roll"
}
```

### Hold and bank points:
```json
{
  "action": "hold"
}
```

## Common Issues

### Port Already in Use

If port 8080 is already in use, start the server on a different port:

```bash
go run main.go -port=8888 -wsurl=ws://localhost:8888
```

### Can't Connect from Browser

If you're running the server on a different machine, update the API URL in `example-client.html`:

```javascript
document.getElementById('apiUrl').value = "http://YOUR_SERVER_IP:8080"
```

### Connection Refused

Make sure the server is running:
```bash
curl http://localhost:8080/api/health
```

## Using Make Commands

If you have `make` installed:

```bash
# Install dependencies
make install

# Run the server
make run

# Run tests
make test

# Build binary
make build

# Open example client
make client
```

## Next Steps

- Read the full [README.md](README.md) for detailed API documentation
- Check out the [example-client.html](example-client.html) source for client implementation examples
- Run tests: `go test ./... -v`
- Build for production: `make build-all`

## Development Tips

### Watch for changes (requires external tools)

Install Air for hot reload:
```bash
go install github.com/cosmtrek/air@latest
air
```

### Run with race detector:
```bash
go run -race main.go
```

### Enable verbose logging:
The server logs all requests and game events to stdout.

## Docker Quick Start

If you prefer Docker:

```bash
# Build and run with docker-compose
docker-compose up --build

# Or with Docker directly
docker build -t pig-server .
docker run -p 8080:8080 pig-server
```

## Testing the API

### Create a game:
```bash
curl -X POST http://localhost:8080/api/match/create \
  -H "Content-Type: application/json" \
  -d '{"player_name": "TestPlayer"}'
```

### Join a game:
```bash
curl -X POST http://localhost:8080/api/match/join \
  -H "Content-Type: application/json" \
  -d '{"player_name": "TestPlayer2"}'
```

### List all matches:
```bash
curl http://localhost:8080/api/matches
```

### Get match details:
```bash
curl http://localhost:8080/api/match/{gameId}
```

## What's Happening Behind the Scenes?

1. **REST API**: Handles player allocation and matchmaking
2. **Match Manager**: Creates and manages game sessions in separate goroutines
3. **WebSocket Server**: Handles real-time game communication
4. **Game Logic**: Processes dice rolls, scoring, and turn management
5. **Broadcasting**: Updates all players when game state changes

## Architecture Flow

```
Client → REST API (Create/Join Match) → Match Manager
           ↓
Client → WebSocket Connection → Match Goroutine
           ↓
Game Actions (Roll/Hold) → Game Logic → Broadcast to All Players
```

## Have Fun!

You're all set! Enjoy playing Pig! 🎲

For more details, see [README.md](README.md)