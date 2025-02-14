package membus

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Instance struct {
	queues map[string][]*queue
	msgCh  chan message
	log    *slog.Logger
}

type queue struct {
	service string
	ch      chan event
	h       func(context.Context, any) error
}

type event struct {
	body any
}

type message struct {
	name string
	body any
}

func NewInstance(log *slog.Logger) *Instance {
	return &Instance{
		queues: map[string][]*queue{},
		msgCh:  make(chan message, 4096),
		log:    log,
	}
}

func Subscribe[T any](bus *Instance, service string, name string, handler func(context.Context, T) error) {
	h := func(ctx context.Context, e any) error {
		te, ok := e.(T)
		if !ok {
			return fmt.Errorf("service %s can't handle wrong type of %s event", service, name)
		}
		return handler(ctx, te)
	}
	if queues, ok := bus.queues[name]; ok {
		bus.queues[name] = append(queues, &queue{service: service, ch: make(chan event), h: h})
	} else {
		bus.queues[name] = []*queue{{service: service, ch: make(chan event, 4096), h: h}}
	}
}

func (bus *Instance) Publish(ctx context.Context, name string, e any) {
	bus.msgCh <- message{name: name, body: e}
}

func (bus *Instance) Start(ctx context.Context) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for name, queues := range bus.queues {
		for _, q := range queues {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case e := <-q.ch:
						bus.processEvent(ctx, *q, name, e)
					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}

	for {
		select {
		case msg := <-bus.msgCh:
			if queues, ok := bus.queues[msg.name]; ok {
				for _, queue := range queues {
					queue.ch <- event{body: msg.body}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (bus *Instance) processEvent(ctx context.Context, q queue, name string, e event) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := q.h(ctx, e.body); err != nil {
		bus.log.ErrorContext(ctx, "error handling event", "service", q.service, "name", name, "error", err)
	}
}
