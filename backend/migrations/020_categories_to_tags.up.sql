CREATE TABLE tags (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug          TEXT NOT NULL UNIQUE,
  name          TEXT NOT NULL UNIQUE,
  description   TEXT NOT NULL DEFAULT '',
  display_order INTEGER,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tags_display_order ON tags (display_order)
  WHERE display_order IS NOT NULL;

CREATE TABLE debate_tags (
  debate_id UUID NOT NULL REFERENCES debates(id) ON DELETE CASCADE,
  tag_id    UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (debate_id, tag_id)
);

CREATE INDEX idx_debate_tags_tag_id ON debate_tags (tag_id);

CREATE TABLE prompt_tags (
  prompt_id UUID NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
  tag_id    UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (prompt_id, tag_id)
);

CREATE INDEX idx_prompt_tags_tag_id ON prompt_tags (tag_id);

INSERT INTO tags (slug, name, description, display_order)
VALUES
  ('philosophy', 'Philosophy', 'Metaphysics, epistemology, logic', 1),
  ('ethics', 'Ethics', 'Moral dilemmas, justice, applied ethics', 2),
  ('politics', 'Politics', 'Governance, elections, public policy', 3),
  ('law', 'Law', 'Legal systems, rights, regulation', 4),
  ('economics', 'Economics', 'Markets, labor, inequality, trade', 5),
  ('business', 'Business', 'Strategy, operations, entrepreneurship', 6),
  ('technology', 'Technology', 'Software, internet, digital culture', 7),
  ('ai', 'AI', 'Machine learning, alignment, automation', 8),
  ('science', 'Science', 'Method, evidence, scientific theory', 9),
  ('health', 'Health', 'Medicine, public health, wellness', 10),
  ('environment', 'Environment', 'Climate, sustainability, conservation', 11),
  ('education', 'Education', 'Learning systems, pedagogy, access', 12),
  ('psychology', 'Psychology', 'Behavior, cognition, mental models', 13),
  ('society', 'Society', 'Social structure, institutions, norms', 14),
  ('culture', 'Culture', 'Identity, media, values', 15),
  ('art', 'Art', 'Aesthetics, expression, criticism', 16),
  ('history', 'History', 'Historical events, interpretation, lessons', 17),
  ('religion', 'Religion', 'Faith, theology, spirituality', 18),
  ('sports', 'Sports', 'Competition, rules, fairness', 19),
  ('international', 'International', 'Geopolitics, diplomacy, global governance', 20),
  ('security', 'Security', 'Defense, cybersecurity, risk', 21),
  ('lifestyle', 'Lifestyle', 'Daily habits, consumer choices, wellbeing', 22),
  ('future', 'Future', 'Forecasts, scenarios, long-term impacts', 23),
  ('meta', 'Meta', 'Debate norms, platform questions, miscellaneous', 24);

WITH category_tag_map(category_slug, tag_slug) AS (
  VALUES
    ('philosophy-ethics', 'philosophy'),
    ('philosophy-ethics', 'ethics'),
    ('politics-law', 'politics'),
    ('politics-law', 'law'),
    ('technology', 'technology'),
    ('technology', 'ai'),
    ('science', 'science'),
    ('society-culture', 'society'),
    ('society-culture', 'culture'),
    ('society-culture', 'education'),
    ('economics', 'economics'),
    ('economics', 'business'),
    ('religion-spirituality', 'religion'),
    ('health-environment', 'health'),
    ('health-environment', 'environment'),
    ('history', 'history'),
    ('sports-entertainment', 'sports'),
    ('sports-entertainment', 'art'),
    ('hypothetical', 'future'),
    ('hypothetical', 'philosophy'),
    ('free-topic', 'meta')
)
INSERT INTO debate_tags (debate_id, tag_id)
SELECT d.id, t.id
FROM debates d
JOIN categories c ON c.id = d.category_id
JOIN category_tag_map m ON m.category_slug = c.slug
JOIN tags t ON t.slug = m.tag_slug
ON CONFLICT DO NOTHING;

WITH category_tag_map(category_slug, tag_slug) AS (
  VALUES
    ('philosophy-ethics', 'philosophy'),
    ('philosophy-ethics', 'ethics'),
    ('politics-law', 'politics'),
    ('politics-law', 'law'),
    ('technology', 'technology'),
    ('technology', 'ai'),
    ('science', 'science'),
    ('society-culture', 'society'),
    ('society-culture', 'culture'),
    ('society-culture', 'education'),
    ('economics', 'economics'),
    ('economics', 'business'),
    ('religion-spirituality', 'religion'),
    ('health-environment', 'health'),
    ('health-environment', 'environment'),
    ('history', 'history'),
    ('sports-entertainment', 'sports'),
    ('sports-entertainment', 'art'),
    ('hypothetical', 'future'),
    ('hypothetical', 'philosophy'),
    ('free-topic', 'meta')
)
INSERT INTO prompt_tags (prompt_id, tag_id)
SELECT p.id, t.id
FROM prompts p
JOIN categories c ON c.id = p.category_id
JOIN category_tag_map m ON m.category_slug = c.slug
JOIN tags t ON t.slug = m.tag_slug
ON CONFLICT DO NOTHING;

ALTER TABLE debates DROP COLUMN category_id;
ALTER TABLE prompts DROP COLUMN category_id;
DROP TABLE categories;
