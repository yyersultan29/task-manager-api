-- Deleted tasks cannot be restored.
ALTER TABLE tasks
    ALTER COLUMN user_id DROP NOT NULL;