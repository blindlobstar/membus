# membus

`membus` is a lightweight, type-safe in-memory message bus implementation for Go that enables publish-subscribe (pub/sub) pattern communication between different parts of your application.

## Features

- Type-safe message handling with Go generics
- Asynchronous message processing
- Multiple subscribers per message type
- Buffered channels for better performance
- Zero external dependencies
- Built-in error logging with `slog`
- Simple, clean API

## Installation

```bash
go get github.com/blindlobstar/membus
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/blindlobstar/membus"
)

// Define your message type
type UserCreated struct {
    ID   string
    Name string
}

func main() {
    // Create a new message bus
    logger := slog.Default()
    bus := membus.NewInstance(logger)

    // Subscribe to messages
    membus.Subscribe(bus, "EmailService", "user.created", func(ctx context.Context, user UserCreated) error {
        fmt.Printf("Sending welcome email to: %s\n", user.Name)
        return nil
    })

    // Create context
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Start the bus
    go bus.Start(ctx)

    // Publish a message
    bus.Publish(ctx, "user.created", UserCreated{
        ID:   "123",
        Name: "John Doe",
    })

    // Wait for processing
    time.Sleep(time.Second)
}
```

## Usage Guide

### Creating a Message Bus

Create a new message bus instance by providing a `slog.Logger`:

```go
logger := slog.Default()
bus := membus.NewInstance(logger)
```

### Subscribing to Messages

Subscribe to messages using the generic `Subscribe` function. Specify the service name, event name, and a handler function:

```go
membus.Subscribe(bus, "ServiceName", "event.name", func(ctx context.Context, msg YourMessageType) error {
    // Handle the message
    return nil
})
```

The handler function is type-safe and will only receive messages of the specified type.

### Publishing Messages

Publish messages to the bus using the `Publish` method:

```go
bus.Publish(ctx, "event.name", message)
```

### Starting the Bus

Start processing messages by calling the `Start` method with a context:

```go
ctx := context.Background()
bus.Start(ctx)
```

The bus will continue processing messages until the context is cancelled.

## Error Handling

The message bus automatically logs errors that occur during message processing using the provided `slog.Logger`. Errors include:
- Type mismatch errors
- Handler execution errors

Error logs include the service name, event name, and the error details for debugging.

## Example: Multiple Subscribers

```go
// Define message type
type OrderCreated struct {
    OrderID string
    Amount  float64
}

// Subscribe multiple services
membus.Subscribe(bus, "EmailService", "order.created", func(ctx context.Context, order OrderCreated) error {
    fmt.Printf("Sending order confirmation email for order: %s\n", order.OrderID)
    return nil
})

membus.Subscribe(bus, "AnalyticsService", "order.created", func(ctx context.Context, order OrderCreated) error {
    fmt.Printf("Recording order metrics for order: %s\n", order.OrderID)
    return nil
})

membus.Subscribe(bus, "InventoryService", "order.created", func(ctx context.Context, order OrderCreated) error {
    fmt.Printf("Updating inventory for order: %s\n", order.OrderID)
    return nil
})
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.