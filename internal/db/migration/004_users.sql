-- +goose Up

    CREATE TABLE users (
        id SERIAl PRIMARY KEY,
        tg_chat_id BIGINT NOT NULL UNIQUE,
        tg_username VARCHAR(255),
        full_name VARCHAR(255) NOT NULL,
        phone VARCHAR(20),
        class_number SMALLINT,
        class_group_id INTEGER REFERENCES class_groups(id),
        is_active BOOLEAN DEFAULT TRUE,
        registered_at TIMESTAMPTZ DEFAULT NOW()
    );

CREATE INDEX idx_users_tg_chat_id ON users(tg_chat_id);


-- +goose Down

DROP TABLE users;