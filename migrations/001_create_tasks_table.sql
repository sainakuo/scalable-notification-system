CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    type TEXT NOT NULL,
    payload TEXT NOT NULL,
    status TEXT NOT NULL
        CHECK (status IN ('pending', 'processing', 'done', 'failed')),
    retry_count INT DEFAULT 0
        CHECK (retry_count >= 0 AND retry_count <= 3),
    created_at TIMESTAMP DEFAULT NOW()
);