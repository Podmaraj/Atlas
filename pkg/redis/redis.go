package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"edgecore/internal/config"
	"edgecore/internal/logger"
)

// ConfigEventType enum for dynamic Pub/Sub notifications
type ConfigEventType string

const (
	EventRouteUpdated   ConfigEventType = "ROUTE_UPDATED"
	EventRouteDeleted   ConfigEventType = "ROUTE_DELETED"
	EventServiceUpdated ConfigEventType = "SERVICE_UPDATED"
	EventServiceDeleted ConfigEventType = "SERVICE_DELETED"
	EventPluginUpdated  ConfigEventType = "PLUGIN_UPDATED"
	EventPluginDeleted  ConfigEventType = "PLUGIN_DELETED"
	EventApiKeyUpdated  ConfigEventType = "APIKEY_UPDATED"
	EventApiKeyDeleted  ConfigEventType = "APIKEY_DELETED"
	EventCachePurged    ConfigEventType = "CACHE_PURGED"
)

// ConfigEvent represents a Pub/Sub payload sent when Gateway config changes
type ConfigEvent struct {
	Type      ConfigEventType `json:"type"`
	TargetID  string          `json:"target_id"`
	Scope     string          `json:"scope,omitempty"`
	Payload   string          `json:"payload,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// Client wraps redis.Client with domain-specific Gateway operations
type Client struct {
	rdb     *redis.Client
	channel string
}

// NewClient initializes a Redis connection pool using app config
func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis server at %s: %w", cfg.Addr, err)
	}

	logger.Info("Redis client connected successfully", zap.String("addr", cfg.Addr))

	return &Client{
		rdb:     rdb,
		channel: cfg.PubSubChannel,
	}, nil
}

// Raw returns the underlying *redis.Client for custom queries
func (c *Client) Raw() *redis.Client {
	return c.rdb
}

// Close closes the Redis connection pool
func (c *Client) Close() error {
	return c.rdb.Close()
}

// PublishConfigEvent broadcasts a configuration change event to all Data Plane instances
func (c *Client) PublishConfigEvent(ctx context.Context, event ConfigEvent) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal config event: %w", err)
	}

	if err := c.rdb.Publish(ctx, c.channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish config event to channel %s: %w", c.channel, err)
	}

	logger.Debug("Published Redis config event",
		zap.String("type", string(event.Type)),
		zap.String("target_id", event.TargetID),
	)

	return nil
}

// SubscribeConfigEvents listens for Pub/Sub events and invokes the provided handler callback
func (c *Client) SubscribeConfigEvents(ctx context.Context, handler func(event ConfigEvent)) error {
	pubsub := c.rdb.Subscribe(ctx, c.channel)

	// Wait for subscription confirmation
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe to redis channel %s: %w", c.channel, err)
	}

	logger.Info("Subscribed to Redis config updates channel", zap.String("channel", c.channel))

	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Stopping Redis PubSub subscription listener")
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var event ConfigEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					logger.Error("Failed to unmarshal received config event",
						zap.Error(err),
						zap.String("payload", msg.Payload),
					)
					continue
				}
				handler(event)
			}
		}
	}()

	return nil
}

// Cache helper functions
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}
