package provider

import (
	"net/http"
)

// Tool and function calling types
type FunctionDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  *Parameters `json:"parameters"`
}

type Tools struct {
}

type Tool struct {
	Type     string                 `json:"type"`
	Function *FunctionDefinition    `json:"function"`
	Example  map[string]interface{} `json:"example,omitempty"`
}

type ToolMetadata struct {
	// Label is a concise, Go-style CamelCase name (e.g., RunCommand).
	Label string
	// Info describes the purpose and context where the tool is useful.
	Info string
	// Category groups tools logically (e.g., Shell, Filesystem, Web, Project, Functions).
	Category string
	// Params specify input arguments, their types, and whether they are required.
	Params []ToolParam
	// Returns describes the expected output type and its meaning.
	Returns ToolReturn
	// Actions explains what the tool does in concrete terms.
	Actions []string
	// Examples provide quick usage hints for the agent.
	Examples []string
}

// ToolParam defines a single input argument that a tool accepts.
type ToolParam struct {
	// Name is the argument name (snake_case recommended for external call schema).
	Name string
	// Type is a logical type hint: string, number, boolean, array, object, path, command, json.
	Type string
	// Description explains what the argument represents and any constraints.
	Description string
	// Required indicates whether the argument must be provided.
	Required bool
}

// ToolReturn defines the expected return value for a tool.
type ToolReturn struct {
	// Type is a logical type hint: string, number, boolean, array, object.
	Type string
	// Description explains the content of the returned value.
	Description string
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Prop struct {
	Type        string   `json:"type"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Items       *Item  `json:"items,omitempty"`
}

type Item struct {
	Type       string          `json:"type"`
	Properties map[string]Prop `json:"properties"`
	Required   []string        `json:"required,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Message types with tool support
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// Request and response types
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
	Stream   bool      `json:"stream,omitempty"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}

// ChatCompletionChunk represents a single streamed delta chunk from OpenRouter.
// It follows the general OpenAI-like schema with choices containing delta updates.
// Only fields relevant for text and tool calls are included for efficiency.
type ChatCompletionChunk struct {
	Choices []struct {
		Delta struct {
			Content   string          `json:"content"`
			ToolCalls []ToolCallDelta `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// ToolCallDelta captures incremental tool call info during streaming.
// Some providers send partial function name/arguments over several chunks.
type ToolCallDelta struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Function struct {
		Name      string `json:"name,omitempty"`
		Arguments string `json:"arguments,omitempty"`
	} `json:"function"`
}

// Client implementation
type Client struct {
	APIKey     string
	HTTPClient *http.Client
	// BaseURL allows overriding the default OpenRouter endpoint, useful for tests.
	BaseURL string
	// OnStreamText, if set, is called with text deltas as they arrive.
	OnStreamText func(chunk string)
	// OnStreamToolCallDelta, if set, is called for tool call delta messages.
	OnStreamToolCallDelta func(call ToolCall)
}
