CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email               TEXT NOT NULL,
  email_verified      BOOLEAN NOT NULL DEFAULT false,
  username            TEXT NOT NULL,
  password_hash       TEXT NOT NULL,
  bio                 TEXT NOT NULL DEFAULT '',
  rating              INTEGER NOT NULL DEFAULT 1200,
  wins                INTEGER NOT NULL DEFAULT 0,
  losses              INTEGER NOT NULL DEFAULT 0,
  draws               INTEGER NOT NULL DEFAULT 0,
  default_reveal      BOOLEAN NOT NULL DEFAULT false,
  username_changed_at TIMESTAMPTZ,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_email ON users (LOWER(email));
CREATE UNIQUE INDEX idx_users_username ON users (LOWER(username));
CREATE INDEX idx_users_rating ON users (rating DESC);
CREATE INDEX idx_users_created_at ON users (created_at DESC);

INSERT INTO users (id, email, username, password_hash)
VALUES (
  '00000000-0000-0000-0000-000000000000',
  'system@respond.im',
  'system',
  'not-a-valid-hash'
);
