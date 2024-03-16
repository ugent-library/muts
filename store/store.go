package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyala/fasttemplate"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	pool         *pgxpool.Pool
	selectTmpl   *fasttemplate.Template
	relFieldTmpl *fasttemplate.Template
}

func New(ctx context.Context, conn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, conn)
	if err != nil {
		return nil, err
	}
	return NewFromPool(pool)
}

func NewFromPool(pool *pgxpool.Pool) (*Store, error) {
	return &Store{
		pool:         pool,
		selectTmpl:   fasttemplate.New(qSelect, "{{", "}}"),
		relFieldTmpl: fasttemplate.New(qRelField, "{{", "}}"),
	}, nil
}

type Rec struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	Attrs     json.RawMessage `json:"attrs"`
	Rels      []*Rel          `json:"rels,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type Rel struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	Attrs     json.RawMessage `json:"attrs"`
	RecID     string          `json:"recID"`
	Rec       *Rec            `json:"rec,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func (s *Store) Rec() Builder {
	return Builder{s: s}
}

// TODO optimistic locking
func (s *Store) Mutate(ctx context.Context, muts ...Mut) error {
	rows := make([][]any, len(muts))

	for i, mut := range muts {
		ops, err := json.Marshal(mut.Ops)
		if err != nil {
			return err
		}
		rows[i] = []any{
			mut.RecordID,
			mut.Author,
			pgtype.Text{Valid: mut.Reason != "", String: mut.Reason},
			ops,
		}
	}

	_, err := s.pool.CopyFrom(
		ctx,
		pgx.Identifier{"mutations"},
		[]string{"record_id", "author", "reason", "ops"},
		pgx.CopyFromRows(rows),
	)

	return err
}

type Builder struct {
	s        *Store
	withRels bool
	withRecs bool
}

func (b Builder) WithRels() Builder {
	b.withRels = true
	return b
}

func (b Builder) WithRecs() Builder {
	b.withRels = true
	b.withRecs = true
	return b
}

func (b Builder) Get(ctx context.Context, id string) (*Rec, error) {
	q := b.buildQuery(true)

	var rec Rec
	err := b.s.pool.QueryRow(ctx, q, id).Scan(
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
	q := b.buildQuery(false)

	rows, err := b.s.pool.Query(ctx, q)
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

func (b Builder) buildQuery(get bool) string {
	vars := map[string]any{
		"relField": qNullField,
	}
	if b.withRels {
		relFieldVars := map[string]any{}
		if b.withRecs {
			relFieldVars["recField"] = qRecField
			vars["recJoin"] = qRecJoin
		}
		vars["relJoin"] = qRelJoin
		vars["relField"] = b.s.relFieldTmpl.ExecuteString(relFieldVars)
	}
	if get {
		vars["where"] = qWhereID
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
GROUP BY records.id;
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
const qWhereID = `
WHERE records.id = $1
`
