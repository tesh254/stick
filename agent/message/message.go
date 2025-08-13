package message

type AgentToolCallMsg struct {
	Name string
	Args string
}

type AgentToolResultMsg struct {
	Name    string
	Result  string
	IsError bool
}
