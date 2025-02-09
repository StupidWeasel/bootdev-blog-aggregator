-- +goose Up
CREATE TABLE feed_follows (
	id SERIAL PRIMARY KEY,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	user_id UUID NOT NULL,
	feed_id INTEGER NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY(feed_id) REFERENCES feeds(id),
	CONSTRAINT UC_userfeed UNIQUE (user_id, feed_id)
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
BEFORE UPDATE ON feed_follows
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS set_updated_at ON feed_follows;
DROP TABLE feed_follows;
