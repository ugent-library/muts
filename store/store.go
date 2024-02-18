package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, conn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, conn)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func NewFromPool(pool *pgxpool.Pool) (*Store, error) {
	return &Store{pool: pool}, nil
}

type Rec struct {
	ID         string          `json:"id"`
	Kind       string          `json:"kind"`
	Attributes json.RawMessage `json:"attributes"`
	Relations  []Rel           `json:"relations"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Rel struct {
	ID         string          `json:"id"`
	Kind       string          `json:"kind"`
	To         string          `json:"to"`
	Attributes json.RawMessage `json:"attributes"`
}

func (s *Store) GetRec(ctx context.Context, id string) (*Rec, error) {
	q := `
	SELECT records.id,
	       records.kind,
		   records.attributes,
	       jsonb_agg(jsonb_build_object('id', r.id, 'kind', r.kind, 'to', r.to_id, 'attributes', r.attributes)) AS relations,
		   records.created_at,
		   records.updated_at
	FROM records
	LEFT JOIN relations r ON r.from_id = records.id
	WHERE records.id = $1
	GROUP BY records.id;
	`

	var rec Rec
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&rec.ID,
		&rec.Kind,
		&rec.Attributes,
		&rec.Relations,
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

// TODO optimistic locking
func (s *Store) Mutate(ctx context.Context, muts ...Mut) error {
	rows := make([][]any, len(muts))

	for i, mut := range muts {
		ops, err := json.Marshal(mut.Ops)
		if err != nil {
			return err
		}
		rows[i] = []any{
			ulid.Make().String(),
			mut.RecordID,
			mut.Author,
			pgtype.Text{Valid: mut.Reason != "", String: mut.Reason},
			ops,
		}
	}

	_, err := s.pool.CopyFrom(
		ctx,
		pgx.Identifier{"mutations"},
		[]string{"id", "record_id", "author", "reason", "ops"},
		pgx.CopyFromRows(rows),
	)

	return err
}
