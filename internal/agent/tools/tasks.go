package tools

import "github.com/tesh254/stick/internal/agent"

var currentSession *agent.Session

// CreateTasks creates tasks in memory for the agent to manage
func CreateTasks(tasks []string) ([]agent.Task, error) {
	if currentSession == nil {
		currentSession = &agent.Session{}
	}

	// Convert string descriptions to Task objects
	newTasks := make([]agent.Task, len(tasks))
	for i, desc := range tasks {
		newTasks[i] = agent.Task{
			Description: desc,
			IsDone:      false,
		}
	}

	// Append new tasks to the current session
	currentSession.Tasks = append(currentSession.Tasks, newTasks...)

	return newTasks, nil
}

// UpdateTasks update a single task based on index in memory
func UpdateTasks(index int, description string) error {
	if currentSession == nil || index < 0 || index >= len(currentSession.Tasks) {
		return nil
	}

	currentSession.Tasks[index].Description = description
	return nil
}

// GetCurrentSessionTasks get's current tasks used in the session
func GetCurrentSessionTasks() []agent.Task {
	if currentSession == nil {
		return []agent.Task{}
	}

	return currentSession.Tasks
}

// UpdateTaskStatus update's the current state of a task
func UpdateTaskStatus(index int, isDone bool) error {
	if currentSession == nil || index < 0 || index >= len(currentSession.Tasks) {
		return nil
	}

	currentSession.Tasks[index].IsDone = isDone
	return nil
}

// MarkAllTasksAsComplete marks all tasks as complete in the session
func MarkAllTasksAsComplete() error {
	if currentSession == nil {
		return nil
	}

	for i := range currentSession.Tasks {
		currentSession.Tasks[i].IsDone = true
	}
	return nil
}

// SetSession sets the current session for task management
func SetSession(session *agent.Session) {
	currentSession = session
}
