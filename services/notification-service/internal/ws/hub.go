package ws

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type Hub struct {
	mu      sync.RWMutex
	orders  map[string]map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{orders: make(map[string]map[*websocket.Conn]bool)}
}

func (h *Hub) Subscribe(orderID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.orders[orderID] == nil {
		h.orders[orderID] = make(map[*websocket.Conn]bool)
	}
	h.orders[orderID][conn] = true
}

func (h *Hub) Unsubscribe(orderID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.orders[orderID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.orders, orderID)
		}
	}
}

func (h *Hub) Broadcast(orderID string, event interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, _ := json.Marshal(event)
	for conn := range h.orders[orderID] {
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
}

type OrderStatusEvent struct {
	Type    string `json:"type"`
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	UserID  string `json:"user_id"`
}

func (h *Hub) BroadcastOrderStatus(orderID, userID, status string) {
	h.Broadcast(orderID, OrderStatusEvent{
		Type: "order_status", OrderID: orderID, Status: status, UserID: userID,
	})
}

func ParseUserID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
