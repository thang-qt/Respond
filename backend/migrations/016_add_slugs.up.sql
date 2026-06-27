ALTER TABLE categories ADD COLUMN slug TEXT;
ALTER TABLE debates ADD COLUMN slug TEXT;

UPDATE categories
SET slug = CASE
  WHEN trim(both '-' from regexp_replace(lower(name), '[^a-z0-9]+', '-', 'g')) = ''
    THEN 'category-' || substring(id::text from 1 for 8)
  ELSE trim(both '-' from regexp_replace(lower(name), '[^a-z0-9]+', '-', 'g'))
END;

WITH base AS (
  SELECT
    id,
    created_at,
    trim(both '-' from regexp_replace(lower(topic), '[^a-z0-9]+', '-', 'g')) AS base_slug
  FROM debates
),
ranked AS (
  SELECT
    id,
    CASE
      WHEN base_slug = '' THEN 'debate'
      ELSE base_slug
    END AS base_slug,
    row_number() OVER (
      PARTITION BY CASE WHEN base_slug = '' THEN 'debate' ELSE base_slug END
      ORDER BY created_at, id
    ) AS rn
  FROM base
)
UPDATE debates d
SET slug = CASE
  WHEN r.rn = 1 THEN r.base_slug
  ELSE r.base_slug || '-' || r.rn
END
FROM ranked r
WHERE d.id = r.id;

ALTER TABLE categories ALTER COLUMN slug SET NOT NULL;
ALTER TABLE debates ALTER COLUMN slug SET NOT NULL;

CREATE UNIQUE INDEX idx_categories_slug_unique ON categories (slug);
CREATE UNIQUE INDEX idx_debates_slug_unique ON debates (slug);
