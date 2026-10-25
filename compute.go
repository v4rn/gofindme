package main

import "go/types"

func computeImplementations(index *Index) []Implementation {
	impl := []Implementation{}

	for _, c := range index.Concrete {
		for _, a := range index.Abstract {
			iface := a.Type().Underlying().(*types.Interface)
			if iface.NumMethods() == 0 {
				continue
			}
			t := c.Type()
			switch {
			case types.Implements(t, iface):
				impl = append(impl, Implementation{t.String(), a.Type().String()})
			case types.Implements(types.NewPointer(t), iface):
				impl = append(impl, Implementation{"*" + t.String(), a.Type().String()})
			}
		}
	}

	return impl
}
