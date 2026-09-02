package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	kubeapiauth "github.com/rancher/kube-api-auth/pkg"
	"github.com/rancher/kube-api-auth/pkg/api/v1/types"
	"github.com/rancher/kube-api-auth/pkg/clients"
	"github.com/rancher/kube-api-auth/pkg/clusterauth"
	clusterv3wr "github.com/rancher/kube-api-auth/pkg/generated/controllers/cluster.cattle.io/v3"
	clusterv3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/retry"
)

type Authenticator struct {
	namespace                  string
	clusterAuthTokens          clusterv3wr.ClusterAuthTokenClient
	clusterAuthTokensCache     clusterv3wr.ClusterAuthTokenCache
	clusterUserAttributes      clusterv3wr.ClusterUserAttributeClient
	clusterUserAttributesCache clusterv3wr.ClusterUserAttributeCache
	configMapLister            corev1listers.ConfigMapNamespaceLister
	secrets                    corev1client.SecretInterface
	secretLister               corev1listers.SecretNamespaceLister
}

func NewAuthenticator(namespace string, c *clients.Clients) *Authenticator {
	return &Authenticator{
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

func (a *Authenticator) Authenticate(w http.ResponseWriter, r *http.Request) {
	response := types.V1AuthnResponse{
		APIVersion: kubeapiauth.DefaultK8sAPIVersion,
		Kind:       kubeapiauth.DefaultAuthnKind,
		Status: types.V1AuthnResponseStatus{
			Authenticated: false,
		},
	}

	accessKey, secretKey, err := v1parseBody(r)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	log.Debugf("Processing authentication request for %s", accessKey)

	user, err := a.v1getAndVerifyUser(r.Context(), accessKey, secretKey)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusUnauthorized, fmt.Sprintf("%v", err))
		return
	}

	response.Status.Authenticated = true
	response.Status.User = user

	responseJSON, err := json.Marshal(response)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusServiceUnavailable, fmt.Sprintf("%v", err))
		return
	}
	if _, err := w.Write(responseJSON); err != nil {
		ReturnHTTPError(w, r, http.StatusServiceUnavailable, fmt.Sprintf("%v", err))
		return
	}
}

func v1parseBody(r *http.Request) (string, string, error) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return "", "", err
	}

	authnReq, err := v1getBodyAuthnRequest(bytes)
	if err != nil {
		return "", "", err
	}

	tokenParts := strings.SplitN(authnReq.Spec.Token, ":", 2)
	if len(tokenParts) != 2 {
		return "", "", fmt.Errorf("found %d parts of token", len(tokenParts))
	}

	accessKey := strings.TrimPrefix(tokenParts[0], "ext/")
	secretKey := tokenParts[1]

	return accessKey, secretKey, nil
}

func v1getBodyAuthnRequest(bytes []byte) (*types.V1AuthnRequest, error) {
	authnReq := new(types.V1AuthnRequest)
	if err := json.Unmarshal(bytes, authnReq); err != nil {
		return nil, err
	}

	if authnReq.Kind != kubeapiauth.DefaultAuthnKind {
		return nil, errors.New("authentication request kind is not TokenReview")
	}

	if authnReq.Spec.Token == "" {
		return nil, errors.New("authentication request is missing Token")
	}

	return authnReq, nil
}

func (a *Authenticator) v1getAndVerifyUser(ctx context.Context, accessKey, secretKey string) (*types.V1AuthnResponseUser, error) {
	clusterAuthToken, err := a.clusterAuthTokensCache.Get(a.namespace, accessKey)
	if err != nil {
		return nil, err
	}
	if !clusterAuthToken.Enabled {
		return nil, fmt.Errorf("token is not enabled")
	}

	clusterUserAttribute, err := a.clusterUserAttributesCache.Get(a.namespace, clusterAuthToken.UserName)
	if err != nil {
		return nil, err
	}
	if !clusterUserAttribute.Enabled {
		return nil, fmt.Errorf("user is not enabled")
	}

	clusterAuthTokenSecret, err := a.secretLister.Get(accessKey)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	migrate, err := clusterauth.VerifyClusterAuthToken(secretKey, clusterAuthToken, clusterAuthTokenSecret)
	if err != nil {
		return nil, err
	}

	if migrate {
		migrated, err := a.migrateHash(ctx, accessKey)
		if err != nil {
			return nil, err
		}
		if migrated != nil {
			clusterAuthToken = migrated
		}
	}

	now := time.Now()
	refreshPeriod := a.getRefreshPeriod()
	if refreshPeriod >= 0 && clusterUserAttribute.LastRefresh != "" && !clusterUserAttribute.NeedsRefresh {
		refresh, err := time.Parse(time.RFC3339, clusterUserAttribute.LastRefresh)
		if err != nil {
			return nil, fmt.Errorf("error parsing lastRefresh: %w", err)
		}

		if refresh.Add(refreshPeriod).Before(now) {
			updated := clusterUserAttribute.DeepCopy()
			updated.NeedsRefresh = true
			if _, err := a.clusterUserAttributes.Update(updated); err != nil {
				// NeedsRefresh is a best-effort hint to the refresh
				// controller; do not fail authentication on it.
				log.Errorf("error setting NeedsRefresh on clusterUserAttribute %s: %s", clusterUserAttribute.Name, err)
			}
		}
	}

	a.touchLastUsedAt(now, clusterAuthToken)

	return &types.V1AuthnResponseUser{
		UserName: clusterAuthToken.UserName,
		Groups:   clusterUserAttribute.Groups,
	}, nil
}

// migrateHash moves a token's hash from the deprecated
// ClusterAuthToken.SecretKeyHash field into a Secret and clears the field on
// the token. Only the token update is retried on conflict: a concurrent
// migration by another pod (or the token controller) may have already cleared
// the field, in which case a fresh API read returns and nothing is updated.
// Returns the migrated token so callers can use it in place of their stale
// copy; returns nil if another actor already migrated.
//
// The read inside the retry loop hits the API server directly rather than
// the informer cache: after a Conflict the cache still holds the stale
// resourceVersion (the informer watch hasn't caught up), so a cache re-read
// would loop until DefaultRetry exhausts.
func (a *Authenticator) migrateHash(ctx context.Context, accessKey string) (*clusterv3.ClusterAuthToken, error) {
	var migrated *clusterv3.ClusterAuthToken
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		token, err := a.clusterAuthTokens.Get(a.namespace, accessKey, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if token.SecretKeyHash == "" { //nolint:staticcheck
			return nil
		}

		secret := clusterauth.NewClusterAuthTokenSecretForName(a.namespace, token.Name, token.SecretKeyHash) //nolint:staticcheck
		if _, err := a.secrets.Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return err
			}
			existing, err := a.secrets.Get(ctx, secret.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("migrating clusterAuthToken secret %s: %w", token.Name, err)
			}
			existing.Data = secret.Data
			if _, err := a.secrets.Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("migrating clusterAuthToken secret %s: %w", token.Name, err)
			}
		}

		token = token.DeepCopy()
		token.SecretKeyHash = "" //nolint:staticcheck
		updated, err := a.clusterAuthTokens.Update(token)
		if err != nil {
			return fmt.Errorf("migrating clusterAuthToken %s: %w", token.Name, err)
		}
		migrated = updated
		return nil
	})
	return migrated, err
}

// touchLastUsedAt bumps ClusterAuthToken.LastUsedAt to now, at most once per
// second, on a defensive copy of the cached object. Update errors are logged
// and swallowed since a stale LastUsedAt should not fail authentication.
func (a *Authenticator) touchLastUsedAt(now time.Time, token *clusterv3.ClusterAuthToken) {
	const precision = time.Second
	now = now.Truncate(precision)

	if token.LastUsedAt != nil && now.Equal(token.LastUsedAt.Truncate(precision)) {
		return
	}

	updated := token.DeepCopy()
	lastUsedAt := metav1.NewTime(now)
	updated.LastUsedAt = &lastUsedAt

	if _, err := a.clusterAuthTokens.Update(updated); err != nil {
		log.Errorf("error updating clusterAuthToken %s: %s", updated.Name, err)
	}
}

func (a *Authenticator) getRefreshPeriod() time.Duration {
	const noDefault = time.Duration(-1)

	configMap, err := a.configMapLister.Get(clusterauth.AuthProviderRefreshDebounceSettingName)
	if err != nil || configMap.Data == nil {
		return noDefault
	}

	refreshStr := configMap.Data["value"]
	if refreshStr == "" {
		return noDefault
	}

	refreshSeconds, err := strconv.ParseInt(refreshStr, 10, 64)
	if err != nil {
		return noDefault
	}

	return time.Duration(refreshSeconds) * time.Second
}
