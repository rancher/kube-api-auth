package hashers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHasherForHash(t *testing.T) {
	const testSecret = "testsecret"
	scryptHash, err := ScryptHasher{}.CreateHash(testSecret)
	assert.NoError(t, err, "error when creating scrypt hash")
	sha256Hash, err := Sha256Hasher{}.CreateHash(testSecret)
	assert.NoError(t, err, "error when creating sha256 hash")
	sha3Hash, err := Sha3Hasher{}.CreateHash(testSecret)
	assert.NoError(t, err, "error when creating sha3 hash")

	tests := []struct {
		name       string
		hash       string
		wantHasher Hasher
		wantErr    bool
	}{
		{
			name:       "scrypt hash",
			hash:       scryptHash,
			wantHasher: ScryptHasher{},
		},
		{
			name:       "sha256 hash",
			hash:       sha256Hash,
			wantHasher: Sha256Hasher{},
		},
		{
			name:       "sha3 hash",
			hash:       sha3Hash,
			wantHasher: Sha3Hasher{},
		},
		{
			name:    "invalid hash",
			hash:    "thisisnotahash",
			wantErr: true,
		},
		{
			name:    "invalid hash version",
			hash:    "$4:some-salt-here:some-secret-here",
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			hasher, err := GetHasherForHash(test.hash)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.IsTypef(t, test.wantHasher, hasher, "did not get the expected hasher")
			}
		})
	}
}

func TestGetHasher(t *testing.T) {
	t.Parallel()
	assert.IsTypef(t, Sha3Hasher{}, GetHasher(), "expected SHA3 to be the default hasher")
}

func TestGetHashVersion(t *testing.T) {
	tests := []struct {
		name            string
		hash            string
		wantHashVersion HashVersion
		wantErr         bool
	}{
		{
			name:            "test valid hash",
			hash:            "$1:some-salt-here:some-secret-here",
			wantHashVersion: ScryptVersion,
		},
		{
			name:    "test bad hash format",
			hash:    "$1:some-secret-here",
			wantErr: true,
		},
		{
			name:    "test bad hash version",
			hash:    "$not-a-number:some-salt-here:some-secret-here",
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			hashVersion, err := GetHashVersion(test.hash)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.wantHashVersion, hashVersion)
			}
		})
	}
}
