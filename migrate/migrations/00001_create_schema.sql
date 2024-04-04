-- +goose Up

CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS btree_gin;

CREATE TABLE muts_records (
  id TEXT PRIMARY KEY CHECK (id <> ''),
  kind LTREE NOT NULL CHECK (kind <> ''),
  attributes JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX muts_records_kind_gist_idx ON muts_records USING gist (kind);
CREATE INDEX muts_records_attributes_gin_idx ON muts_records USING gin (attributes jsonb_path_ops);

-- TODO using an int position is not very efficient when rearranging large lists
-- possible solutions: use a floating point or alphanumeric position, or a linked list
-- https://stackoverflow.com/questions/9536262/best-representation-of-an-ordered-list-in-a-database
-- https://stackoverflow.com/questions/38923376/return-a-new-string-that-sorts-between-two-given-strings/38927158#38927158
CREATE TABLE muts_relations (
  id TEXT PRIMARY KEY CHECK (id <> ''),
  kind LTREE NOT NULL CHECK (kind <> ''),
  from_id TEXT NOT NULL REFERENCES muts_records (id) ON DELETE CASCADE,
  to_id TEXT NOT NULL REFERENCES muts_records (id) ON DELETE CASCADE,
  position INT NOT NULL,
  attributes JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX muts_relations_from_id_fkey ON muts_relations (from_id);
CREATE INDEX muts_relations_to_id_fkey ON muts_relations (to_id);
CREATE INDEX muts_relations_kind_gist_idx ON muts_relations USING gist (kind);
CREATE INDEX muts_relations_position_idx ON muts_relations (position);
CREATE INDEX muts_relations_attributes_gin_idx ON muts_relations USING gin (attributes jsonb_path_ops);

CREATE TABLE muts_mutations (
  id BIGSERIAL PRIMARY KEY,
  record_id TEXT NOT NULL CHECK (record_id <> ''),
  author TEXT NOT NULL CHECK (author <> ''),
  reason TEXT CHECK (reason <> ''),
  ops JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT transaction_timestamp()
);

CREATE INDEX muts_mutations_record_id_key ON muts_mutations (record_id);
CREATE INDEX muts_mutations_created_at_key ON muts_mutations (created_at);

-- +goose Down

DROP TABLE muts_relations CASCADE;
DROP TABLE muts_records CASCADE;
DROP TABLE muts_mutations CASCADE;
