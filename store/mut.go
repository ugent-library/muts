package store

type Mut struct {
	RecordID string
	Author   string
	Reason   string
	Ops      []Op
}

type Op struct {
	Name string `json:"name"`
	Args Args   `json:"args,omitempty"`
}

type Args = map[string]string

func AddRec(recType string) Op {
	return Op{Name: "add-rec", Args: Args{"type": recType}}
}

func AddAttr(name, value string) Op {
	return Op{Name: "add-attr", Args: Args{"name": name, "value": value}}
}

func DelAttrs() Op {
	return Op{Name: "del-attrs"}
}

func AddRel(name, to string) Op {
	return Op{Name: "add-rel", Args: Args{"name": name, "to": to}}
}

func DelRel(name, to string) Op {
	return Op{Name: "del-rel", Args: Args{"name": name, "to": to}}
}
