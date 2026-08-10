package task

import (
	"context"
	"errors"

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
	`, options.Done,options.Limit,options.Offset)

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
