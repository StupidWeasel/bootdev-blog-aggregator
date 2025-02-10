-- +goose Up
CREATE TABLE posts (
	id SERIAL PRIMARY KEY,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	title text NOT NULL,
	url text NOT NULL UNIQUE,
	description text DEFAULT NULL,
	published_at TIMESTAMPTZ NULL, -- sql type:"timestamp null"
	feed_id INTEGER NOT NULL,
	FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
	IF NEW.updated_at = OLD.updated_at THEN
		NEW.updated_at = CURRENT_TIMESTAMP;
	END IF;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON posts
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS set_updated_at ON posts;
DROP TABLE posts;
