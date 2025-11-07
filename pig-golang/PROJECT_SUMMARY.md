# Pig Game Server - Project Summary

## Overview

A production-ready, real-time multiplayer backend server for the dice game **Pig**, built with Go. The server uses WebSockets for real-time gameplay and a REST API for player matchmaking and game allocation.

## Architecture

### Technology Stack
- **Language**: Go 1.21+
- **WebSocket Library**: gorilla/websocket
- **HTTP Router**: gorilla/mux
- **UUID Generation**: google/uuid

### Key Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Layer                             │
│  (Web Browsers, Mobile Apps, CLI Tools, Bots)              │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     HTTP Server                              │
│  ┌─────────────────┐         ┌──────────────────┐          │
│  │   REST API      │         │  WebSocket       │          │
│  │  Endpoints      │         │  Handlers        │          │
│  └─────────────────┘         └──────────────────┘          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  Match Manager                               │
│  • Creates and manages game sessions                         │
│  • Handles matchmaking                                       │
│  • Cleans up inactive matches                               │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│               Individual Match (Goroutine)                   │
│  ┌─────────────────────────────────────────────────┐        │
│  │  Channels:                                       │        │
│  │  • Register (player joins)                       │        │
│  │  • Unregister (player leaves)                    │        │
│  │  • PlayerActions (roll, hold)                    │        │
│  │  • Broadcast (send to all players)               │        │
│  └─────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     Game Logic                               │
│  • Pig game rules implementation                             │
│  • Dice rolling (1-6)                                        │
│  • Score tracking                                            │
│  • Turn management                                           │
│  • Win condition checking                                    │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
pig-golang/
├── main.go                 # Server entry point, routing, middleware
├── models/
│   └── types.go           # Shared data structures (Player, GameState, etc.)
├── game/
│   ├── pig.go            # Core Pig game logic
│   └── pig_test.go       # Unit tests for game logic
├── server/
│   ├── match.go          # Match/session management
│   ├── rest.go           # REST API handlers
│   └── websocket.go      # WebSocket connection handlers
├── go.mod                 # Go module dependencies
├── go.sum                 # Dependency checksums
├── Makefile              # Build and development commands
├── Dockerfile            # Container build instructions
├── docker-compose.yml    # Docker Compose configuration
├── example-client.html   # Interactive web client demo
├── README.md             # Comprehensive documentation
├── QUICKSTART.md         # Quick start guide
├── API_EXAMPLES.md       # API usage examples
└── .gitignore           # Git ignore rules
```

## Core Features

### 1. REST API
- **Create Match**: Start a new game
- **Join Match**: Automatic matchmaking
- **List Matches**: View all active games
- **Get Match Details**: Query game state
- **Health Check**: Server status

### 2. WebSocket Communication
- Real-time bidirectional communication
- Automatic game start when enough players join
- Live game state updates
- Player connection/disconnection handling
- Ping/pong keepalive mechanism

### 3. Game Logic
- Thread-safe game state management
- Dice rolling with proper randomization
- Turn-based gameplay
- Score tracking and win conditions
- Player management (active/inactive)

### 4. Match Management
- Concurrent game sessions (each in own goroutine)
- Automatic matchmaking
- Idle match cleanup
- Broadcast system for player updates

## Concurrency Model

### Thread Safety
- **Mutexes**: Protect shared state in game logic and match management
- **Channels**: Safe communication between goroutines
  - `Register`: Add new player connections
  - `Unregister`: Remove disconnected players
  - `PlayerActions`: Process game actions
  - `Broadcast`: Distribute messages to all players

### Goroutine Structure
```
Main Thread
├── HTTP Server (gorilla/mux)
├── Match Manager Cleanup Routine
└── For each match:
    ├── Match.Run() goroutine
    │   ├── Handles registration
    │   ├── Processes player actions
    │   └── Broadcasts updates
    └── For each player:
        ├── readPump() goroutine (read from WebSocket)
        └── writePump() goroutine (write to WebSocket)
```

## Game Rules: Pig

1. Players take turns rolling a single die
2. Each turn, players can:
   - **Roll**: Add die value to turn score (unless roll is 1)
   - **Hold**: Bank turn score and end turn
3. Rolling a **1** loses all turn points and ends turn
4. First player to reach 100 points wins
5. 2-4 players per game

## API Endpoints

### REST API
```
GET  /api/health              - Health check
GET  /api/matches             - List active matches
POST /api/match/create        - Create new match
POST /api/match/join          - Join/create match (matchmaking)
GET  /api/match/{gameId}      - Get match details
```

### WebSocket
```
WS   /ws/game/{gameId}?playerId={id}&playerName={name}
```

## Message Protocol

### Client → Server
```json
{"action": "roll"}  // Roll the dice
{"action": "hold"}  // Bank points and end turn
```

### Server → Client
```json
{"type": "joined", "game_state": {...}}        // Player joined
{"type": "game_start", "game_state": {...}}    // Game started
{"type": "game_update", "game_state": {...}}   // State updated
{"type": "game_over", "game_state": {...}}     // Game ended
{"type": "player_left", "game_state": {...}}   // Player disconnected
{"type": "error", "error": "..."}              // Error occurred
```

## Testing

### Unit Tests
- **19 test cases** covering game logic
- Tests for concurrent access
- Edge cases (invalid players, full games, etc.)
- Win condition validation

### Run Tests
```bash
go test ./... -v
```

## Deployment Options

### Local Development
```bash
go run main.go
```

### Binary Build
```bash
go build -o pig-server main.go
./pig-server -port=8080
```

### Docker
```bash
docker build -t pig-server .
docker run -p 8080:8080 pig-server
```

### Docker Compose
```bash
docker-compose up --build
```

## Performance Characteristics

- **Scalability**: Multiple concurrent games (limited by system resources)
- **Latency**: Sub-millisecond game action processing
- **Connection Handling**: Automatic cleanup of idle connections
- **Memory**: O(n) where n = number of active players across all games
- **CPU**: Minimal - event-driven architecture

## Security Considerations

### Current Implementation
- CORS enabled for development (allow all origins)
- No authentication/authorization
- No rate limiting
- No input validation beyond basic checks

### Production Recommendations
1. Implement proper origin checking for WebSockets
2. Add player authentication (JWT, OAuth)
3. Implement rate limiting per IP/user
4. Add input validation and sanitization
5. Use TLS/SSL (HTTPS/WSS)
6. Configure restrictive CORS policies
7. Add request logging and monitoring
8. Implement DDoS protection

## Configuration

### Command Line Flags
```bash
-port=8080                        # Server port
-wsurl=ws://localhost:8080        # WebSocket URL for clients
```

### Environment Variables
```bash
PORT=8080
WS_URL=ws://localhost:8080
```

## Monitoring & Observability

### Logging
- All requests logged with timing
- Game events logged (player joins, leaves, wins)
- Error conditions logged with context

### Health Endpoint
```bash
curl http://localhost:8080/api/health
```

Returns:
- Server status
- Number of active matches
- Timestamp

## Extension Points

### Easy Extensions
1. **Custom Win Scores**: Already parameterized
2. **More Players**: Increase `MaxPlayers` in match.go
3. **Game Variants**: Modify game/pig.go rules
4. **Persistence**: Add database for game history
5. **Leaderboards**: Track wins/losses per player
6. **Spectator Mode**: Read-only WebSocket connections
7. **Private Games**: Add game passwords/invite codes
8. **Replay System**: Record game events

### Integration Points
1. **Frontend Frameworks**: React, Vue, Angular via WebSocket
2. **Mobile Apps**: iOS, Android WebSocket clients
3. **Discord Bots**: Bot framework integration
4. **Analytics**: Game statistics and player behavior
5. **Authentication**: JWT middleware
6. **Database**: PostgreSQL, MongoDB for persistence

## Development Tools

### Makefile Commands
```bash
make install          # Install dependencies
make build           # Build binary
make run             # Run server
make test            # Run tests
make test-coverage   # Generate coverage report
make clean           # Remove build artifacts
make fmt             # Format code
make lint            # Lint code
make client          # Open example client
```

## Known Limitations

1. **No Persistence**: Games lost on server restart
2. **No Authentication**: Anyone can join any game
3. **No Reconnection**: Disconnected players cannot rejoin
4. **No Spectators**: Only active players can connect
5. **Single Server**: No horizontal scaling support
6. **In-Memory Only**: All state in RAM

## Future Improvements

1. Database integration for game persistence
2. Redis for distributed state management
3. Authentication and authorization system
4. WebSocket reconnection with session recovery
5. Spectator mode
6. Game replays
7. Player statistics and leaderboards
8. Multiple game variants
9. Private/password-protected games
10. Tournament mode

## Dependencies

```
github.com/google/uuid v1.5.0       # UUID generation
github.com/gorilla/mux v1.8.1       # HTTP routing
github.com/gorilla/websocket v1.5.1 # WebSocket support
```

## License

MIT License - See repository for details

## Getting Started

See [QUICKSTART.md](QUICKSTART.md) for a 5-minute setup guide.

See [README.md](README.md) for comprehensive documentation.

See [API_EXAMPLES.md](API_EXAMPLES.md) for API usage examples.

## Support

- Open an issue on GitHub
- Check server logs for debugging
- Use `/api/health` to verify server status
- Review example client code in `example-client.html`

---

**Built with ❤️ using Go**