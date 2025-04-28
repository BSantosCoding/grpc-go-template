CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL, -- Ensure emails are unique
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL -- Added NOT NULL for consistency
);