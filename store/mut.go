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
	Name string `json:"name"`
	Args any    `json:"args,omitempty"`
}

func AddRec(kind string, attrs any) Op {
	return Op{Name: "add-rec", Args: struct {
		Kind  string `json:"kind"`
		Attrs any    `json:"attrs,omitempty"`
	}{
		Kind:  kind,
		Attrs: attrs,
	}}
}

func SetAttr(key string, val any) Op {
	return Op{Name: "set-attr", Args: struct {
		Key string `json:"key"`
		Val any    `json:"val"`
	}{
		Key: key,
		Val: val,
	}}
}

func DelAttr(key string) Op {
	return Op{Name: "del-attr", Args: struct {
		Key string `json:"key"`
	}{
		Key: key,
	}}
}

func ClearAttrs() Op {
	return Op{Name: "clear-attrs"}
}

func AddRel(kind, to string, attrs any) Op {
	return Op{Name: "add-rel", Args: struct {
		ID    string `json:"id"`
		Kind  string `json:"kind"`
		To    string `json:"to"`
		Attrs any    `json:"attrs,omitempty"`
	}{
		ID:    ulid.Make().String(),
		Kind:  kind,
		To:    to,
		Attrs: attrs,
	}}
}

func DelRel(id string) Op {
	return Op{Name: "del-rel", Args: struct {
		ID string `json:"id"`
	}{
		ID: id,
	}}
}
