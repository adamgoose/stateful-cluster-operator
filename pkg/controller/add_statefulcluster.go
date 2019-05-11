package controller

import (
	"github.com/adamgoose/stateful-cluster-operator/pkg/controller/statefulcluster"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, statefulcluster.Add)
}
