package tui

import (
	"context"
	"encoding/json"

	"github.com/dombox/uuidv7"
	"github.com/tesh254/stick/internal/db"
	"github.com/tesh254/stick/internal/functions"
	"github.com/tesh254/stick/internal/utils"
)

// LoadConversationFromDB rebuilds the display and storage slices from persisted data.
// It also attempts to replay styled function calls using recorded metadata.
func (m *model) LoadConversationFromDB(conversationID uuidv7.UUID) error {
	if m.repoManager == nil {
		return nil
	}
	ctx := context.Background()
	_, msgs, err := m.repoManager.Conversations().GetWithMessages(ctx, conversationID)
	if err != nil {
		return err
	}

	// Reset current buffers
	m.messages = []string{}
	m.storageMessages = []*db.Message{}

	// Fetch all call events in this conversation, index by parent message ID
	calls, _ := m.repoManager.Calls().GetByConversationID(ctx, conversationID)
	callIndex := map[string][]*db.CallEvent{}
	for _, ev := range calls {
		pmid := ""
		if ev.ParentMessage != nil {
			pmid = ev.ParentMessage.String()
		}
		callIndex[pmid] = append(callIndex[pmid], ev)
	}

	// Rebuild messages chronologically using seq when available.
	lastUserWasFunc := false
	for _, msg := range msgs {
		// Populate storage slice
		m.storageMessages = append(m.storageMessages, msg)
		// Render display line according to role
		if msg.Role == db.User {
			username := utils.GetUser()
			m.messages = append(m.messages, m.senderStyle.Render("{"+username+"}: ")+msg.Content)
			// Attempt to replay any calls that originated from this message
			if evs, ok := callIndex[msg.ID.String()]; ok {
				for _, ev := range evs {
					// Decode parameters and render using stored metadata
					var args []string
					_ = json.Unmarshal([]byte(ev.ParamsJSON), &args)
					fr := NewFunctionRenderer()
					nameBlock := fr.RenderFunctionName(ev.Name, args)
					isErr := ev.Status == db.CallStatusError
					result := ev.ResultRaw
					if isErr {
						result = ev.Error
					}
					rendered := fr.renderFunctionOrToolResult(ev.Name, joinArgs(args), result, isErr)
					m.messages = append(m.messages, nameBlock)
					m.messages = append(m.messages, rendered)
				}
				lastUserWasFunc = true
			} else {
				// Fallback: parse and render if this user message is a function call
				p := functions.Parser{}
				parsed := p.ParseDetailed(msg.Content)
				if parsed.Error == nil && parsed.HasFunction && parsed.FunctionName != "" {
					funcs := m.functionRegistry.GetFunctions()
					if _, exists := funcs[parsed.FunctionName]; exists {
						fr := NewFunctionRenderer()
						nameBlock := fr.RenderFunctionName(parsed.FunctionName, parsed.Arguments)
						rendered, err := fr.ExecuteAndRender(m.functionRegistry, parsed.FunctionName, parsed.Arguments, &CallOptions{CaseSensitive: true})
						m.messages = append(m.messages, nameBlock)
						if err != nil {
							m.messages = append(m.messages, fr.renderFunctionOrToolResult(parsed.FunctionName, joinArgs(parsed.Arguments), err.Error(), true))
						} else {
							m.messages = append(m.messages, rendered)
						}
						lastUserWasFunc = true
					} else {
						lastUserWasFunc = false
					}
				} else {
					lastUserWasFunc = false
				}
			}
		} else {
			// Assistant messages: append only if previous user message was not a function call
			if !lastUserWasFunc {
				m.messages = append(m.messages, msg.Content)
			}
			lastUserWasFunc = false
		}
	}

	// Update viewport after reconstruction
	m.viewport.SetContent(m.wrapStyle.Render(formatMessages(m.messages)))
	m.viewport.GotoBottom()
	return nil
}

// ReplayCallFromEvent executes and renders a function call from stored metadata.
func (m *model) ReplayCallFromEvent(ev *db.CallEvent) (string, string, error) {
	var args []string
	_ = json.Unmarshal([]byte(ev.ParamsJSON), &args)
	fr := NewFunctionRenderer()
	name := fr.RenderFunctionName(ev.Name, args)
	res, err := fr.ExecuteAndRender(m.functionRegistry, ev.Name, args, &CallOptions{CaseSensitive: true})
	if err != nil {
		return name, fr.renderFunctionOrToolResult(ev.Name, joinArgs(args), err.Error(), true), err
	}
	return name, res, nil
}

func joinArgs(args []string) string {
	s := ""
	for i, a := range args {
		if i > 0 {
			s += ", "
		}
		s += a
	}
	return s
}
