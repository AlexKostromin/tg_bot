-- +goose Up
CREATE TABLE admin_users (
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(50)  NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMP    NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS admin_users;
