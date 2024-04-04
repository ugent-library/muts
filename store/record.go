package store

import (
	"encoding/json"
	"time"
)

type Record struct {
	ID         string          `json:"id"`
	Kind       string          `json:"kind"`
	Attributes json.RawMessage `json:"attributes"`
	Relations  []*Relation     `json:"relations,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Relation struct {
	ID         string          `json:"id"`
	Kind       string          `json:"kind"`
	Attributes json.RawMessage `json:"attributes"`
	Record     *Record         `json:"record,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
