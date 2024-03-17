package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Builder struct {
	store        *Store
	withRels     bool
	withRelRecs  bool
	idEq         string
	kindEq       string
	kindMatches  string
	hasAttr      []string
	attrContains []map[string]any
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

func (b Builder) Kind(kind string) Builder {
	b.kindEq = kind
	return b
}

func (b Builder) KindMatches(kind string) Builder {
	b.kindMatches = kind
	return b
}

func (b Builder) HasAttr(key string) Builder {
	b.hasAttr = append(b.hasAttr, key)
	return b
}

func (b Builder) AttrContains(key string, val any) Builder {
	b.attrContains = append(b.attrContains, map[string]any{key: val})
	return b
}

func (b Builder) Get(ctx context.Context, id string) (*Rec, error) {
	b.idEq = id
	q, args, err := b.buildQuery()
	if err != nil {
		return nil, err
	}

	var rec Rec
	err = b.store.pool.QueryRow(ctx, q, args...).Scan(
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
	q, args, err := b.buildQuery()
	if err != nil {
		return err
	}

	rows, err := b.store.pool.Query(ctx, q, args...)
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

func (b Builder) buildQuery() (string, []any, error) {
	var args []any
	var preds []string
	var i int

	if b.idEq != "" {
		i++
		preds = append(preds, fmt.Sprintf(`records.id = $%d`, i))
		args = append(args, b.idEq)
	}
	if b.kindEq != "" {
		i++
		preds = append(preds, fmt.Sprintf(`records.kind = $%d`, i))
		args = append(args, b.kindEq)
	}
	if b.kindMatches != "" {
		i++
		preds = append(preds, fmt.Sprintf(`records.kind ~ $%d`, i))
		args = append(args, b.kindMatches)
	}
	for _, pred := range b.hasAttr {
		i++
		preds = append(preds, fmt.Sprintf(`records.attributes ? $%d`, i))
		args = append(args, pred)
	}
	for _, pred := range b.attrContains {
		i++
		preds = append(preds, fmt.Sprintf(`records.attributes @> $%d::jsonb`, i))

		j, err := json.Marshal(pred)
		if err != nil {
			return "", nil, err
		}
		args = append(args, string(j))
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
		vars["relField"] = b.store.relFieldTmpl.ExecuteString(relFieldVars)
	}
	if len(preds) > 0 {
		vars["where"] = ` WHERE ` + strings.Join(preds, ` AND `)
	}
	return b.store.selectTmpl.ExecuteString(vars), args, nil
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
;`
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
