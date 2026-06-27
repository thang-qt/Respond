DROP INDEX IF EXISTS idx_debates_slug_unique;
DROP INDEX IF EXISTS idx_categories_slug_unique;

ALTER TABLE debates DROP COLUMN IF EXISTS slug;
ALTER TABLE categories DROP COLUMN IF EXISTS slug;
