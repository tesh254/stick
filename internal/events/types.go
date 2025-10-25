package events

import (
	"context"
	"sync"
)

type Listener struct {
	ch     chan string
	cancel context.CancelFunc
}

type EventEmitter struct {
	mu        sync.RWMutex
	listeners map[string]map[*Listener]struct{}
	closed    bool
}
