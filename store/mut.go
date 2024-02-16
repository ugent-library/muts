package store

type Mut struct {
	RecordID string
	Author   string
	Reason   string
	Ops      []Op
}

type Op struct {
	Name string            `json:"name"`
	Args map[string]string `json:"args,omitempty"`
}

func AddRec(recType string) Op {
	return Op{Name: "add-rec", Args: map[string]string{"type": recType}}
}

func AddAttr(name, value string) Op {
	return Op{Name: "add-attr", Args: map[string]string{"name": name, "value": value}}
}

func DelAttrs() Op {
	return Op{Name: "del-attrs"}
}

func AddRel(name, to string) Op {
	return Op{Name: "add-rel", Args: map[string]string{"name": name, "to": to}}
}

func DelRel(name, to string) Op {
	return Op{Name: "del-rel", Args: map[string]string{"name": name, "to": to}}
}
