-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION muts_check_rev(id text, rev uuid)
RETURNS VOID
LANGUAGE PLPGSQL STABLE PARALLEL SAFE AS
$$
  DECLARE
    _rev uuid;
  BEGIN
    SELECT n.rev
    FROM muts_nodes n
    WHERE n.id = muts_check_rev.id
    INTO _rev;

    IF _rev IS NULL THEN
      RAISE EXCEPTION 'muts:notfound';
    END IF;
    IF _rev <> muts_check_rev.rev THEN
      RAISE EXCEPTION 'muts:conflict';
    END IF;
  END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_create_node(id text, kind ltree, properties jsonb = '{}', ts timestamptz = transaction_timestamp())
RETURNS VOID
LANGUAGE SQL VOLATILE PARALLEL SAFE AS
$$
  INSERT INTO muts_nodes (id, kind, properties, rev, created_at, updated_at)
  VALUES (muts_create_node.id,
          muts_create_node.kind,
          muts_create_node.properties,
          gen_random_uuid(),
          muts_create_node.ts,
          muts_create_node.ts);
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_update_node(id text, kind ltree = null, properties jsonb = null, ts timestamptz = transaction_timestamp())
RETURNS VOID
LANGUAGE SQL VOLATILE PARALLEL SAFE AS
$$
  UPDATE muts_nodes
  SET kind = coalesce(muts_update_node.kind, kind),
      properties = coalesce(muts_update_node.properties, properties),
      rev = gen_random_uuid(),
      updated_at = muts_update_node.ts
  WHERE id = muts_update_node.id;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_set_node_property(id text, key text, value jsonb, ts timestamptz = transaction_timestamp())
RETURNS VOID
LANGUAGE SQL VOLATILE PARALLEL SAFE AS
$$
  UPDATE muts_nodes
  SET properties = properties || jsonb_build_object(muts_set_node_property.key, muts_set_node_property.value),
      rev = gen_random_uuid(),
      updated_at = muts_set_node_property.ts
  WHERE id = muts_set_node_property.id;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_delete_node_property(id text, key text, ts timestamptz = transaction_timestamp())
RETURNS VOID
LANGUAGE SQL VOLATILE PARALLEL SAFE AS
$$
  UPDATE muts_nodes
  SET properties = properties - muts_delete_node_property.key,
      rev = gen_random_uuid(),
      updated_at = muts_delete_node_property.ts
  WHERE id = muts_delete_node_property.id;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_delete_node(id text)
RETURNS VOID
LANGUAGE SQL VOLATILE PARALLEL SAFE AS
$$
  DELETE FROM muts_nodes
  WHERE id = muts_delete_node.id;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_create_link(id text, kind ltree, from_id text, to_id text, properties jsonb = '{}', ts timestamptz = transaction_timestamp())
RETURNS VOID
LANGUAGE SQL VOLATILE PARALLEL SAFE AS
$$
  INSERT INTO muts_links (id, kind, from_id, to_id, position, properties, created_at, updated_at)
  VALUES (
    muts_create_link.id,
    muts_create_link.kind,
    muts_create_link.from_id,
    muts_create_link.to_id,
    (SELECT COUNT(*) FROM muts_links WHERE from_id = muts_create_link.from_id AND kind = muts_create_link.kind),
    muts_create_link.properties,
    muts_create_link.ts,
    muts_create_link.ts
  );

  UPDATE muts_nodes
  SET rev = gen_random_uuid(),
      updated_at = muts_create_link.ts
  WHERE id = muts_create_link.from_id;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_update_link(id text, kind ltree = null, properties jsonb = null, ts timestamptz = transaction_timestamp())
RETURNS VOID
LANGUAGE SQL VOLATILE PARALLEL SAFE AS
$$
  WITH link AS (
    UPDATE muts_links
    SET kind = coalesce(muts_update_link.kind, kind),
        properties = coalesce(muts_update_link.properties, properties),
        updated_at = muts_update_link.ts
    WHERE id = muts_update_link.id
    RETURNING from_id
  )
  UPDATE muts_nodes
  SET rev = gen_random_uuid(),
      updated_at = muts_update_link.ts
  FROM link
  WHERE id = link.from_id;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_set_link_property(id text, key text, value jsonb, ts timestamptz = transaction_timestamp())
RETURNS VOID
LANGUAGE SQL VOLATILE PARALLEL SAFE AS
$$
  WITH link AS (
    UPDATE muts_links
    SET properties = properties || jsonb_build_object(muts_set_link_property.key, muts_set_link_property.value),
        updated_at = muts_set_link_property.ts
    WHERE id = muts_set_link_property.id
    RETURNING from_id
  )
  UPDATE muts_nodes
  SET rev = gen_random_uuid(),
      updated_at = muts_set_link_property.ts
  FROM link
  WHERE id = link.from_id;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_delete_link_property(id text, key text, ts timestamptz = transaction_timestamp())
RETURNS VOID
LANGUAGE SQL VOLATILE PARALLEL SAFE AS
$$
  WITH link AS (
    UPDATE muts_links
    SET properties = properties - muts_delete_link_property.key,
        updated_at = muts_delete_link_property.ts
    WHERE id = muts_delete_link_property.id
    RETURNING from_id
  )
  UPDATE muts_nodes
  SET rev = gen_random_uuid(),
      updated_at = muts_delete_link_property.ts
  FROM link
  WHERE id = link.from_id;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION muts_delete_link(id text, ts timestamptz = transaction_timestamp())
RETURNS VOID
LANGUAGE SQL VOLATILE PARALLEL SAFE AS
$$
  WITH link AS (
    DELETE FROM muts_links
    WHERE id = muts_delete_link.id
    RETURNING from_id, kind, position
  ), update_positions AS (
    UPDATE muts_links l
    SET position = l.position - 1
    FROM link
    WHERE l.from_id = link.from_id AND l.kind = link.kind AND l.position > link.position
  )
  UPDATE muts_nodes
  SET rev = gen_random_uuid(),
      updated_at = muts_delete_link.ts
  FROM link
  WHERE id = link.from_id;
$$;
-- +goose StatementEnd

-- -- +goose StatementBegin
-- CREATE FUNCTION muts_mutate(
--   _node_id text,
--   _ops jsonb,
--   _ts timestamptz = transaction_timestamp()
-- )
-- RETURNS VOID
-- LANGUAGE PLPGSQL
-- AS $$
-- DECLARE
-- 	op jsonb;
-- BEGIN
--   FOR op IN SELECT * FROM jsonb_array_elements(_ops)
--   LOOP
--     CASE op->>'name'
 
--     WHEN 'add_rec' THEN
--       INSERT INTO muts_nodes (id, kind, properties, created_at, updated_at)
--       VALUES (
--         _node_id,
--         (op->'args'->>'kind')::ltree,
--         coalesce(op->'args'->'attrs', '{}'::jsonb),
--         _ts,
--         _ts
--       );

--     WHEN 'set_attr' THEN
--       UPDATE muts_nodes
--       SET properties = properties || jsonb_build_object(op->'args'->>'key', op->'args'->'val'), updated_at = _ts
--       WHERE id = _node_id;

--     WHEN 'del_attr' THEN
--       UPDATE muts_nodes
--       SET properties = properties - op->'args'->>'key', updated_at = _ts
--       WHERE id = _node_id;

--     WHEN 'clear_attrs' THEN
--       UPDATE muts_nodes
--       SET properties = '{}'::jsonb, updated_at = _ts
--       WHERE id = _node_id;

--     WHEN 'add_rel' THEN
--       INSERT INTO muts_links (id, kind, from_id, to_id, position, properties, created_at, updated_at)
--       VALUES (
--         op->'args'->>'id',
--         (op->'args'->>'kind')::ltree,
--         _node_id,
--         op->'args'->>'to',
--         (SELECT COUNT(*) FROM muts_links WHERE from_id = _node_id AND kind = (op->'args'->>'kind')::ltree),
--         coalesce(op->'args'->'attrs', '{}'::jsonb),
--         _ts,
--         _ts
--       );

--       UPDATE muts_nodes
--       SET updated_at = _ts
--       WHERE id = _node_id;

--     WHEN 'set_rel_attr' THEN
--       UPDATE muts_links
--       SET properties = properties || jsonb_build_object(op->'args'->>'key', op->'args'->'val'), updated_at = _ts
--       WHERE id = op->'args'->>'id';

--       UPDATE muts_nodes
--       SET updated_at = _ts
--       WHERE id = _node_id;
  
--     WHEN 'del_rel_attr' THEN
--       UPDATE muts_links
--       SET properties = properties - op->'args'->>'key', updated_at = _ts
--       WHERE id = op->'args'->>'id';

--       UPDATE muts_nodes
--       SET updated_at = _ts
--       WHERE id = _node_id;

--     WHEN 'clear_rel_attrs' THEN
--       UPDATE muts_links
--       SET properties = '{}'::jsonb, updated_at = _ts
--       WHERE id = op->'args'->>'id';

--       UPDATE muts_nodes
--       SET updated_at = _ts
--       WHERE id = _node_id;

--     WHEN 'del_rel' THEN
--       WITH rel AS (
--         DELETE FROM muts_links
--         WHERE id = op->'args'->>'id'
--         RETURNING kind, position
--       )
--       UPDATE muts_links r
--       SET position = r.position - 1
--       FROM rel
--       WHERE from_id = _node_id AND r.kind = rel.kind AND r.position > rel.position;

--       UPDATE muts_nodes
--       SET updated_at = _ts
--       WHERE id = _node_id;

--     ELSE
--       RAISE EXCEPTION 'Unknown operation "%"', op->>'name';

--     END CASE;
--   END LOOP;
-- END;
-- $$;
-- -- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION _muts_links_tree(_rec_id text, _conds jsonb = '{}', _level int = 7)
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
          rl.properties,
          rl.position,
          rl.created_at,
          rl.updated_at,
          jsonb_build_object(
          	'id', r.id,
          	'kind', r.kind,
          	'properties', r.properties,
          	'created_at', r.created_at,
          	'updated_at', r.updated_at,
          	'links', (CASE WHEN _level > 1 THEN _muts_links_tree(r.id, _conds, _level - 1) END)
   		  ) AS node
   FROM   muts_links rl
   JOIN   muts_nodes r ON r.id = rl.to_id
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
	out properties jsonb,
	out created_at timestamptz,
	out updated_at timestamptz,
	out links jsonb
)
  RETURNS SETOF record
  LANGUAGE sql STABLE PARALLEL SAFE AS
$$
   SELECT id,
          kind,
          properties,
          created_at,
          updated_at,
          _muts_links_tree(id, _conds, _level) as links -- TODO can be NULL

   FROM muts_nodes
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
   	 WHEN _conds ? 'attr' THEN properties @? (_conds->>'attr')::jsonpath
     ELSE TRUE
   END)
$$;
-- +goose StatementEnd

-- -- +goose StatementBegin
-- CREATE FUNCTION muts_apply_mutation() 
-- RETURNS TRIGGER 
-- LANGUAGE PLPGSQL
-- AS $$
-- BEGIN
--   perform muts_mutate(NEW.node_id, NEW.ops, NEW.created_at);
--   RETURN NEW;
-- END;
-- $$;
-- -- +goose StatementEnd

-- CREATE TRIGGER muts_apply_mutation_trigger AFTER INSERT
-- ON muts_mutations
-- FOR EACH ROW
-- EXECUTE PROCEDURE muts_apply_mutation();

-- +goose Down

DROP FUNCTION muts_create_node;
DROP FUNCTION muts_update_node;
DROP FUNCTION muts_set_node_property;
DROP FUNCTION muts_delete_node_property;
DROP FUNCTION muts_delete_node;

DROP FUNCTION muts_create_link;
DROP FUNCTION muts_update_link;
DROP FUNCTION muts_set_link_property;
DROP FUNCTION muts_delete_link_property;
DROP FUNCTION muts_delete_link;

-- DROP FUNCTION muts_mutate;
DROP FUNCTION _muts_links_tree;
DROP FUNCTION muts_select;
-- DROP TRIGGER muts_apply_mutation_trigger ON muts_mutations;
-- DROP FUNCTION muts_apply_mutation;
