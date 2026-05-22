package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	kubeapiauth "github.com/rancher/kube-api-auth/pkg"
	"github.com/rancher/kube-api-auth/pkg/api/v1/types"
	clusterv3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	"github.com/rancher/rancher/pkg/auth/tokens/hashers"
	"github.com/rancher/rancher/pkg/controllers/managementuser/clusterauthtoken/common"
	clusterfakes "github.com/rancher/rancher/pkg/generated/norman/cluster.cattle.io/v3/fakes"
	corefakes "github.com/rancher/rancher/pkg/generated/norman/core/v1/fakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	testNamespace = "test-ns"
	testAccessKey = "test-token"
	testSecretKey = "test-secret-key"
	testUserName  = "user-abc123"
)

var testSecretKeyHash string

// Pre-compute the scrypt hash once — too expensive (~100ms) to repeat per subtest.
func TestMain(m *testing.M) {
	hash, err := hashers.ScryptHasher{}.CreateHash(testSecretKey)
	if err != nil {
		panic("failed to create test hash: " + err.Error())
	}
	testSecretKeyHash = hash
	os.Exit(m.Run())
}

func tokenReviewRequest(t *testing.T, token string) *http.Request {
	t.Helper()

	body := types.V1AuthnRequest{
		APIVersion: kubeapiauth.DefaultK8sAPIVersion,
		Kind:       kubeapiauth.DefaultAuthnKind,
		Spec:       types.V1AuthnRequestSpec{Token: token},
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	return httptest.NewRequest(http.MethodPost, "/v1/authenticate", bytes.NewReader(data))
}

func notFound(name string) error {
	return apierrors.NewNotFound(schema.GroupResource{}, name)
}

func newTestToken() *clusterv3.ClusterAuthToken {
	return &clusterv3.ClusterAuthToken{
		ObjectMeta: metav1.ObjectMeta{Name: testAccessKey},
		UserName:   testUserName,
		Enabled:    true,
	}
}

func newTestUser(groups ...string) *clusterv3.ClusterUserAttribute {
	return &clusterv3.ClusterUserAttribute{
		ObjectMeta: metav1.ObjectMeta{Name: testUserName},
		Groups:     groups,
		Enabled:    true,
	}
}

func newTestSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.ClusterAuthTokenSecretName(testAccessKey),
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			common.ClusterAuthSecretHashField: []byte(testSecretKeyHash),
		},
	}
}

func noRefreshConfigMap() *corefakes.ConfigMapListerMock {
	return &corefakes.ConfigMapListerMock{
		GetFunc: func(ns, name string) (*corev1.ConfigMap, error) {
			return nil, notFound(name)
		},
	}
}

func TestV1parseBody(t *testing.T) {
	t.Parallel()

	t.Run("valid tokens", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name       string
			token      string
			wantKey    string
			wantSecret string
		}{
			{
				name:       "legacy token",
				token:      "tokenName:secretValue",
				wantKey:    "tokenName",
				wantSecret: "secretValue",
			},
			{
				name:       "ext token",
				token:      "ext/token-abc123:secretValue",
				wantKey:    "token-abc123",
				wantSecret: "secretValue",
			},
			{
				name:       "ext token with colons in secret",
				token:      "ext/token-abc123:secret:with:colons",
				wantKey:    "token-abc123",
				wantSecret: "secret:with:colons",
			},
			{
				name:       "legacy token with colons in secret",
				token:      "tokenName:secret:with:colons",
				wantKey:    "tokenName",
				wantSecret: "secret:with:colons",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				r := tokenReviewRequest(t, tt.token)
				accessKey, secretKey, err := v1parseBody(r)
				require.NoError(t, err)
				assert.Equal(t, tt.wantKey, accessKey)
				assert.Equal(t, tt.wantSecret, secretKey)
			})
		}
	})

	t.Run("missing colon", func(t *testing.T) {
		t.Parallel()

		r := tokenReviewRequest(t, "nocolonhere")

		_, _, err := v1parseBody(r)
		require.Error(t, err)
	})

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", strings.NewReader(""))

		_, _, err := v1parseBody(r)
		require.Error(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", strings.NewReader("{invalid"))

		_, _, err := v1parseBody(r)
		require.Error(t, err)
	})

	t.Run("wrong kind", func(t *testing.T) {
		t.Parallel()

		body := types.V1AuthnRequest{
			Kind: "NotTokenReview",
			Spec: types.V1AuthnRequestSpec{Token: "key:secret"},
		}
		data, _ := json.Marshal(body)
		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", bytes.NewReader(data))

		_, _, err := v1parseBody(r)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not TokenReview")
	})

	t.Run("empty token", func(t *testing.T) {
		t.Parallel()

		body := types.V1AuthnRequest{
			Kind: kubeapiauth.DefaultAuthnKind,
			Spec: types.V1AuthnRequestSpec{Token: ""},
		}
		data, _ := json.Marshal(body)
		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", bytes.NewReader(data))

		_, _, err := v1parseBody(r)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing Token")
	})
}

func TestGetRefreshPeriod(t *testing.T) {
	t.Parallel()

	t.Run("configmap not found", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace:       testNamespace,
			configMapLister: noRefreshConfigMap(),
		}

		assert.Equal(t, time.Duration(-1), h.getRefreshPeriod())
	})

	t.Run("valid value", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			configMapLister: &corefakes.ConfigMapListerMock{
				GetFunc: func(ns, name string) (*corev1.ConfigMap, error) {
					return &corev1.ConfigMap{Data: map[string]string{"value": "60"}}, nil
				},
			},
		}

		assert.Equal(t, 60*time.Second, h.getRefreshPeriod())
	})

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			configMapLister: &corefakes.ConfigMapListerMock{
				GetFunc: func(ns, name string) (*corev1.ConfigMap, error) {
					return &corev1.ConfigMap{Data: map[string]string{"value": "0"}}, nil
				},
			},
		}

		assert.Equal(t, time.Duration(0), h.getRefreshPeriod())
	})

	t.Run("empty value", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			configMapLister: &corefakes.ConfigMapListerMock{
				GetFunc: func(ns, name string) (*corev1.ConfigMap, error) {
					return &corev1.ConfigMap{Data: map[string]string{"value": ""}}, nil
				},
			},
		}

		assert.Equal(t, time.Duration(-1), h.getRefreshPeriod())
	})

	t.Run("non-numeric value", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			configMapLister: &corefakes.ConfigMapListerMock{
				GetFunc: func(ns, name string) (*corev1.ConfigMap, error) {
					return &corev1.ConfigMap{Data: map[string]string{"value": "abc"}}, nil
				},
			},
		}

		assert.Equal(t, time.Duration(-1), h.getRefreshPeriod())
	})

	t.Run("nil data", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			configMapLister: &corefakes.ConfigMapListerMock{
				GetFunc: func(ns, name string) (*corev1.ConfigMap, error) {
					return &corev1.ConfigMap{}, nil
				},
			},
		}

		assert.Equal(t, time.Duration(-1), h.getRefreshPeriod())
	})
}

func TestGetAndVerifyUser(t *testing.T) {
	t.Parallel()

	t.Run("valid token and user", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()
			user := newTestUser("group1", "group2")
			secret := newTestSecret()

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return secret, nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return in1, nil
					},
				},
			}

			result, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)
			assert.Equal(t, []string{"group1", "group2"}, result.Groups)
		})
	})

	t.Run("token not found", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return nil, notFound(name)
				},
			},
		}

		_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.True(t, apierrors.IsNotFound(err))
	})

	t.Run("token disabled", func(t *testing.T) {
		t.Parallel()

		token := newTestToken()
		token.Enabled = false

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return token, nil
				},
			},
		}

		_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not enabled")
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return newTestToken(), nil
				},
			},
			clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return nil, notFound(name)
				},
			},
		}

		_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.True(t, apierrors.IsNotFound(err))
	})

	t.Run("user disabled", func(t *testing.T) {
		t.Parallel()

		user := newTestUser()
		user.Enabled = false

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return newTestToken(), nil
				},
			},
			clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return user, nil
				},
			},
		}

		_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not enabled")
	})

	t.Run("wrong secret key", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return newTestToken(), nil
				},
			},
			clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return newTestUser(), nil
				},
			},
			secretLister: &corefakes.SecretListerMock{
				GetFunc: func(ns, name string) (*corev1.Secret, error) {
					return newTestSecret(), nil
				},
			},
		}

		_, err := h.v1getAndVerifyUser(testAccessKey, "wrong-secret")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match")
	})

	t.Run("expired token", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()
			token.ExpiresAt = time.Now().Add(-time.Hour).Format(time.RFC3339)

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser(), nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
			}

			_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "expired")
		})
	})

	t.Run("secret missing and no hash", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return newTestToken(), nil
				},
			},
			clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return newTestUser(), nil
				},
			},
			secretLister: &corefakes.SecretListerMock{
				GetFunc: func(ns, name string) (*corev1.Secret, error) {
					return nil, notFound(name)
				},
			},
		}

		_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing")
	})

	t.Run("migration creates secret and clears token hash", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()
			token.SecretKeyHash = testSecretKeyHash // nolint:staticcheck

			var createdSecret *corev1.Secret
			var updatedToken *clusterv3.ClusterAuthToken

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser("admins"), nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return nil, notFound(name)
					},
				},
				secrets: &corefakes.SecretInterfaceMock{
					CreateFunc: func(in1 *corev1.Secret) (*corev1.Secret, error) {
						createdSecret = in1
						return in1, nil
					},
				},
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						updatedToken = in1.DeepCopy()
						return in1, nil
					},
				},
				configMapLister: noRefreshConfigMap(),
			}

			result, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)

			require.NotNil(t, createdSecret)
			assert.Equal(t, testSecretKeyHash, string(createdSecret.Data[common.ClusterAuthSecretHashField]))

			require.NotNil(t, updatedToken)
			assert.Empty(t, updatedToken.SecretKeyHash) // nolint:staticcheck
		})
	})

	t.Run("migration overwrites existing secret", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()
			token.SecretKeyHash = testSecretKeyHash // nolint:staticcheck

			var updatedSecretData map[string][]byte

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser(), nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return nil, notFound(name)
					},
				},
				secrets: &corefakes.SecretInterfaceMock{
					CreateFunc: func(in1 *corev1.Secret) (*corev1.Secret, error) {
						return nil, apierrors.NewAlreadyExists(schema.GroupResource{}, in1.Name)
					},
					GetNamespacedFunc: func(ns, name string, opts metav1.GetOptions) (*corev1.Secret, error) {
						return &corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
							Data:       map[string][]byte{"hash": []byte("old-hash")},
						}, nil
					},
					UpdateFunc: func(in1 *corev1.Secret) (*corev1.Secret, error) {
						updatedSecretData = in1.Data
						return in1, nil
					},
				},
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return in1, nil
					},
				},
				configMapLister: noRefreshConfigMap(),
			}

			result, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)

			require.NotNil(t, updatedSecretData)
			assert.Equal(t, testSecretKeyHash, string(updatedSecretData[common.ClusterAuthSecretHashField]))
		})
	})

	t.Run("migration secret create fails", func(t *testing.T) {
		t.Parallel()

		token := newTestToken()
		token.SecretKeyHash = testSecretKeyHash // nolint:staticcheck

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return token, nil
				},
			},
			clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return newTestUser(), nil
				},
			},
			secretLister: &corefakes.SecretListerMock{
				GetFunc: func(ns, name string) (*corev1.Secret, error) {
					return nil, notFound(name)
				},
			},
			secrets: &corefakes.SecretInterfaceMock{
				CreateFunc: func(in1 *corev1.Secret) (*corev1.Secret, error) {
					return nil, fmt.Errorf("storage unavailable")
				},
			},
		}

		_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "storage unavailable")
	})

	t.Run("migration token update fails", func(t *testing.T) {
		t.Parallel()

		token := newTestToken()
		token.SecretKeyHash = testSecretKeyHash // nolint:staticcheck

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return token, nil
				},
			},
			clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return newTestUser(), nil
				},
			},
			secretLister: &corefakes.SecretListerMock{
				GetFunc: func(ns, name string) (*corev1.Secret, error) {
					return nil, notFound(name)
				},
			},
			secrets: &corefakes.SecretInterfaceMock{
				CreateFunc: func(in1 *corev1.Secret) (*corev1.Secret, error) {
					return in1, nil
				},
			},
			clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
				UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
					return nil, fmt.Errorf("conflict")
				},
			},
		}

		_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
		require.Error(t, err)
	})

	t.Run("refresh triggered when overdue", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			user := newTestUser("editors")
			user.LastRefresh = time.Now().Add(-2 * time.Hour).Format(time.RFC3339)

			var userUpdated bool

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: &corefakes.ConfigMapListerMock{
					GetFunc: func(ns, name string) (*corev1.ConfigMap, error) {
						return &corev1.ConfigMap{Data: map[string]string{"value": "3600"}}, nil
					},
				},
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return in1, nil
					},
				},
				clusterUserAttribute: &clusterfakes.ClusterUserAttributeInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterUserAttribute) (*clusterv3.ClusterUserAttribute, error) {
						userUpdated = true
						assert.True(t, in1.NeedsRefresh)
						return in1, nil
					},
				},
			}

			result, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)
			assert.True(t, userUpdated)
		})
	})

	t.Run("refresh not triggered when recent", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			user := newTestUser("viewers")
			user.LastRefresh = time.Now().Add(-10 * time.Minute).Format(time.RFC3339)

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: &corefakes.ConfigMapListerMock{
					GetFunc: func(ns, name string) (*corev1.ConfigMap, error) {
						return &corev1.ConfigMap{Data: map[string]string{"value": "3600"}}, nil
					},
				},
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return in1, nil
					},
				},
			}

			result, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)
			assert.False(t, user.NeedsRefresh)
		})
	})

	t.Run("refresh skipped when period disabled", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			user := newTestUser()
			user.LastRefresh = time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return in1, nil
					},
				},
			}

			result, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)
			assert.False(t, user.NeedsRefresh)
		})
	})

	t.Run("invalid last refresh format", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			user := newTestUser()
			user.LastRefresh = "not-a-date"

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: &corefakes.ConfigMapListerMock{
					GetFunc: func(ns, name string) (*corev1.ConfigMap, error) {
						return &corev1.ConfigMap{Data: map[string]string{"value": "3600"}}, nil
					},
				},
			}

			_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "parsing lastRefresh")
		})
	})

	t.Run("user attribute update fails during refresh", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			user := newTestUser()
			user.LastRefresh = time.Now().Add(-2 * time.Hour).Format(time.RFC3339)

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: &corefakes.ConfigMapListerMock{
					GetFunc: func(ns, name string) (*corev1.ConfigMap, error) {
						return &corev1.ConfigMap{Data: map[string]string{"value": "3600"}}, nil
					},
				},
				clusterUserAttribute: &clusterfakes.ClusterUserAttributeInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterUserAttribute) (*clusterv3.ClusterUserAttribute, error) {
						return nil, fmt.Errorf("update failed")
					},
				},
			}

			_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "update failed")
		})
	})

	t.Run("last used at updated on first use", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()
			var lastUsedAtSet bool

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser(), nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						lastUsedAtSet = true
						require.NotNil(t, in1.LastUsedAt)
						return in1, nil
					},
				},
			}

			_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.True(t, lastUsedAtSet)
		})
	})

	t.Run("last used at throttled within same second", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			now := time.Now().Truncate(time.Second)
			lastUsedAt := metav1.NewTime(now)
			token := newTestToken()
			token.LastUsedAt = &lastUsedAt

			var tokenUpdateCalled bool

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser(), nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						tokenUpdateCalled = true
						return in1, nil
					},
				},
			}

			_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.False(t, tokenUpdateCalled)
		})
	})

	t.Run("last used at updated after time passes", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			past := time.Now().Add(-2 * time.Second).Truncate(time.Second)
			lastUsedAt := metav1.NewTime(past)
			token := newTestToken()
			token.LastUsedAt = &lastUsedAt

			var tokenUpdateCalled bool

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser(), nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						tokenUpdateCalled = true
						return in1, nil
					},
				},
			}

			_, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.True(t, tokenUpdateCalled)
		})
	})

	t.Run("last used at update failure is silent", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()

			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser("admins"), nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return nil, fmt.Errorf("transient error")
					},
				},
			}

			result, err := h.v1getAndVerifyUser(testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)
			assert.Equal(t, []string{"admins"}, result.Groups)
		})
	})
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns authenticated response", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			h := &KubeAPIHandlers{
				namespace: testNamespace,
				clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributeLister: &clusterfakes.ClusterUserAttributeListerMock{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser("group1"), nil
					},
				},
				secretLister: &corefakes.SecretListerMock{
					GetFunc: func(ns, name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &clusterfakes.ClusterAuthTokenInterfaceMock{
					UpdateFunc: func(in1 *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return in1, nil
					},
				},
			}

			w := httptest.NewRecorder()
			r := tokenReviewRequest(t, testAccessKey+":"+testSecretKey)

			h.V1AuthenticateHandler().ServeHTTP(w, r)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp types.V1AuthnResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, kubeapiauth.DefaultK8sAPIVersion, resp.APIVersion)
			assert.Equal(t, kubeapiauth.DefaultAuthnKind, resp.Kind)
			assert.True(t, resp.Status.Authenticated)
			require.NotNil(t, resp.Status.User)
			assert.Equal(t, testUserName, resp.Status.User.UserName)
			assert.Equal(t, []string{"group1"}, resp.Status.User.Groups)
		})
	})

	t.Run("malformed body returns 400", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", strings.NewReader("not json"))

		h.V1AuthenticateHandler().ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid credentials returns 401", func(t *testing.T) {
		t.Parallel()

		h := &KubeAPIHandlers{
			namespace: testNamespace,
			clusterAuthTokensLister: &clusterfakes.ClusterAuthTokenListerMock{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return nil, notFound(name)
				},
			},
		}

		w := httptest.NewRecorder()
		r := tokenReviewRequest(t, "unknown-token:secret")

		h.V1AuthenticateHandler().ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
