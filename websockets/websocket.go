package websockets

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Nitesh-04/realtime-racing/config"
	"github.com/Nitesh-04/realtime-racing/models"
	"github.com/gofiber/websocket/v2"
)

type GameHub struct {
	connections map[string][]*Connection
	stats       map[string]map[string]PlayerStats
	timers      map[string]*time.Timer
	gameStates  map[string]GameState
	mu          sync.RWMutex
}

type Connection struct {
	Conn     *websocket.Conn
	RoomCode string
	Username string
	writeMu  sync.Mutex // Add write mutex for each connection
}

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type PlayerStats struct {
	WPM      int     `json:"wpm"`
	Accuracy float64 `json:"accuracy"`
	Error   float64 `json:"errors"`
}

type GameState struct {
	Stage        string    // "waiting", "countdown", "racing", "finished"
	StartTime    time.Time
	CountdownEnd time.Time
}

var Hub = &GameHub{
	connections: make(map[string][]*Connection),
	stats:       make(map[string]map[string]PlayerStats),
	timers:      make(map[string]*time.Timer),
	gameStates:  make(map[string]GameState),
}

func (c *Connection) SafeWriteMessage(messageType int, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.Conn.WriteMessage(messageType, data)
}

func (h *GameHub) HandleConnection(c *websocket.Conn) {
	log.Printf("New WebSocket connection attempt from: %s", c.RemoteAddr())
	
	roomCode := c.Params("room_code")
	username := c.Query("username")

	log.Printf("Connection details - room_code: '%s', username: '%s'", roomCode, username)

	if roomCode == "" || username == "" {
		log.Printf("Missing room_code or username: room_code='%s' username='%s'", roomCode, username)
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Missing room_code or username"))
		c.Close()
		return
	}

	// Verify room exists
	var room models.Room
	if err := config.DB.Where("room_code = ?", roomCode).First(&room).Error; err != nil {
		log.Printf("Room not found for code '%s': %v", roomCode, err)
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Room not found"))
		c.Close()
		return
	}

	log.Printf("Room found: %s (ID: %v)", room.RoomCode, room.ID)

	// Check if username is already connected to this room
	h.mu.Lock()
	for i := range h.connections[roomCode] {
		existingConn := h.connections[roomCode][i]
		if existingConn.Username == username {
			log.Printf("Username '%s' already connected to room '%s'", username, roomCode)
			h.mu.Unlock()
			c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Username already connected"))
			c.Close()
			return
		}
	}
	h.mu.Unlock()

	conn := &Connection{Conn: c, RoomCode: room.RoomCode, Username: username}
	
	// Add connection and handle game state
	h.addConnection(conn)
	
	log.Printf("Player '%s' successfully connected to room '%s'", username, roomCode)

	defer func() {
		log.Printf("Player '%s' disconnecting from room '%s'", conn.Username, conn.RoomCode)
		h.removeConnection(conn)
		conn.Conn.Close()
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("Connection closed for %s in room %s: %v", conn.Username, conn.RoomCode, err)
			break
		}
		h.handleMessage(conn, msg)
	}
}

func (h *GameHub) BroadcastToRoom(roomCode, msgType string, payload interface{}) {
	message := Message{Type: msgType, Payload: payload}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	log.Printf("Broadcasting to room %s: type='%s'", roomCode, msgType)

	h.mu.RLock()
	connections := make([]*Connection, len(h.connections[roomCode]))
	copy(connections, h.connections[roomCode])
	h.mu.RUnlock()

	if len(connections) == 0 {
		log.Printf("No connections found for room %s", roomCode)
		return
	}

	// Use goroutines to send messages concurrently but safely
	var wg sync.WaitGroup
	for _, conn := range connections {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()
			if err := c.SafeWriteMessage(websocket.TextMessage, jsonMessage); err != nil {
				log.Printf("Error sending message to %s: %v", c.Username, err)
			} else {
				log.Printf("Message sent successfully to %s", c.Username)
			}
		}(conn)
	}
	wg.Wait()

	log.Printf("Successfully sent message to %d players in room %s", len(connections), roomCode)
}

func BroadcastCountdown(roomCode string, seconds int) {
	Hub.BroadcastToRoom(roomCode, "countdown", map[string]interface{}{
		"seconds": seconds,
	})
}

func BroadcastStart(roomCode string) {
	Hub.BroadcastToRoom(roomCode, "start", nil)
}

func BroadcastStatsUpdate(roomCode, username string, stats PlayerStats) {
	Hub.BroadcastToRoom(roomCode, "stats_update", map[string]PlayerStats{
		username: stats,
	})
}

func BroadcastGameOver(roomCode string, winner string, stats map[string]PlayerStats, reason string) {
	payload := map[string]interface{}{
		"winner": winner,
		"stats":  stats,
	}
	if reason != "" {
		payload["reason"] = reason
	}
	Hub.BroadcastToRoom(roomCode, "game_over", payload)
}

func (h *GameHub) BroadcastPlayerList(roomCode string) {
	h.mu.RLock()
	connections := h.connections[roomCode]
	h.mu.RUnlock()

	var players []string
	for _, conn := range connections {
		players = append(players, conn.Username)
	}

	log.Printf("Broadcasting player list for room %s: %v", roomCode, players)
	h.BroadcastToRoom(roomCode, "player_list", players)
}

func (h *GameHub) addConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connections[conn.RoomCode] = append(h.connections[conn.RoomCode], conn)

	if _, exists := h.stats[conn.RoomCode]; !exists {
		h.stats[conn.RoomCode] = make(map[string]PlayerStats)
	}

	// Initialize game state if it doesn't exist
	if _, exists := h.gameStates[conn.RoomCode]; !exists {
		h.gameStates[conn.RoomCode] = GameState{Stage: "waiting"}
	}

	log.Printf("Added player '%s' to room '%s'. Total players: %d", 
		conn.Username, conn.RoomCode, len(h.connections[conn.RoomCode]))

	// Get current game state
	gameState := h.gameStates[conn.RoomCode]
	playerCount := len(h.connections[conn.RoomCode])

	// Broadcast player list
	var players []string
	for _, c := range h.connections[conn.RoomCode] {
		players = append(players, c.Username)
	}

	log.Printf("Current game state: %s, Player count: %d", gameState.Stage, playerCount)

	// Release lock before async operations
	h.mu.Unlock()
	
	// Always broadcast player list first
	h.BroadcastToRoom(conn.RoomCode, "player_list", players)

	h.mu.Lock() // Re-acquire for state checks

	// Handle game state logic
	switch gameState.Stage {
	case "countdown":
		// Game is in countdown - inform new player of remaining time
		remaining := time.Until(gameState.CountdownEnd).Seconds()
		if remaining > 0 {
			log.Printf("Game in countdown, informing new player. Remaining: %.0f seconds", remaining)
			go h.sendCountdownToPlayer(conn, int(remaining))
		}
	case "racing":
		// Game already started - send start immediately to new player
		log.Printf("Game already racing, sending start to new player")
		go func() {
			conn.SafeWriteMessage(websocket.TextMessage, []byte(`{"type":"start","payload":null}`))
		}()
	case "waiting":
		// Only start countdown if we have exactly 2 players and no timer exists
		if playerCount == 2 && h.timers[conn.RoomCode] == nil {
			log.Printf("Starting pre-game countdown for room %s", conn.RoomCode)
			// Update game state
			countdownEnd := time.Now().Add(20 * time.Second)
			h.gameStates[conn.RoomCode] = GameState{
				Stage:        "countdown",
				CountdownEnd: countdownEnd,
			}
			// Start countdown without holding the lock
			h.mu.Unlock()
			h.startPreGame(conn.RoomCode)
			h.mu.Lock()
		}
	}
}

func (h *GameHub) sendCountdownToPlayer(conn *Connection, startingSeconds int) {
	for i := startingSeconds; i > 0; i-- {
		time.Sleep(1 * time.Second)
		message := fmt.Sprintf(`{"type":"countdown","payload":{"seconds":%d}}`, i)
		if err := conn.SafeWriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			return // Connection closed
		}
	}
	// Send start message
	conn.SafeWriteMessage(websocket.TextMessage, []byte(`{"type":"start","payload":null}`))
}

func (h *GameHub) removeConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.connections[conn.RoomCode]
	for i, c := range conns {
		if c.Conn == conn.Conn {
			h.connections[conn.RoomCode] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	log.Printf("Removed player '%s' from room '%s'. Remaining players: %d", 
		conn.Username, conn.RoomCode, len(h.connections[conn.RoomCode]))

	// Broadcast updated player list
	var players []string
	for _, c := range h.connections[conn.RoomCode] {
		players = append(players, c.Username)
	}

	// Release lock before broadcasting
	h.mu.Unlock()
	h.BroadcastToRoom(conn.RoomCode, "player_list", players)
	h.mu.Lock()

	// Stop timer and reset game state if less than 2 players remain
	if len(h.connections[conn.RoomCode]) < 2 {
		if t, ok := h.timers[conn.RoomCode]; ok {
			t.Stop()
			delete(h.timers, conn.RoomCode)
			log.Printf("Stopped timer for room %s due to insufficient players", conn.RoomCode)
		}
		// Reset game state to waiting
		h.gameStates[conn.RoomCode] = GameState{Stage: "waiting"}
		log.Printf("Reset game state to waiting for room %s", conn.RoomCode)
	}
}

func (h *GameHub) startPreGame(roomCode string) {
	log.Printf("Starting 20-second countdown for room %s", roomCode)
	
	countdownDuration := 20
	
	// Create timer first
	h.mu.Lock()
	h.timers[roomCode] = time.NewTimer(time.Duration(countdownDuration) * time.Second)
	timer := h.timers[roomCode]
	h.mu.Unlock()

	// Send countdown updates
	go func() {
		for i := countdownDuration; i > 0; i-- {
			// Check if room still exists and has enough players
			h.mu.RLock()
			connections, exists := h.connections[roomCode]
			gameState := h.gameStates[roomCode]
			h.mu.RUnlock()
			
			if !exists || len(connections) < 2 || gameState.Stage != "countdown" {
				log.Printf("Room %s countdown cancelled - not enough players or state changed", roomCode)
				return
			}
			
			BroadcastCountdown(roomCode, i)
			time.Sleep(1 * time.Second)
		}
		
		// Countdown finished - transition to racing
		h.mu.Lock()
		if gameState, exists := h.gameStates[roomCode]; exists && gameState.Stage == "countdown" {
			h.gameStates[roomCode] = GameState{
				Stage:     "racing",
				StartTime: time.Now(),
			}
			log.Printf("Transitioning room %s to racing state", roomCode)
		}
		h.mu.Unlock()
		
		log.Printf("Countdown finished, starting race for room %s", roomCode)
		BroadcastStart(roomCode)
		h.startRace(roomCode)
	}()

	// Wait for timer completion
	go func() {
		<-timer.C
		log.Printf("Timer completed for room %s", roomCode)
	}()
}

func (h *GameHub) startRace(roomCode string) {
	h.mu.Lock()
	// Replace the countdown timer with race timer
	if oldTimer, exists := h.timers[roomCode]; exists {
		oldTimer.Stop()
	}
	
	h.timers[roomCode] = time.AfterFunc(15*time.Second, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		
		// Update game state to finished
		if gameState, exists := h.gameStates[roomCode]; exists {
			gameState.Stage = "finished"
			h.gameStates[roomCode] = gameState
		}
		
		log.Printf("Race finished for room %s, declaring winner", roomCode)
		h.declareWinner(roomCode)
	})
	h.mu.Unlock()
	
	log.Printf("Race timer started for room %s (15 seconds)", roomCode)
}

func (h *GameHub) handleMessage(conn *Connection, rawMsg []byte) {
	var msg Message
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		log.Printf("Invalid message: %v", err)
		return
	}

	switch msg.Type {
	case "stats_update":
		var stats PlayerStats
		if err := mapToStruct(msg.Payload, &stats); err != nil {
			log.Printf("Invalid stats payload: %v", err)
			return
		}

		h.mu.Lock()
		h.stats[conn.RoomCode][conn.Username] = stats
		h.mu.Unlock()

		BroadcastStatsUpdate(conn.RoomCode, conn.Username, stats)
	}
}

func (h *GameHub) declareWinner(roomCode string) {
	stats := h.stats[roomCode]
	var winnerUsername string
	var bestStats PlayerStats

	for user, s := range stats {
		if winnerUsername == "" ||
			s.WPM > bestStats.WPM ||
			(s.WPM == bestStats.WPM && s.Accuracy > bestStats.Accuracy) ||
			(s.WPM == bestStats.WPM && s.Accuracy == bestStats.Accuracy && s.Error < bestStats.Error) {
			winnerUsername = user
			bestStats = s
		}
	}

	if winnerUsername != "" {
		var room models.Room
		if err := config.DB.Where("room_code = ?", roomCode).First(&room).Error; err == nil {
			var winnerUser models.User
			if err := config.DB.Where("username = ?", winnerUsername).First(&winnerUser).Error; err == nil {
				// Update room winner
				config.DB.Model(&models.Room{}).
					Where("id = ?", room.ID).
					Update("winner_id", winnerUser.ID)

				// Insert results for each player
				for username, s := range stats {
					var user models.User
					if err := config.DB.Where("username = ?", username).First(&user).Error; err == nil {
						opponentName := ""
						for opName := range stats {
							if opName != username {
								opponentName = opName
								break
							}
						}
						var opponent models.User
						config.DB.Where("username = ?", opponentName).First(&opponent)

						result := models.Results{
							UserID:     user.ID,
							OpponentID: opponent.ID,
							Won:        (username == winnerUsername),
							WPM:        s.WPM,
							Accuracy:   s.Accuracy,
							Error:      s.Error,
						}
						config.DB.Create(&result)
					}
				}
			}
		}
	}

	BroadcastGameOver(roomCode, winnerUsername, stats, "")

	// Clean up timers
	if t, ok := h.timers[roomCode]; ok {
		t.Stop()
		delete(h.timers, roomCode)
	}

	// Clean up room after delay
	go func() {
		time.Sleep(5 * time.Second)
		var room models.Room
		if err := config.DB.Where("room_code = ?", roomCode).First(&room).Error; err == nil {
			config.DB.Delete(&room)
		}

		h.mu.Lock()
		delete(h.connections, roomCode)
		delete(h.stats, roomCode)
		delete(h.gameStates, roomCode)
		if t, ok := h.timers[roomCode]; ok {
			t.Stop()
			delete(h.timers, roomCode)
		}
		h.mu.Unlock()

		log.Printf("Cleaned up room %s", roomCode)
	}()
}

func mapToStruct(input interface{}, output interface{}) error {
	bytes, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, output)
}

func (h *GameHub) PeriodicCleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	var emptyRooms []string
	
	// Find rooms with no active connections
	for roomCode, connections := range h.connections {
		if len(connections) == 0 {
			emptyRooms = append(emptyRooms, roomCode)
		}
	}
	
	// Clean up empty rooms
	for _, roomCode := range emptyRooms {
		log.Printf("Periodic cleanup: removing empty room %s", roomCode)

		config.DB.Where("room_code = ?", roomCode).Delete(&models.Room{})

		delete(h.connections, roomCode)
		delete(h.stats, roomCode)
		delete(h.gameStates, roomCode)
		if timer, exists := h.timers[roomCode]; exists {
			timer.Stop()
			delete(h.timers, roomCode)
		}
	}
	
	if len(emptyRooms) > 0 {
		log.Printf("Periodic cleanup completed: removed %d empty rooms", len(emptyRooms))
	}
}

func StartPeriodicCleanup() {
	ticker := time.NewTicker(2 * time.Minute)
	log.Printf("Starting periodic cleanup every 2 minutes")
	go func() {
		for range ticker.C {
			Hub.PeriodicCleanup()
		}
	}()
}