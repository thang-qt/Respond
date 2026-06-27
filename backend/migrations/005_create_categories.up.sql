CREATE TABLE categories (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          TEXT NOT NULL UNIQUE,
  description   TEXT NOT NULL DEFAULT '',
  display_order INTEGER NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_categories_display_order ON categories (display_order);

INSERT INTO categories (name, description, display_order)
VALUES
  ('Philosophy', 'Ethics, metaphysics, epistemology, logic', 1),
  ('Technology', 'AI, software, internet, digital culture', 2),
  ('Politics', 'Governance, policy, law, democracy', 3),
  ('Science', 'Research, method, theory, evidence', 4),
  ('Ethics', 'Moral dilemmas, applied ethics, justice', 5),
  ('Society', 'Culture, education, media, social norms', 6),
  ('Economics', 'Markets, labor, inequality, systems', 7),
  ('Sports', 'Rules, fairness, records, competition', 8),
  ('Hypothetical', 'Thought experiments, \"what if\" scenarios', 9),
  ('Free Topic', 'Anything that doesn''t fit elsewhere', 10);
