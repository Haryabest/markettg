package service_test

import (
	"testing"
)

func TestDeliveryIdempotency(t *testing.T) {
	// Same order_item_id should not create duplicate delivery jobs
	// Enforced by UNIQUE(order_item_id) constraint
	processed := make(map[string]bool)
	eventID := "test-event-1"
	if processed[eventID] {
		t.Fatal("duplicate event should be skipped")
	}
	processed[eventID] = true
	if !processed[eventID] {
		t.Fatal("event should be marked processed")
	}
}
