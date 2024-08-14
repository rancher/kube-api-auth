package handlers

import (
	mgmtcontrollers "github.com/rancher/rancher/pkg/generated/controllers/management.cattle.io/v3"
	clusterv3 "github.com/rancher/rancher/pkg/generated/norman/cluster.cattle.io/v3"
	corev1 "github.com/rancher/rancher/pkg/generated/norman/core/v1"
	"github.com/rancher/rancher/pkg/types/config"
	"github.com/rancher/rancher/pkg/wrangler"
)

type KubeAPIHandlers struct {
	namespace                  string
	clusterAuthTokens          clusterv3.ClusterAuthTokenInterface
	clusterAuthTokensLister    clusterv3.ClusterAuthTokenLister
	clusterUserAttribute       clusterv3.ClusterUserAttributeInterface
	clusterUserAttributeLister clusterv3.ClusterUserAttributeLister
	configMapLister            corev1.ConfigMapLister
	tokenWClient               mgmtcontrollers.TokenController
}

func NewKubeAPIHandlers(namespace string, apiContext *config.UserOnlyContext, wranglerCtx *wrangler.Context) *KubeAPIHandlers {
	return &KubeAPIHandlers{
		namespace:                  namespace,
		clusterAuthTokens:          apiContext.Cluster.ClusterAuthTokens(namespace),
		clusterAuthTokensLister:    apiContext.Cluster.ClusterAuthTokens(namespace).Controller().Lister(),
		clusterUserAttribute:       apiContext.Cluster.ClusterUserAttributes(namespace),
		clusterUserAttributeLister: apiContext.Cluster.ClusterUserAttributes(namespace).Controller().Lister(),
		configMapLister:            apiContext.Core.ConfigMaps(namespace).Controller().Lister(),
		tokenWClient:               wranglerCtx.Mgmt.Token(),
	}
}
