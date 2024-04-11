package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	pool *pgxpool.Pool
	// selectTmpl   *fasttemplate.Template
	// relFieldTmpl *fasttemplate.Template
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
		pool: pool,
		// selectTmpl:   fasttemplate.New(qSelect, "{{", "}}"),
		// relFieldTmpl: fasttemplate.New(qRelField, "{{", "}}"),
	}, nil
}

//	func (s *Store) Rec() Builder {
//		return Builder{store: s}
//	}

type Query struct {
	Limit  int      `json:"-"`
	Offset int      `json:"-"`
	ID     string   `json:"id,omitempty"`
	IDIn   []string `json:"id_in,omitempty"`
	Kind   string   `json:"kind,omitempty"`
	Attr   string   `json:"attr,omitempty"`
	Follow string   `json:"follow,omitempty"`
}

type OneOpts struct {
	ID     string `json:"id,omitempty"`
	Kind   string `json:"kind,omitempty"`
	Attr   string `json:"attr,omitempty"`
	Follow string `json:"follow,omitempty"`
}

func (s *Store) One(ctx context.Context, opts OneOpts) (*Record, error) {
	o, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}
	row := s.pool.QueryRow(ctx, "select * from muts_select($1) limit 1", o)
	var rec Record
	err = row.Scan(
		&rec.ID,
		&rec.Kind,
		&rec.Attributes,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&rec.Relations,
	)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Store) Many(ctx context.Context, query Query) ([]*Record, error) {
	q, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `select * from muts_select($1) limit $2 offset $3`, q, query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (*Record, error) {
		rec := Record{}
		err := row.Scan(
			&rec.ID,
			&rec.Kind,
			&rec.Attributes,
			&rec.CreatedAt,
			&rec.UpdatedAt,
			&rec.Relations,
		)
		return &rec, err
	})
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
		pgx.Identifier{"muts_mutations"},
		[]string{"record_id", "author", "reason", "ops"},
		pgx.CopyFromRows(rows),
	)

	return err
}
