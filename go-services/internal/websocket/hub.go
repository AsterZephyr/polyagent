package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/polyagent/go-services/internal/metrics"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 在生产环境中应该更严格地检查源
	},
}

// Message WebSocket消息结构
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	SessionID string      `json:"session_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
}

// Client WebSocket客户端
type Client struct {
	ID        string
	SessionID string
	UserID    string
	Conn      *websocket.Conn
	Send      chan Message
	Hub       *Hub
}

// Hub WebSocket连接管理中心
type Hub struct {
	clients    map[*Client]bool
	sessions   map[string][]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	mutex      sync.RWMutex
}

// NewHub 创建新的WebSocket Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		sessions:   make(map[string][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
	}
}

// Run 运行Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			if client.SessionID != "" {
				h.sessions[client.SessionID] = append(h.sessions[client.SessionID], client)
			}
			h.mutex.Unlock()
			
			metrics.IncWebSocketConnections()
			log.Printf("Client %s connected (session: %s, user: %s)", 
				client.ID, client.SessionID, client.UserID)
			
			// 发送欢迎消息
			select {
			case client.Send <- Message{
				Type:      "connection",
				Data:      map[string]string{"status": "connected"},
				Timestamp: time.Now(),
			}:
			default:
				h.closeClient(client)
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.closeClient(client)
				metrics.DecWebSocketConnections()
				log.Printf("Client %s disconnected", client.ID)
			}

		case message := <-h.broadcast:
			h.mutex.RLock()
			if message.SessionID != "" {
				// 发送给特定会话的所有客户端
				for _, client := range h.sessions[message.SessionID] {
					select {
					case client.Send <- message:
					default:
						h.closeClient(client)
					}
				}
			} else {
				// 广播给所有客户端
				for client := range h.clients {
					select {
					case client.Send <- message:
					default:
						h.closeClient(client)
					}
				}
			}
			h.mutex.RUnlock()
			
			metrics.RecordWebSocketMessage(message.Type, "sent")
		}
	}
}

// closeClient 关闭客户端连接
func (h *Hub) closeClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		
		// 从会话中移除客户端
		if client.SessionID != "" {
			sessionClients := h.sessions[client.SessionID]
			for i, c := range sessionClients {
				if c == client {
					h.sessions[client.SessionID] = append(sessionClients[:i], sessionClients[i+1:]...)
					break
				}
			}
			
			// 如果会话中没有客户端了，删除会话
			if len(h.sessions[client.SessionID]) == 0 {
				delete(h.sessions, client.SessionID)
			}
		}
		
		close(client.Send)
		client.Conn.Close()
	}
}

// Broadcast 广播消息
func (h *Hub) Broadcast(message Message) {
	h.broadcast <- message
}

// BroadcastToSession 向特定会话广播消息
func (h *Hub) BroadcastToSession(sessionID string, message Message) {
	message.SessionID = sessionID
	h.broadcast <- message
}

// GetSessionClients 获取会话的所有客户端
func (h *Hub) GetSessionClients(sessionID string) []*Client {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	clients := make([]*Client, len(h.sessions[sessionID]))
	copy(clients, h.sessions[sessionID])
	return clients
}

// GetClientCount 获取连接的客户端数量
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// GetSessionCount 获取活跃会话数量
func (h *Hub) GetSessionCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.sessions)
}

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(hub *Hub, c *gin.Context) {
	sessionID := c.Param("session_id")
	userID := c.Query("user_id")
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		metrics.RecordWebSocketMessage("connection", "failed")
		return
	}

	client := &Client{
		ID:        generateClientID(),
		SessionID: sessionID,
		UserID:    userID,
		Conn:      conn,
		Send:      make(chan Message, 256),
		Hub:       hub,
	}

	// 注册客户端
	hub.register <- client

	// 启动读写goroutines
	go client.writePump()
	go client.readPump()
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// readPump 处理从WebSocket连接读取消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageData, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var message Message
		if err := json.Unmarshal(messageData, &message); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			metrics.RecordWebSocketMessage("message", "error")
			continue
		}

		message.Timestamp = time.Now()
		message.SessionID = c.SessionID
		message.UserID = c.UserID

		// 处理不同类型的消息
		switch message.Type {
		case "ping":
			c.Send <- Message{
				Type:      "pong",
				Data:      map[string]string{"status": "ok"},
				Timestamp: time.Now(),
			}
		case "chat":
			// 处理聊天消息，这里可以调用AI服务
			c.handleChatMessage(message)
		default:
			log.Printf("Unknown message type: %s", message.Type)
		}

		metrics.RecordWebSocketMessage(message.Type, "received")
	}
}

// writePump 处理向WebSocket连接写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			messageData, err := json.Marshal(message)
			if err != nil {
				log.Printf("JSON marshal error: %v", err)
				continue
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, messageData); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleChatMessage 处理聊天消息
func (c *Client) handleChatMessage(message Message) {
	// 这里应该调用聊天处理逻辑，暂时只是回显
	response := Message{
		Type:      "chat_response",
		Data:      map[string]interface{}{
			"message": "Received: " + message.Data.(string),
			"status":  "processed",
		},
		Timestamp: time.Now(),
		SessionID: c.SessionID,
		UserID:    c.UserID,
	}

	// 发送响应
	select {
	case c.Send <- response:
	default:
		c.Hub.unregister <- c
	}
}