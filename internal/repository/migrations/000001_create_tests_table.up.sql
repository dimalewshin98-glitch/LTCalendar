CREATE TABLE tests (
    uuid VARCHAR(255) PRIMARY KEY,
    test_name VARCHAR(255) NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    tps DECIMAL(10, 3),
    additional_params TEXT NOT NULL,
    user_id INTEGER,
    is_deleted BOOLEAN
);

CREATE INDEX idx_test_name ON tests(test_name);
CREATE INDEX idx_start_time ON tests(start_time);