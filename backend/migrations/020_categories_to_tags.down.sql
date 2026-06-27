CREATE TABLE categories (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          TEXT NOT NULL UNIQUE,
  description   TEXT NOT NULL DEFAULT '',
  slug          TEXT NOT NULL,
  display_order INTEGER NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_categories_display_order ON categories (display_order);
CREATE UNIQUE INDEX idx_categories_slug_unique ON categories (slug);

INSERT INTO categories (name, description, slug, display_order)
VALUES
  ('Philosophy & Ethics', 'Metaphysics, epistemology, moral dilemmas, logic', 'philosophy-ethics', 1),
  ('Politics & Law', 'Governance, policy, democracy, legal systems', 'politics-law', 2),
  ('Technology', 'AI, software, internet, digital culture', 'technology', 3),
  ('Science', 'Research, method, theory, evidence', 'science', 4),
  ('Society & Culture', 'Education, media, social norms, identity', 'society-culture', 5),
  ('Economics', 'Markets, labor, inequality, trade', 'economics', 6),
  ('Religion & Spirituality', 'Faith, theology, secularism, meaning', 'religion-spirituality', 7),
  ('Health & Environment', 'Medicine, wellness, climate, sustainability', 'health-environment', 8),
  ('History', 'Events, revisionism, counterfactuals, lessons', 'history', 9),
  ('Sports & Entertainment', 'Competition, pop culture, games, media', 'sports-entertainment', 10),
  ('Hypothetical', 'Thought experiments, "what if" scenarios', 'hypothetical', 11),
  ('Free Topic', 'Anything that doesn''t fit elsewhere', 'free-topic', 12);

ALTER TABLE debates ADD COLUMN category_id UUID;
ALTER TABLE prompts ADD COLUMN category_id UUID;

WITH picked_debate_tag AS (
  SELECT DISTINCT ON (dt.debate_id)
    dt.debate_id,
    t.slug AS tag_slug
  FROM debate_tags dt
  JOIN tags t ON t.id = dt.tag_id
  ORDER BY dt.debate_id, COALESCE(t.display_order, 1000000), t.slug
), mapped_debate_category AS (
  SELECT
    pdt.debate_id,
    CASE
      WHEN pdt.tag_slug IN ('philosophy', 'ethics') THEN 'philosophy-ethics'
      WHEN pdt.tag_slug IN ('politics', 'law', 'international', 'security') THEN 'politics-law'
      WHEN pdt.tag_slug IN ('technology', 'ai') THEN 'technology'
      WHEN pdt.tag_slug = 'science' THEN 'science'
      WHEN pdt.tag_slug IN ('society', 'culture', 'education', 'psychology') THEN 'society-culture'
      WHEN pdt.tag_slug IN ('economics', 'business') THEN 'economics'
      WHEN pdt.tag_slug = 'religion' THEN 'religion-spirituality'
      WHEN pdt.tag_slug IN ('health', 'environment', 'lifestyle') THEN 'health-environment'
      WHEN pdt.tag_slug = 'history' THEN 'history'
      WHEN pdt.tag_slug IN ('sports', 'art') THEN 'sports-entertainment'
      WHEN pdt.tag_slug = 'future' THEN 'hypothetical'
      ELSE 'free-topic'
    END AS category_slug
  FROM picked_debate_tag pdt
)
UPDATE debates d
SET category_id = c.id
FROM mapped_debate_category mdc
JOIN categories c ON c.slug = mdc.category_slug
WHERE d.id = mdc.debate_id;

UPDATE debates d
SET category_id = c.id
FROM categories c
WHERE d.category_id IS NULL
  AND c.slug = 'free-topic';

WITH picked_prompt_tag AS (
  SELECT DISTINCT ON (pt.prompt_id)
    pt.prompt_id,
    t.slug AS tag_slug
  FROM prompt_tags pt
  JOIN tags t ON t.id = pt.tag_id
  ORDER BY pt.prompt_id, COALESCE(t.display_order, 1000000), t.slug
), mapped_prompt_category AS (
  SELECT
    ppt.prompt_id,
    CASE
      WHEN ppt.tag_slug IN ('philosophy', 'ethics') THEN 'philosophy-ethics'
      WHEN ppt.tag_slug IN ('politics', 'law', 'international', 'security') THEN 'politics-law'
      WHEN ppt.tag_slug IN ('technology', 'ai') THEN 'technology'
      WHEN ppt.tag_slug = 'science' THEN 'science'
      WHEN ppt.tag_slug IN ('society', 'culture', 'education', 'psychology') THEN 'society-culture'
      WHEN ppt.tag_slug IN ('economics', 'business') THEN 'economics'
      WHEN ppt.tag_slug = 'religion' THEN 'religion-spirituality'
      WHEN ppt.tag_slug IN ('health', 'environment', 'lifestyle') THEN 'health-environment'
      WHEN ppt.tag_slug = 'history' THEN 'history'
      WHEN ppt.tag_slug IN ('sports', 'art') THEN 'sports-entertainment'
      WHEN ppt.tag_slug = 'future' THEN 'hypothetical'
      ELSE 'free-topic'
    END AS category_slug
  FROM picked_prompt_tag ppt
)
UPDATE prompts p
SET category_id = c.id
FROM mapped_prompt_category mpc
JOIN categories c ON c.slug = mpc.category_slug
WHERE p.id = mpc.prompt_id;

UPDATE prompts p
SET category_id = c.id
FROM categories c
WHERE p.category_id IS NULL
  AND c.slug = 'free-topic';

ALTER TABLE debates ALTER COLUMN category_id SET NOT NULL;
ALTER TABLE prompts ALTER COLUMN category_id SET NOT NULL;

ALTER TABLE debates
  ADD CONSTRAINT debates_category_id_fkey
  FOREIGN KEY (category_id) REFERENCES categories(id);

ALTER TABLE prompts
  ADD CONSTRAINT prompts_category_id_fkey
  FOREIGN KEY (category_id) REFERENCES categories(id);

CREATE INDEX idx_debates_category_id ON debates (category_id);
CREATE INDEX idx_prompts_category_id ON prompts (category_id);

DROP TABLE prompt_tags;
DROP TABLE debate_tags;
DROP INDEX IF EXISTS idx_tags_display_order;
DROP TABLE tags;
