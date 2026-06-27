CREATE TYPE user_role AS ENUM ('user', 'moderator', 'admin');

ALTER TABLE users
  ADD COLUMN role user_role NOT NULL DEFAULT 'user';

CREATE INDEX idx_users_role ON users (role);
