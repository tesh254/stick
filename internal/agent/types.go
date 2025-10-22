package agent

type Task struct {
	Description string `json:"description"`
	IsDone      bool   `json:"is_done"`
}

type Session struct {
	Tasks []Task
}
