-- +goose Up

-- +goose StatementBegin
CREATE FUNCTION muts_apply_mutation() 
RETURNS TRIGGER 
LANGUAGE PLPGSQL
AS $$
BEGIN
  perform muts_mutate(NEW.record_id, NEW.ops, NEW.created_at);
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
