package handlers

import (
	clusterv3 "github.com/rancher/rancher/pkg/generated/norman/cluster.cattle.io/v3"
	corev1 "github.com/rancher/rancher/pkg/generated/norman/core/v1"
	"github.com/rancher/rancher/pkg/types/config"
)

type KubeAPIHandlers struct {
	namespace                  string
	clusterAuthTokens          clusterv3.ClusterAuthTokenInterface
	clusterAuthTokensLister    clusterv3.ClusterAuthTokenLister
	clusterUserAttribute       clusterv3.ClusterUserAttributeInterface
	clusterUserAttributeLister clusterv3.ClusterUserAttributeLister
	configMapLister            corev1.ConfigMapLister
	secrets                    corev1.SecretInterface
	secretLister               corev1.SecretLister
}

func NewKubeAPIHandlers(namespace string, apiContext *config.UserOnlyContext) *KubeAPIHandlers {
	return &KubeAPIHandlers{
		namespace:                  namespace,
		clusterAuthTokens:          apiContext.Cluster.ClusterAuthTokens(namespace),
		clusterAuthTokensLister:    apiContext.Cluster.ClusterAuthTokens(namespace).Controller().Lister(),
		clusterUserAttribute:       apiContext.Cluster.ClusterUserAttributes(namespace),
		clusterUserAttributeLister: apiContext.Cluster.ClusterUserAttributes(namespace).Controller().Lister(),
		configMapLister:            apiContext.Core.ConfigMaps(namespace).Controller().Lister(),
		secrets:                    apiContext.Core.Secrets(namespace),
		secretLister:               apiContext.Core.Secrets(namespace).Controller().Lister(),
	}
}
