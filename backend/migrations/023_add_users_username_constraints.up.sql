ALTER TABLE users
  ADD CONSTRAINT chk_users_username_length
  CHECK (char_length(username) BETWEEN 5 AND 20) NOT VALID;

ALTER TABLE users
  ADD CONSTRAINT chk_users_username_format
  CHECK (username ~ '^[A-Za-z0-9_]+$') NOT VALID;
