package handlers

import (
	"bytes"
	"context"
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
	"github.com/rancher/kube-api-auth/pkg/auth/hashers"
	"github.com/rancher/kube-api-auth/pkg/clusterauth"
	clusterv3wr "github.com/rancher/kube-api-auth/pkg/generated/controllers/cluster.cattle.io/v3"
	clusterv3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
)

const (
	testNamespace = "test-ns"
	testAccessKey = "test-token"
	testSecretKey = "test-secret-key"
	testUserName  = "user-abc123"
)

var testSecretKeyHash string

func TestMain(m *testing.M) {
	hash, err := hashers.GetHasher().CreateHash(testSecretKey)
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
			Name:      testAccessKey,
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			clusterauth.ClusterAuthSecretHashField: []byte(testSecretKeyHash),
		},
	}
}

type fakeClusterAuthTokenCache struct {
	clusterv3wr.ClusterAuthTokenCache
	GetFunc func(namespace, name string) (*clusterv3.ClusterAuthToken, error)
}

func (f *fakeClusterAuthTokenCache) Get(namespace, name string) (*clusterv3.ClusterAuthToken, error) {
	return f.GetFunc(namespace, name)
}

type fakeClusterUserAttributeCache struct {
	clusterv3wr.ClusterUserAttributeCache
	GetFunc func(namespace, name string) (*clusterv3.ClusterUserAttribute, error)
}

func (f *fakeClusterUserAttributeCache) Get(namespace, name string) (*clusterv3.ClusterUserAttribute, error) {
	return f.GetFunc(namespace, name)
}

type fakeClusterAuthTokenClient struct {
	clusterv3wr.ClusterAuthTokenClient
	GetFunc    func(namespace, name string, opts metav1.GetOptions) (*clusterv3.ClusterAuthToken, error)
	UpdateFunc func(*clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error)
}

func (f *fakeClusterAuthTokenClient) Get(namespace, name string, opts metav1.GetOptions) (*clusterv3.ClusterAuthToken, error) {
	return f.GetFunc(namespace, name, opts)
}

func (f *fakeClusterAuthTokenClient) Update(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
	return f.UpdateFunc(obj)
}

type fakeClusterUserAttributeClient struct {
	clusterv3wr.ClusterUserAttributeClient
	UpdateFunc func(*clusterv3.ClusterUserAttribute) (*clusterv3.ClusterUserAttribute, error)
}

func (f *fakeClusterUserAttributeClient) Update(obj *clusterv3.ClusterUserAttribute) (*clusterv3.ClusterUserAttribute, error) {
	return f.UpdateFunc(obj)
}

type fakeSecretLister struct {
	corev1listers.SecretNamespaceLister
	GetFunc func(name string) (*corev1.Secret, error)
}

func (f *fakeSecretLister) Get(name string) (*corev1.Secret, error) {
	return f.GetFunc(name)
}

func (f *fakeSecretLister) List(_ labels.Selector) ([]*corev1.Secret, error) {
	return nil, nil
}

type fakeConfigMapLister struct {
	corev1listers.ConfigMapNamespaceLister
	GetFunc func(name string) (*corev1.ConfigMap, error)
}

func (f *fakeConfigMapLister) Get(name string) (*corev1.ConfigMap, error) {
	return f.GetFunc(name)
}

func (f *fakeConfigMapLister) List(_ labels.Selector) ([]*corev1.ConfigMap, error) {
	return nil, nil
}

type fakeSecretClient struct {
	corev1client.SecretInterface
	CreateFunc func(ctx context.Context, s *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error)
	GetFunc    func(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Secret, error)
	UpdateFunc func(ctx context.Context, s *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error)
}

func (f *fakeSecretClient) Create(ctx context.Context, s *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error) {
	return f.CreateFunc(ctx, s, opts)
}

func (f *fakeSecretClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Secret, error) {
	return f.GetFunc(ctx, name, opts)
}

func (f *fakeSecretClient) Update(ctx context.Context, s *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
	return f.UpdateFunc(ctx, s, opts)
}

func noRefreshConfigMap() *fakeConfigMapLister {
	return &fakeConfigMapLister{
		GetFunc: func(name string) (*corev1.ConfigMap, error) {
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
		require.ErrorContains(t, err, "found 1 parts of token")
	})

	t.Run("empty body", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", strings.NewReader(""))

		_, _, err := v1parseBody(r)
		require.ErrorContains(t, err, "unexpected end of JSON input")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", strings.NewReader("{invalid"))

		_, _, err := v1parseBody(r)
		require.ErrorContains(t, err, "invalid character")
	})

	t.Run("wrong kind", func(t *testing.T) {
		t.Parallel()

		body := types.V1AuthnRequest{
			Kind: "NotTokenReview",
			Spec: types.V1AuthnRequestSpec{Token: "key:secret"},
		}
		data, err := json.Marshal(body)
		require.NoError(t, err)
		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", bytes.NewReader(data))

		_, _, err = v1parseBody(r)
		require.Error(t, err)
		assert.ErrorContains(t, err, "not TokenReview")
	})

	t.Run("empty token", func(t *testing.T) {
		t.Parallel()

		body := types.V1AuthnRequest{
			Kind: kubeapiauth.DefaultAuthnKind,
			Spec: types.V1AuthnRequestSpec{Token: ""},
		}
		data, err := json.Marshal(body)
		require.NoError(t, err)
		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", bytes.NewReader(data))

		_, _, err = v1parseBody(r)
		require.Error(t, err)
		assert.ErrorContains(t, err, "missing Token")
	})
}

func TestGetRefreshPeriod(t *testing.T) {
	t.Parallel()

	t.Run("configmap not found", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace:       testNamespace,
			configMapLister: noRefreshConfigMap(),
		}

		assert.Equal(t, time.Duration(-1), h.getRefreshPeriod())
	})

	t.Run("valid value", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace: testNamespace,
			configMapLister: &fakeConfigMapLister{
				GetFunc: func(name string) (*corev1.ConfigMap, error) {
					return &corev1.ConfigMap{Data: map[string]string{"value": "60"}}, nil
				},
			},
		}

		assert.Equal(t, 60*time.Second, h.getRefreshPeriod())
	})

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace: testNamespace,
			configMapLister: &fakeConfigMapLister{
				GetFunc: func(name string) (*corev1.ConfigMap, error) {
					return &corev1.ConfigMap{Data: map[string]string{"value": "0"}}, nil
				},
			},
		}

		assert.Equal(t, time.Duration(0), h.getRefreshPeriod())
	})

	t.Run("empty value", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace: testNamespace,
			configMapLister: &fakeConfigMapLister{
				GetFunc: func(name string) (*corev1.ConfigMap, error) {
					return &corev1.ConfigMap{Data: map[string]string{"value": ""}}, nil
				},
			},
		}

		assert.Equal(t, time.Duration(-1), h.getRefreshPeriod())
	})

	t.Run("non-numeric value", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace: testNamespace,
			configMapLister: &fakeConfigMapLister{
				GetFunc: func(name string) (*corev1.ConfigMap, error) {
					return &corev1.ConfigMap{Data: map[string]string{"value": "abc"}}, nil
				},
			},
		}

		assert.Equal(t, time.Duration(-1), h.getRefreshPeriod())
	})

	t.Run("nil data", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace: testNamespace,
			configMapLister: &fakeConfigMapLister{
				GetFunc: func(name string) (*corev1.ConfigMap, error) {
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

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return secret, nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return obj, nil
					},
				},
			}

			result, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)
			assert.Equal(t, []string{"group1", "group2"}, result.Groups)
		})
	})

	t.Run("token not found", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace: testNamespace,
			clusterAuthTokensCache: &fakeClusterAuthTokenCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return nil, notFound(name)
				},
			},
		}

		_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.True(t, apierrors.IsNotFound(err))
	})

	t.Run("token disabled", func(t *testing.T) {
		t.Parallel()

		token := newTestToken()
		token.Enabled = false

		h := &Authenticator{
			namespace: testNamespace,
			clusterAuthTokensCache: &fakeClusterAuthTokenCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return token, nil
				},
			},
		}

		_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.ErrorContains(t, err, "not enabled")
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace: testNamespace,
			clusterAuthTokensCache: &fakeClusterAuthTokenCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return newTestToken(), nil
				},
			},
			clusterUserAttributesCache: &fakeClusterUserAttributeCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return nil, notFound(name)
				},
			},
		}

		_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.True(t, apierrors.IsNotFound(err))
	})

	t.Run("user disabled", func(t *testing.T) {
		t.Parallel()

		user := newTestUser()
		user.Enabled = false

		h := &Authenticator{
			namespace: testNamespace,
			clusterAuthTokensCache: &fakeClusterAuthTokenCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return newTestToken(), nil
				},
			},
			clusterUserAttributesCache: &fakeClusterUserAttributeCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return user, nil
				},
			},
		}

		_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.ErrorContains(t, err, "not enabled")
	})

	t.Run("wrong secret key", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace: testNamespace,
			clusterAuthTokensCache: &fakeClusterAuthTokenCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return newTestToken(), nil
				},
			},
			clusterUserAttributesCache: &fakeClusterUserAttributeCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return newTestUser(), nil
				},
			},
			secretLister: &fakeSecretLister{
				GetFunc: func(name string) (*corev1.Secret, error) {
					return newTestSecret(), nil
				},
			},
		}

		_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, "wrong-secret")
		require.Error(t, err)
		assert.ErrorContains(t, err, "does not match")
	})

	t.Run("expired token", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()
			token.ExpiresAt = time.Now().Add(-time.Hour).Format(time.RFC3339)

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser(), nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
			}

			_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
			require.Error(t, err)
			assert.ErrorContains(t, err, "expired")
		})
	})

	t.Run("secret missing and no hash", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace: testNamespace,
			clusterAuthTokensCache: &fakeClusterAuthTokenCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return newTestToken(), nil
				},
			},
			clusterUserAttributesCache: &fakeClusterUserAttributeCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return newTestUser(), nil
				},
			},
			secretLister: &fakeSecretLister{
				GetFunc: func(name string) (*corev1.Secret, error) {
					return nil, notFound(name)
				},
			},
		}

		_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.ErrorContains(t, err, "missing")
	})

	t.Run("migration creates secret and clears token hash", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()
			token.SecretKeyHash = testSecretKeyHash //nolint:staticcheck

			var createdSecret *corev1.Secret
			var updatedToken *clusterv3.ClusterAuthToken

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser("admins"), nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return nil, notFound(name)
					},
				},
				secrets: &fakeSecretClient{
					CreateFunc: func(ctx context.Context, s *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error) {
						createdSecret = s
						return s, nil
					},
				},
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					GetFunc: func(ns, name string, opts metav1.GetOptions) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						updatedToken = obj.DeepCopy()
						return obj, nil
					},
				},
				configMapLister: noRefreshConfigMap(),
			}

			result, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)

			require.NotNil(t, createdSecret)
			assert.Equal(t, testSecretKeyHash, string(createdSecret.Data[clusterauth.ClusterAuthSecretHashField]))

			require.NotNil(t, updatedToken)
			assert.Empty(t, updatedToken.SecretKeyHash) //nolint:staticcheck
		})
	})

	t.Run("migration overwrites existing secret", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()
			token.SecretKeyHash = testSecretKeyHash //nolint:staticcheck

			var updatedSecretData map[string][]byte

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser(), nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return nil, notFound(name)
					},
				},
				secrets: &fakeSecretClient{
					CreateFunc: func(ctx context.Context, s *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error) {
						return nil, apierrors.NewAlreadyExists(schema.GroupResource{}, s.Name)
					},
					GetFunc: func(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Secret, error) {
						return &corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace},
							Data:       map[string][]byte{"hash": []byte("old-hash")},
						}, nil
					},
					UpdateFunc: func(ctx context.Context, s *corev1.Secret, opts metav1.UpdateOptions) (*corev1.Secret, error) {
						updatedSecretData = s.Data
						return s, nil
					},
				},
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					GetFunc: func(ns, name string, opts metav1.GetOptions) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return obj, nil
					},
				},
				configMapLister: noRefreshConfigMap(),
			}

			result, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)

			require.NotNil(t, updatedSecretData)
			assert.Equal(t, testSecretKeyHash, string(updatedSecretData[clusterauth.ClusterAuthSecretHashField]))
		})
	})

	t.Run("migration secret create fails", func(t *testing.T) {
		t.Parallel()

		token := newTestToken()
		token.SecretKeyHash = testSecretKeyHash //nolint:staticcheck

		h := &Authenticator{
			namespace: testNamespace,
			clusterAuthTokensCache: &fakeClusterAuthTokenCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return token, nil
				},
			},
			clusterUserAttributesCache: &fakeClusterUserAttributeCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return newTestUser(), nil
				},
			},
			secretLister: &fakeSecretLister{
				GetFunc: func(name string) (*corev1.Secret, error) {
					return nil, notFound(name)
				},
			},
			secrets: &fakeSecretClient{
				CreateFunc: func(ctx context.Context, s *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error) {
					return nil, fmt.Errorf("storage unavailable")
				},
			},
			clusterAuthTokens: &fakeClusterAuthTokenClient{
				GetFunc: func(ns, name string, opts metav1.GetOptions) (*clusterv3.ClusterAuthToken, error) {
					return token, nil
				},
			},
		}

		_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
		require.Error(t, err)
		assert.ErrorContains(t, err, "storage unavailable")
	})

	t.Run("migration token update fails", func(t *testing.T) {
		t.Parallel()

		token := newTestToken()
		token.SecretKeyHash = testSecretKeyHash //nolint:staticcheck

		h := &Authenticator{
			namespace: testNamespace,
			clusterAuthTokensCache: &fakeClusterAuthTokenCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return token, nil
				},
			},
			clusterUserAttributesCache: &fakeClusterUserAttributeCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
					return newTestUser(), nil
				},
			},
			secretLister: &fakeSecretLister{
				GetFunc: func(name string) (*corev1.Secret, error) {
					return nil, notFound(name)
				},
			},
			secrets: &fakeSecretClient{
				CreateFunc: func(ctx context.Context, s *corev1.Secret, opts metav1.CreateOptions) (*corev1.Secret, error) {
					return s, nil
				},
			},
			clusterAuthTokens: &fakeClusterAuthTokenClient{
				GetFunc: func(ns, name string, opts metav1.GetOptions) (*clusterv3.ClusterAuthToken, error) {
					return token, nil
				},
				UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
					return nil, fmt.Errorf("conflict")
				},
			},
		}

		_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
		require.Error(t, err)
	})

	t.Run("refresh triggered when overdue", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			user := newTestUser("editors")
			user.LastRefresh = time.Now().Add(-2 * time.Hour).Format(time.RFC3339)

			var userUpdated bool

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: &fakeConfigMapLister{
					GetFunc: func(name string) (*corev1.ConfigMap, error) {
						return &corev1.ConfigMap{Data: map[string]string{"value": "3600"}}, nil
					},
				},
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return obj, nil
					},
				},
				clusterUserAttributes: &fakeClusterUserAttributeClient{
					UpdateFunc: func(obj *clusterv3.ClusterUserAttribute) (*clusterv3.ClusterUserAttribute, error) {
						userUpdated = true
						assert.True(t, obj.NeedsRefresh)
						return obj, nil
					},
				},
			}

			result, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
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

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: &fakeConfigMapLister{
					GetFunc: func(name string) (*corev1.ConfigMap, error) {
						return &corev1.ConfigMap{Data: map[string]string{"value": "3600"}}, nil
					},
				},
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return obj, nil
					},
				},
			}

			result, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
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

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return obj, nil
					},
				},
			}

			result, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
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

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: &fakeConfigMapLister{
					GetFunc: func(name string) (*corev1.ConfigMap, error) {
						return &corev1.ConfigMap{Data: map[string]string{"value": "3600"}}, nil
					},
				},
			}

			_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
			require.Error(t, err)
			assert.ErrorContains(t, err, "parsing lastRefresh")
		})
	})

	t.Run("user attribute update failure during refresh is silent", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			user := newTestUser()
			user.LastRefresh = time.Now().Add(-2 * time.Hour).Format(time.RFC3339)

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return user, nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: &fakeConfigMapLister{
					GetFunc: func(name string) (*corev1.ConfigMap, error) {
						return &corev1.ConfigMap{Data: map[string]string{"value": "3600"}}, nil
					},
				},
				clusterUserAttributes: &fakeClusterUserAttributeClient{
					UpdateFunc: func(obj *clusterv3.ClusterUserAttribute) (*clusterv3.ClusterUserAttribute, error) {
						return nil, fmt.Errorf("update failed")
					},
				},
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return obj, nil
					},
				},
			}

			result, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.Equal(t, testUserName, result.UserName)
		})
	})

	t.Run("last used at updated on first use", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()
			var lastUsedAtSet bool

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser(), nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						lastUsedAtSet = true
						require.NotNil(t, obj.LastUsedAt)
						return obj, nil
					},
				},
			}

			_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
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

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser(), nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						tokenUpdateCalled = true
						return obj, nil
					},
				},
			}

			_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
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

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser(), nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						tokenUpdateCalled = true
						return obj, nil
					},
				},
			}

			_, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
			require.NoError(t, err)
			assert.True(t, tokenUpdateCalled)
		})
	})

	t.Run("last used at update failure is silent", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			token := newTestToken()

			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return token, nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser("admins"), nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return nil, fmt.Errorf("transient error")
					},
				},
			}

			result, err := h.v1getAndVerifyUser(t.Context(), testAccessKey, testSecretKey)
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
			h := &Authenticator{
				namespace: testNamespace,
				clusterAuthTokensCache: &fakeClusterAuthTokenCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
						return newTestToken(), nil
					},
				},
				clusterUserAttributesCache: &fakeClusterUserAttributeCache{
					GetFunc: func(ns, name string) (*clusterv3.ClusterUserAttribute, error) {
						return newTestUser("group1"), nil
					},
				},
				secretLister: &fakeSecretLister{
					GetFunc: func(name string) (*corev1.Secret, error) {
						return newTestSecret(), nil
					},
				},
				configMapLister: noRefreshConfigMap(),
				clusterAuthTokens: &fakeClusterAuthTokenClient{
					UpdateFunc: func(obj *clusterv3.ClusterAuthToken) (*clusterv3.ClusterAuthToken, error) {
						return obj, nil
					},
				},
			}

			w := httptest.NewRecorder()
			r := tokenReviewRequest(t, testAccessKey+":"+testSecretKey)

			h.Authenticate(w, r)

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

		h := &Authenticator{}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/v1/authenticate", strings.NewReader("not json"))

		h.Authenticate(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid credentials returns 401", func(t *testing.T) {
		t.Parallel()

		h := &Authenticator{
			namespace: testNamespace,
			clusterAuthTokensCache: &fakeClusterAuthTokenCache{
				GetFunc: func(ns, name string) (*clusterv3.ClusterAuthToken, error) {
					return nil, notFound(name)
				},
			},
		}

		w := httptest.NewRecorder()
		r := tokenReviewRequest(t, "unknown-token:secret")

		h.Authenticate(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
