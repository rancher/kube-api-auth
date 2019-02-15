package handlers

import (
	clusterv3 "github.com/rancher/types/apis/cluster.cattle.io/v3"
	corev1 "github.com/rancher/types/apis/core/v1"
	"github.com/rancher/types/config"
)

type KubeAPIHandlers struct {
	namespace                  string
	clusterAuthTokens          clusterv3.ClusterAuthTokenInterface
	clusterAuthTokensLister    clusterv3.ClusterAuthTokenLister
	clusterUserAttribute       clusterv3.ClusterUserAttributeInterface
	clusterUserAttributeLister clusterv3.ClusterUserAttributeLister
	configMapLister            corev1.ConfigMapLister
}

func NewKubeAPIHandlers(namespace string, apiContext *config.UserOnlyContext) *KubeAPIHandlers {
	return &KubeAPIHandlers{
		namespace:                  namespace,
		clusterAuthTokens:          apiContext.Cluster.ClusterAuthTokens(namespace),
		clusterAuthTokensLister:    apiContext.Cluster.ClusterAuthTokens(namespace).Controller().Lister(),
		clusterUserAttribute:       apiContext.Cluster.ClusterUserAttributes(namespace),
		clusterUserAttributeLister: apiContext.Cluster.ClusterUserAttributes(namespace).Controller().Lister(),
		configMapLister:            apiContext.Core.ConfigMaps(namespace).Controller().Lister(),
	}
}
