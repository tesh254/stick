package agent

import (
	"github.com/tesh254/stick/internal/provider"
)

type Task struct {
	Description string `json:"description"`
	IsDone      bool   `json:"is_done"`
}

type Session struct {
	Tasks   []Task
	Client  *provider.Client
	History []provider.Message
	Mode    string
}
