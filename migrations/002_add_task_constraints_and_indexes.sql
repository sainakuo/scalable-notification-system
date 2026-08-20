CREATE INDEX IF NOT EXISTS idx_tasks_status
ON tasks(status);

CREATE INDEX IF NOT EXISTS idx_tasks_created_at
ON tasks(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id
ON tasks(user_id);

ALTER TABLE tasks
ADD CONSTRAINT chk_tasks_status
CHECK (status IN ('pending', 'processing', 'done', 'failed'));

ALTER TABLE tasks
ADD CONSTRAINT chk_tasks_retry_count
CHECK (retry_count >= 0 AND retry_count <= 3);