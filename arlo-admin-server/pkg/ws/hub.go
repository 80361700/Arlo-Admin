package ws

import (
	"encoding/json"
	"sync"

	"arlo-admin/pkg/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// EventUnreadChanged 通知客户端重新拉取未读数
const EventUnreadChanged = "unread_changed"

type envelope struct {
	Type string `json:"type"`
}

// Client 单个 WebSocket 连接
type Client struct {
	UserID uint64
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub 按用户维护连接（同一用户可多标签多连接）
type Hub struct {
	mu      sync.RWMutex
	clients map[uint64]map[*Client]struct{}
}

var defaultHub = NewHub()

// Default 全局 Hub（单进程）
func Default() *Hub { return defaultHub }

func NewHub() *Hub {
	return &Hub{clients: make(map[uint64]map[*Client]struct{})}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.UserID] == nil {
		h.clients[c.UserID] = make(map[*Client]struct{})
	}
	h.clients[c.UserID][c] = struct{}{}
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set := h.clients[c.UserID]
	if set == nil {
		return
	}
	if _, ok := set[c]; ok {
		delete(set, c)
		close(c.Send)
	}
	if len(set) == 0 {
		delete(h.clients, c.UserID)
	}
}

func (h *Hub) NotifyUsers(userIDs ...uint64) {
	if len(userIDs) == 0 {
		return
	}
	payload, err := json.Marshal(envelope{Type: EventUnreadChanged})
	if err != nil {
		return
	}
	seen := make(map[uint64]struct{}, len(userIDs))
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, uid := range userIDs {
		if uid == 0 {
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		for c := range h.clients[uid] {
			select {
			case c.Send <- payload:
			default:
				logger.Logger.Warn("ws send buffer full, drop", zap.Uint64("userId", uid))
			}
		}
	}
}

// NotifyAll 广播未读变更（全员站内信）
func (h *Hub) NotifyAll() {
	payload, err := json.Marshal(envelope{Type: EventUnreadChanged})
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for uid, set := range h.clients {
		for c := range set {
			select {
			case c.Send <- payload:
			default:
				logger.Logger.Warn("ws send buffer full, drop", zap.Uint64("userId", uid))
			}
		}
	}
}
