# Pig Game Server - API Examples

This document provides comprehensive examples for interacting with the Pig Game Server API.

## Base URL

```
http://localhost:8080
```

## Table of Contents

1. [Health Check](#health-check)
2. [Create a New Match](#create-a-new-match)
3. [Join an Available Match](#join-an-available-match)
4. [List Active Matches](#list-active-matches)
5. [Get Match Details](#get-match-details)
6. [WebSocket Connection](#websocket-connection)
7. [Complete Game Flow Example](#complete-game-flow-example)

---

## Health Check

Check if the server is running and healthy.

### Request

```http
GET /api/health
```

### cURL

```bash
curl http://localhost:8080/api/health
```

### Response

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "matches": 3
}
```

---

## Create a New Match

Create a new game match with you as the first player.

### Request

```http
POST /api/match/create
Content-Type: application/json

{
  "player_name": "Alice"
}
```

### cURL

```bash
curl -X POST http://localhost:8080/api/match/create \
  -H "Content-Type: application/json" \
  -d '{"player_name": "Alice"}'
```

### Response

```json
{
  "game_id": "550e8400-e29b-41d4-a716-446655440000",
  "player_id": "660e8400-e29b-41d4-a716-446655440001",
  "ws_url": "ws://localhost:8080/ws/game/550e8400-e29b-41d4-a716-446655440000?playerId=660e8400-e29b-41d4-a716-446655440001&playerName=Alice",
  "message": "Match created successfully. Connect via WebSocket to join.",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Notes

- The `game_id` is used to identify this match
- The `player_id` is your unique player identifier
- The `ws_url` is used to connect via WebSocket
- Save these values for WebSocket connection

---

## Join an Available Match

Join an existing match or create a new one if none are available.

### Request

```http
POST /api/match/join
Content-Type: application/json

{
  "player_name": "Bob"
}
```

### cURL

```bash
curl -X POST http://localhost:8080/api/match/join \
  -H "Content-Type: application/json" \
  -d '{"player_name": "Bob"}'
```

### Response

```json
{
  "game_id": "550e8400-e29b-41d4-a716-446655440000",
  "player_id": "770e8400-e29b-41d4-a716-446655440002",
  "ws_url": "ws://localhost:8080/ws/game/550e8400-e29b-41d4-a716-446655440000?playerId=770e8400-e29b-41d4-a716-446655440002&playerName=Bob",
  "message": "Match found. Connect via WebSocket to join.",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Notes

- This endpoint implements automatic matchmaking
- It will join an existing match with available slots
- If no matches are available, it creates a new one
- Ideal for quick-play functionality

---

## List Active Matches

Get a list of all active game matches.

### Request

```http
GET /api/matches
```

### cURL

```bash
curl http://localhost:8080/api/matches
```

### Response

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
  },
  {
    "game_id": "660e8400-e29b-41d4-a716-446655440003",
    "player_count": 1,
    "max_players": 4,
    "is_started": false,
    "is_game_over": false,
    "created_at": "2024-01-15T10:28:00Z",
    "winning_score": 100
  }
]
```

### Notes

- Shows all currently active matches
- `is_started`: false means waiting for more players
- `player_count` < `max_players` means match has available slots

---

## Get Match Details

Get detailed information about a specific match.

### Request

```http
GET /api/match/{gameId}
```

### cURL

```bash
# Replace {gameId} with actual game ID
curl http://localhost:8080/api/match/550e8400-e29b-41d4-a716-446655440000
```

### Response

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

### Notes

- Provides full game state snapshot
- `current_player`: index of player whose turn it is
- `turn_score`: points accumulated in current turn
- `last_roll`: last dice roll value

---

## WebSocket Connection

Connect to a match via WebSocket for real-time gameplay.

### Connection URL

```
ws://localhost:8080/ws/game/{gameId}?playerId={playerId}&playerName={playerName}
```

### JavaScript Example

```javascript
// Get connection details from REST API first
const response = await fetch('http://localhost:8080/api/match/join', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ player_name: 'Alice' })
});

const { ws_url, player_id, game_id } = await response.json();

// Connect via WebSocket
const ws = new WebSocket(ws_url);

ws.onopen = () => {
  console.log('Connected to game!');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
  handleGameMessage(message);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from game');
};
```

### Node.js Example (with ws package)

```javascript
const WebSocket = require('ws');
const fetch = require('node-fetch');

async function playGame() {
  // Join a match
  const response = await fetch('http://localhost:8080/api/match/join', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ player_name: 'BotPlayer' })
  });

  const { ws_url } = await response.json();

  // Connect via WebSocket
  const ws = new WebSocket(ws_url);

  ws.on('open', () => {
    console.log('Connected to game!');
  });

  ws.on('message', (data) => {
    const message = JSON.parse(data);
    console.log('Received:', message);

    // Simple bot logic: roll if turn_score < 20, else hold
    if (message.type === 'game_update' || message.type === 'game_start') {
      const state = message.game_state;
      if (state.players[state.current_player].id === playerId) {
        if (state.turn_score < 20) {
          ws.send(JSON.stringify({ action: 'roll' }));
        } else {
          ws.send(JSON.stringify({ action: 'hold' }));
        }
      }
    }
  });
}

playGame();
```

### Message Types (Server → Client)

#### 1. Joined

Player successfully joined the game.

```json
{
  "type": "joined",
  "player_id": "660e8400-e29b-41d4-a716-446655440001",
  "game_state": { /* full game state */ },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### 2. Game Start

Game has started with enough players.

```json
{
  "type": "game_start",
  "game_state": { /* full game state */ },
  "timestamp": "2024-01-15T10:30:15Z"
}
```

#### 3. Game Update

Game state updated after a player action.

```json
{
  "type": "game_update",
  "action": "roll",
  "player_id": "660e8400-e29b-41d4-a716-446655440001",
  "data": {
    "roll": 5,
    "action": "roll"
  },
  "game_state": { /* full game state */ },
  "timestamp": "2024-01-15T10:30:20Z"
}
```

#### 4. Game Over

Game has ended with a winner.

```json
{
  "type": "game_over",
  "game_state": {
    "winner": {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "Alice",
      "score": 105,
      "is_active": true
    },
    "is_game_over": true,
    /* rest of game state */
  },
  "timestamp": "2024-01-15T10:35:00Z"
}
```

#### 5. Player Left

A player disconnected from the game.

```json
{
  "type": "player_left",
  "player_id": "770e8400-e29b-41d4-a716-446655440002",
  "game_state": { /* updated game state */ },
  "timestamp": "2024-01-15T10:32:00Z"
}
```

#### 6. Error

An error occurred processing your action.

```json
{
  "type": "error",
  "error": "it is not this player's turn",
  "player_id": "660e8400-e29b-41d4-a716-446655440001",
  "timestamp": "2024-01-15T10:30:25Z"
}
```

### Actions (Client → Server)

#### Roll the Dice

```json
{
  "action": "roll"
}
```

#### Hold and Bank Points

```json
{
  "action": "hold"
}
```

---

## Complete Game Flow Example

Here's a complete example of a two-player game using cURL and websocat.

### Terminal 1: Player 1 (Alice)

```bash
# 1. Create a match
curl -X POST http://localhost:8080/api/match/create \
  -H "Content-Type: application/json" \
  -d '{"player_name": "Alice"}' | jq

# Copy the ws_url from response

# 2. Connect via WebSocket (requires websocat)
websocat "ws://localhost:8080/ws/game/{gameId}?playerId={playerId}&playerName=Alice"

# 3. Wait for another player to join...

# 4. When game starts and it's your turn, roll:
{"action": "roll"}

# 5. Roll again or hold:
{"action": "hold"}
```

### Terminal 2: Player 2 (Bob)

```bash
# 1. Join a match
curl -X POST http://localhost:8080/api/match/join \
  -H "Content-Type: application/json" \
  -d '{"player_name": "Bob"}' | jq

# Copy the ws_url from response

# 2. Connect via WebSocket
websocat "ws://localhost:8080/ws/game/{gameId}?playerId={playerId}&playerName=Bob"

# Game starts automatically when 2 players connect!

# 3. When it's your turn:
{"action": "roll"}

# 4. Continue playing...
{"action": "hold"}
```

### Terminal 3: Monitor (Optional)

```bash
# Watch all matches
watch -n 2 'curl -s http://localhost:8080/api/matches | jq'

# Or check specific match
curl http://localhost:8080/api/match/{gameId} | jq
```

---

## Python Example

Complete Python client using `requests` and `websocket-client`:

```python
import json
import requests
from websocket import WebSocketApp

BASE_URL = "http://localhost:8080"
player_id = None

def on_message(ws, message):
    msg = json.loads(message)
    print(f"Received: {msg['type']}")
    
    if msg['type'] in ['game_update', 'game_start', 'joined']:
        state = msg['game_state']
        print(f"Current player: {state['players'][state['current_player']]['name']}")
        print(f"Turn score: {state['turn_score']}")
        
        # Check if it's our turn
        current = state['players'][state['current_player']]
        if current['id'] == player_id:
            print("It's our turn!")
            # Simple strategy: roll if turn_score < 20, else hold
            if state['turn_score'] < 20:
                ws.send(json.dumps({"action": "roll"}))
            else:
                ws.send(json.dumps({"action": "hold"}))
    
    elif msg['type'] == 'game_over':
        winner = msg['game_state']['winner']
        print(f"Game over! Winner: {winner['name']} with {winner['score']} points")
        ws.close()

def on_error(ws, error):
    print(f"Error: {error}")

def on_close(ws, close_status_code, close_msg):
    print("Connection closed")

def on_open(ws):
    print("Connected to game!")

def main():
    global player_id
    
    # Join a match
    response = requests.post(
        f"{BASE_URL}/api/match/join",
        json={"player_name": "PythonBot"}
    )
    data = response.json()
    
    player_id = data['player_id']
    ws_url = data['ws_url']
    
    print(f"Joined game: {data['game_id']}")
    print(f"Player ID: {player_id}")
    
    # Connect via WebSocket
    ws = WebSocketApp(
        ws_url,
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close
    )
    
    ws.run_forever()

if __name__ == "__main__":
    main()
```

Install dependencies:
```bash
pip install requests websocket-client
```

Run:
```bash
python pig_client.py
```

---

## Testing Tools

### Using websocat

Install websocat:
```bash
# macOS
brew install websocat

# Linux
cargo install websocat

# Or download binary from GitHub releases
```

Connect to a game:
```bash
websocat "ws://localhost:8080/ws/game/{gameId}?playerId={playerId}&playerName=Test"
```

Send actions:
```json
{"action": "roll"}
{"action": "hold"}
```

### Using wscat

Install wscat:
```bash
npm install -g wscat
```

Connect:
```bash
wscat -c "ws://localhost:8080/ws/game/{gameId}?playerId={playerId}&playerName=Test"
```

---

## Error Responses

### 400 Bad Request

```json
{
  "error": "Invalid request body"
}
```

### 404 Not Found

```json
{
  "error": "Game not found"
}
```

### 405 Method Not Allowed

```json
{
  "error": "Method not allowed"
}
```

---

## Rate Limiting

Currently, there is no rate limiting implemented. For production use, consider:

- Implementing rate limiting middleware
- Setting connection limits per IP
- Adding authentication/authorization

---

## Tips and Best Practices

1. **Save Connection Details**: Store `game_id` and `player_id` from REST API response
2. **Handle Reconnections**: Implement reconnection logic for dropped WebSocket connections
3. **Parse All Messages**: Always parse and handle all message types
4. **Error Handling**: Check for errors in every API response
5. **Concurrent Games**: Each game runs independently, so you can have multiple games running
6. **Clean Disconnections**: Close WebSocket connections properly when done

---

## Need Help?

- Check server logs for detailed information
- Use the `/api/health` endpoint to verify server status
- Monitor active matches with `/api/matches`
- Review the full documentation in [README.md](README.md)