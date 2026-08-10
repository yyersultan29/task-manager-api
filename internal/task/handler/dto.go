package handler

import (
	"task-manager-api/internal/task"
)

type errorResponse struct {
	Error string `json:"error"`
}

type createTaskRequest struct {
	Title *string `json:"title"`
}

type taskResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func toTaskResponse(t task.Task) taskResponse {
	return taskResponse{
		ID:    t.ID,
		Title: t.Title,
		Done:  t.Done,
	}
}

func toTaskResponses(tasks []task.Task) []taskResponse {
	response := make([]taskResponse, 0, len(tasks))

	for _, t := range tasks {
		response = append(response, toTaskResponse(t))
	}

	return response
}
