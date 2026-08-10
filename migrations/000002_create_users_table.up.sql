CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT not null unique,
    password_hash TEXT not null,
    created_at TIMESTAMPTZ not null default now()
);
