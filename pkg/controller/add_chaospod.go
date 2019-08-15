package controller

import (
	"github.com/wtfjoke/ordered-chaos-monkey-operator/pkg/controller/chaospod"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, chaospod.Add)
}
