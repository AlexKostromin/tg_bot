-- +goose Up

CREATE TABLE tutors
(
    id         SERIAL PRIMARY KEY,
    full_name  VARCHAR(255) NOT NULL,
    tg_chat_id BIGINT UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TABLE tutor_groups
(
    tutor_id       INTEGER REFERENCES tutors (id) ON DELETE CASCADE,
    class_group_id INTEGER REFERENCES class_groups (id) ON DELETE CASCADE,
    PRIMARY KEY (tutor_id, class_group_id)
);
CREATE TABLE tutor_subjects
(
    tutor_id       INTEGER REFERENCES tutors (id) ON DELETE CASCADE,
    subject_id INTEGER REFERENCES subjects (id) ON DELETE CASCADE,
    PRIMARY KEY (tutor_id, subject_id)
);

-- +goose Down
DROP TABLE tutors;
DROP TABLE tutor_groups;
DROP TABLE tutor_subjects

