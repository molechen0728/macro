package macro

import (
	"bytes"
	"fmt"
	"go/ast"
	"gss"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type AtVisitor struct {
	PkgName  string
	Pos, End int
	gss.FuncMetaData
	Smd       map[string]gss.StructMetaData
	StructMap map[string]map[string]struct{}
	noAlign   bool
}

func (a *AtVisitor) Fast(f *ast.File) {
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok {
			for _, s := range gd.Specs {
				switch t := s.(type) {
				case *ast.ImportSpec:
					a.parseImport(t)
				case *ast.TypeSpec:
					if st, ok := t.Type.(*ast.StructType); ok {
						m := map[string]struct{}{}
						for _, f := range st.Fields.List {
							for _, i2 := range f.Names {
								m[i2.Name] = struct{}{}
							}
						}
						a.StructMap[t.Name.Name] = m
					}
				}
			}
		}

		if fd, ok := d.(*ast.FuncDecl); ok {
			if fd.Recv == nil {
				a.parseFunDecl(fd)
				break
			}

			f := fd.Recv.List[0]
			if m, ok := a.StructMap[f.Type.(*ast.Ident).Name]; ok {
				a.parseMethodDecl(fd, m)
			}
		}
	}
}

func (a *AtVisitor) Visit(node ast.Node) ast.Visitor {

	switch t := node.(type) {
	case *ast.File:
		a.PkgName = t.Name.Name
		a.Fast(t)
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

	return a
}

func (a *AtVisitor) parseImport(node *ast.ImportSpec) {
	if node.Path.Value == `"gss"` {
		a.noAlign = true
	}
}

func (a *AtVisitor) parseMethodDecl(node *ast.FuncDecl, m map[string]struct{}) {
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
					a.Smd[k] = smd
				}
			}
		}
	}
}

func (a *AtVisitor) parseFunDecl(node *ast.FuncDecl) {

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

					a.Pos = int(es.Pos())
					a.End = int(es.End())

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
						a.kv(kv.Key, kv.Value)
					}
				}
			}
		}
	}
}

func (a *AtVisitor) kv(k, v ast.Expr) {

	name := k.(*ast.Ident).Name
	if name == "Name" {
		a.Name = v.(*ast.BasicLit).Value
	}

	if name == "Requires" {
		if v.(*ast.CompositeLit).Type.(*ast.ArrayType).Elt.(*ast.Ident).Name != "string" {
			return
		}
		for _, e := range v.(*ast.CompositeLit).Elts {
			if a.Requires == nil {
				a.Requires = make([]string, 0)
			}
			a.Requires = append(a.Requires, e.(*ast.BasicLit).Value)
		}
	}

	if name == "Hidden" {
		if v.(*ast.Ident).String() == "true" {
			a.Hidden = true
		} else {
			a.Hidden = false
		}
	}

	if name == "DelayInmillis" {
		i, _ := strconv.ParseInt(v.(*ast.BasicLit).Value, 10, 64)
		a.DelayInmillis = int(i)
	}
}

func (a AtVisitor) String() string {

	var buf bytes.Buffer
	for k, smd := range a.Smd {
		buf.WriteString(fmt.Sprintf("\n     %s: {\n               Name: %s, \n               No: %d, \n               Mutable: %v\n      }", k, smd.Name, smd.No, smd.Mutable))
	}

	return fmt.Sprintf("PkgName: %s\n"+
		"Pos, End: %d, %d\n"+
		"Name: %s\n"+
		"Requires: %v\n"+
		"Hidden: %v\n"+
		"DelayInmillis: %d\n"+
		"Smd: %v",
		a.PkgName, a.Pos, a.End, a.Name, a.Requires, a.Hidden, a.DelayInmillis, buf.String(),
	)
}

type GenErrorRetrunVisitor struct {
	Lines []struct {
		Tpl  string
		Line []int
	}
}

func (g *GenErrorRetrunVisitor) Visit(node ast.Node) ast.Visitor {
	switch t := node.(type) {
	case *ast.FuncDecl:
		g.visitFuncDecl(t)
	}
	return g
}

func (g *GenErrorRetrunVisitor) visitFuncDecl(fd *ast.FuncDecl) {
	var tpl = `
if %s != nil {
	return %s
}`
	var ret string
	if fd.Type.Results != nil {
		for i, f := range fd.Type.Results.List {
			switch f.Type.(*ast.Ident).Name {
			case "error":
				ret += "%s,"
			}
			if i == len(fd.Type.Results.List)-1 {
				ret = strings.TrimRight(ret, ",")
			}
		}
	}

	var erri *ast.Ident
	var iline = struct {
		Tpl  string
		Line []int
	}{
		Line: make([]int, 0, 10),
	}

	for _, s := range fd.Body.List {
		switch t := s.(type) {
		case *ast.AssignStmt:
			if erri == nil {
				break
			}
			for _, e := range t.Lhs {
				i, ok := e.(*ast.Ident)
				if ok {
					if i.Name == erri.Name {
						iline.Line = append(iline.Line, int(t.End()))
					}
				}
			}
		case *ast.DeclStmt:
			gd, ok := t.Decl.(*ast.GenDecl)
			if !ok {
				break
			}

			if len(gd.Specs) != 1 {
				break
			}

			vs, ok := gd.Specs[0].(*ast.ValueSpec)
			if !ok {
				break
			}

			se, ok := vs.Type.(*ast.SelectorExpr)
			if !ok {
				break
			}

			if se.X.(*ast.Ident).Name != "gss" || se.Sel.Name != "IfErrorNotNilReturn" {
				break
			}

			if len(vs.Names) != 1 {
				break
			}

			erri = vs.Names[0]
			if ret != "" {
				ret = fmt.Sprintf(ret, erri)
			}

			iline.Tpl = fmt.Sprintf(tpl, erri, ret)
		}
	}
	if iline.Tpl != "" {
		g.Lines = append(g.Lines, iline)
	}
}

func (g *GenErrorRetrunVisitor) Replace(old, new string) {
	bs, _ := ioutil.ReadFile(old)
	s := string(bs)

	var offset = 0

	for _, v := range g.Lines {
		for _, v2 := range v.Line {
			left := s[:v2+offset-1]
			right := s[v2+offset-1:]
			s = left + v.Tpl + right
			offset += len(v.Tpl)
		}
	}
	ioutil.WriteFile(new, []byte(s), os.ModePerm)
}

func (g GenErrorRetrunVisitor) String() string {
	var buf bytes.Buffer
	for _, v := range g.Lines {
		for _, v2 := range v.Line {
			buf.WriteString(strconv.FormatInt(int64(v2), 10) + ",")
		}
		buf.WriteString(v.Tpl + "\n")
	}
	return buf.String()
}

func tof(i any) {
	fmt.Println(reflect.TypeOf(i))
}

func fp(i ...any) {
	fmt.Println(i...)
}
