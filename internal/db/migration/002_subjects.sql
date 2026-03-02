-- +goose Up

CREATE TABLE subjects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    class_group_id INTEGER REFERENCES class_groups(id)
);

-- +goose Down

DROP TABLE subjects;