package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

func collectPackageDocs(pkg *packages.Package, docs map[types.Object]string) {
	add := func(id *ast.Ident, groups ...*ast.CommentGroup) {
		obj := pkg.TypesInfo.Defs[id]
		if obj == nil {
			return
		}
		for _, g := range groups {
			if g != nil {
				docs[obj] = g.Text()
				return
			}
		}
	}

	for _, f := range pkg.Syntax {
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				// package-level functions and methods with receivers
				add(d.Name, d.Doc)
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				for _, s := range d.Specs {
					ts := s.(*ast.TypeSpec)
					// a lone `type X ...` keeps its doc on the GenDecl
					add(ts.Name, ts.Doc, d.Doc)

					if it, ok := ts.Type.(*ast.InterfaceType); ok {
						for _, m := range it.Methods.List {
							// embedded interfaces have no Names
							for _, name := range m.Names {
								add(name, m.Doc)
							}
						}
					}
				}
			}
		}
	}
}

func skipPackage(p *packages.Package) bool {
	if p.Name == "main" {
		return true
	}
	for _, elem := range strings.Split(p.PkgPath, "/") {
		if elem == "internal" || elem == "vendor" || elem == "testdata" {
			return true
		}
	}
	return strings.HasPrefix(p.PkgPath, "cmd/")
}

func loadPackages(pattern string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkg, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(pkg) > 0 {
		return nil, fmt.Errorf("standard err has non-zero errors")
	}
	return pkg, nil
}

func classifyPackage(pkg *packages.Package, index *Index) {
	for _, name := range pkg.Types.Scope().Names() {
		if !token.IsExported(name) {
			continue
		}
		switch obj := pkg.Types.Scope().Lookup(name).(type) {
		case *types.Func:
			index.Funcs = append(index.Funcs, obj)
		case *types.TypeName:
			if types.IsInterface(obj.Type()) {
				index.Abstract = append(index.Abstract, obj)
			} else {
				index.Concrete = append(index.Concrete, obj)
			}

			// receiver functions
			if named, ok := obj.Type().(*types.Named); ok {
				for i := 0; i < named.NumMethods(); i++ {
					if m := named.Method(i); m.Exported() {
						index.Funcs = append(index.Funcs, m)
					}
				}
			}
		}
	}
}

func indexPackages(pattern string) (*Index, error) {
	pkgs, err := loadPackages(pattern)
	if err != nil {
		return nil, err
	}

	index := Index{Docs: make(map[types.Object]string)}

	for _, pkg := range pkgs {
		if skipPackage(pkg) {
			continue
		}

		collectPackageDocs(pkg, index.Docs)
		classifyPackage(pkg, &index)
	}
	return &index, nil
}
