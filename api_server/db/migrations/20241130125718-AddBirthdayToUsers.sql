
-- +migrate Up
ALTER TABLE users ADD birthday DATE AFTER password;

-- +migrate Down
ALTER TABLE users DROP COLUMN birthday;
