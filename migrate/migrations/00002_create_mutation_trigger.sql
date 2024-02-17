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
      INSERT INTO records (id, type) VALUES (NEW.record_id, (op->'args'->>'type')::ltree);

    WHEN 'add-attr' THEN
      INSERT INTO attributes (record_id, name, value) VALUES (NEW.record_id, (op->'args'->>'name')::ltree, op->'args'->>'value');

    WHEN 'del-attrs' THEN
      DELETE FROM attributes WHERE record_id = NEW.record_id;

    WHEN 'add-rel' THEN
      INSERT INTO relations (from_id, to_id, name, position)
      VALUES (
        NEW.record_id,
        op->'args'->>'to',
        op->'args'->>'name',
        (SELECT COUNT(*) FROM relations WHERE from_id = NEW.record_id AND name = op->'args'->>'name')
      );

    WHEN 'del-rel' THEN
      WITH del_rel AS (
        DELETE FROM relations
        WHERE from_id = NEW.record_id AND to_id = op->'args'->>'to' AND name = op->'args'->>'name'
        RETURNING position
      )
      UPDATE relations r
      SET position = r.position - 1
      FROM del_rel
      WHERE from_id = NEW.record_id AND name = op->'args'->>'name' AND r.position > del_rel.position;
  
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