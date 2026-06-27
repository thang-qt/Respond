CREATE TABLE prompts (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  topic         TEXT NOT NULL,
  category_id   UUID NOT NULL REFERENCES categories(id),
  used_on       DATE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_prompts_category_id ON prompts (category_id);
CREATE INDEX idx_prompts_used_on ON prompts (used_on);
CREATE UNIQUE INDEX idx_prompts_used_on_unique ON prompts (used_on)
  WHERE used_on IS NOT NULL;
