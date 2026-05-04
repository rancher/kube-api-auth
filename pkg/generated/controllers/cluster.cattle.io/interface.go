package cluster

import (
	v3 "github.com/rancher/kube-api-auth/pkg/generated/controllers/cluster.cattle.io/v3"
	"github.com/rancher/lasso/pkg/controller"
)

type Interface interface {
	V3() v3.Interface
}

func New(controllerFactory controller.SharedControllerFactory) Interface {
	return &version{
		controllerFactory: controllerFactory,
	}
}

type version struct {
	controllerFactory controller.SharedControllerFactory
}

func (v *version) V3() v3.Interface {
	return v3.New(v.controllerFactory)
}
