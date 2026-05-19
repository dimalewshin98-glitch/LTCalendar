CREATE TABLE users (
    user_id INTEGER PRIMARY KEY,
    user_login VARCHAR(255) NOT NULL UNIQUE,
    user_password VARCHAR(255) NOT NULL
);