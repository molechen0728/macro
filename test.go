package main

import (
	"gss"
	"os"
)

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

func Foo() error {

	gss.Ox40(gss.FuncMetaData{
		Name:          "foo",
		Requires:      []string{"image", "document", "video"},
		Hidden:        true,
		DelayInmillis: 1000,
	})

	var e gss.IfErrorNotNilReturn

	_, e = os.ReadFile("")

	e = os.Chdir("")
	_, e = os.ReadFile("")

	e = os.Chdir("")

	_, e = os.ReadFile("")

	e = os.Chdir("")
	_, e = os.ReadFile("")

	e = os.Chdir("")
	_ = e
	return nil
}

func Foo2() error {

	gss.Ox40(gss.FuncMetaData{
		Name:          "foo",
		Requires:      []string{"image", "document", "video"},
		Hidden:        true,
		DelayInmillis: 1000,
	})

	var eeeeeeeeee gss.IfErrorNotNilReturn

	_, eeeeeeeeee = os.ReadFile("")

	eeeeeeeeee = os.Chdir("")
	_, eeeeeeeeee = os.ReadFile("")

	eeeeeeeeee = os.Chdir("")

	_, eeeeeeeeee = os.ReadFile("")

	eeeeeeeeee = os.Chdir("")
	_, eeeeeeeeee = os.ReadFile("")

	eeeeeeeeee = os.Chdir("")
	_ = eeeeeeeeee
	return nil
}
