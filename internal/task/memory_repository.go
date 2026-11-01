package task

import (
	"context"
	"sort"
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

func (r *MemoryRepository) List(ctx context.Context, options ListOptions) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	tasks := append([]Task(nil), r.tasks...)
	r.mu.RUnlock()

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })

	filtered := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		if options.Done != nil && task.Done != *options.Done {
			continue
		}
		filtered = append(filtered, task)
	}
	if options.Offset >= len(filtered) || options.Limit <= 0 {
		return []Task{}, nil
	}
	filtered = filtered[options.Offset:]
	if options.Limit < len(filtered) {
		filtered = filtered[:options.Limit]
	}
	return filtered, nil
}

func (r *MemoryRepository) ListAfterID(ctx context.Context, afterID int, limit int) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		return []Task{}, nil
	}
	r.mu.RLock()
	tasks := append([]Task(nil), r.tasks...)
	r.mu.RUnlock()

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	items := make([]Task, 0, limit)
	for _, task := range tasks {
		if task.ID <= afterID {
			continue
		}
		items = append(items, task)
		if len(items) == limit {
			break
		}
	}
	return items, nil
}

func (r *MemoryRepository) Upsert(ctx context.Context, task Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.tasks {
		if r.tasks[i].ID == task.ID {
			r.tasks[i] = task
			return nil
		}
	}
	r.tasks = append(r.tasks, task)
	sort.Slice(r.tasks, func(i, j int) bool { return r.tasks[i].ID < r.tasks[j].ID })
	return nil
}

func (r *MemoryRepository) InsertIfAbsent(ctx context.Context, task Task) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.tasks {
		if r.tasks[i].ID == task.ID {
			return false, nil
		}
	}
	r.tasks = append(r.tasks, task)
	sort.Slice(r.tasks, func(i, j int) bool { return r.tasks[i].ID < r.tasks[j].ID })
	return true, nil
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
