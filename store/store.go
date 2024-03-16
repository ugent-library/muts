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
