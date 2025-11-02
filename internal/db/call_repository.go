package db

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/dombox/uuidv7"
)

// CallRepository defines operations for storing and retrieving call events.
type CallRepository interface {
    Create(ctx context.Context, call *CallEvent) error
    GetByConversationID(ctx context.Context, conversationID uuidv7.UUID) ([]*CallEvent, error)
    GetByParentMessageID(ctx context.Context, parentID uuidv7.UUID) ([]*CallEvent, error)
    UpdateStatus(ctx context.Context, id uuidv7.UUID, status CallStatus, resultRaw, errText string, completedAt time.Time, durationMS int64) error
    DeleteByConversationID(ctx context.Context, conversationID uuidv7.UUID) error
}

type callRepository struct {
    db *sql.DB
    tx *sql.Tx
}

func (r *callRepository) getExecer() interface {
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
} {
    if r.tx != nil {
        return r.tx
    }
    return r.db
}

func (r *callRepository) Create(ctx context.Context, call *CallEvent) error {
    query := `
        INSERT INTO calls (
            id, conversation_id, parent_message_id, type, name, params_json,
            status, result_raw, error, started_at, completed_at, duration_ms
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    var parent string
    if call.ParentMessage != nil {
        parent = call.ParentMessage.String()
    }
    _, err := r.getExecer().ExecContext(ctx, query,
        call.ID.String(), call.Conversation.String(), parent, string(call.Type), call.Name,
        call.ParamsJSON, string(call.Status), call.ResultRaw, call.Error, call.StartedAt, call.CompletedAt, call.DurationMS,
    )
    if err != nil {
        return fmt.Errorf("failed to create call event: %w", err)
    }
    return nil
}

func (r *callRepository) GetByConversationID(ctx context.Context, conversationID uuidv7.UUID) ([]*CallEvent, error) {
    query := `SELECT id, conversation_id, parent_message_id, type, name, params_json, status, result_raw, error, started_at, completed_at, duration_ms FROM calls WHERE conversation_id = ? ORDER BY started_at ASC`
    rows, err := r.getExecer().QueryContext(ctx, query, conversationID.String())
    if err != nil {
        return nil, fmt.Errorf("failed to get calls: %w", err)
    }
    defer rows.Close()

    var calls []*CallEvent
    for rows.Next() {
        var idStr, convStr, parentStr, typeStr, name, paramsJSON, statusStr, resultRaw, errText string
        var started, completed time.Time
        var dur int64
        if err := rows.Scan(&idStr, &convStr, &parentStr, &typeStr, &name, &paramsJSON, &statusStr, &resultRaw, &errText, &started, &completed, &dur); err != nil {
            return nil, fmt.Errorf("failed to scan call event: %w", err)
        }
        id, err := uuidv7.Parse(idStr)
        if err != nil { return nil, err }
        conv, err := uuidv7.Parse(convStr)
        if err != nil { return nil, err }
        var parentPtr *uuidv7.UUID
        if parentStr != "" {
            pid, err := uuidv7.Parse(parentStr)
            if err != nil { return nil, err }
            parentPtr = &pid
        }
        call := &CallEvent{
            ID: id,
            Conversation: conv,
            ParentMessage: parentPtr,
            Type: CallType(typeStr),
            Name: name,
            ParamsJSON: paramsJSON,
            Status: CallStatus(statusStr),
            ResultRaw: resultRaw,
            Error: errText,
            StartedAt: started,
            CompletedAt: completed,
            DurationMS: dur,
        }
        calls = append(calls, call)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating calls: %w", err)
    }
    return calls, nil
}

func (r *callRepository) GetByParentMessageID(ctx context.Context, parentID uuidv7.UUID) ([]*CallEvent, error) {
    query := `SELECT id, conversation_id, parent_message_id, type, name, params_json, status, result_raw, error, started_at, completed_at, duration_ms FROM calls WHERE parent_message_id = ? ORDER BY started_at ASC`
    rows, err := r.getExecer().QueryContext(ctx, query, parentID.String())
    if err != nil {
        return nil, fmt.Errorf("failed to get calls: %w", err)
    }
    defer rows.Close()
    var calls []*CallEvent
    for rows.Next() {
        var idStr, convStr, parentStr, typeStr, name, paramsJSON, statusStr, resultRaw, errText string
        var started, completed time.Time
        var dur int64
        if err := rows.Scan(&idStr, &convStr, &parentStr, &typeStr, &name, &paramsJSON, &statusStr, &resultRaw, &errText, &started, &completed, &dur); err != nil {
            return nil, fmt.Errorf("failed to scan call event: %w", err)
        }
        id, err := uuidv7.Parse(idStr)
        if err != nil { return nil, err }
        conv, err := uuidv7.Parse(convStr)
        if err != nil { return nil, err }
        var parentPtr *uuidv7.UUID
        if parentStr != "" {
            pid, err := uuidv7.Parse(parentStr)
            if err != nil { return nil, err }
            parentPtr = &pid
        }
        call := &CallEvent{
            ID: id,
            Conversation: conv,
            ParentMessage: parentPtr,
            Type: CallType(typeStr),
            Name: name,
            ParamsJSON: paramsJSON,
            Status: CallStatus(statusStr),
            ResultRaw: resultRaw,
            Error: errText,
            StartedAt: started,
            CompletedAt: completed,
            DurationMS: dur,
        }
        calls = append(calls, call)
    }
    if err := rows.Err(); err != nil { return nil, fmt.Errorf("error iterating calls: %w", err) }
    return calls, nil
}

func (r *callRepository) UpdateStatus(ctx context.Context, id uuidv7.UUID, status CallStatus, resultRaw, errText string, completedAt time.Time, durationMS int64) error {
    query := `UPDATE calls SET status = ?, result_raw = ?, error = ?, completed_at = ?, duration_ms = ? WHERE id = ?`
    res, err := r.getExecer().ExecContext(ctx, query, string(status), resultRaw, errText, completedAt, durationMS, id.String())
    if err != nil { return fmt.Errorf("failed to update call status: %w", err) }
    if rows, _ := res.RowsAffected(); rows == 0 { return fmt.Errorf("call event %s not found", id.String()) }
    return nil
}

func (r *callRepository) DeleteByConversationID(ctx context.Context, conversationID uuidv7.UUID) error {
    query := `DELETE FROM calls WHERE conversation_id = ?`
    _, err := r.getExecer().ExecContext(ctx, query, conversationID.String())
    if err != nil { return fmt.Errorf("failed to delete calls: %w", err) }
    return nil
}