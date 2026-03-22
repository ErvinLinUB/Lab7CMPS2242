CREATE TABLE IF NOT EXISTS courses (
    id         bigserial PRIMARY KEY,
    code       varchar(20) NOT NULL,
    title      text NOT NULL,
    credits    int NOT NULL,
    enrolled   int NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_courses_code ON courses (code);

INSERT INTO courses (code, title, credits, enrolled) VALUES
    ('CMPS2212', 'GUI Programming',  3, 28),
    ('CMPS3412', 'Database Systems', 3, 22);