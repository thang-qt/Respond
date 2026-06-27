-- Add turn_nudge_sent to debates to track whether the 75% turn expiry
-- notification has been sent for the current turn. Reset to false on each
-- new turn submission.
ALTER TABLE debates ADD COLUMN turn_nudge_sent BOOLEAN NOT NULL DEFAULT false;
