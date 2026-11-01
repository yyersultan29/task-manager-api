package task

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const (
	defaultBackfillBatch = 100
	mirrorTimeout        = 2 * time.Second
)

// Store is a repository that can also accept a row whose id was minted
// somewhere else. Create cannot be used for the copy: each store would
// generate its own id, and Complete/Delete would then hit different rows.
type Store interface {
	Repository
	Upsert(ctx context.Context, task Task) error
	InsertIfAbsent(ctx context.Context, task Task) (bool, error)
	ListAfterID(ctx context.Context, afterID int, limit int) ([]Task, error)
}

// MigratingRepository is the Discord cutover, scaled down.
//
// Reads and the user's write result come from primary. Every successful
// write is also mirrored to secondary. A secondary failure is logged and
// does not fail the request: the new store is still unproven.
//
// Backfill copies rows with id greater than the checkpoint. The checkpoint
// moves only after the copy succeeds, so a crash retries the same id.
// InsertIfAbsent makes that retry harmless and refuses to overwrite a newer
// dual-write with a stale snapshot.
//
// Delete records a tombstone before touching secondary. Backfill will not
// insert that id, which is what stops a row already read into a batch from
// coming back to life. Tombstones live in this process. Shadow reads are
// how a lost tombstone is noticed.
type MigratingRepository struct {
	logger *slog.Logger

	swapMu    sync.RWMutex
	primary   Store
	secondary Store

	cpMu      sync.Mutex
	lastID    int
	deleted   map[int]struct{}
	shadow    bool
	batchSize int
}

func NewMigratingRepository(primary Store, secondary Store, shadow bool, logger *slog.Logger) (*MigratingRepository, error) {
	if primary == nil || secondary == nil {
		return nil, errors.New("primary and secondary stores are required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &MigratingRepository{
		logger:    logger,
		primary:   primary,
		secondary: secondary,
		deleted:   make(map[int]struct{}),
		shadow:    shadow,
	}, nil
}

// Cutover makes the current secondary the source of truth and keeps
// mirroring to the old primary, so one more call rolls the read path back.
// Stop Backfill first. The checkpoint belongs to the old direction and is reset.
func (r *MigratingRepository) Cutover() {
	r.swapMu.Lock()
	r.primary, r.secondary = r.secondary, r.primary
	r.swapMu.Unlock()

	r.cpMu.Lock()
	r.lastID = 0
	r.deleted = make(map[int]struct{})
	r.cpMu.Unlock()
}

func (r *MigratingRepository) SetShadow(enabled bool) {
	r.cpMu.Lock()
	r.shadow = enabled
	r.cpMu.Unlock()
}

func (r *MigratingRepository) LastID() int {
	r.cpMu.Lock()
	defer r.cpMu.Unlock()
	return r.lastID
}

func (r *MigratingRepository) Create(ctx context.Context, title string) (Task, error) {
	primary, secondary := r.stores()
	task, err := primary.Create(ctx, title)
	if err != nil {
		return Task{}, err
	}
	r.mirror(ctx, "create", task.ID, func(ctx context.Context) error {
		return secondary.Upsert(ctx, task)
	})
	return task, nil
}

func (r *MigratingRepository) Complete(ctx context.Context, id int) (Task, error) {
	primary, secondary := r.stores()
	task, err := primary.Complete(ctx, id)
	if err != nil {
		return Task{}, err
	}
	// Upsert, not Complete: the row may not have been copied yet.
	// Writing the primary's result creates it in the completed state.
	r.mirror(ctx, "complete", task.ID, func(ctx context.Context) error {
		return secondary.Upsert(ctx, task)
	})
	return task, nil
}

func (r *MigratingRepository) Delete(ctx context.Context, id int) error {
	primary, secondary := r.stores()
	if err := primary.Delete(ctx, id); err != nil {
		return err
	}
	// Tombstone first. If we crash before the secondary delete, Backfill
	// still refuses to copy this id.
	r.markDeleted(id)
	r.mirror(ctx, "delete", id, func(ctx context.Context) error {
		err := secondary.Delete(ctx, id)
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	})
	return nil
}

func (r *MigratingRepository) FindByID(ctx context.Context, id int) (Task, error) {
	primary, secondary := r.stores()
	task, err := primary.FindByID(ctx, id)
	if err != nil {
		return Task{}, err
	}
	if r.shadowEnabled() {
		r.shadowOne(ctx, secondary, task)
	}
	return task, nil
}

func (r *MigratingRepository) List(ctx context.Context, options ListOptions) ([]Task, error) {
	primary, secondary := r.stores()
	items, err := primary.List(ctx, options)
	if err != nil {
		return nil, err
	}
	if r.shadowEnabled() {
		r.shadowList(ctx, secondary, options, items)
	}
	return items, nil
}

// Backfill copies primary rows after the checkpoint. An error leaves the
// checkpoint on the last row that made it, and the next call continues there.
func (r *MigratingRepository) Backfill(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		primary, secondary := r.stores()
		afterID := r.LastID()
		batch, err := primary.ListAfterID(ctx, afterID, r.batch())
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		for _, snapshot := range batch {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := r.copyOne(ctx, primary, secondary, snapshot.ID); err != nil {
				return err
			}
		}
	}
}

func (r *MigratingRepository) copyOne(ctx context.Context, primary Store, secondary Store, id int) error {
	if r.isDeleted(id) {
		if err := secondary.Delete(ctx, id); err != nil && !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("backfill delete %d: %w", id, err)
		}
		r.advance(id)
		return nil
	}

	// Read again. The batch snapshot may be older than a dual-write that
	// already landed, and InsertIfAbsent will leave that newer row alone.
	current, err := primary.FindByID(ctx, id)
	if errors.Is(err, ErrNotFound) || r.isDeleted(id) {
		if err := secondary.Delete(ctx, id); err != nil && !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("backfill delete %d: %w", id, err)
		}
		r.advance(id)
		return nil
	}
	if err != nil {
		return fmt.Errorf("backfill read %d: %w", id, err)
	}

	if _, err := secondary.InsertIfAbsent(ctx, current); err != nil {
		return fmt.Errorf("backfill write %d: %w", id, err)
	}
	// Delete can land while the insert is in flight. The tombstone is the
	// signal to remove the row we just copied.
	if r.isDeleted(id) {
		if err := secondary.Delete(ctx, id); err != nil && !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("backfill delete %d: %w", id, err)
		}
	}
	r.advance(id)
	return nil
}

func (r *MigratingRepository) mirror(ctx context.Context, op string, id int, write func(context.Context) error) {
	mirrorCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), mirrorTimeout)
	defer cancel()
	if err := write(mirrorCtx); err != nil {
		r.logger.Error("migration secondary write failed", "op", op, "id", id, "err", err)
	}
}

func (r *MigratingRepository) shadowOne(ctx context.Context, secondary Store, primaryTask Task) {
	mirrorCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), mirrorTimeout)
	defer cancel()
	other, err := secondary.FindByID(mirrorCtx, primaryTask.ID)
	if errors.Is(err, ErrNotFound) {
		r.logger.Warn("migration shadow mismatch", "op", "find", "id", primaryTask.ID, "secondary", "missing")
		return
	}
	if err != nil {
		r.logger.Error("migration shadow read failed", "op", "find", "id", primaryTask.ID, "err", err)
		return
	}
	if !sameTask(primaryTask, other) {
		r.logger.Warn("migration shadow mismatch", "op", "find", "id", primaryTask.ID, "primary", primaryTask, "secondary", other)
	}
}

func (r *MigratingRepository) shadowList(ctx context.Context, secondary Store, options ListOptions, primaryItems []Task) {
	mirrorCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), mirrorTimeout)
	defer cancel()
	other, err := secondary.List(mirrorCtx, options)
	if err != nil {
		r.logger.Error("migration shadow read failed", "op", "list", "err", err)
		return
	}
	if sameTasks(primaryItems, other) {
		return
	}
	r.logger.Warn("migration shadow mismatch", "op", "list", "primary", summarize(primaryItems), "secondary", summarize(other))
}

func (r *MigratingRepository) stores() (Store, Store) {
	r.swapMu.RLock()
	defer r.swapMu.RUnlock()
	return r.primary, r.secondary
}

func (r *MigratingRepository) shadowEnabled() bool {
	r.cpMu.Lock()
	defer r.cpMu.Unlock()
	return r.shadow
}

func (r *MigratingRepository) markDeleted(id int) {
	r.cpMu.Lock()
	r.deleted[id] = struct{}{}
	r.cpMu.Unlock()
}

func (r *MigratingRepository) isDeleted(id int) bool {
	r.cpMu.Lock()
	_, ok := r.deleted[id]
	r.cpMu.Unlock()
	return ok
}

func (r *MigratingRepository) advance(id int) {
	r.cpMu.Lock()
	if id > r.lastID {
		r.lastID = id
	}
	r.cpMu.Unlock()
}

func (r *MigratingRepository) batch() int {
	if r.batchSize > 0 {
		return r.batchSize
	}
	return defaultBackfillBatch
}

func sameTask(a Task, b Task) bool {
	return a.ID == b.ID && a.Title == b.Title && a.Done == b.Done && a.CreatedAt.Equal(b.CreatedAt)
}

func sameTasks(a []Task, b []Task) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !sameTask(a[i], b[i]) {
			return false
		}
	}
	return true
}

func summarize(tasks []Task) []int {
	ids := make([]int, len(tasks))
	for i, task := range tasks {
		ids[i] = task.ID
	}
	return ids
}

var (
	_ Repository = (*MigratingRepository)(nil)
	_ Store      = (*MemoryRepository)(nil)
	_ Store      = (*PostgresRepository)(nil)
)
