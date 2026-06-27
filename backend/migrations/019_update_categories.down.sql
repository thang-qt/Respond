-- Remove new categories
DELETE FROM categories WHERE name IN ('Religion & Spirituality', 'Health & Environment', 'History');

-- Rename categories back
UPDATE categories SET name = 'Philosophy', description = 'Ethics, metaphysics, epistemology, logic', slug = 'philosophy' WHERE name = 'Philosophy & Ethics';
UPDATE categories SET name = 'Politics', description = 'Governance, policy, law, democracy', slug = 'politics' WHERE name = 'Politics & Law';
UPDATE categories SET name = 'Society', description = 'Culture, education, media, social norms', slug = 'society' WHERE name = 'Society & Culture';
UPDATE categories SET name = 'Sports', description = 'Rules, fairness, records, competition', slug = 'sports' WHERE name = 'Sports & Entertainment';

-- Re-create Ethics category
INSERT INTO categories (name, description, slug, display_order)
VALUES ('Ethics', 'Moral dilemmas, applied ethics, justice', 'ethics', 5);

-- Restore original display order
UPDATE categories SET display_order = 1 WHERE name = 'Philosophy';
UPDATE categories SET display_order = 2 WHERE name = 'Technology';
UPDATE categories SET display_order = 3 WHERE name = 'Politics';
UPDATE categories SET display_order = 4 WHERE name = 'Science';
UPDATE categories SET display_order = 5 WHERE name = 'Ethics';
UPDATE categories SET display_order = 6 WHERE name = 'Society';
UPDATE categories SET display_order = 7 WHERE name = 'Economics';
UPDATE categories SET display_order = 8 WHERE name = 'Sports';
UPDATE categories SET display_order = 9 WHERE name = 'Hypothetical';
UPDATE categories SET display_order = 10 WHERE name = 'Free Topic';
