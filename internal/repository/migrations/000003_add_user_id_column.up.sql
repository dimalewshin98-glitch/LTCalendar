ALTER TABLE urls ADD COLUMN user_id INTEGER;
CREATE INDEX idx_user_id ON urls(user_id);