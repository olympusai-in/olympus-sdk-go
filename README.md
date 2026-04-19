# Olympus Go SDK

Official Go SDK for [Olympus](https://olympusai.in) — AI-powered log analysis and debugging platform.

## Installation

```bash
go get github.com/olympusai-in/olympus-sdk-go
```

## Quick Start

```go
import (
    "time"
    olympus "github.com/olympusai-in/olympus-sdk-go"
)

client := olympus.New(olympus.Config{
    APIKey:  "ol_your_key_here",
    Service: "payment-svc",
})
defer client.Close()

client.Info("Payment processed — $14.99")
client.Warn("Retry attempt 2/3")
client.Error("Connection timeout after 30s")
client.Debug("Query executed in 42ms")
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `APIKey` | string | required | Your Olympus API key (`ol_...`) |
| `Service` | string | required | Service name for log grouping |
| `Endpoint` | string | `https://api.olympusai.in` | Olympus server URL |
| `FlushInterval` | time.Duration | `10s` | Duration between auto-flushes |
| `BatchSize` | int | `100` | Max logs per flush |

## Features

- Thread-safe (mutex-protected buffer)
- Auto-flushes every 10s in a goroutine
- Auto-flushes when buffer hits batch size
- Retries on failure (puts logs back in buffer)
- `Close()` flushes and stops the goroutine
- Zero dependencies (stdlib only)
