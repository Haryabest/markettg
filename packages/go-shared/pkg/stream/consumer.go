package stream

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	client  *redis.Client
	stream  string
	group   string
	consumer string
}

func NewConsumer(client *redis.Client, stream, group, consumer string) *Consumer {
	return &Consumer{client: client, stream: stream, group: group, consumer: consumer}
}

func (c *Consumer) EnsureGroup(ctx context.Context) error {
	err := c.client.XGroupCreateMkStream(ctx, c.stream, c.group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}

type Handler func(ctx context.Context, eventID string, payload []byte) error

func (c *Consumer) Run(ctx context.Context, handler Handler) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    c.group,
				Consumer: c.consumer,
				Streams:  []string{c.stream, ">"},
				Count:    10,
				Block:    5 * time.Second,
			}).Result()
			if err != nil {
				if err == redis.Nil {
					continue
				}
				time.Sleep(time.Second)
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					eventRaw, _ := msg.Values["event"].(string)
					if err := handler(ctx, msg.ID, []byte(eventRaw)); err != nil {
						continue
					}
					_ = c.client.XAck(ctx, c.stream, c.group, msg.ID).Err()
				}
			}
		}
	}
}

func UnmarshalEvent[T any](data []byte) (T, error) {
	var wrapper struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}
	var zero T
	if err := json.Unmarshal(data, &wrapper); err != nil {
		// try direct unmarshal
		var direct T
		if err2 := json.Unmarshal(data, &direct); err2 != nil {
			return zero, err
		}
		return direct, nil
	}
	if err := json.Unmarshal(wrapper.Payload, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}
