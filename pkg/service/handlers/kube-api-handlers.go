package handlers

import (
	"github.com/rancher/kube-api-auth/pkg/clients"
	clusterv3 "github.com/rancher/kube-api-auth/pkg/generated/controllers/cluster.cattle.io/v3"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
)

type KubeAPIHandlers struct {
	namespace                  string
	clusterAuthTokens          clusterv3.ClusterAuthTokenClient
	clusterAuthTokensCache     clusterv3.ClusterAuthTokenCache
	clusterUserAttributes      clusterv3.ClusterUserAttributeClient
	clusterUserAttributesCache clusterv3.ClusterUserAttributeCache
	configMapLister            corev1listers.ConfigMapNamespaceLister
	secrets                    corev1client.SecretInterface
	secretLister               corev1listers.SecretNamespaceLister
}

func NewKubeAPIHandlers(namespace string, c *clients.Clients) *KubeAPIHandlers {
	return &KubeAPIHandlers{
		namespace:                  namespace,
		clusterAuthTokens:          c.ClusterAuthTokens,
		clusterAuthTokensCache:     c.ClusterAuthTokenCache,
		clusterUserAttributes:      c.ClusterUserAttributes,
		clusterUserAttributesCache: c.ClusterUserAttributeCache,
		configMapLister:            c.ConfigMapLister,
		secrets:                    c.Secrets,
		secretLister:               c.SecretLister,
	}
}
