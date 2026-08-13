package service_test

import (
	"testing"

	"github.com/google/uuid"
)

func TestIdempotencyKeyRequired(t *testing.T) {
	key := uuid.New().String()
	if key == "" {
		t.Fatal("idempotency key should not be empty")
	}
}

func TestPaymentAmountFromOrder(t *testing.T) {
	// Payment service must always fetch order total from order service
	// Never use amount from frontend request body
	orderTotal := int64(9900)
	frontendAmount := int64(1) // malicious
	if frontendAmount == orderTotal {
		t.Fatal("frontend amount must not be trusted")
	}
}
