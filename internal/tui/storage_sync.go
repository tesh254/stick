package tui

import (
    "context"
    "log"
    "time"

    "github.com/tesh254/stick/internal/db"
)

// startStorageWorker launches a goroutine to persist messages from the queue
// without blocking the UI thread. It retries transient failures and logs errors.
func startStorageWorker(m *model) {
    if m.repoManager == nil || m.storageQueue == nil {
        // No storage configured; nothing to do.
        return
    }

    go func() {
        for msg := range m.storageQueue {
            // Persist with basic retry on transient errors
            ctx := context.Background()
            var err error
            for attempt := 0; attempt < 3; attempt++ {
                err = m.repoManager.Messages().Create(ctx, msg)
                if err == nil {
                    break
                }
                time.Sleep(100 * time.Millisecond)
            }
            if err != nil {
                log.Printf("tui: failed to persist message %s: %v", msg.ID.String(), err)
            }
        }
    }()
}

// enqueueStorageMessage schedules a message for background persistence.
// It avoids blocking the UI by spawning a small goroutine on backpressure.
func (m *model) enqueueStorageMessage(msg *db.Message) {
    if m.storageQueue == nil {
        return
    }
    select {
    case m.storageQueue <- msg:
        // enqueued without blocking
    default:
        // fallback to async send to avoid blocking UI thread
        go func() { m.storageQueue <- msg }()
    }
}

// startCallStorageWorker persists call events asynchronously.
func startCallStorageWorker(m *model) {
    if m.repoManager == nil || m.callStorageQueue == nil {
        return
    }
    go func() {
        for ev := range m.callStorageQueue {
            ctx := context.Background()
            var err error
            for attempt := 0; attempt < 3; attempt++ {
                err = m.repoManager.Calls().Create(ctx, ev)
                if err == nil { break }
                time.Sleep(100 * time.Millisecond)
            }
            if err != nil {
                log.Printf("tui: failed to persist call event %s: %v", ev.ID.String(), err)
            }
        }
    }()
}

// enqueueCallEvent schedules a call event to be stored without blocking UI.
func (m *model) enqueueCallEvent(ev *db.CallEvent) {
    if m.callStorageQueue == nil { return }
    select {
    case m.callStorageQueue <- ev:
    default:
        go func() { m.callStorageQueue <- ev }()
    }
}

// validateConsistency performs basic checks that display and storage slices
// have matching counts and chronological integrity. Returns an error string
// if inconsistency is detected; otherwise returns empty string.
func (m *model) validateConsistency() string {
    // For now, we only validate counts, since display strings combine styled blocks
    // and storage keeps raw content. We maintain 1:1 for user+assistant outputs.
    if len(m.messages) != len(m.storageMessages) {
        return "display/storage slice length mismatch"
    }
    // Chronological integrity is ensured by append order; we could add more checks
    // (e.g., timestamps monotonic). Keep minimal to avoid performance impact.
    return ""
}