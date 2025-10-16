package handlers

import (
	clusterv3 "github.com/rancher/rancher/pkg/generated/norman/cluster.cattle.io/v3"
	corev1 "github.com/rancher/rancher/pkg/generated/norman/core/v1"
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

func NewKubeAPIHandlers(namespace string, clusterAPI clusterv3.Interface, coreAPI corev1.Interface) *KubeAPIHandlers {
	return &KubeAPIHandlers{
		namespace:                  namespace,
		clusterAuthTokens:          clusterAPI.ClusterAuthTokens(namespace),
		clusterAuthTokensLister:    clusterAPI.ClusterAuthTokens(namespace).Controller().Lister(),
		clusterUserAttribute:       clusterAPI.ClusterUserAttributes(namespace),
		clusterUserAttributeLister: clusterAPI.ClusterUserAttributes(namespace).Controller().Lister(),
		configMapLister:            coreAPI.ConfigMaps(namespace).Controller().Lister(),
		secrets:                    coreAPI.Secrets(namespace),
		secretLister:               coreAPI.Secrets(namespace).Controller().Lister(),
	}
}
