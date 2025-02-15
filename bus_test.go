package membus

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNewInstance(t *testing.T) {
	logger := slog.Default()
	bus := NewInstance(logger)

	if bus == nil {
		t.Fatal("NewInstance returned nil")
	}
	if bus.queues == nil {
		t.Error("queues map not initialized")
	}
	if bus.msgCh == nil {
		t.Error("message channel not initialized")
	}
	if bus.log == nil {
		t.Error("logger not initialized")
	}
}

func TestSubscribe(t *testing.T) {
	logger := slog.Default()
	bus := NewInstance(logger)

	type TestMessage struct {
		Data string
	}

	Subscribe(bus, "TestService", "test.event", func(ctx context.Context, msg TestMessage) error {
		return nil
	})

	queues, exists := bus.queues["test.event"]
	if !exists {
		t.Fatal("event queue not created")
	}
	if len(queues) != 1 {
		t.Fatal("expected 1 queue, got", len(queues))
	}
	if queues[0].service != "TestService" {
		t.Fatal("wrong service name")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	logger := slog.Default()
	bus := NewInstance(logger)

	type TestMessage struct {
		Data string
	}

	Subscribe(bus, "Service1", "test.event", func(ctx context.Context, msg TestMessage) error {
		return nil
	})
	Subscribe(bus, "Service2", "test.event", func(ctx context.Context, msg TestMessage) error {
		return nil
	})

	queues, exists := bus.queues["test.event"]
	if !exists {
		t.Fatal("event queue not created")
	}
	if len(queues) != 2 {
		t.Fatal("expected 2 queues, got", len(queues))
	}
}

func TestPublishAndReceive(t *testing.T) {
	logger := slog.Default()
	bus := NewInstance(logger)

	type TestMessage struct {
		Data string
	}

	var wg sync.WaitGroup
	wg.Add(1)

	receivedData := ""
	Subscribe(bus, "TestService", "test.event", func(ctx context.Context, msg TestMessage) error {
		defer wg.Done()
		receivedData = msg.Data
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go bus.Start(ctx)

	testMsg := TestMessage{Data: "test data"}
	bus.Publish(ctx, "test.event", testMsg)

	wg.Wait()

	if receivedData != testMsg.Data {
		t.Errorf("expected to receive %q, got %q", testMsg.Data, receivedData)
	}
}

func TestContextCancellation(t *testing.T) {
	logger := slog.Default()
	bus := NewInstance(logger)

	type TestMessage struct {
		Data string
	}

	processed := false
	Subscribe(bus, "TestService", "test.event", func(ctx context.Context, msg TestMessage) error {
		processed = true
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	go bus.Start(ctx)

	// Cancel context immediately
	cancel()

	// Try to publish after cancellation
	bus.Publish(ctx, "test.event", TestMessage{Data: "test"})

	// Give some time for potential processing
	time.Sleep(100 * time.Millisecond)

	if processed {
		t.Error("message was processed after context cancellation")
	}
}

func TestConcurrentPublish(t *testing.T) {
	logger := slog.Default()
	bus := NewInstance(logger)

	type TestMessage struct {
		Data string
	}

	messageCount := 100
	var wg sync.WaitGroup
	wg.Add(messageCount)

	receivedCount := 0
	var mu sync.Mutex

	Subscribe(bus, "TestService", "test.event", func(ctx context.Context, msg TestMessage) error {
		mu.Lock()
		defer mu.Unlock()
		defer wg.Done()
		receivedCount++
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go bus.Start(ctx)

	// Publish messages concurrently
	for i := 0; i < messageCount; i++ {
		go func(i int) {
			bus.Publish(ctx, "test.event", TestMessage{Data: strconv.Itoa(i)})
		}(i)
	}

	wg.Wait()

	if receivedCount != messageCount {
		t.Errorf("expected to receive %d messages, got %d", messageCount, receivedCount)
	}
}
