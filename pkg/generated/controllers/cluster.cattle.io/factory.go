package cluster

import (
	"github.com/rancher/wrangler/v3/pkg/generic"
	"k8s.io/client-go/rest"
)

type Factory struct {
	*generic.Factory
}

type FactoryOptions = generic.FactoryOptions

func NewFactoryFromConfigWithOptions(config *rest.Config, opts *FactoryOptions) (*Factory, error) {
	f, err := generic.NewFactoryFromConfigWithOptions(config, opts)
	return &Factory{Factory: f}, err
}

func (c *Factory) Cluster() Interface {
	return New(c.ControllerFactory())
}
