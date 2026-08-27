CREATE INDEX IF NOT EXISTS idx_tasks_status
ON tasks(status);

CREATE INDEX IF NOT EXISTS idx_tasks_created_at
ON tasks(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id
ON tasks(user_id);