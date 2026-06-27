-- 043_create_challenge_lobby.up.sql
-- Adds challenge_lobby_entries and challenge_lobby_entry_tags tables.
-- One entry per user (PK on user_id). Tags via join table (0–15 per entry,
-- enforced at application layer).

CREATE TABLE challenge_lobby_entries (
  user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  bio_note   TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE challenge_lobby_entries
  ADD CONSTRAINT chk_lobby_bio_note_length
  CHECK (char_length(bio_note) <= 300);

CREATE TABLE challenge_lobby_entry_tags (
  user_id UUID NOT NULL REFERENCES challenge_lobby_entries(user_id) ON DELETE CASCADE,
  tag_id  UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, tag_id)
);

CREATE INDEX idx_challenge_lobby_entry_tags_tag_id
  ON challenge_lobby_entry_tags (tag_id);
