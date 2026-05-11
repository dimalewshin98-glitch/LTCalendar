CREATE TABLE urls (
    uuid VARCHAR(255) PRIMARY KEY,
    short_url VARCHAR(255) NOT NULL,
    original_url TEXT NOT NULL
);

CREATE INDEX idx_short_url ON urls(short_url);

CREATE INDEX idx_original_url ON urls(original_url);