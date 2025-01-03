
-- +migrate Up
CREATE TABLE IF NOT EXISTS todos(
	id BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	user_id BIGINT NOT NULL,
	title VARCHAR(255) NOT NULL,
	content TEXT,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS todos;
