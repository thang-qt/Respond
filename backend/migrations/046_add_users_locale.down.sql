ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_locale_supported;
ALTER TABLE users DROP COLUMN IF EXISTS locale;
