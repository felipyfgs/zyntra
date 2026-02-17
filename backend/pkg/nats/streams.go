package nats

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// Stream names
const (
	StreamMessages    = "MESSAGES"
	StreamConnections = "CONNECTIONS"
	StreamQR          = "QR"
)

// SetupStreams creates the JetStream streams for the application
func (c *Client) SetupStreams(ctx context.Context) error {
	log.Printf("[NATS] Setting up JetStream streams...")

	// Use a longer timeout for stream creation
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Messages stream - for chat messages
	if err := c.createOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        StreamMessages,
		Description: "Chat messages from WhatsApp connections",
		Subjects:    []string{"zyntra.messages.>"},
		Retention:   jetstream.LimitsPolicy,
		MaxAge:      7 * 24 * time.Hour, // Keep messages for 7 days
		MaxMsgs:     -1,                 // Unlimited messages
		MaxBytes:    1024 * 1024 * 1024, // 1GB max
		Discard:     jetstream.DiscardOld,
		Storage:     jetstream.FileStorage,
		Replicas:    1,
	}); err != nil {
		return fmt.Errorf("failed to create MESSAGES stream: %w", err)
	}

	// Connections stream - for connection status updates
	if err := c.createOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        StreamConnections,
		Description: "WhatsApp connection status updates",
		Subjects:    []string{"zyntra.connections.>"},
		Retention:   jetstream.LimitsPolicy,
		MaxAge:      24 * time.Hour, // Keep for 24 hours
		MaxMsgs:     10000,
		Discard:     jetstream.DiscardOld,
		Storage:     jetstream.FileStorage,
		Replicas:    1,
	}); err != nil {
		return fmt.Errorf("failed to create CONNECTIONS stream: %w", err)
	}

	// QR stream - for QR codes (short-lived)
	if err := c.createOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        StreamQR,
		Description: "QR codes for WhatsApp authentication",
		Subjects:    []string{"zyntra.qr.>"},
		Retention:   jetstream.LimitsPolicy,
		MaxAge:      5 * time.Minute, // QR codes expire quickly
		MaxMsgs:     1000,
		Discard:     jetstream.DiscardOld,
		Storage:     jetstream.MemoryStorage, // Use memory for speed
		Replicas:    1,
	}); err != nil {
		return fmt.Errorf("failed to create QR stream: %w", err)
	}

	log.Printf("[NATS] JetStream streams setup complete")
	return nil
}

// createOrUpdateStream creates or updates a stream
func (c *Client) createOrUpdateStream(ctx context.Context, config jetstream.StreamConfig) error {
	// Try to create or update stream
	_, err := c.js.CreateOrUpdateStream(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create/update stream %s: %w", config.Name, err)
	}
	log.Printf("[NATS] Created/updated stream: %s", config.Name)
	return nil
}

// CreateConsumer creates a consumer for a stream
func (c *Client) CreateConsumer(ctx context.Context, streamName, consumerName string, filterSubject string) (jetstream.Consumer, error) {
	stream, err := c.js.Stream(ctx, streamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream %s: %w", streamName, err)
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:          consumerName,
		Durable:       consumerName,
		FilterSubject: filterSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverNewPolicy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer %s: %w", consumerName, err)
	}

	return consumer, nil
}

// DeleteConsumer deletes a consumer
func (c *Client) DeleteConsumer(ctx context.Context, streamName, consumerName string) error {
	stream, err := c.js.Stream(ctx, streamName)
	if err != nil {
		return fmt.Errorf("failed to get stream %s: %w", streamName, err)
	}

	return stream.DeleteConsumer(ctx, consumerName)
}

// GetStreamInfo returns info about a stream
func (c *Client) GetStreamInfo(ctx context.Context, streamName string) (*jetstream.StreamInfo, error) {
	stream, err := c.js.Stream(ctx, streamName)
	if err != nil {
		return nil, err
	}
	return stream.Info(ctx)
}
