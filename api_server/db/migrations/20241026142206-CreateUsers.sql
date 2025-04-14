
-- +migrate Up
CREATE TABLE IF NOT EXISTS users(
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	first_name VARCHAR(255) NOT NULL,
	last_name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	password VARCHAR(255) NOT NULL,
	front_identification VARCHAR(255) NOT NULL,
	back_identification VARCHAR(255) NOT NULL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
