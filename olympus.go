package olympus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	LevelInfo  = 0
	LevelWarn  = 1
	LevelError = 2
	LevelDebug = 3
)

type Config struct {
	APIKey        string
	Service       string
	Endpoint      string        // default: https://api.olympusai.in
	FlushInterval time.Duration // default: 10s
	BatchSize     int           // default: 100
}

type logEntry struct {
	Level     int    `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type ingestPayload struct {
	Service string     `json:"service"`
	Logs    []logEntry `json:"logs"`
}

type Client struct {
	config Config
	buffer []logEntry
	mu     sync.Mutex
	done   chan struct{}
	client *http.Client
}

func New(cfg Config) *Client {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.olympusai.in"
	}
	if cfg.FlushInterval == 0 {
		cfg.FlushInterval = 10 * time.Second
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}

	c := &Client{
		config: cfg,
		buffer: make([]logEntry, 0, cfg.BatchSize),
		done:   make(chan struct{}),
		client: &http.Client{Timeout: 10 * time.Second},
	}

	go c.autoFlush()

	return c
}

func (c *Client) Info(message string) {
	c.push(LevelInfo, message)
}

func (c *Client) Warn(message string) {
	c.push(LevelWarn, message)
}

func (c *Client) Error(message string) {
	c.push(LevelError, message)
}

func (c *Client) Debug(message string) {
	c.push(LevelDebug, message)
}

func (c *Client) push(level int, message string) {
	c.mu.Lock()
	c.buffer = append(c.buffer, logEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	})
	shouldFlush := len(c.buffer) >= c.config.BatchSize
	c.mu.Unlock()

	if shouldFlush {
		c.Flush()
	}
}

func (c *Client) Flush() error {
	c.mu.Lock()
	if len(c.buffer) == 0 {
		c.mu.Unlock()
		return nil
	}

	logs := make([]logEntry, len(c.buffer))
	copy(logs, c.buffer)
	c.buffer = c.buffer[:0]
	c.mu.Unlock()

	payload := ingestPayload{
		Service: c.config.Service,
		Logs:    logs,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		c.mu.Lock()
		c.buffer = append(logs, c.buffer...)
		c.mu.Unlock()
		return fmt.Errorf("[Olympus] marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.Endpoint+"/api/v1/ingest", bytes.NewReader(body))
	if err != nil {
		c.mu.Lock()
		c.buffer = append(logs, c.buffer...)
		c.mu.Unlock()
		return fmt.Errorf("[Olympus] request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		c.mu.Lock()
		c.buffer = append(logs, c.buffer...)
		c.mu.Unlock()
		return fmt.Errorf("[Olympus] network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		c.mu.Lock()
		c.buffer = append(logs, c.buffer...)
		c.mu.Unlock()
		return fmt.Errorf("[Olympus] server returned %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) autoFlush() {
	ticker := time.NewTicker(c.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.Flush(); err != nil {
				fmt.Println(err)
			}
		case <-c.done:
			return
		}
	}
}

// Close flushes remaining logs and stops the auto-flush goroutine.
func (c *Client) Close() error {
	close(c.done)
	return c.Flush()
}
