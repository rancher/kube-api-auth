// Package clusterauth verifies cluster auth tokens against their stored hash.
//
// Adapted from github.com/rancher/rancher/pkg/controllers/managementuser/clusterauthtoken/common at
// v0.0.0-20260625140903-06a1727c1d54. Token-creation helpers were dropped because kube-api-auth
// only verifies tokens; the VerifyClusterAuthToken return order was normalized to (result, error).
package clusterauth

import (
	"fmt"
	"time"

	"github.com/rancher/kube-api-auth/pkg/auth/hashers"
	clusterv3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterAuthTokenSecretValue extracts the token hash stored in the secret.
func ClusterAuthTokenSecretValue(clusterAuthSecret *corev1.Secret) string {
	return string(clusterAuthSecret.Data[ClusterAuthSecretHashField])
}

// NewClusterAuthTokenSecretForName creates a new secret from the given token and its hash value.
// The cluster auth token is managed separately. Does not create the secret in the remote cluster.
func NewClusterAuthTokenSecretForName(ns, name, hashedValue string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.Version,
			Kind:       "Secret",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			ClusterAuthSecretHashField: []byte(hashedValue),
		},
	}
}

// VerifyClusterAuthToken verifies that a provided secret key is valid for the
// given clusterAuthToken and hashed value. Also determines if the hashed value
// requires migration from cluster auth token to cluster auth token secret.
func VerifyClusterAuthToken(secretKey string, clusterAuthToken *clusterv3.ClusterAuthToken, clusterAuthTokenSecret *corev1.Secret) (migrate bool, err error) {
	if !clusterAuthToken.Enabled {
		return false, fmt.Errorf("token is not enabled")
	}

	expiresAt := clusterAuthToken.ExpiresAt
	if expiresAt != "" {
		expires, err := time.Parse(time.RFC3339, expiresAt)
		if err != nil {
			return false, err
		}
		if expires.Before(time.Now()) {
			return false, fmt.Errorf("auth expired at %s", expiresAt)
		}
	}

	hashedValue := clusterAuthToken.SecretKeyHash //nolint:staticcheck
	migrate = true

	if hashedValue == "" {
		if clusterAuthTokenSecret == nil {
			return false, fmt.Errorf("hash secret is missing")
		}

		hashedValue = ClusterAuthTokenSecretValue(clusterAuthTokenSecret)
		migrate = false
	}

	hasher, err := hashers.GetHasherForHash(hashedValue)
	if err != nil {
		return false, fmt.Errorf("unable to get hasher for clusterAuthToken %s/%s, err: %w",
			clusterAuthToken.Name, clusterAuthToken.Namespace, err)
	}

	if err := hasher.VerifyHash(hashedValue, secretKey); err != nil {
		return migrate, err
	}
	return migrate, nil
}
