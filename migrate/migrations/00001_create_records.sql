-- +goose Up

CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS btree_gin;

CREATE TABLE records (
  id TEXT PRIMARY KEY CHECK (id <> ''),
  kind LTREE NOT NULL CHECK (kind <> ''),
  attributes JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX records_kind_gist_idx ON records USING gist (kind);
CREATE INDEX records_attributes_gin_idx ON records USING gin (attributes jsonb_path_ops);

-- TODO using an int position is not very efficient when rearranging large lists
-- possible solutions: use a floating point or alphanumeric position, or a linked list
-- https://stackoverflow.com/questions/9536262/best-representation-of-an-ordered-list-in-a-database
-- https://stackoverflow.com/questions/38923376/return-a-new-string-that-sorts-between-two-given-strings/38927158#38927158
CREATE TABLE relations (
  id TEXT PRIMARY KEY CHECK (id <> ''),
  from_id TEXT NOT NULL REFERENCES records (id) ON DELETE CASCADE,
  to_id TEXT NOT NULL REFERENCES records (id) ON DELETE CASCADE,
  kind LTREE NOT NULL CHECK (kind <> ''),
  position INT NOT NULL,
  attributes JSONB NOT NULL
);

CREATE INDEX relations_from_id_fkey ON relations (from_id);
CREATE INDEX relations_to_id_fkey ON relations (to_id);
CREATE INDEX relations_kind_gist_idx ON relations USING gist (kind);
CREATE INDEX relations_position_idx ON relations (position);
CREATE INDEX relations_attributes_gin_idx ON relations USING gin (attributes jsonb_path_ops);

-- TODO guarantee correct ordering through a monotonic version number
-- or rely on ulid ordering?
CREATE TABLE mutations (
  id BIGSERIAL PRIMARY KEY,
  record_id TEXT NOT NULL CHECK (record_id <> ''),
  author TEXT NOT NULL CHECK (author <> ''),
  reason TEXT CHECK (reason <> ''),
  ops JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT transaction_timestamp()
);

CREATE INDEX mutations_record_id_key ON mutations (record_id);
CREATE INDEX mutations_created_at_key ON mutations (created_at);

-- +goose Down

DROP TABLE relations CASCADE;
DROP TABLE records CASCADE;
DROP TABLE mutations CASCADE;
