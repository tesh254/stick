package agent

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/tesh254/stick/agent/message"
)

// mockAIClient is a mock implementation of the AIClient interface for testing.
type mockAIClient struct {
	callCount int
	responses []*ChatCompletionResponse
	t         *testing.T
}

func (m *mockAIClient) Create(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if m.callCount >= len(m.responses) {
		m.t.Fatalf("mockAIClient.Create called more times than expected (%d)", m.callCount)
		return nil, fmt.Errorf("unexpected call to Create")
	}
	response := m.responses[m.callCount]
	m.callCount++
	return response, nil
}

func (m *mockAIClient) Stream(req ChatCompletionRequest) (chan string, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestAgentSession_MultiStepWorkflow(t *testing.T) {
	// 1. Define the sequence of AI responses.
	responses := []*ChatCompletionResponse{
		// 1a. AI decides to print a code block.
		{
			Choices: []struct {
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				} `json:"message"`
			}{
				{Message: struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				}{
					Role: "assistant",
					ToolCalls: []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					}{{
						ID: "call_1", Type: "function", Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{Name: "print_code_block", Arguments: `{"content": "hello"}`},
					}},
				}},
			},
		},
		// 2a. After printing, AI decides to render markdown.
		{
			Choices: []struct {
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				} `json:"message"`
			}{
				{Message: struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				}{
					Role: "assistant",
					ToolCalls: []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					}{{
						ID: "call_2", Type: "function", Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{Name: "render_markdown", Arguments: `{"content": "## world"}`},
					}},
				}},
			},
		},
		// 3a. After rendering, AI decides task is complete.
		{
			Choices: []struct {
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				} `json:"message"`
			}{
				{Message: struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				}{
					Role: "assistant",
					ToolCalls: []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					}{{
						ID: "call_3", Type: "function", Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{Name: "task_complete", Arguments: `{"message": "Done."}`},
					}},
				}},
			},
		},
	}

	// 2. Set up the AgentSession.
	mockClient := &mockAIClient{responses: responses, t: t}
	responseChan := make(chan any, 10)
	session := &AgentSession{
		provider:      "mock",
		client:        mockClient,
		messages:      []Message{{Role: "system", Content: "test"}},
		responseChan:  responseChan,
		userInputChan: make(chan string),
		promptChan:    make(chan string),
	}

	// 3. Run processPrompt in a goroutine.
	go session.processPrompt("test prompt")

	// 4. Assert the sequence of events.
	expectedEvents := []any{
		message.AgentToolCallMsg{Name: "print_code_block", Args: `{"content": "hello"}`},
		message.AgentToolResultMsg{Name: "print_code_block", Result: "hello", IsError: false},
		message.AgentToolCallMsg{Name: "render_markdown", Args: `{"content": "## world"}`},
		message.AgentToolResultMsg{Name: "render_markdown", Result: "## world", IsError: false},
		message.AgentToolCallMsg{Name: "task_complete", Args: `{"message": "Done."}`},
		"Done.",
		"AGENT_DONE",
	}

	for i, expectedEvent := range expectedEvents {
		select {
		case event := <-responseChan:
			if !reflect.DeepEqual(event, expectedEvent) {
				t.Errorf("Event %d: unexpected event.\nExpected: %#v (%T)\nGot:      %#v (%T)", i, expectedEvent, expectedEvent, event, event)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Timed out waiting for event %d", i)
		}
	}

	if len(responseChan) > 0 {
		t.Errorf("Unexpected extra events on channel: %#v", <-responseChan)
	}
}

func TestAgentSession_HandlesEmptyToolOutput(t *testing.T) {
	// 1. Define the sequence of AI responses.
	responses := []*ChatCompletionResponse{
		// 1a. AI decides to run a command that produces no output.
		{
			Choices: []struct {
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				} `json:"message"`
			}{
				{Message: struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				}{
					Role: "assistant",
					ToolCalls: []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					}{{
						ID: "call_1", Type: "function", Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{Name: "run_tool", Arguments: `{"command": "mkdir -p ./temp_test_dir"}`},
					}},
				}},
			},
		},
		// 2a. After the silent command, AI decides the task is complete.
		{
			Choices: []struct {
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				} `json:"message"`
			}{
				{Message: struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				}{
					Role: "assistant",
					ToolCalls: []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					}{{
						ID: "call_2", Type: "function", Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{Name: "task_complete", Arguments: `{"message": "Directory created."}`},
					}},
				}},
			},
		},
	}

	// 2. Set up the AgentSession.
	mockClient := &mockAIClient{responses: responses, t: t}
	responseChan := make(chan any, 10)
	session := &AgentSession{
		provider:      "mock",
		client:        mockClient,
		messages:      []Message{{Role: "system", Content: "test"}},
		responseChan:  responseChan,
		userInputChan: make(chan string),
		promptChan:    make(chan string),
	}

	// 3. Run processPrompt in a goroutine.
	go session.processPrompt("test prompt")

	// 4. Assert the sequence of events.
	expectedEvents := []any{
		message.AgentToolCallMsg{Name: "run_tool", Args: `{"command": "mkdir -p ./temp_test_dir"}`},
		message.AgentToolResultMsg{Name: "run_tool", Result: "", IsError: false},
		message.AgentToolCallMsg{Name: "task_complete", Args: `{"message": "Directory created."}`},
		"Directory created.",
		"AGENT_DONE",
	}

	for i, expectedEvent := range expectedEvents {
		select {
		case event := <-responseChan:
			if !reflect.DeepEqual(event, expectedEvent) {
				t.Errorf("Event %d: unexpected event.\nExpected: %#v (%T)\nGot:      %#v (%T)", i, expectedEvent, expectedEvent, event, event)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Timed out waiting for event %d", i)
		}
	}

	if len(responseChan) > 0 {
		t.Errorf("Unexpected extra events on channel: %#v", <-responseChan)
	}
}
