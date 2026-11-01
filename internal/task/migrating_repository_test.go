package task

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
)

func TestBackfillResumesAfterSecondaryFailure(t *testing.T) {
	ctx := context.Background()
	primary := NewMemoryRepository(nil)
	for _, title := range []string{"one", "two", "three"} {
		if _, err := primary.Create(ctx, title); err != nil {
			t.Fatal(err)
		}
	}
	secondary := &flakyStore{MemoryRepository: NewMemoryRepository(nil), failOnCall: 2}

	repo, err := NewMigratingRepository(primary, secondary, false, discardLogger())
	if err != nil {
		t.Fatal(err)
	}
	repo.batchSize = 1

	err = repo.Backfill(ctx)
	if err == nil {
		t.Fatal("expected backfill to stop on the secondary failure")
	}
	if got := repo.LastID(); got != 1 {
		t.Fatalf("checkpoint = %d, want 1", got)
	}

	secondary.failOnCall = 0
	if err := repo.Backfill(ctx); err != nil {
		t.Fatal(err)
	}
	if got := repo.LastID(); got != 3 {
		t.Fatalf("checkpoint = %d, want 3", got)
	}
	items, err := secondary.List(ctx, ListOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 || items[0].ID != 1 || items[2].Title != "three" {
		t.Fatalf("secondary = %+v", items)
	}
}

func TestDeleteDuringBackfillDoesNotResurrect(t *testing.T) {
	ctx := context.Background()
	primary := NewMemoryRepository(nil)
	for _, title := range []string{"keep", "drop", "keep-too"} {
		if _, err := primary.Create(ctx, title); err != nil {
			t.Fatal(err)
		}
	}
	secondary := NewMemoryRepository(nil)
	repo, err := NewMigratingRepository(primary, secondary, false, discardLogger())
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete(ctx, 2); err != nil {
		t.Fatal(err)
	}
	if err := repo.Backfill(ctx); err != nil {
		t.Fatal(err)
	}

	_, err = secondary.FindByID(ctx, 2)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted row came back: %v", err)
	}
	items, err := secondary.List(ctx, ListOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].ID != 1 || items[1].ID != 3 {
		t.Fatalf("secondary = %+v", items)
	}
}

func TestSecondaryWriteFailureStaysOffTheUserPath(t *testing.T) {
	ctx := context.Background()
	primary := NewMemoryRepository(nil)
	secondary := &flakyStore{MemoryRepository: NewMemoryRepository(nil), failUpsert: true}
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	repo, err := NewMigratingRepository(primary, secondary, false, logger)
	if err != nil {
		t.Fatal(err)
	}

	task, err := repo.Create(ctx, "ship it")
	if err != nil {
		t.Fatal(err)
	}
	if task.ID == 0 {
		t.Fatal("primary create returned an empty task")
	}
	if _, err := secondary.FindByID(ctx, task.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("secondary stored the row despite the failure: %v", err)
	}
	if logs.Len() == 0 {
		t.Fatal("secondary failure was not logged")
	}
}

func TestShadowReadReportsMismatchAndReturnsPrimary(t *testing.T) {
	ctx := context.Background()
	primary := NewMemoryRepository(nil)
	created, err := primary.Create(ctx, "original")
	if err != nil {
		t.Fatal(err)
	}
	secondary := NewMemoryRepository(nil)
	copied := created
	copied.Title = "drift"
	if err := secondary.Upsert(ctx, copied); err != nil {
		t.Fatal(err)
	}

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	repo, err := NewMigratingRepository(primary, secondary, true, logger)
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "original" {
		t.Fatalf("user saw %q, want the primary title", got.Title)
	}
	if logs.Len() == 0 {
		t.Fatal("shadow mismatch was not logged")
	}
}

func TestCutoverSwapsTheReadPath(t *testing.T) {
	ctx := context.Background()
	oldPrimary := NewMemoryRepository(nil)
	newPrimary := NewMemoryRepository(nil)
	repo, err := NewMigratingRepository(oldPrimary, newPrimary, false, discardLogger())
	if err != nil {
		t.Fatal(err)
	}

	first, err := repo.Create(ctx, "before")
	if err != nil {
		t.Fatal(err)
	}
	repo.Cutover()

	got, err := repo.FindByID(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "before" {
		t.Fatalf("title = %q", got.Title)
	}

	second, err := repo.Create(ctx, "after")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newPrimary.FindByID(ctx, second.ID); err != nil {
		t.Fatalf("new primary missed the write: %v", err)
	}
	if _, err := oldPrimary.FindByID(ctx, second.ID); err != nil {
		t.Fatalf("old primary was not mirrored after cutover: %v", err)
	}
}

func TestBackfillDoesNotOverwriteNewerDualWrite(t *testing.T) {
	ctx := context.Background()
	primary := NewMemoryRepository(nil)
	created, err := primary.Create(ctx, "open")
	if err != nil {
		t.Fatal(err)
	}
	secondary := NewMemoryRepository(nil)
	completed := created
	completed.Done = true
	if err := secondary.Upsert(ctx, completed); err != nil {
		t.Fatal(err)
	}

	repo, err := NewMigratingRepository(primary, secondary, false, discardLogger())
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Backfill(ctx); err != nil {
		t.Fatal(err)
	}

	got, err := secondary.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Done {
		t.Fatal("backfill replaced the newer completed row")
	}
}

type flakyStore struct {
	*MemoryRepository
	failUpsert bool
	failOnCall int
	calls      int
}

func (f *flakyStore) Upsert(ctx context.Context, task Task) error {
	if f.failUpsert {
		return errors.New("disk full")
	}
	return f.MemoryRepository.Upsert(ctx, task)
}

func (f *flakyStore) InsertIfAbsent(ctx context.Context, task Task) (bool, error) {
	f.calls++
	if f.failOnCall > 0 && f.calls == f.failOnCall {
		return false, errors.New("disk full")
	}
	return f.MemoryRepository.InsertIfAbsent(ctx, task)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
}
