package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"touchdown-tally/internal/config"
	"touchdown-tally/internal/models"
	"touchdown-tally/pkg/logger"
	"touchdown-tally/pkg/response"
)

type ChatHandler struct {
	db       *sql.DB
	config   *config.Config
	logger   *logger.Logger
	upgrader websocket.Upgrader
	clients  map[string]map[*websocket.Conn]string // poolID -> conn -> userID
	messages chan models.ChatMessage
}

func NewChatHandler(db *sql.DB, config *config.Config, logger *logger.Logger) *ChatHandler {
	handler := &ChatHandler{
		db:     db,
		config: config,
		logger: logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, you should check the origin properly
				return true
			},
		},
		clients:  make(map[string]map[*websocket.Conn]string),
		messages: make(chan models.ChatMessage, 256),
	}

	// Start the message broadcasting goroutine
	go handler.handleMessages()

	return handler
}

// WebSocket endpoint for real-time chat
func (h *ChatHandler) WebSocketHandler(c *gin.Context) {
	poolID := c.Param("id")
	userID := c.GetString("user_id")

	// Verify user has access to this pool
	var memberRole string
	err := h.db.QueryRow(`
		SELECT role FROM pool_memberships 
		WHERE pool_id = $1 AND user_id = $2 AND is_active = true
	`, poolID, userID).Scan(&memberRole)

	if err == sql.ErrNoRows {
		response.Error(c, http.StatusForbidden, "Access denied to this pool")
		return
	}
	if err != nil {
		h.logger.Error("Failed to check pool membership", "pool_id", poolID, "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to validate access")
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to websocket", "error", err)
		return
	}
	defer conn.Close()

	// Register client
	h.registerClient(poolID, userID, conn)
	defer h.unregisterClient(poolID, conn)

	// Get user display name
	var displayName string
	err = h.db.QueryRow(`
		SELECT display_name FROM user_profiles WHERE id = $1
	`, userID).Scan(&displayName)
	if err != nil {
		h.logger.Error("Failed to get user display name", "user_id", userID, "error", err)
		displayName = "Unknown User"
	}

	// Send join notification
	joinMessage := models.ChatMessage{
		PoolID:      poolID,
		UserID:      userID,
		DisplayName: displayName,
		Message:     displayName + " joined the chat",
		MessageType: "system",
		Timestamp:   time.Now(),
	}
	h.messages <- joinMessage

	// Listen for messages from this client
	for {
		var msg struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		}
		
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket error", "error", err)
			}
			break
		}

		// Validate message
		if len(msg.Message) == 0 || len(msg.Message) > 1000 {
			continue
		}

		// Default to user message type
		if msg.Type == "" {
			msg.Type = "user"
		}

		// Create chat message
		chatMessage := models.ChatMessage{
			PoolID:      poolID,
			UserID:      userID,
			DisplayName: displayName,
			Message:     msg.Message,
			MessageType: msg.Type,
			Timestamp:   time.Now(),
		}

		// Save to database
		err = h.saveMessage(chatMessage)
		if err != nil {
			h.logger.Error("Failed to save chat message", "error", err)
			continue
		}

		// Broadcast to all clients in this pool
		h.messages <- chatMessage
	}

	// Send leave notification
	leaveMessage := models.ChatMessage{
		PoolID:      poolID,
		UserID:      userID,
		DisplayName: displayName,
		Message:     displayName + " left the chat",
		MessageType: "system",
		Timestamp:   time.Now(),
	}
	h.messages <- leaveMessage
}

// GetChatHistory returns chat message history for a pool
func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	poolID := c.Param("id")
	userID := c.GetString("user_id")

	// Verify user has access to this pool
	var memberRole string
	err := h.db.QueryRow(`
		SELECT role FROM pool_memberships 
		WHERE pool_id = $1 AND user_id = $2 AND is_active = true
	`, poolID, userID).Scan(&memberRole)

	if err == sql.ErrNoRows {
		response.Error(c, http.StatusForbidden, "Access denied to this pool")
		return
	}
	if err != nil {
		h.logger.Error("Failed to check pool membership", "pool_id", poolID, "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to validate access")
		return
	}

	// Get pagination parameters
	limit := 50 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get chat messages
	messages, err := h.getChatHistory(poolID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get chat history", "pool_id", poolID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve chat history")
		return
	}

	response.Success(c, gin.H{
		"pool_id":  poolID,
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
	})
}

// SendMessage allows sending a chat message via REST API (alternative to WebSocket)
func (h *ChatHandler) SendMessage(c *gin.Context) {
	poolID := c.Param("id")
	userID := c.GetString("user_id")

	// Verify user has access to this pool
	var memberRole string
	err := h.db.QueryRow(`
		SELECT role FROM pool_memberships 
		WHERE pool_id = $1 AND user_id = $2 AND is_active = true
	`, poolID, userID).Scan(&memberRole)

	if err == sql.ErrNoRows {
		response.Error(c, http.StatusForbidden, "Access denied to this pool")
		return
	}
	if err != nil {
		h.logger.Error("Failed to check pool membership", "pool_id", poolID, "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to validate access")
		return
	}

	var req struct {
		Message string `json:"message" binding:"required,min=1,max=1000"`
		Type    string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Default to user message type
	if req.Type == "" {
		req.Type = "user"
	}

	// Get user display name
	var displayName string
	err = h.db.QueryRow(`
		SELECT display_name FROM user_profiles WHERE id = $1
	`, userID).Scan(&displayName)
	if err != nil {
		h.logger.Error("Failed to get user display name", "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get user information")
		return
	}

	// Create chat message
	chatMessage := models.ChatMessage{
		PoolID:      poolID,
		UserID:      userID,
		DisplayName: displayName,
		Message:     req.Message,
		MessageType: req.Type,
		Timestamp:   time.Now(),
	}

	// Save to database
	err = h.saveMessage(chatMessage)
	if err != nil {
		h.logger.Error("Failed to save chat message", "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to send message")
		return
	}

	// Broadcast to WebSocket clients
	h.messages <- chatMessage

	response.Success(c, gin.H{
		"message": "Message sent successfully",
		"id":      chatMessage.ID,
	})
}

// Helper functions

func (h *ChatHandler) registerClient(poolID, userID string, conn *websocket.Conn) {
	if h.clients[poolID] == nil {
		h.clients[poolID] = make(map[*websocket.Conn]string)
	}
	h.clients[poolID][conn] = userID
	h.logger.Info("Client registered", "pool_id", poolID, "user_id", userID)
}

func (h *ChatHandler) unregisterClient(poolID string, conn *websocket.Conn) {
	if h.clients[poolID] != nil {
		if userID, exists := h.clients[poolID][conn]; exists {
			delete(h.clients[poolID], conn)
			h.logger.Info("Client unregistered", "pool_id", poolID, "user_id", userID)
			
			if len(h.clients[poolID]) == 0 {
				delete(h.clients, poolID)
			}
		}
	}
}

func (h *ChatHandler) handleMessages() {
	for {
		message := <-h.messages
		
		// Broadcast to all clients in the pool
		if clients, exists := h.clients[message.PoolID]; exists {
			for conn := range clients {
				err := conn.WriteJSON(message)
				if err != nil {
					h.logger.Error("Failed to send message to client", "error", err)
					conn.Close()
					delete(clients, conn)
				}
			}
		}
	}
}

func (h *ChatHandler) saveMessage(message models.ChatMessage) error {
	_, err := h.db.Exec(`
		INSERT INTO chat_messages (pool_id, user_id, message, message_type, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, message.PoolID, message.UserID, message.Message, message.MessageType, message.Timestamp)
	return err
}

func (h *ChatHandler) getChatHistory(poolID string, limit, offset int) ([]models.ChatMessage, error) {
	rows, err := h.db.Query(`
		SELECT 
			cm.id, cm.pool_id, cm.user_id, up.display_name,
			cm.message, cm.message_type, cm.created_at
		FROM chat_messages cm
		JOIN user_profiles up ON cm.user_id = up.id
		WHERE cm.pool_id = $1
		ORDER BY cm.created_at DESC
		LIMIT $2 OFFSET $3
	`, poolID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(
			&msg.ID, &msg.PoolID, &msg.UserID, &msg.DisplayName,
			&msg.Message, &msg.MessageType, &msg.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse the slice to show oldest messages first
	for i := len(messages)/2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
	}

	return messages, rows.Err()
}
