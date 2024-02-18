-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION apply_mutation() 
RETURNS TRIGGER 
LANGUAGE PLPGSQL
AS $$
DECLARE
	op jsonb;
BEGIN
  FOR op IN SELECT * FROM jsonb_array_elements(NEW.ops)
  LOOP
    CASE op->>'name'
 
    WHEN 'add-rec' THEN
      INSERT INTO records (id, kind, attributes)
      VALUES (
        NEW.record_id,
        (op->'args'->>'kind')::ltree,
        coalesce(op->'args'->'attrs', '{}'::jsonb)
      );

    WHEN 'set-attr' THEN
      UPDATE records
      SET attributes = attributes || jsonb_build_object(op->'args'->>'key', op->'args'->'val')
      WHERE id = NEW.record_id;

    WHEN 'del-attr' THEN
      UPDATE records
      SET attributes = attributes - op->'args'->>'key'
      WHERE id = NEW.record_id;

    WHEN 'clear-attrs' THEN
      UPDATE records
      SET attributes = '{}'::jsonb
      WHERE id = NEW.record_id;

    WHEN 'add-rel' THEN
      INSERT INTO relations (id, from_id, to_id, kind, position, attributes)
      VALUES (
        op->'args'->>'id',
        NEW.record_id,
        op->'args'->>'to',
        (op->'args'->>'kind')::ltree,
        (SELECT COUNT(*) FROM relations WHERE from_id = NEW.record_id AND kind = (op->'args'->>'kind')::ltree),
        coalesce(op->'args'->'attrs', '{}'::jsonb)
      );

    WHEN 'del-rel' THEN
      WITH rel AS (
        DELETE FROM relations
        WHERE id = op->'args'->>'id'
        RETURNING kind, position
      )
      UPDATE relations r
      SET position = r.position - 1
      FROM rel
      WHERE from_id = NEW.record_id AND r.kind = rel.kind AND r.position > rel.position;

    ELSE
      RAISE EXCEPTION 'Unknown operation "%"', op->>'name';

    END CASE;
  END LOOP;
  
  RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER apply_mutation_trigger AFTER INSERT
ON mutations
FOR EACH ROW
EXECUTE PROCEDURE apply_mutation();

-- +goose Down

DROP TRIGGER apply_mutation_trigger ON mutations;
DROP FUNCTION apply_mutation;