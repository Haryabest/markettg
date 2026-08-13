package events

const (
	UserCreated        = "UserCreated"
	OrderCreated       = "OrderCreated"
	OrderCompleted     = "OrderCompleted"
	OrderCancelled     = "OrderCancelled"
	PaymentCreated     = "PaymentCreated"
	PaymentSucceeded   = "PaymentSucceeded"
	PaymentFailed      = "PaymentFailed"
	DeliveryStarted    = "DeliveryStarted"
	DeliveryCompleted  = "DeliveryCompleted"
	DeliveryFailed     = "DeliveryFailed"
)

type OrderCreatedPayload struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Total   int64  `json:"total_kopecks"`
}

type PaymentSucceededPayload struct {
	PaymentID string `json:"payment_id"`
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	Amount    int64  `json:"amount_kopecks"`
	Method    string `json:"method"`
}

type DeliveryCompletedPayload struct {
	OrderID     string `json:"order_id"`
	OrderItemID string `json:"order_item_id"`
	UserID      string `json:"user_id"`
}

type OrderStatusPayload struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Status  string `json:"status"`
}
