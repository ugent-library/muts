package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Builder struct {
	store       *Store
	withRels    bool
	withRelRecs bool
	filters     []string
	relFilters  []string
	args        []any
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
	i := len(b.args) + 1
	b.filters = append(b.filters, fmt.Sprintf(`r.kind = $%d`, i))
	b.args = append(b.args, kind)
	return b
}

func (b Builder) KindMatches(kind string) Builder {
	i := len(b.args) + 1
	b.filters = append(b.filters, fmt.Sprintf(`r.kind ~ $%d`, i))
	b.args = append(b.args, kind)
	return b
}

func (b Builder) HasAttr(key string) Builder {
	i := len(b.args) + 1
	b.filters = append(b.filters, fmt.Sprintf(`r.attributes ? $%d`, i))
	b.args = append(b.args, key)
	return b
}

// TODO handle error
func (b Builder) AttrContains(key string, val any) Builder {
	j, _ := json.Marshal(map[string]any{key: val})
	i := len(b.args) + 1
	b.filters = append(b.filters, fmt.Sprintf(`r.attributes @> $%d::jsonb`, i))
	b.args = append(b.args, string(j))
	return b
}

// TOOD should be in own Rel() builder so we can test each rel on multiple conditions
func (b Builder) RelKind(kind string) Builder {
	i := len(b.args) + 1
	b.relFilters = append(b.relFilters, fmt.Sprintf(`rl.kind = $%d`, i))
	b.args = append(b.args, kind)
	return b
}

// TOOD should be in own Rel() builder so we can test each rel on multiple conditions
func (b Builder) RelKindMatches(kind string) Builder {
	i := len(b.args) + 1
	b.relFilters = append(b.relFilters, fmt.Sprintf(`rl.kind ~ $%d`, i))
	b.args = append(b.args, kind)
	return b
}

// TOOD should be in own Rel() builder so we can test each rel on multiple conditions
func (b Builder) RelHasAttr(key string) Builder {
	i := len(b.args) + 1
	b.relFilters = append(b.relFilters, fmt.Sprintf(`rl.attributes ? $%d`, i))
	b.args = append(b.args, key)
	return b
}

// TODO handle error
// TOOD should be in own Rel() builder so we can test each rel on multiple conditions
func (b Builder) RelAttrContains(key string, val any) Builder {
	j, _ := json.Marshal(map[string]any{key: val})
	i := len(b.args) + 1
	b.relFilters = append(b.relFilters, fmt.Sprintf(`rl.attributes @> $%d::jsonb`, i))
	b.args = append(b.args, string(j))
	return b
}

func (b Builder) ID(id string) Builder {
	i := len(b.args) + 1
	b.filters = append(b.filters, fmt.Sprintf(`r.id = $%d`, i))
	b.args = append(b.args, id)
	return b
}

func (b Builder) One(ctx context.Context) (*Rec, error) {
	var rec Rec
	err := b.store.pool.QueryRow(ctx, b.Query(), b.args...).Scan(
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
	rows, err := b.store.pool.Query(ctx, b.Query(), b.args...)
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

func (b Builder) Query() string {
	vars := map[string]any{
		"relField": qNullField,
	}

	if b.withRels {
		relVars := map[string]any{}
		if b.withRelRecs {
			relVars["recField"] = qRecField
			vars["recJoin"] = qRecJoin
		}
		vars["relJoin"] = qRelJoin
		vars["relGroup"] = qRelGroup
		vars["relField"] = b.store.relFieldTmpl.ExecuteString(relVars)
	}
	if len(b.filters) > 0 {
		vars["where"] = ` WHERE ` + strings.Join(b.filters, ` AND `)
	}
	if len(b.relFilters) > 0 {
		vars["relWhere"] = `JOIN relations rl ON rl.from_id = r.id AND ` + strings.Join(b.relFilters, ` AND `)
	}

	return b.store.selectTmpl.ExecuteString(vars)
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
