package main

import (
	"fmt"
	"go/types"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

func skip(p *packages.Package) bool {
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

func main() {

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, "std")
	if err != nil {
		log.Fatal(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	var abstractTypes []*types.TypeName
	var concreteTypes []*types.TypeName

	for _, pkg := range pkgs {
		if skip(pkg) {
			continue
		}

		if pkg.PkgPath == "io" || pkg.PkgPath == "os" {
			for _, name := range pkg.Types.Scope().Names() {
				tn, ok := pkg.Types.Scope().Lookup(name).(*types.TypeName)
				if !ok {
					// This is not a valid type, it is probably a constant or a function
					continue
				}
				if types.IsInterface(tn.Type()) {
					abstractTypes = append(abstractTypes, tn)
				} else {
					concreteTypes = append(concreteTypes, tn)
				}
			}
		}
	}

	// All types have been gathered
	for _, c := range concreteTypes {
		for _, a := range abstractTypes {
			iface := a.Type().Underlying().(*types.Interface)
			if iface.NumMethods() == 0 {
				continue
			}
			t := c.Type()
			switch {
			case types.Implements(t, iface):
				fmt.Println(t, "implements", a.Type())
			case types.Implements(types.NewPointer(t), iface):
				fmt.Println("*"+t.String(), "implements", a.Type())
			}
		}
	}
}
