package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/SoVii11/auth-service/internal/entities"
	"github.com/SoVii11/auth-service/internal/infrastructure/repository"
	sharedJWT "github.com/SoVii11/shared/pkg/jwt"
	"github.com/SoVii11/shared/pkg/response"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // разрешаем все origins для разработки
	},
}

// клиент — одно WebSocket соединение
type client struct {
	conn    *websocket.Conn
	userID  int64
	isAdmin bool
}

// ChatController хранит все активные соединения
type ChatController struct {
	clients     map[int64]*client // userID -> client
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

// HandleUserChat godoc
// @Summary      WebSocket чат пользователя
// @Description  Подключение пользователя к чату с администратором через WebSocket
// @Tags         chat
// @Param        token  query  string  true  "JWT токен"
// @Router       /ws/chat [get]
func (c *ChatController) HandleUserChat(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	claims, err := sharedJWT.ParseToken(tokenStr, c.jwtSecret)
	if err != nil {
		response.Unauthorized(w, "unauthorized")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.log.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	cl := &client{conn: conn, userID: claims.UserID, isAdmin: false}

	c.mu.Lock()
	c.clients[claims.UserID] = cl
	c.mu.Unlock()

	c.log.Info("user connected to chat", zap.Int64("user_id", claims.UserID))

	// отправляем историю сообщений
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

		// сохраняем в БД
		dbMsg := &entities.Message{
			UserID:  claims.UserID,
			Text:    msg.Text,
			IsAdmin: false,
		}
		if err := c.messageRepo.Create(dbMsg); err != nil {
			c.log.Error("failed to save message", zap.Error(err))
			continue
		}

		c.log.Info("message from user", zap.Int64("user_id", claims.UserID), zap.String("text", msg.Text))

		// если админ онлайн — отправляем ему
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

// HandleAdminChat godoc
// @Summary      WebSocket чат администратора
// @Description  Подключение администратора к чату через WebSocket
// @Tags         chat
// @Param        token  query  string  true  "JWT токен администратора"
// @Router       /ws/admin/chat [get]
func (c *ChatController) HandleAdminChat(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	claims, err := sharedJWT.ParseToken(tokenStr, c.jwtSecret)
	if err != nil || claims.Role != "admin" {
		response.Unauthorized(w, "unauthorized")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.log.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	admin := &client{conn: conn, userID: claims.UserID, isAdmin: true}

	c.mu.Lock()
	c.adminClient = admin
	c.mu.Unlock()

	c.log.Info("admin connected to chat")

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

		// сохраняем в БД
		dbMsg := &entities.Message{
			UserID:  msg.UserID,
			Text:    msg.Text,
			IsAdmin: true,
		}
		if err := c.messageRepo.Create(dbMsg); err != nil {
			c.log.Error("failed to save message", zap.Error(err))
			continue
		}

		c.log.Info("message from admin", zap.Int64("to_user", msg.UserID), zap.String("text", msg.Text))

		// отправляем пользователю если онлайн
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

// GetChatHistory godoc
// @Summary      История чата пользователя (админ)
// @Description  Возвращает историю сообщений конкретного пользователя
// @Tags         chat
// @Param        Authorization  header  string  true  "Bearer токен"
// @Param        id             path    int     true  "ID пользователя"
// @Success      200            {object}  map[string]any
// @Failure      403            {object}  map[string]string
// @Router       /admin/chat/{id} [get]
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

// GetAllChats godoc
// @Summary      Список всех чатов (админ)
// @Description  Возвращает последнее сообщение от каждого пользователя
// @Tags         chat
// @Param        Authorization  header  string  true  "Bearer токен"
// @Success      200            {object}  map[string]any
// @Failure      403            {object}  map[string]string
// @Router       /admin/chats [get]
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

	_ = claims

	chats, err := c.messageRepo.GetAllChats()
	if err != nil {
		response.Internal(w, "failed to get chats")
		return
	}

	// сериализуем вручную чтобы добавить json тег
	type chatItem struct {
		UserID  int64  `json:"user_id"`
		LastMsg string `json:"last_message"`
		IsAdmin bool   `json:"is_admin"`
	}

	result := make([]chatItem, 0, len(chats))
	for _, m := range chats {
		result = append(result, chatItem{
			UserID:  m.UserID,
			LastMsg: m.Text,
			IsAdmin: m.IsAdmin,
		})
	}

	response.Success(w, result)
}

// GetMyChat godoc
// @Summary      История моего чата
// @Description  Возвращает историю сообщений авторизованного пользователя
// @Tags         chat
// @Param        Authorization  header  string  true  "Bearer токен"
// @Success      200            {object}  map[string]any
// @Failure      401            {object}  map[string]string
// @Router       /chat/history [get]
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

// ExportChatJSON godoc
// @Summary      Экспорт чата в JSON (админ)
// @Description  Возвращает историю чата пользователя в формате JSON файла
// @Tags         chat
// @Param        Authorization  header  string  true  "Bearer токен"
// @Param        id             path    int     true  "ID пользователя"
// @Success      200
// @Failure      403  {object}  map[string]string
// @Router       /admin/chat/{id}/export [get]
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
