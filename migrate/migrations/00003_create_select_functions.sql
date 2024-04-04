-- +goose Up

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
$func$
   SELECT 
          	id,
          	kind,
          	attributes,
          	created_at,
          	updated_at,
          	muts_relations_tree(id, _conds) as relations -- TODO can be NULL
   		  
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
$func$;
-- +goose StatementEnd

-- +goose Down

DROP FUNCTION muts_relations_tree;
DROP FUNCTION muts_select;
