CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE debate_search_documents (
  debate_id UUID PRIMARY KEY REFERENCES debates(id) ON DELETE CASCADE,
  document TSVECTOR NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_debate_search_documents_document
  ON debate_search_documents
  USING GIN (document);

CREATE INDEX idx_tags_name_trgm
  ON tags
  USING GIN (lower(name) gin_trgm_ops);

CREATE INDEX idx_tags_slug_trgm
  ON tags
  USING GIN (lower(slug) gin_trgm_ops);

CREATE INDEX idx_users_username_trgm
  ON users
  USING GIN (lower(username) gin_trgm_ops);

CREATE OR REPLACE FUNCTION refresh_debate_search_document(p_debate_id UUID)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
  INSERT INTO debate_search_documents (debate_id, document, updated_at)
  SELECT
    d.id,
    setweight(to_tsvector('english', COALESCE(d.topic, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(d.context, '')), 'B') ||
    setweight(
      to_tsvector(
        'english',
        COALESCE(
          (
            SELECT string_agg(t.content, ' ')
            FROM turns t
            WHERE t.debate_id = d.id
              AND t.is_system = false
          ),
          ''
        )
      ),
      'C'
    ),
    now()
  FROM debates d
  WHERE d.id = p_debate_id
  ON CONFLICT (debate_id)
  DO UPDATE
    SET document = EXCLUDED.document,
        updated_at = now();
END;
$$;

CREATE OR REPLACE FUNCTION trg_refresh_debate_search_document_from_debates()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    DELETE FROM debate_search_documents
    WHERE debate_id = OLD.id;
    RETURN OLD;
  END IF;

  PERFORM refresh_debate_search_document(NEW.id);
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION trg_refresh_debate_search_document_from_turns()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
  v_debate_id UUID;
BEGIN
  v_debate_id := COALESCE(NEW.debate_id, OLD.debate_id);

  IF v_debate_id IS NOT NULL THEN
    PERFORM refresh_debate_search_document(v_debate_id);
  END IF;

  RETURN COALESCE(NEW, OLD);
END;
$$;

CREATE TRIGGER tr_refresh_debate_search_document_on_debates
AFTER INSERT OR UPDATE OF topic, context OR DELETE
ON debates
FOR EACH ROW
EXECUTE FUNCTION trg_refresh_debate_search_document_from_debates();

CREATE TRIGGER tr_refresh_debate_search_document_on_turns
AFTER INSERT OR UPDATE OF content, is_system OR DELETE
ON turns
FOR EACH ROW
EXECUTE FUNCTION trg_refresh_debate_search_document_from_turns();

INSERT INTO debate_search_documents (debate_id, document, updated_at)
SELECT
  d.id,
  setweight(to_tsvector('english', COALESCE(d.topic, '')), 'A') ||
  setweight(to_tsvector('english', COALESCE(d.context, '')), 'B') ||
  setweight(
    to_tsvector(
      'english',
      COALESCE(
        (
          SELECT string_agg(t.content, ' ')
          FROM turns t
          WHERE t.debate_id = d.id
            AND t.is_system = false
        ),
        ''
      )
    ),
    'C'
  ),
  now()
FROM debates d
ON CONFLICT (debate_id)
DO UPDATE
  SET document = EXCLUDED.document,
      updated_at = now();
