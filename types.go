package main

import "go/types"

type Implementation struct {
	Concrete string
	Abstract string
}

type Index struct {
	Abstract   []*types.TypeName
	Concrete   []*types.TypeName
	Funcs      []*types.Func
	Docs       map[types.Object]string
	Implements []Implementation
}
