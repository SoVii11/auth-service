CREATE TABLE IF NOT EXISTS psychologist_schedules (
    id               BIGSERIAL PRIMARY KEY,
    psychologist_id  BIGINT    NOT NULL REFERENCES psychologists(id) ON DELETE CASCADE,
    date_from        DATE      NOT NULL,
    date_to          DATE      NOT NULL,
    reason           VARCHAR(255) NOT NULL DEFAULT 'vacation',
    created_at       TIMESTAMP NOT NULL DEFAULT NOW()
);