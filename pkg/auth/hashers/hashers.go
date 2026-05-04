package hashers

import (
	"fmt"
	"strconv"
	"strings"
)

type HashVersion int

const (
	ScryptVersion HashVersion = iota + 1
	SHA256Version
	SHA3Version
)

type Hasher interface {
	CreateHash(secretKey string) (string, error)
	VerifyHash(hash, secretKey string) error
}

func GetHasherForHash(hash string) (Hasher, error) {
	version, err := GetHashVersion(hash)
	if err != nil {
		return nil, fmt.Errorf("unable to determine version for hash, %w", err)
	}
	switch HashVersion(version) {
	case ScryptVersion:
		return ScryptHasher{}, nil
	case SHA256Version:
		return Sha256Hasher{}, nil
	case SHA3Version:
		return Sha3Hasher{}, nil
	default:
		return nil, fmt.Errorf("invalid version %d, no hasher exists for that version", version)
	}
}

func GetHasher() Hasher {
	return Sha3Hasher{}
}

func GetHashVersion(hash string) (HashVersion, error) {
	splitHash := strings.SplitN(strings.TrimPrefix(hash, "$"), ":", 3)
	if len(splitHash) != 3 {
		return 0, fmt.Errorf("hash format invalid")
	}
	version, err := strconv.Atoi(splitHash[0])
	if err != nil {
		return 0, fmt.Errorf("unable to convert hash version")
	}
	return HashVersion(version), nil
}
