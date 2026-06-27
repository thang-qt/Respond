ALTER TABLE debates
  DROP COLUMN IF EXISTS side_b_rating_delta,
  DROP COLUMN IF EXISTS side_a_rating_delta;
