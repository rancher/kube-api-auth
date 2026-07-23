package clusterauth

import (
	"strings"
	"testing"
	"time"

	"github.com/rancher/kube-api-auth/pkg/auth/hashers"
	clusterv3 "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testNamespace = "test-namespace"

func newTestClusterAuthToken(name, userID string) *clusterv3.ClusterAuthToken {
	return &clusterv3.ClusterAuthToken{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "ClusterAuthToken",
		},
		UserName: userID,
		Enabled:  true,
	}
}

func TestValidUser(t *testing.T) {
	t.Parallel()
	secretKey := strings.Repeat("A", 72)
	hashedValue, err := hashers.GetHasher().CreateHash(secretKey)
	assert.NoError(t, err)

	token := newTestClusterAuthToken("test-token", "me")
	secret := NewClusterAuthTokenSecretForName(testNamespace, token.Name, hashedValue)

	migrate, verifyErr := VerifyClusterAuthToken(secretKey, token, secret)
	assert.NoError(t, verifyErr)
	assert.False(t, migrate)
}

func TestUnmigrated(t *testing.T) {
	t.Parallel()
	secretKey := strings.Repeat("A", 72)
	hashedValue, err := hashers.GetHasher().CreateHash(secretKey)
	assert.NoError(t, err)

	token := newTestClusterAuthToken("test-token", "me")
	token.SecretKeyHash = hashedValue //nolint:staticcheck

	migrate, verifyErr := VerifyClusterAuthToken(secretKey, token, nil)
	assert.NoError(t, verifyErr)
	assert.True(t, migrate)
}

func TestMissingSecret(t *testing.T) {
	t.Parallel()
	token := newTestClusterAuthToken("test-token", "me")

	migrate, verifyErr := VerifyClusterAuthToken("anything", token, nil)
	assert.Error(t, verifyErr)
	assert.False(t, migrate)
}

func TestInvalidPassword(t *testing.T) {
	t.Parallel()
	secretKey := strings.Repeat("A", 72)
	hashedValue, err := hashers.GetHasher().CreateHash(secretKey)
	assert.NoError(t, err)

	token := newTestClusterAuthToken("test-token", "me")
	secret := NewClusterAuthTokenSecretForName(testNamespace, token.Name, hashedValue)

	migrate, verifyErr := VerifyClusterAuthToken(secretKey+":wrong!", token, secret)
	assert.Error(t, verifyErr)
	assert.False(t, migrate)
}

func TestExpired(t *testing.T) {
	t.Parallel()
	secretKey := strings.Repeat("A", 72)
	hashedValue, err := hashers.GetHasher().CreateHash(secretKey)
	assert.NoError(t, err)

	token := newTestClusterAuthToken("test-token", "me")
	token.ExpiresAt = time.Now().Add(-time.Minute).Format(time.RFC3339)
	secret := NewClusterAuthTokenSecretForName(testNamespace, token.Name, hashedValue)

	migrate, verifyErr := VerifyClusterAuthToken(secretKey, token, secret)
	assert.Error(t, verifyErr)
	assert.False(t, migrate)
}

func TestNotExpired(t *testing.T) {
	t.Parallel()
	secretKey := strings.Repeat("A", 72)
	hashedValue, err := hashers.GetHasher().CreateHash(secretKey)
	assert.NoError(t, err)

	token := newTestClusterAuthToken("test-token", "me")
	token.ExpiresAt = time.Now().Add(time.Minute).Format(time.RFC3339)
	secret := NewClusterAuthTokenSecretForName(testNamespace, token.Name, hashedValue)

	migrate, verifyErr := VerifyClusterAuthToken(secretKey, token, secret)
	assert.NoError(t, verifyErr)
	assert.False(t, migrate)
}

func TestUnknownHashVersion(t *testing.T) {
	t.Parallel()
	token := newTestClusterAuthToken("test-token", "me")
	secret := NewClusterAuthTokenSecretForName(testNamespace, token.Name, "$9:not-a-real-version:salt:hash")

	migrate, verifyErr := VerifyClusterAuthToken("anything", token, secret)
	assert.Error(t, verifyErr)
	assert.ErrorContains(t, verifyErr, "unable to get hasher")
	assert.False(t, migrate)
}

func TestMalformedHashFormat(t *testing.T) {
	t.Parallel()
	token := newTestClusterAuthToken("test-token", "me")
	secret := NewClusterAuthTokenSecretForName(testNamespace, token.Name, "not-a-valid-hash-format")

	migrate, verifyErr := VerifyClusterAuthToken("anything", token, secret)
	assert.Error(t, verifyErr)
	assert.False(t, migrate)
}

func TestInvalidExpiresAt(t *testing.T) {
	t.Parallel()
	secretKey := strings.Repeat("A", 72)
	hashedValue, err := hashers.GetHasher().CreateHash(secretKey)
	assert.NoError(t, err)

	token := newTestClusterAuthToken("test-token", "me")
	token.ExpiresAt = "some invalid time stamp"
	secret := NewClusterAuthTokenSecretForName(testNamespace, token.Name, hashedValue)

	migrate, verifyErr := VerifyClusterAuthToken(secretKey, token, secret)
	assert.Error(t, verifyErr)
	assert.False(t, migrate)
}
