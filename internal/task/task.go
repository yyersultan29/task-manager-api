package task

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Task struct {
	ID        int
	Title     string
	Done      bool
	CreatedAt time.Time
}

var (
	ErrNotFound      = errors.New("task not found")
	ErrTitleRequired = errors.New("title is required")
	ErrInvalidID     = errors.New("invalid task id")
	ErrTitleTooLong  = errors.New("title length too long")
)

const maxTitleLength = 120

func (task Task) Print() {
	fmt.Printf("%d. %s, done: %t\n", task.ID, task.Title, task.Done)
}
func (task *Task) MarkDone() {
	task.Done = true
}

func CountCompleted(tasks []Task) int {
	count := 0

	for _, task := range tasks {
		if task.Done {
			count++
		}
	}
	return count
}

func Create(tasks []Task, title string) ([]Task, error) {

	title = strings.TrimSpace(title)

	if title == "" {
		return tasks, ErrTitleRequired
	}

	nextID := 1

	for _, task := range tasks {
		if task.ID >= nextID {
			nextID = task.ID + 1
		}
	}

	newTask := Task{ID: nextID, Title: title, Done: false}

	newTasks := append(tasks, newTask)

	return newTasks, nil
}

func FindByID(tasks []Task, id int) (Task, error) {

	for _, task := range tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return Task{}, ErrNotFound
}

func MarkDone(tasks []Task, id int) error {

	for i, task := range tasks {
		if task.ID == id {
			tasks[i].MarkDone()
			return nil
		}
	}

	return ErrNotFound
}

func Delete(tasks []Task, id int) ([]Task, error) {
	for i, task := range tasks {
		if id == task.ID {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return tasks, nil
		}
	}
	return tasks, ErrNotFound
}

func Complete(tasks []Task, id int) (Task, error) {
	for i, task := range tasks {
		if id == task.ID {
			tasks[i].Done = true
			return tasks[i], nil
		}
	}
	return Task{}, ErrNotFound

}
