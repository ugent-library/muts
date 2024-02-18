package store

import "github.com/oklog/ulid/v2"

type Mut struct {
	RecordID string
	Author   string
	Reason   string
	Ops      []Op
}

// TODO make ops opaque
type Op struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
}

func AddRec(kind string, attrs map[string]any) Op {
	args := map[string]any{"kind": kind}
	if attrs != nil {
		args["attrs"] = attrs
	}
	return Op{Name: "add-rec", Args: args}
}

func SetAttr(key string, val any) Op {
	return Op{Name: "set-attr", Args: map[string]any{"key": key, "val": val}}
}

func DelAttr(key string) Op {
	return Op{Name: "del-attr", Args: map[string]any{"key": key}}
}

func ClearAttrs() Op {
	return Op{Name: "clear-attrs"}
}

func AddRel(kind, to string, attrs map[string]any) Op {
	args := map[string]any{"id": ulid.Make().String(), "kind": kind, "to": to}
	if attrs != nil {
		args["attrs"] = attrs
	}
	return Op{Name: "add-rel", Args: args}
}

func DelRel(id string) Op {
	return Op{Name: "del-rel", Args: map[string]any{"id": id}}
}
