DROP TRIGGER IF EXISTS tr_refresh_debate_search_document_on_turns ON turns;
DROP TRIGGER IF EXISTS tr_refresh_debate_search_document_on_debates ON debates;

DROP FUNCTION IF EXISTS trg_refresh_debate_search_document_from_turns();
DROP FUNCTION IF EXISTS trg_refresh_debate_search_document_from_debates();
DROP FUNCTION IF EXISTS refresh_debate_search_document(UUID);

DROP TABLE IF EXISTS debate_search_documents;

DROP INDEX IF EXISTS idx_users_username_trgm;
DROP INDEX IF EXISTS idx_tags_slug_trgm;
DROP INDEX IF EXISTS idx_tags_name_trgm;
