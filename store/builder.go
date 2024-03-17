package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Builder struct {
	store          *Store
	withRels       bool
	withRelRecs    bool
	idEq           string
	kindEq         string
	kindMatches    string
	hasAttr        []string
	attrContains   []map[string]any
	relKindEq      string
	relKindMatches string
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

// TOOD should be in own Rel() builder so we can test each rel on multiple conditions
func (b Builder) RelKind(kind string) Builder {
	b.relKindEq = kind
	return b
}

// TOOD should be in own Rel() builder so we can test each rel on multiple conditions
func (b Builder) RelKindMatches(kind string) Builder {
	b.relKindMatches = kind
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

	log.Printf("query: %s", q)

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
	var relPreds []string
	var i int

	if b.idEq != "" {
		i++
		preds = append(preds, fmt.Sprintf(`r.id = $%d`, i))
		args = append(args, b.idEq)
	}
	if b.kindEq != "" {
		i++
		preds = append(preds, fmt.Sprintf(`r.kind = $%d`, i))
		args = append(args, b.kindEq)
	}
	if b.kindMatches != "" {
		i++
		preds = append(preds, fmt.Sprintf(`r.kind ~ $%d`, i))
		args = append(args, b.kindMatches)
	}
	for _, pred := range b.hasAttr {
		i++
		preds = append(preds, fmt.Sprintf(`r.attributes ? $%d`, i))
		args = append(args, pred)
	}
	for _, pred := range b.attrContains {
		i++
		preds = append(preds, fmt.Sprintf(`r.attributes @> $%d::jsonb`, i))

		j, err := json.Marshal(pred)
		if err != nil {
			return "", nil, err
		}
		args = append(args, string(j))
	}

	if b.relKindEq != "" {
		i++
		relPreds = append(relPreds, fmt.Sprintf(`rl.kind = $%d`, i))
		args = append(args, b.relKindEq)
	}
	if b.relKindMatches != "" {
		i++
		relPreds = append(relPreds, fmt.Sprintf(`rl.kind ~ $%d`, i))
		args = append(args, b.relKindMatches)
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
	if len(relPreds) > 0 {
		vars["relWhere"] = `JOIN relations rl ON rl.from_id = r.id AND ` + strings.Join(relPreds, ` AND `)
	}
	if len(preds) > 0 {
		vars["where"] = ` WHERE ` + strings.Join(preds, ` AND `)
	}

	return b.store.selectTmpl.ExecuteString(vars), args, nil
}

const qSelect = `
SELECT r.id,
	  r.kind,
	  r.attributes,
	  {{relField}}
	  r.created_at,
	  r.updated_at
FROM records r
{{relWhere}}
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
LEFT JOIN relations rels ON rels.from_id = r.id

`
const qRelGroup = `
GROUP BY r.id;
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
