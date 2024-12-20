CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    jwt_id TEXT NOT NULL,
    user_id UUID NOT NULL,
    refresh_token TEXT NOT NULL,
    ip TEXT NOT NULL
);