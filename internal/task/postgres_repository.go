package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) List(ctx context.Context, options ListOptions) ([]Task, error) {
	rows, err := r.db.Query(ctx, `
	    SELECT id, title, done, created_at 
		FROM tasks 
		WHERE ($1::boolean IS NULL OR done = $1)
		ORDER BY id
		LIMIT $2 OFFSET $3
	`, options.Done, options.Limit, options.Offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items []Task
	for rows.Next() {
		var item Task

		err = rows.Scan(&item.ID, &item.Title, &item.Done, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id int) (Task, error) {

	var task Task
	err := r.db.QueryRow(ctx, `SELECT id, title, done, created_at FROM tasks WHERE id = $1`, id).
		Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}

	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func (r *PostgresRepository) Create(ctx context.Context, title string) (Task, error) {
	var task Task
	err := r.db.QueryRow(ctx, `
	INSERT INTO tasks (title) 
	VALUES ($1) 
	RETURNING id, title, done, created_at`,
		title).Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt)

	if err != nil {
		return Task{}, err
	}

	return task, nil
}

// todo auth token

func (r *PostgresRepository) ListAfterID(ctx context.Context, afterID int, limit int) ([]Task, error) {
	if limit <= 0 {
		return []Task{}, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, title, done, created_at
		FROM tasks
		WHERE id > $1
		ORDER BY id
		LIMIT $2
	`, afterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Task
	for rows.Next() {
		var item Task
		if err := rows.Scan(&item.ID, &item.Title, &item.Done, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// Upsert writes a row whose id already exists on the other store.
// The sequence moves forward with the explicit id, so a later Create
// does not collide with a copied row.
func (r *PostgresRepository) Upsert(ctx context.Context, task Task) error {
	_, err := r.writeWithID(ctx, task, `
		ON CONFLICT (id) DO UPDATE
		SET title = EXCLUDED.title,
		    done = EXCLUDED.done,
		    created_at = EXCLUDED.created_at`)
	return err
}

// InsertIfAbsent leaves an existing row untouched. Backfill uses this so a
// newer dual-write is not replaced by an older snapshot.
func (r *PostgresRepository) InsertIfAbsent(ctx context.Context, task Task) (bool, error) {
	return r.writeWithID(ctx, task, `ON CONFLICT (id) DO NOTHING`)
}

// conflictAction is a fixed SQL fragment owned by this file.
func (r *PostgresRepository) writeWithID(ctx context.Context, task Task, conflictAction string) (bool, error) {
	query := fmt.Sprintf(`
		WITH written AS (
			INSERT INTO tasks (id, title, done, created_at)
			VALUES ($1, $2, $3, $4)
			%s
			RETURNING id
		)
		SELECT
			(SELECT count(*) FROM written),
			setval(
				pg_get_serial_sequence('tasks', 'id'),
				GREATEST((SELECT COALESCE(MAX(id), 1) FROM tasks), 1)
			)
	`, conflictAction)

	var inserted int64
	var sequence int64
	err := r.db.QueryRow(ctx, query, task.ID, task.Title, task.Done, task.CreatedAt).Scan(&inserted, &sequence)
	if err != nil {
		return false, err
	}
	return inserted == 1, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id int) error {

	tag, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) Complete(ctx context.Context, id int) (Task, error) {
	var task Task

	err := r.db.QueryRow(ctx, `
		UPDATE tasks 
		SET done = $2 
		WHERE id = $1 
		RETURNING id, title, done, created_at
	`, id, true).Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}

	if err != nil {
		return Task{}, err
	}

	return task, nil
}
