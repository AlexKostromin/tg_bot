-- +goose Up

CREATE TABLE subjects (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

-- Связь many-to-many: один предмет доступен нескольким группам,
-- одна группа может заниматься несколькими предметами.
CREATE TABLE subject_groups (
    subject_id     INTEGER REFERENCES subjects(id)     ON DELETE CASCADE,
    class_group_id INTEGER REFERENCES class_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (subject_id, class_group_id)
);

-- Наполняем предметами
INSERT INTO subjects (name) VALUES
    ('Математика'),
    ('Физика'),
    ('Информатика');

-- Привязываем все три предмета ко всем трём группам
INSERT INTO subject_groups (subject_id, class_group_id)
SELECT s.id, g.id
FROM subjects s
CROSS JOIN class_groups g;

-- +goose Down

DROP TABLE subject_groups;
DROP TABLE subjects;
