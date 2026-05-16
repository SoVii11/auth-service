CREATE TABLE IF NOT EXISTS appointments (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    psychologist_id BIGINT     NOT NULL REFERENCES psychologists(id) ON DELETE CASCADE,
    status         VARCHAR(50) NOT NULL DEFAULT 'pending',
    comment        TEXT,
    created_at     TIMESTAMP   NOT NULL DEFAULT NOW()
);