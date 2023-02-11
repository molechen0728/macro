package main

import "gss"

type S struct {
	Cpu string
	Ram string
}

func (s S) Ox24() {
	gss.Ox24(map[string]gss.StructMetaData{
		"Cpu": {Name: "cpu", No: 1, Mutable: true},
		"Ram": {Name: "Ram", No: 2, Mutable: false},
	})
}

func Foo() {

	gss.Ox40(gss.FuncMetaData{
		Name:          "foo",
		Requires:      []string{"image", "document", "video"},
		Hidden:        true,
		DelayInmillis: 1000,
	})
}
