-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION muts_mutate(
  _record_id text,
  _ops jsonb,
  _ts timestamptz = transaction_timestamp()
)
RETURNS VOID
LANGUAGE PLPGSQL
AS $$
DECLARE
	op jsonb;
BEGIN
  FOR op IN SELECT * FROM jsonb_array_elements(_ops)
  LOOP
    CASE op->>'name'
 
    WHEN 'add_rec' THEN
      INSERT INTO muts_records (id, kind, attributes, created_at, updated_at)
      VALUES (
        _record_id,
        (op->'args'->>'kind')::ltree,
        coalesce(op->'args'->'attrs', '{}'::jsonb),
        _ts,
        _ts
      );

    WHEN 'set_attr' THEN
      UPDATE muts_records
      SET attributes = attributes || jsonb_build_object(op->'args'->>'key', op->'args'->'val'), updated_at = _ts
      WHERE id = _record_id;

    WHEN 'del_attr' THEN
      UPDATE muts_records
      SET attributes = attributes - op->'args'->>'key', updated_at = _ts
      WHERE id = _record_id;

    WHEN 'clear_attrs' THEN
      UPDATE muts_records
      SET attributes = '{}'::jsonb, updated_at = _ts
      WHERE id = _record_id;

    WHEN 'add_rel' THEN
      INSERT INTO muts_relations (id, kind, from_id, to_id, position, attributes, created_at, updated_at)
      VALUES (
        op->'args'->>'id',
        (op->'args'->>'kind')::ltree,
        _record_id,
        op->'args'->>'to',
        (SELECT COUNT(*) FROM muts_relations WHERE from_id = _record_id AND kind = (op->'args'->>'kind')::ltree),
        coalesce(op->'args'->'attrs', '{}'::jsonb),
        _ts,
        _ts
      );

      UPDATE muts_records
      SET updated_at = _ts
      WHERE id = _record_id;

    WHEN 'set_rel_attr' THEN
      UPDATE muts_relations
      SET attributes = attributes || jsonb_build_object(op->'args'->>'key', op->'args'->'val'), updated_at = _ts
      WHERE id = op->'args'->>'id';

      UPDATE muts_records
      SET updated_at = _ts
      WHERE id = _record_id;
  
    WHEN 'del_rel_attr' THEN
      UPDATE muts_relations
      SET attributes = attributes - op->'args'->>'key', updated_at = _ts
      WHERE id = op->'args'->>'id';

      UPDATE muts_records
      SET updated_at = _ts
      WHERE id = _record_id;

    WHEN 'clear_rel_attrs' THEN
      UPDATE muts_relations
      SET attributes = '{}'::jsonb, updated_at = _ts
      WHERE id = op->'args'->>'id';

      UPDATE muts_records
      SET updated_at = _ts
      WHERE id = _record_id;

    WHEN 'del_rel' THEN
      WITH rel AS (
        DELETE FROM muts_relations
        WHERE id = op->'args'->>'id'
        RETURNING kind, position
      )
      UPDATE muts_relations r
      SET position = r.position - 1
      FROM rel
      WHERE from_id = _record_id AND r.kind = rel.kind AND r.position > rel.position;

      UPDATE muts_records
      SET updated_at = _ts
      WHERE id = _record_id;

    ELSE
      RAISE EXCEPTION 'Unknown operation "%"', op->>'name';

    END CASE;
  END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_relations_tree(_rec_id text, _conds jsonb = '{}', _level int = 7)
  RETURNS jsonb
  LANGUAGE sql STABLE PARALLEL SAFE AS
$$
SELECT CASE WHEN _level > 1
            THEN jsonb_strip_nulls(jsonb_agg(sub))
--            ELSE CASE WHEN count(*) > 0 THEN to_jsonb(count(*) || ' - stopped recursion at max level') END
       END
FROM  (
   SELECT rl.id,
          rl.kind, 
          rl.attributes,
          rl.position,
          rl.created_at,
          rl.updated_at,
          jsonb_build_object(
          	'id', r.id,
          	'kind', r.kind,
          	'attributes', r.attributes,
          	'created_at', r.created_at,
          	'updated_at', r.updated_at,
          	'relations', (CASE WHEN _level > 1 THEN muts_relations_tree(r.id, _conds, _level - 1) END)
   		  ) AS record
   FROM   muts_relations rl
   JOIN   muts_records r ON r.id = rl.to_id
   WHERE
     rl.from_id = _rec_id
     AND
    (CASE
      WHEN _conds ? 'follow' THEN rl.kind ~ (_conds->>'follow')::lquery
      ELSE TRUE
    END) 
  
   ORDER  BY r.id
   ) sub
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_select(
	in _conds jsonb = '{}',
	in _level int = 7,
	out id text,
	out kind ltree,
	out attributes jsonb,
	out created_at timestamptz,
	out updated_at timestamptz,
	out relations jsonb
)
  RETURNS SETOF record
  LANGUAGE sql STABLE PARALLEL SAFE AS
$$
   SELECT 
          	id,
          	kind,
          	attributes,
          	created_at,
          	updated_at,
          	muts_relations_tree(id, _conds, _level) as relations -- TODO can be NULL

   FROM muts_records
   WHERE
   -- select by id
   (CASE
   	 WHEN _conds ? 'id' THEN id = _conds->>'id'
   	 WHEN _conds ? 'id_in' THEN id = any(SELECT jsonb_array_elements_text(_conds->'id_in'))
     ELSE TRUE
   END)
   AND
   -- select by kind
   (CASE
   	 WHEN _conds ? 'kind' THEN kind ~ (_conds->>'kind')::lquery
     ELSE TRUE
   END) 
   AND
   -- select by attribute
   (CASE
   	 WHEN _conds ? 'attr' THEN attributes @? (_conds->>'attr')::jsonpath
     ELSE TRUE
   END)
$$;
-- +goose StatementEnd

-- +goose Down

DROP FUNCTION muts_mutate;
DROP FUNCTION muts_relations_tree;
DROP FUNCTION muts_select;
