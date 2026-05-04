package v3

import (
	v3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	"github.com/rancher/wrangler/v3/pkg/generic"
)

type ClusterUserAttributeController interface {
	generic.ControllerInterface[*v3.ClusterUserAttribute, *v3.ClusterUserAttributeList]
}

type ClusterUserAttributeClient interface {
	generic.ClientInterface[*v3.ClusterUserAttribute, *v3.ClusterUserAttributeList]
}

type ClusterUserAttributeCache interface {
	generic.CacheInterface[*v3.ClusterUserAttribute]
}
