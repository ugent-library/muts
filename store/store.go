package store

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
)

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

// TODO optimistic locking
func (s *Store) Mutate(ctx context.Context, muts ...Mut) error {
	rows := make([][]any, 0, len(muts))
	for _, mut := range muts {
		ops, err := json.Marshal(mut.Ops)
		if err != nil {
			return err
		}
		rows = append(rows, []any{ulid.Make().String(), mut.Author, mut.Reason, ops})
	}

	_, err := s.pool.CopyFrom(
		ctx,
		pgx.Identifier{"mutations"},
		[]string{"id", "author", "reason", "ops"},
		pgx.CopyFromRows(rows),
	)

	return err
}
