package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/kube-api-auth/pkg"
	"github.com/rancher/kube-api-auth/pkg/api/v1/types"
	"github.com/rancher/rancher/pkg/controllers/managementuser/clusterauthtoken/common"
	log "github.com/sirupsen/logrus"
)

func (kube *KubeAPIHandlers) V1AuthenticateHandler() http.HandlerFunc {
	return kube.v1Authenticate
}

func (kube *KubeAPIHandlers) v1Authenticate(w http.ResponseWriter, r *http.Request) {
	log.Info("Processing v1Authenticate request...")

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
	log.Infof("  ...looking up token for %s", accessKey)

	user, err := kube.v1getAndVerifyUser(accessKey, secretKey)
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
	log.Infof(string(responseJSON))
	log.Infof("  ...authenticated %s!", accessKey)
}

func v1parseBody(r *http.Request) (string, string, error) {
	bytes, err := ioutil.ReadAll(r.Body)
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
	accessKey := tokenParts[0]
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

func (kube *KubeAPIHandlers) v1getAndVerifyUser(accessKey, secretKey string) (*types.V1AuthnResponseUser, error) {
	clusterAuthToken, err := kube.clusterAuthTokensLister.Get(kube.namespace, accessKey)
	if err != nil {
		return nil, err
	}
	if !clusterAuthToken.Enabled {
		return nil, fmt.Errorf("token is not enabled")
	}

	userName := clusterAuthToken.UserName
	clusterUserAttribute, err := kube.clusterUserAttributeLister.Get(kube.namespace, userName)
	if err != nil {
		return nil, err
	}
	if !clusterUserAttribute.Enabled {
		return nil, fmt.Errorf("user is not enabled")
	}

	err = common.VerifyClusterAuthToken(secretKey, clusterAuthToken)
	if err != nil {
		return nil, err
	}

	refreshPeriod := kube.getRefreshPeriod()
	if refreshPeriod >= time.Duration(0) && clusterUserAttribute.LastRefresh != "" && !clusterUserAttribute.NeedsRefresh {
		refresh, err := time.Parse(time.RFC3339, clusterUserAttribute.LastRefresh)
		if err != nil {
			return nil, err
		}
		if refresh.Add(refreshPeriod).Before(time.Now()) {
			clusterUserAttribute.NeedsRefresh = true
			_, err := kube.clusterUserAttribute.Update(clusterUserAttribute)
			if err != nil {
				return nil, err
			}
		}
	}

	response := &types.V1AuthnResponseUser{
		UserName: userName,
		Groups:   clusterUserAttribute.Groups,
	}

	return response, nil
}

func (kube *KubeAPIHandlers) getRefreshPeriod() time.Duration {
	const noDefault = time.Duration(-1)
	configMap, err := kube.configMapLister.Get(kube.namespace, common.AuthProviderRefreshDebounceSettingName)
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
