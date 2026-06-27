-- Merge Ethics into Philosophy & Ethics
-- Move all debates from Ethics to Philosophy before renaming
UPDATE debates
SET category_id = (SELECT id FROM categories WHERE name = 'Philosophy')
WHERE category_id = (SELECT id FROM categories WHERE name = 'Ethics');

-- Move all prompts from Ethics to Philosophy before deleting
UPDATE prompts
SET category_id = (SELECT id FROM categories WHERE name = 'Philosophy')
WHERE category_id = (SELECT id FROM categories WHERE name = 'Ethics');

-- Delete the now-empty Ethics category
DELETE FROM categories WHERE name = 'Ethics';

-- Rename existing categories
UPDATE categories SET name = 'Philosophy & Ethics', description = 'Metaphysics, epistemology, moral dilemmas, logic', slug = 'philosophy-ethics' WHERE name = 'Philosophy';
UPDATE categories SET name = 'Politics & Law', description = 'Governance, policy, democracy, legal systems', slug = 'politics-law' WHERE name = 'Politics';
UPDATE categories SET name = 'Society & Culture', description = 'Education, media, social norms, identity', slug = 'society-culture' WHERE name = 'Society';
UPDATE categories SET name = 'Sports & Entertainment', description = 'Competition, pop culture, games, media', slug = 'sports-entertainment' WHERE name = 'Sports';

-- Update descriptions for unchanged-name categories
UPDATE categories SET description = 'AI, software, internet, digital culture' WHERE name = 'Technology';
UPDATE categories SET description = 'Research, method, theory, evidence' WHERE name = 'Science';
UPDATE categories SET description = 'Markets, labor, inequality, trade' WHERE name = 'Economics';
UPDATE categories SET description = 'Thought experiments, "what if" scenarios' WHERE name = 'Hypothetical';
UPDATE categories SET description = 'Anything that doesn''t fit elsewhere' WHERE name = 'Free Topic';

-- Add new categories
INSERT INTO categories (name, description, slug, display_order)
VALUES
  ('Religion & Spirituality', 'Faith, theology, secularism, meaning', 'religion-spirituality', 7),
  ('Health & Environment', 'Medicine, wellness, climate, sustainability', 'health-environment', 8),
  ('History', 'Events, revisionism, counterfactuals, lessons', 'history', 9);

-- Reorder all categories
UPDATE categories SET display_order = 1 WHERE name = 'Philosophy & Ethics';
UPDATE categories SET display_order = 2 WHERE name = 'Politics & Law';
UPDATE categories SET display_order = 3 WHERE name = 'Technology';
UPDATE categories SET display_order = 4 WHERE name = 'Science';
UPDATE categories SET display_order = 5 WHERE name = 'Society & Culture';
UPDATE categories SET display_order = 6 WHERE name = 'Economics';
UPDATE categories SET display_order = 7 WHERE name = 'Religion & Spirituality';
UPDATE categories SET display_order = 8 WHERE name = 'Health & Environment';
UPDATE categories SET display_order = 9 WHERE name = 'History';
UPDATE categories SET display_order = 10 WHERE name = 'Sports & Entertainment';
UPDATE categories SET display_order = 11 WHERE name = 'Hypothetical';
UPDATE categories SET display_order = 12 WHERE name = 'Free Topic';
