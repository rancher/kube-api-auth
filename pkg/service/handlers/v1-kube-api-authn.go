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
	"github.com/rancher/kube-api-auth/pkg/clusterauth"
	clusterv3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

func (h *KubeAPIHandlers) V1AuthenticateHandler() http.HandlerFunc {
	return h.v1Authenticate
}

func (h *KubeAPIHandlers) v1Authenticate(w http.ResponseWriter, r *http.Request) {
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

	log.Debugf("Processing v1Authenticate request for %s", accessKey)

	user, err := h.v1getAndVerifyUser(r.Context(), accessKey, secretKey)
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

func (h *KubeAPIHandlers) v1getAndVerifyUser(ctx context.Context, accessKey, secretKey string) (*types.V1AuthnResponseUser, error) {

	var clusterUserAttribute *clusterv3.ClusterUserAttribute
	var clusterAuthToken *clusterv3.ClusterAuthToken

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		clusterAuthTokenLocal, err := h.clusterAuthTokensCache.Get(h.namespace, accessKey)
		if err != nil {
			return err
		}
		clusterAuthToken = clusterAuthTokenLocal
		if !clusterAuthTokenLocal.Enabled {
			return fmt.Errorf("token is not enabled")
		}

		clusterUserAttributeLocal, err := h.clusterUserAttributesCache.Get(h.namespace, clusterAuthTokenLocal.UserName)
		if err != nil {
			return err
		}

		clusterUserAttribute = clusterUserAttributeLocal
		if !clusterUserAttributeLocal.Enabled {
			return fmt.Errorf("user is not enabled")
		}

		clusterAuthTokenSecret, err := h.secretLister.Get(accessKey)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}

		migrate, err := clusterauth.VerifyClusterAuthToken(secretKey, clusterAuthTokenLocal, clusterAuthTokenSecret)
		if err != nil || !migrate {
			return err
		}

		// Migrate an un-migrated cluster auth token. This is done by creating
		// or writing over the secret to store the hash, and then removing the
		// hash from the cluster auth token. The token controller performs the
		// same actions.
		// go linting notes: this section of code intentionally reads/writes a deprecated resource field
		clusterAuthTokenSecret = clusterauth.NewClusterAuthTokenSecretForName(h.namespace, clusterAuthTokenLocal.Name, clusterAuthTokenLocal.SecretKeyHash) // nolint:staticcheck

		if _, err = h.secrets.Create(ctx, clusterAuthTokenSecret, metav1.CreateOptions{}); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return err
			}

			existing, err := h.secrets.Get(ctx, clusterAuthTokenSecret.Name, metav1.GetOptions{})
			if err != nil {
				log.Errorf("error migrating clusterAuthToken's secret %s: %s", clusterAuthTokenLocal.Name, err)
				return err
			}
			existing.Data = clusterAuthTokenSecret.Data
			if _, err := h.secrets.Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
				log.Errorf("error migrating clusterAuthToken's secret %s: %s", clusterAuthTokenLocal.Name, err)
				return err
			}
		}

		clusterAuthTokenLocal.SecretKeyHash = "" // nolint:staticcheck
		if _, err = h.clusterAuthTokens.Update(clusterAuthTokenLocal); err != nil {
			log.Errorf("error migrating clusterAuthToken %s: %s", clusterAuthTokenLocal.Name, err)
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	now := time.Now()
	refreshPeriod := h.getRefreshPeriod()
	if refreshPeriod >= 0 && clusterUserAttribute.LastRefresh != "" && !clusterUserAttribute.NeedsRefresh {
		refresh, err := time.Parse(time.RFC3339, clusterUserAttribute.LastRefresh)
		if err != nil {
			return nil, fmt.Errorf("error parsing lastRefresh: %w", err)
		}

		if refresh.Add(refreshPeriod).Before(now) {
			clusterUserAttribute.NeedsRefresh = true
			if _, err := h.clusterUserAttributes.Update(clusterUserAttribute); err != nil {
				return nil, fmt.Errorf("error updating clusterUserAttribute %s: %w", clusterUserAttribute.Name, err)
			}
		}
	}

	func() {
		precision := time.Second
		now = now.Truncate(precision)

		if clusterAuthToken.LastUsedAt != nil {
			if now.Equal(clusterAuthToken.LastUsedAt.Truncate(precision)) {
				return
			}
		}

		lastUsedAt := metav1.NewTime(now)
		clusterAuthToken.LastUsedAt = &lastUsedAt

		if _, err = h.clusterAuthTokens.Update(clusterAuthToken); err != nil {
			log.Errorf("error updating clusterAuthToken %s: %s", clusterAuthToken.Name, err)
		}
	}()

	return &types.V1AuthnResponseUser{
		UserName: clusterAuthToken.UserName,
		Groups:   clusterUserAttribute.Groups,
	}, nil
}

func (h *KubeAPIHandlers) getRefreshPeriod() time.Duration {
	const noDefault = time.Duration(-1)

	configMap, err := h.configMapLister.Get(clusterauth.AuthProviderRefreshDebounceSettingName)
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
