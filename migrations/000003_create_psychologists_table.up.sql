CREATE TABLE IF NOT EXISTS psychologists (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    photo_url   VARCHAR(500),
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);