package message

type AgentToolCallMsg struct {
	ToolID string
	Name   string
	Args   string
}

type AgentToolResultMsg struct {
	ToolID  string
	Name    string
	Result  string
	IsError bool
}
