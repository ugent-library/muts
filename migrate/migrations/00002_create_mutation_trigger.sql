-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION muts_apply_mutation() 
RETURNS TRIGGER 
LANGUAGE PLPGSQL
AS $$
DECLARE
	op jsonb;
BEGIN
  FOR op IN SELECT * FROM jsonb_array_elements(NEW.ops)
  LOOP
    CASE op->>'name'
 
    WHEN 'add_rec' THEN
      INSERT INTO muts_records (id, kind, attributes, created_at, updated_at)
      VALUES (
        NEW.record_id,
        (op->'args'->>'kind')::ltree,
        coalesce(op->'args'->'attrs', '{}'::jsonb),
        NEW.created_at,
        NEW.created_at
      );

    WHEN 'set_attr' THEN
      UPDATE muts_records
      SET attributes = attributes || jsonb_build_object(op->'args'->>'key', op->'args'->'val'), updated_at = NEW.created_at
      WHERE id = NEW.record_id;

    WHEN 'del_attr' THEN
      UPDATE muts_records
      SET attributes = attributes - op->'args'->>'key', updated_at = NEW.created_at
      WHERE id = NEW.record_id;

    WHEN 'clear_attrs' THEN
      UPDATE muts_records
      SET attributes = '{}'::jsonb, updated_at = NEW.created_at
      WHERE id = NEW.record_id;

    WHEN 'add_rel' THEN
      INSERT INTO muts_relations (id, kind, from_id, to_id, position, attributes, created_at, updated_at)
      VALUES (
        op->'args'->>'id',
        (op->'args'->>'kind')::ltree,
        NEW.record_id,
        op->'args'->>'to',
        (SELECT COUNT(*) FROM muts_relations WHERE from_id = NEW.record_id AND kind = (op->'args'->>'kind')::ltree),
        coalesce(op->'args'->'attrs', '{}'::jsonb),
        NEW.created_at,
        NEW.created_at
      );

      UPDATE muts_records
      SET updated_at = NEW.created_at
      WHERE id = NEW.record_id;

    WHEN 'set_rel_attr' THEN
      UPDATE muts_relations
      SET attributes = attributes || jsonb_build_object(op->'args'->>'key', op->'args'->'val'), updated_at = NEW.created_at
      WHERE id = op->'args'->>'id';

      UPDATE muts_records
      SET updated_at = NEW.created_at
      WHERE id = NEW.record_id;
  
    WHEN 'del_rel_attr' THEN
      UPDATE muts_relations
      SET attributes = attributes - op->'args'->>'key', updated_at = NEW.created_at
      WHERE id = op->'args'->>'id';

      UPDATE muts_records
      SET updated_at = NEW.created_at
      WHERE id = NEW.record_id;

    WHEN 'clear_rel_attrs' THEN
      UPDATE muts_relations
      SET attributes = '{}'::jsonb, updated_at = NEW.created_at
      WHERE id = op->'args'->>'id';

      UPDATE muts_records
      SET updated_at = NEW.created_at
      WHERE id = NEW.record_id;

    WHEN 'del_rel' THEN
      WITH rel AS (
        DELETE FROM muts_relations
        WHERE id = op->'args'->>'id'
        RETURNING kind, position
      )
      UPDATE muts_relations r
      SET position = r.position - 1
      FROM rel
      WHERE from_id = NEW.record_id AND r.kind = rel.kind AND r.position > rel.position;

      UPDATE muts_records
      SET updated_at = NEW.created_at
      WHERE id = NEW.record_id;

    ELSE
      RAISE EXCEPTION 'Unknown operation "%"', op->>'name';

    END CASE;
  END LOOP;
  
  RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER muts_apply_mutation_trigger AFTER INSERT
ON muts_mutations
FOR EACH ROW
EXECUTE PROCEDURE muts_apply_mutation();

-- +goose Down

DROP TRIGGER muts_apply_mutation_trigger ON muts_mutations;
DROP FUNCTION muts_apply_mutation;
