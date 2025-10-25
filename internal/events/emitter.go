package events

import "context"

// New creates a new event emitter
func New() *EventEmitter {
	return &EventEmitter{
		listeners: make(map[string]map[*Listener]struct{}),
	}
}

// On subscribes to an event. It returns
// - recv channel for string payloads
// - unsubscribe func to remove this listener
//
// Provide a context to control the listener lifecycle. When ctx is cancelled,
// the listener is automatically removed and it's channel is closed
//
// buffer controls channel capacity to avoid blocking producers
func (e *EventEmitter) On(ctx context.Context, event string, buffer int) (<-chan string, func()) {
	if buffer <= 0 {
		buffer = 1
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		// return a closed channel and no-op unsubscribe when event emitter is closed
		ch := make(chan string)
		close(ch)
		return ch, func() {}
	}

	lctx, cancel := context.WithCancel(ctx)
	l := &Listener{
		ch:     make(chan string, buffer),
		cancel: cancel,
	}

	if _, ok := e.listeners[event]; !ok {
		e.listeners[event] = make(map[*Listener]struct{})
	}
	e.listeners[event][l] = struct{}{}

	// auto-remove when context is done.
	go func() {
		<-lctx.Done()
		e.removeListener(event, l)
	}()

	// unsubscribe function explicitly removes the listener.
	unsub := func() {
		cancel()
	}

	return l.ch, unsub
}

// Emit sends a payload to all listeners of the event.
// It uses non-blocking send: if a listener's buffer is full, it drops the message
// for that listener to avoid blocking. This mirrors Node EventEmitter's best-effort
// semantics, but with backpressure via buffer.
//
// If you need strict delivery, consider increasing buffer or implementing an
// ack/retry mechanism.
func (e *EventEmitter) Emit(event string, payload string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return
	}

	listeners := e.listeners[event]
	for l := range listeners {
		select {
		case l.ch <- payload:
			// delivered
		default:
			// drop to avoid blocking; alternatively, spawn a goroutine to block-send
			// but that risks goroutine leaks under pressure.
		}
	}
}

// Off removes all listeners for an event.
func (e *EventEmitter) Off(event string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return
	}

	if ls, ok := e.listeners[event]; ok {
		for l := range ls {
			// Cancel each listener to close its channel and stop goroutine.
			l.cancel()
			delete(ls, l)
			// Channel is closed by the goroutine that waits on context.
		}
		delete(e.listeners, event)
	}
}

// Close shuts down the emitter and removes all listeners across all events.
func (e *EventEmitter) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return
	}
	e.closed = true

	for event, ls := range e.listeners {
		for l := range ls {
			l.cancel()
			delete(ls, l)
		}
		delete(e.listeners, event)
	}
}

// removeListener removes a specific listener and closes its channel.
func (e *EventEmitter) removeListener(event string, l *Listener) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ls, ok := e.listeners[event]; ok {
		if _, exists := ls[l]; exists {
			delete(ls, l)
		}
		if len(ls) == 0 {
			delete(e.listeners, event)
		}
	}

	// Safe close: ensure channel is closed exactly once.
	select {
	case <-l.ch:
		// We probed a receive; not reliable for "closed" detection. Instead, use recover on close.
	default:
	}
	// Use defer/recover to avoid panic on double-close if some race occurs.
	defer func() {
		_ = recover()
	}()
	close(l.ch)
}
