-- +goose Up

CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS btree_gin;

CREATE TABLE muts_nodes (
  id text PRIMARY KEY CHECK (id <> ''),
  kind ltree NOT NULL CHECK (kind <> ''),
  rev uuid NOT NULL,
  properties jsonb NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE INDEX muts_nodes_kind_gist_idx ON muts_nodes USING gist (kind);
CREATE INDEX muts_nodes_properties_gin_idx ON muts_nodes USING gin (properties jsonb_path_ops);

-- TODO using an int position is not very efficient when rearranging large lists
-- possible solutions: use a floating point or alphanumeric position, or a linked list
-- https://stackoverflow.com/questions/9536262/best-representation-of-an-ordered-list-in-a-database
-- https://stackoverflow.com/questions/38923376/return-a-new-string-that-sorts-between-two-given-strings/38927158#38927158
CREATE TABLE muts_links (
  id TEXT PRIMARY KEY CHECK (id <> ''),
  kind LTREE NOT NULL CHECK (kind <> ''),
  from_id TEXT NOT NULL REFERENCES muts_nodes (id) ON DELETE CASCADE,
  to_id TEXT NOT NULL REFERENCES muts_nodes (id) ON DELETE CASCADE,
  position INT NOT NULL,
  properties JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  CHECK (from_id <> to_id)
);

CREATE INDEX muts_links_from_id_fkey ON muts_links (from_id);
CREATE INDEX muts_links_to_id_fkey ON muts_links (to_id);
CREATE INDEX muts_links_kind_gist_idx ON muts_links USING gist (kind);
CREATE INDEX muts_links_position_idx ON muts_links (position);
CREATE INDEX muts_links_properties_gin_idx ON muts_links USING gin (properties jsonb_path_ops);

-- +goose Down

DROP TABLE muts_links CASCADE;
DROP TABLE muts_nodes CASCADE;
