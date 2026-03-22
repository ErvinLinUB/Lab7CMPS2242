CREATE TABLE IF NOT EXISTS students (
    id         bigserial PRIMARY KEY,
    name       varchar(100) NOT NULL,
    programme  text NOT NULL,
    year       smallint NOT NULL CHECK (year BETWEEN 1 AND 4),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_students_name ON students (name);

INSERT INTO students (name, programme, year) VALUES
    ('Eve Castillo',   'BSc Computer Science',    2),
    ('Marco Tillett',  'BSc Computer Science',    3),
    ('Aisha Gentle',   'BSc Information Systems', 1),
    ('Raj Palacio',    'BSc Computer Science',    4);