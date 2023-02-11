package macro

import (
	"bytes"
	"fmt"
	"go/ast"
	"gss"
	"reflect"
	"strconv"
	"strings"
)

type Node struct {
	PkgName  string
	Pos, End int
	gss.FuncMetaData
	Smd       map[string]gss.StructMetaData
	StructMap map[string]map[string]struct{}
	noAlign   bool
}

func (n *Node) Fast(f *ast.File) {
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok {
			for _, s := range gd.Specs {
				switch t := s.(type) {
				case *ast.ImportSpec:
					n.parseImport(t)
				case *ast.TypeSpec:
					if st, ok := t.Type.(*ast.StructType); ok {
						m := map[string]struct{}{}
						for _, f := range st.Fields.List {
							for _, i2 := range f.Names {
								m[i2.Name] = struct{}{}
							}
						}
						n.StructMap[t.Name.Name] = m
					}
				}
			}
		}

		if fd, ok := d.(*ast.FuncDecl); ok {
			if fd.Recv == nil {
				n.parseFunDecl(fd)
				break
			}

			f := fd.Recv.List[0]
			if m, ok := n.StructMap[f.Type.(*ast.Ident).Name]; ok {
				n.parseMethodDecl(fd, m)
			}
		}
	}
}

func (n *Node) Visit(node ast.Node) ast.Visitor {

	switch t := node.(type) {
	case *ast.File:
		n.PkgName = t.Name.Name
		n.Fast(t)
		// case *ast.ImportSpec:
		// 	n.parseImport(t)
		// case *ast.TypeSpec:
		// 	if st, ok := t.Type.(*ast.StructType); ok {
		// 		m := map[string]struct{}{}
		// 		for _, f := range st.Fields.List {
		// 			for _, i2 := range f.Names {
		// 				m[i2.Name] = struct{}{}
		// 			}
		// 		}
		// 		n.StructMap[t.Name.Name] = m
		// 	}
		// case *ast.FuncDecl:
		// 	if t.Recv == nil {
		// 		n.parseFunDecl(t)
		// 		break
		// 	}

		// 	f := t.Recv.List[0]
		// 	if m, ok := n.StructMap[f.Type.(*ast.Ident).Name]; ok {
		// 		n.parseMethodDecl(t, m)
		// 	}
	}

	return n
}

func (n *Node) parseImport(node *ast.ImportSpec) {
	if node.Path.Value == `"gss"` {
		n.noAlign = true
	}
}

func (n *Node) parseMethodDecl(node *ast.FuncDecl, m map[string]struct{}) {
	s := node.Body.List[0]
	if es, ok := s.(*ast.ExprStmt); ok {
		if ce, ok := es.X.(*ast.CallExpr); ok {
			if se, ok := ce.Fun.(*ast.SelectorExpr); ok {
				i, ok := se.X.(*ast.Ident)
				if !ok || i.Name != "gss" {
					return
				}

				if se.Sel.Name != "Ox24" {
					return
				}

				cl, ok := ce.Args[0].(*ast.CompositeLit)
				if !ok {
					return
				}

				if mt, ok := cl.Type.(*ast.MapType); !ok || mt.Key.(*ast.Ident).Name != "string" ||
					mt.Value.(*ast.SelectorExpr).X.(*ast.Ident).Name != "gss" ||
					mt.Value.(*ast.SelectorExpr).Sel.Name != "StructMetaData" {
					return
				}

				for _, e := range cl.Elts {
					kv := e.(*ast.KeyValueExpr)
					k := kv.Key.(*ast.BasicLit).Value
					k = strings.Trim(k, "\"")
					if _, ok := m[k]; !ok {
						continue
					}

					smd := gss.StructMetaData{}
					for _, e2 := range kv.Value.(*ast.CompositeLit).Elts {
						kv2 := e2.(*ast.KeyValueExpr)
						switch kv2.Key.(*ast.Ident).Name {
						case "Name":
							s := strings.Trim(kv2.Value.(*ast.BasicLit).Value, "\"")
							smd.Name = s
						case "No":
							i, _ := strconv.ParseInt(kv2.Value.(*ast.BasicLit).Value, 10, 64)
							smd.No = int(i)
						case "Mutable":
							if kv2.Value.(*ast.Ident).Name == "true" {
								smd.Mutable = true
							} else {
								smd.Mutable = false
							}
						}
					}
					n.Smd[k] = smd
				}
			}
		}
	}
}

func (n *Node) parseFunDecl(node *ast.FuncDecl) {

	for _, s := range node.Body.List {
		if es, ok := s.(*ast.ExprStmt); ok {
			if ce, ok := es.X.(*ast.CallExpr); ok {
				if se, ok := ce.Fun.(*ast.SelectorExpr); ok {
					i, ok := se.X.(*ast.Ident)
					if !ok || i.Name != "gss" {
						return
					}
					if se.Sel.Name != "Ox40" {
						return
					}

					n.Pos = int(es.Pos())
					n.End = int(es.End())

					if len(ce.Args) != 1 {
						return
					}

					cl, ok := ce.Args[0].(*ast.CompositeLit)
					if !ok {
						return
					}
					if se, ok := cl.Type.(*ast.SelectorExpr); !ok || se.Sel.Name != "FuncMetaData" || se.X.(*ast.Ident).Name != "gss" {
						return
					}

					for _, e := range cl.Elts {
						kv := e.(*ast.KeyValueExpr)
						n.kv(kv.Key, kv.Value)
					}
				}
			}
		}
	}
}

func (n *Node) kv(k, v ast.Expr) {

	name := k.(*ast.Ident).Name
	if name == "Name" {
		n.Name = v.(*ast.BasicLit).Value
	}

	if name == "Requires" {
		if v.(*ast.CompositeLit).Type.(*ast.ArrayType).Elt.(*ast.Ident).Name != "string" {
			return
		}
		for _, e := range v.(*ast.CompositeLit).Elts {
			if n.Requires == nil {
				n.Requires = make([]string, 0)
			}
			n.Requires = append(n.Requires, e.(*ast.BasicLit).Value)
		}
	}

	if name == "Hidden" {
		if v.(*ast.Ident).String() == "true" {
			n.Hidden = true
		} else {
			n.Hidden = false
		}
	}

	if name == "DelayInmillis" {
		i, _ := strconv.ParseInt(v.(*ast.BasicLit).Value, 10, 64)
		n.DelayInmillis = int(i)
	}
}

func (n Node) String() string {

	var buf bytes.Buffer
	for k, smd := range n.Smd {
		buf.WriteString(fmt.Sprintf("\n     %s: {\n               Name: %s, \n               No: %d, \n               Mutable: %v\n      }", k, smd.Name, smd.No, smd.Mutable))
	}

	return fmt.Sprintf("PkgName: %s\n"+
		"Pos, End: %d, %d\n"+
		"Name: %s\n"+
		"Requires: %v\n"+
		"Hidden: %v\n"+
		"DelayInmillis: %d\n"+
		"Smd: %v",
		n.PkgName, n.Pos, n.End, n.Name, n.Requires, n.Hidden, n.DelayInmillis, buf.String(),
	)
}

func tof(i any) {
	fmt.Println(reflect.TypeOf(i))
}

func fp(i ...any) {
	fmt.Println(i...)
}
