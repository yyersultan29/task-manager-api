ALTER TABLE tasks
    ADD COLUMN user_id BIGINT NOT NULL,
    ADD CONSTRAINT tasks_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users(id);

CREATE INDEX idx_tasks_user_id_id ON tasks (user_id, id);
