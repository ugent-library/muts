package store

type Mut struct {
	Author string
	Reason string
	Ops    []Op
}

type Op struct {
	ID   string            `json:"id"`
	Name string            `json:"name"`
	Args map[string]string `json:"args,omitempty"`
}

func AddRec(id, recType string) Op {
	return Op{ID: id, Name: "add-rec", Args: map[string]string{"type": recType}}
}

func AddAttr(id, name, value string) Op {
	return Op{ID: id, Name: "add-attr", Args: map[string]string{"name": name, "value": value}}
}

func DelAttrs(id string) Op {
	return Op{ID: id, Name: "del-attrs"}
}

func AddRel(id, name, to string) Op {
	return Op{ID: id, Name: "add-rel", Args: map[string]string{"name": name, "to": to}}
}

func DelRel(id, name, to string) Op {
	return Op{ID: id, Name: "del-rel", Args: map[string]string{"name": name, "to": to}}
}
