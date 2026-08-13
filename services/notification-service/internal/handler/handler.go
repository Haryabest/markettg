package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/markettg/markettg/services/notification-service/internal/ws"
)

type Handler struct {
	hub *ws.Hub
}

func New(hub *ws.Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) OrderWebSocket(c *websocket.Conn) {
	orderID := c.Params("id")
	h.hub.Subscribe(orderID, c)
	defer h.hub.Unsubscribe(orderID, c)

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			break
		}
	}
}

func WebSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}
