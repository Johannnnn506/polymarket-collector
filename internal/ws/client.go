package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// DefaultWSURL is the default WebSocket URL for the CLOB market feed.
	DefaultWSURL = "wss://ws-subscriptions-clob.polymarket.com/ws/market"

	// Default reconnection parameters
	defaultInitialBackoff = 1 * time.Second
	defaultMaxBackoff     = 30 * time.Second
	defaultBackoffFactor  = 2.0
)

// MessageHandler is a callback function for handling parsed WebSocket messages.
type MessageHandler func(messages []WSMessage)

// ReconnectConfig configures the reconnection behavior.
type ReconnectConfig struct {
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
	MaxRetries     int // 0 = infinite
}

// DefaultReconnectConfig returns the default reconnection configuration.
func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		InitialBackoff: defaultInitialBackoff,
		MaxBackoff:     defaultMaxBackoff,
		BackoffFactor:  defaultBackoffFactor,
		MaxRetries:     0,
	}
}

// Client is a WebSocket client for the Polymarket CLOB L2 feed.
type Client struct {
	url             string
	handler         MessageHandler
	reconnectConfig ReconnectConfig

	mu          sync.Mutex
	conn        *websocket.Conn
	tokenIDs    []string
	isConnected bool
}

// NewWSClient creates a new WebSocket client.
func NewWSClient(handler MessageHandler) *Client {
	return &Client{
		url:             DefaultWSURL,
		handler:         handler,
		reconnectConfig: DefaultReconnectConfig(),
	}
}

// WithURL sets a custom WebSocket URL.
func (c *Client) WithURL(url string) *Client {
	c.url = url
	return c
}

// WithReconnectConfig sets the reconnection configuration.
func (c *Client) WithReconnectConfig(config ReconnectConfig) *Client {
	c.reconnectConfig = config
	return c
}

// Connect establishes the WebSocket connection and starts the read loop.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	if c.isConnected {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	return c.connectWithBackoff(ctx)
}

func (c *Client) connectWithBackoff(ctx context.Context) error {
	backoff := c.reconnectConfig.InitialBackoff
	retries := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.url, nil)
		if err == nil {
			c.mu.Lock()
			c.conn = conn
			c.isConnected = true
			c.mu.Unlock()

			// Re-subscribe to any previously subscribed tokens
			if len(c.tokenIDs) > 0 {
				if err := c.sendSubscribe(c.tokenIDs); err != nil {
					log.Printf("Warning: failed to resubscribe: %v", err)
				}
			}

			// Start the read loop
			go c.readLoop(ctx)
			return nil
		}

		retries++
		if c.reconnectConfig.MaxRetries > 0 && retries >= c.reconnectConfig.MaxRetries {
			return fmt.Errorf("max retries (%d) exceeded: %w", c.reconnectConfig.MaxRetries, err)
		}

		log.Printf("WebSocket connection failed (attempt %d): %v. Retrying in %v...", retries, err, backoff)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		backoff = time.Duration(float64(backoff) * c.reconnectConfig.BackoffFactor)
		if backoff > c.reconnectConfig.MaxBackoff {
			backoff = c.reconnectConfig.MaxBackoff
		}
	}
}

// Subscribe subscribes to updates for the given token IDs.
func (c *Client) Subscribe(tokenIDs []string) error {
	c.mu.Lock()
	c.tokenIDs = tokenIDs
	c.mu.Unlock()

	return c.sendSubscribe(tokenIDs)
}

func (c *Client) sendSubscribe(tokenIDs []string) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := SubscribeMessage{AssetsIDs: tokenIDs}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling subscribe message: %w", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("writing subscribe message: %w", err)
	}

	return nil
}

func (c *Client) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()

		if conn == nil {
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			c.mu.Lock()
			c.isConnected = false
			c.mu.Unlock()

			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("WebSocket closed normally")
				return
			}

			log.Printf("WebSocket read error: %v. Attempting reconnect...", err)

			// Attempt to reconnect
			go func() {
				if reconnErr := c.connectWithBackoff(ctx); reconnErr != nil {
					log.Printf("Reconnection failed: %v", reconnErr)
				}
			}()
			return
		}

		messages, err := Parse(data)
		if err != nil {
			log.Printf("Error parsing WebSocket message: %v", err)
			continue
		}

		if c.handler != nil && len(messages) > 0 {
			c.handler(messages)
		}
	}
}

// Close closes the WebSocket connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isConnected = false
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// IsConnected returns whether the client is currently connected.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isConnected
}
