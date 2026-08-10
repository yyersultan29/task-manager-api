package task

import (
	"context"
	"sync"
)

type MemoryRepository struct {
	tasks []Task
	mu    sync.RWMutex
}

func NewMemoryRepository(tasks []Task) *MemoryRepository {

	return &MemoryRepository{
		tasks: tasks,
	}
}

func (r *MemoryRepository) List(ctx context.Context) ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tasks := make([]Task, len(r.tasks))

	copy(tasks, r.tasks)
	return tasks, nil
}

func (r *MemoryRepository) Create(ctx context.Context, title string) (Task, error) {
	var err error

	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks, err = Create(r.tasks, title)
	if err != nil {
		return Task{}, err
	}

	return r.tasks[len(r.tasks)-1], nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id int) error {
	var err error

	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks, err = Delete(r.tasks, id)
	return err
}

func (r *MemoryRepository) FindByID(ctx context.Context, id int) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return FindByID(r.tasks, id)
}

func (r *MemoryRepository) Complete(ctx context.Context, id int) (Task, error) {

	r.mu.Lock()
	defer r.mu.Unlock()
	return Complete(r.tasks, id)
}
