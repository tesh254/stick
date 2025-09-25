package agent

// Task represents a single task in a task list.
type Task struct {
	Description string `json:"description"`
	Done        bool   `json:"done"`
}
