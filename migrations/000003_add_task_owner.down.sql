DROP INDEX IF EXISTS idx_tasks_user_id_id;

ALTER TABLE tasks
    DROP CONSTRAINT tasks_user_id_fkey,
    DROP COLUMN user_id;
