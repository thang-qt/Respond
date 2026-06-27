CREATE TYPE time_mode AS ENUM (
  'marathon',
  'standard',
  'rapid',
  'blitz'
);

CREATE TYPE debate_status AS ENUM (
  'waiting',
  'active',
  'pending_extension',
  'waiting_replacement',
  'finished'
);

CREATE TYPE debate_outcome AS ENUM (
  'concede',
  'resign',
  'draw',
  'walkover',
  'expiry',
  'turn_limit'
);

CREATE TYPE debate_side AS ENUM (
  'a',
  'b'
);
