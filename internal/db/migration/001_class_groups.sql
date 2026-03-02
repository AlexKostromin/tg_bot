-- +goose Up

CREATE TABLE class_groups
(
    id          SERIAL primary key,
    name        VARCHAR(20) NOT NULL UNIQUE,
    description TEXT
);

INSERT INTO class_groups (name, description)
VALUES ('5-6', 'Ученики 5 и 6 классов'),
       ('7-9', 'Ученики 7, 8 и 9 классов'),
       ('10-11', 'Ученики 10 и 11 классов');

-- +goose Down

DROP TABLE class_groups;