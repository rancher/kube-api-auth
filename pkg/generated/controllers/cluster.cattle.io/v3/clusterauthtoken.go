package v3

import (
	v3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	"github.com/rancher/wrangler/v3/pkg/generic"
)

type ClusterAuthTokenController interface {
	generic.ControllerInterface[*v3.ClusterAuthToken, *v3.ClusterAuthTokenList]
}

type ClusterAuthTokenClient interface {
	generic.ClientInterface[*v3.ClusterAuthToken, *v3.ClusterAuthTokenList]
}

type ClusterAuthTokenCache interface {
	generic.CacheInterface[*v3.ClusterAuthToken]
}
