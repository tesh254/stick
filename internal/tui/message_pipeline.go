package tui

import (
    "encoding/json"
    "fmt"
    "strings"
    "time"

    "github.com/charmbracelet/lipgloss"
    "github.com/tesh254/stick/internal/db"
    "github.com/tesh254/stick/internal/functions"
    "github.com/dombox/uuidv7"
)

// PipelineResult captures both display and storage-oriented outputs
// derived from a parsed function call.
type PipelineResult struct {
    Display   string
    RawResult string
    Err       error
    IsError   bool
    FunctionName string
    Arguments    []string
    StartedAt    time.Time
    CompletedAt  time.Time
}

// processFunctionCallPipeline parses and executes a function call producing both
// styled display output and raw result for storage. Returns ok=false if input
// is not a function call.
func (m *model) processFunctionCallPipeline(input string) (res *PipelineResult, ok bool) {
    s := strings.TrimSpace(input)
    if s == "" {
        return nil, false
    }

    openIdx := strings.Index(s, "(")
    if openIdx == -1 {
        return nil, false
    }
    if openIdx > 0 {
        prev := s[openIdx-1]
        if prev == ' ' || prev == '\t' || prev == '\n' || prev == '\r' {
            return nil, false
        }
    }

    p := functions.Parser{}
    parsed := p.ParseDetailed(s)
    if parsed.Error != nil {
        closeIdx := strings.LastIndex(s, ")")
        if openIdx != -1 && (closeIdx == -1 || closeIdx < openIdx) {
            return &PipelineResult{Display: "", RawResult: "", Err: fmt.Errorf("syntax error: missing closing ')' for '(' at position %d", openIdx), IsError: true}, true
        }
        return &PipelineResult{Display: "", RawResult: "", Err: parsed.Error, IsError: true}, true
    }

    if parsed.HasFunction && parsed.FunctionName != "" {
        funcs := m.functionRegistry.GetFunctions()
        if _, exists := funcs[parsed.FunctionName]; !exists {
            fr := NewFunctionRenderer()
            styled := fr.renderFunctionOrToolResult(parsed.FunctionName, strings.Join(parsed.Arguments, ", "), "", true)
            return &PipelineResult{Display: styled, RawResult: "", Err: fmt.Errorf("unknown function: %s", parsed.FunctionName), IsError: true, FunctionName: parsed.FunctionName, Arguments: parsed.Arguments, StartedAt: time.Now(), CompletedAt: time.Now()}, true
        }

        // Execute raw result for storage
        started := time.Now()
        raw, callErr := m.functionRegistry.Call(parsed.FunctionName, parsed.Arguments)
        fr := NewFunctionRenderer()
        nameBlock := fr.RenderFunctionName(parsed.FunctionName, parsed.Arguments)
        resultBlock, renderErr := fr.ExecuteAndRender(m.functionRegistry, parsed.FunctionName, parsed.Arguments, &CallOptions{CaseSensitive: true})
        combined := lipgloss.JoinVertical(lipgloss.Left, nameBlock, resultBlock)
        completed := time.Now()

        if callErr != nil || renderErr != nil {
            if callErr != nil {
                return &PipelineResult{Display: combined, RawResult: callErr.Error(), Err: callErr, IsError: true, FunctionName: parsed.FunctionName, Arguments: parsed.Arguments, StartedAt: started, CompletedAt: completed}, true
            }
            return &PipelineResult{Display: combined, RawResult: raw, Err: renderErr, IsError: true, FunctionName: parsed.FunctionName, Arguments: parsed.Arguments, StartedAt: started, CompletedAt: completed}, true
        }
        return &PipelineResult{Display: combined, RawResult: raw, Err: nil, IsError: false, FunctionName: parsed.FunctionName, Arguments: parsed.Arguments, StartedAt: started, CompletedAt: completed}, true
    }

    return nil, false
}

// buildStorageMessage constructs a db.Message for the current conversation.
func (m *model) buildStorageMessage(role db.Role, content string) *db.Message {
    mid, _ := uuidv7.New()
    msg := &db.Message{
        ID:           mid,
        Conversation: m.conversationID,
        Content:      content,
        Role:         role,
        CreatedAt:    time.Now(),
    }
    return msg
}

// buildCallEvent constructs a db.CallEvent for function/tool invocations.
func (m *model) buildCallEvent(parent *db.Message, funcName string, args []string, res *PipelineResult) *db.CallEvent {
    id, _ := uuidv7.New()
    // Encode parameters as JSON array of strings for future-proofing
    paramsBytes, _ := json.Marshal(args)
    status := db.CallStatusSuccess
    var errStr string
    if res != nil && res.IsError {
        status = db.CallStatusError
        if res.Err != nil { errStr = res.Err.Error() }
    }
    dur := res.CompletedAt.Sub(res.StartedAt)
    ev := &db.CallEvent{
        ID:              id,
        Conversation:    m.conversationID,
        ParentMessage:   &parent.ID,
        Type:            db.CallTypeFunction,
        Name:            funcName,
        ParamsJSON:      string(paramsBytes),
        Status:          status,
        ResultRaw:       res.RawResult,
        Error:           errStr,
        StartedAt:       res.StartedAt,
        CompletedAt:     res.CompletedAt,
        DurationMS:      dur.Milliseconds(),
    }
    return ev
}