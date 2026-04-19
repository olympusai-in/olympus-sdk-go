package main

import (
	"fmt"
	"os"
	"time"

	olympus "github.com/olympusai-in/olympus-sdk-go"
)

func main() {
	apiKey := os.Getenv("OLYMPUS_API_KEY")
	if apiKey == "" {
		apiKey = "ol_YOUR_KEY_HERE"
	}

	client := olympus.New(olympus.Config{
		APIKey:        apiKey,
		Service:       "my-go-app",
		Endpoint:      "https://api.olympusai.in",
		FlushInterval: 5 * time.Second,
	})
	defer client.Close()

	client.Info("Application started")
	client.Info("Connected to database")
	client.Warn("Cache miss for key=user:profile:123")
	client.Error("Failed to process payment — timeout after 30s")
	client.Debug("Query executed in 42ms")

	fmt.Println("Logs buffered. Flushing...")
	client.Flush()
	fmt.Println("Done!")
}
