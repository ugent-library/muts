package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Builder struct {
	s           *Store
	withRels    bool
	withRelRecs bool
	whereID     string
	whereAttr   []map[string]any
}

func (b Builder) WithRels() Builder {
	b.withRels = true
	return b
}

func (b Builder) WithRelRecs() Builder {
	b.withRels = true
	b.withRelRecs = true
	return b
}

func (b Builder) WhereAttr(key string, val any) Builder {
	b.whereAttr = append(b.whereAttr, map[string]any{key: val})
	return b
}

func (b Builder) Get(ctx context.Context, id string) (*Rec, error) {
	b.whereID = id
	q := b.buildQuery()
	args, err := b.buildArgs()
	if err != nil {
		return nil, err
	}

	var rec Rec
	err = b.s.pool.QueryRow(ctx, q, args...).Scan(
		&rec.ID,
		&rec.Kind,
		&rec.Attrs,
		&rec.Rels,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (b Builder) Each(ctx context.Context, fn func(*Rec) bool) error {
	q := b.buildQuery()
	args, err := b.buildArgs()
	if err != nil {
		return err
	}

	rows, err := b.s.pool.Query(ctx, q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var rec Rec
		err := rows.Scan(
			&rec.ID,
			&rec.Kind,
			&rec.Attrs,
			&rec.Rels,
			&rec.CreatedAt,
			&rec.UpdatedAt,
		)
		if err != nil {
			return err
		}
		if !fn(&rec) {
			break
		}
	}

	return nil
}

func (b Builder) buildArgs() ([]any, error) {
	var args []any
	if b.whereID != "" {
		args = append(args, b.whereID)
	}
	for _, pred := range b.whereAttr {
		j, err := json.Marshal(pred)
		if err != nil {
			return nil, err
		}
		args = append(args, string(j))
	}
	return args, nil
}

func (b Builder) buildQuery() string {
	var preds []string
	var i int
	if b.whereID != "" {
		i++
		preds = append(preds, fmt.Sprintf(`records.id = $%d`, i))
	}
	for range b.whereAttr {
		i++
		preds = append(preds, fmt.Sprintf(`records.attributes @> $%d::jsonb`, i))
	}
	vars := map[string]any{
		"relField": qNullField,
	}
	if b.withRels {
		relFieldVars := map[string]any{}
		if b.withRelRecs {
			relFieldVars["recField"] = qRecField
			vars["recJoin"] = qRecJoin
		}
		vars["relJoin"] = qRelJoin
		vars["relGroup"] = qRelGroup
		vars["relField"] = b.s.relFieldTmpl.ExecuteString(relFieldVars)
	}
	if len(preds) > 0 {
		vars["where"] = ` WHERE ` + strings.Join(preds, ` AND `)
	}
	return b.s.selectTmpl.ExecuteString(vars)
}

const qSelect = `
SELECT records.id,
	records.kind,
	records.attributes,
	{{relField}}
	records.created_at,
	records.updated_at
FROM records
{{relJoin}}
{{recJoin}}
{{where}}
{{relGroup}}
;
`
const qNullField = `
	NULL,
`
const qRelField = `
jsonb_agg(jsonb_build_object(
	'id', rels.id,
	'kind', rels.kind,
	'recID', rels.to_id,
	{{recField}}
	'attrs', rels.attributes,
	'createdAt', rels.created_at,
	'updatedAt', rels.updated_at
)),
`
const qRelJoin = `
LEFT JOIN relations rels ON rels.from_id = records.id

`
const qRelGroup = `
GROUP BY records.id;
`
const qRecField = `
'rec', jsonb_build_object(
	'id', recs.id,
	'kind', recs.kind,
	'attrs', recs.attributes,
	'createdAt', recs.created_at,
	'updatedAt', recs.updated_at
),
`
const qRecJoin = `
INNER JOIN records recs ON rels.to_id = recs.id
`
