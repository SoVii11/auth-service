package ws

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/SoVii11/auth-service/services/appointment-service/internal/domain"
	"github.com/SoVii11/auth-service/services/appointment-service/internal/repository"
	sharedJWT "github.com/SoVii11/shared/pkg/jwt"
	"github.com/SoVii11/shared/pkg/response"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type client struct {
	conn    *websocket.Conn
	userID  int64
	isAdmin bool
}

type ChatController struct {
	clients     map[int64]*client
	adminClient *client
	mu          sync.Mutex
	messageRepo *repository.MessageRepository
	log         *zap.Logger
	jwtSecret   string
}

func NewChatController(messageRepo *repository.MessageRepository, log *zap.Logger, jwtSecret string) *ChatController {
	return &ChatController{
		clients:     make(map[int64]*client),
		messageRepo: messageRepo,
		log:         log,
		jwtSecret:   jwtSecret,
	}
}

type wsMessage struct {
	Text   string `json:"text"`
	UserID int64  `json:"user_id,omitempty"`
}

func (c *ChatController) HandleUserChat(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	claims, err := sharedJWT.ParseToken(tokenStr, c.jwtSecret)
	if err != nil {
		response.Unauthorized(w, "unauthorized")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	cl := &client{conn: conn, userID: claims.UserID}
	c.mu.Lock()
	c.clients[claims.UserID] = cl
	c.mu.Unlock()

	messages, err := c.messageRepo.GetByUserID(claims.UserID)
	if err == nil {
		for _, msg := range messages {
			_ = conn.WriteJSON(msg)
		}
	}

	defer func() {
		c.mu.Lock()
		delete(c.clients, claims.UserID)
		c.mu.Unlock()
		conn.Close()
	}()

	for {
		var msg wsMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		dbMsg := &domain.Message{UserID: claims.UserID, Text: msg.Text, IsAdmin: false}
		if err := c.messageRepo.Create(dbMsg); err != nil {
			continue
		}

		c.mu.Lock()
		admin := c.adminClient
		c.mu.Unlock()

		if admin != nil {
			_ = admin.conn.WriteJSON(map[string]any{
				"user_id":  claims.UserID,
				"text":     msg.Text,
				"is_admin": false,
			})
		}
	}
}

func (c *ChatController) HandleAdminChat(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	claims, err := sharedJWT.ParseToken(tokenStr, c.jwtSecret)
	if err != nil || claims.Role != "admin" {
		response.Unauthorized(w, "unauthorized")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	admin := &client{conn: conn, userID: claims.UserID, isAdmin: true}
	c.mu.Lock()
	c.adminClient = admin
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.adminClient = nil
		c.mu.Unlock()
		conn.Close()
	}()

	for {
		var msg wsMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		dbMsg := &domain.Message{UserID: msg.UserID, Text: msg.Text, IsAdmin: true}
		if err := c.messageRepo.Create(dbMsg); err != nil {
			continue
		}

		c.mu.Lock()
		userClient, ok := c.clients[msg.UserID]
		c.mu.Unlock()

		if ok {
			_ = userClient.conn.WriteJSON(map[string]any{
				"user_id":  msg.UserID,
				"text":     msg.Text,
				"is_admin": true,
			})
		}
	}
}

func (c *ChatController) GetMyChat(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}
	claims, err := sharedJWT.ParseToken(tokenStr, c.jwtSecret)
	if err != nil {
		response.Unauthorized(w, "unauthorized")
		return
	}

	messages, err := c.messageRepo.GetByUserID(claims.UserID)
	if err != nil {
		response.Internal(w, "failed to get messages")
		return
	}
	response.Success(w, messages)
}

func (c *ChatController) GetAllChats(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}
	claims, err := sharedJWT.ParseToken(tokenStr, c.jwtSecret)
	if err != nil || claims.Role != "admin" {
		response.Forbidden(w, "admins only")
		return
	}

	chats, err := c.messageRepo.GetAllChats()
	if err != nil {
		response.Internal(w, "failed to get chats")
		return
	}
	response.Success(w, chats)
}

func (c *ChatController) GetChatHistory(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}
	claims, err := sharedJWT.ParseToken(tokenStr, c.jwtSecret)
	if err != nil || claims.Role != "admin" {
		response.Forbidden(w, "admins only")
		return
	}

	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid user id")
		return
	}

	messages, err := c.messageRepo.GetByUserID(userID)
	if err != nil {
		response.Internal(w, "failed to get messages")
		return
	}
	response.Success(w, messages)
}

func (c *ChatController) ExportChatJSON(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}
	claims, err := sharedJWT.ParseToken(tokenStr, c.jwtSecret)
	if err != nil || claims.Role != "admin" {
		response.Forbidden(w, "admins only")
		return
	}

	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid user id")
		return
	}

	messages, err := c.messageRepo.GetByUserID(userID)
	if err != nil {
		response.Internal(w, "failed to get messages")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=chat.json")
	_ = json.NewEncoder(w).Encode(messages)
}
