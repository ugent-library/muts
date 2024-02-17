-- +goose Up

CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE records (
  id TEXT PRIMARY KEY,
  type ltree
);

CREATE INDEX records_type_gist_idx ON records USING GIST (type);

CREATE TABLE attributes (
  record_id TEXT NOT NULL REFERENCES records (id) ON DELETE CASCADE,
  name LTREE NOT NULL CHECK (name <> ''),
  value TEXT NOT NULL CHECK (value <> '')
);

CREATE INDEX attributes_record_id_fkey ON attributes (record_id);
CREATE INDEX attributes_name_gist_idx ON attributes USING GIST (name);
CREATE INDEX attributes_value_gist_idx ON attributes USING GIST (value);

CREATE TABLE relations (
  from_id TEXT NOT NULL REFERENCES records (id) ON DELETE CASCADE,
  to_id TEXT NOT NULL REFERENCES records (id) ON DELETE CASCADE,
  name TEXT NOT NULL CHECK (name <> ''),
  -- TODO using an int position is nog very efficient when rearranging large lists
  -- possible solutions: use a floating point or alphanumeric position; or a linked list
  position INT NOT NULL,
  -- TODO uniqueness only makes sense if we don't introduce relation attributes
  UNIQUE (from_id, to_id, name)
);

CREATE TABLE mutations (
  id TEXT PRIMARY KEY,
  record_id TEXT NOT NULL,
  author TEXT NOT NULL,
  reason TEXT,
  -- ops JSONB NOT NULL CHECK (jsonb_typeof(ops) = 'array' AND jsonb_array_length(ops) > 0),
  ops JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX mutations_record_id_fkey ON mutations (record_id);
CREATE INDEX mutations_created_at_key ON mutations (created_at);

-- +goose Down

DROP TABLE records CASCADE;
DROP TABLE mutations CASCADE;
DROP TABLE attributes CASCADE;
DROP TABLE relations CASCADE;
