package hashers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

const (
	sha3HashFormat  = "$%d:%d:%s:%s"
	sha512variation = 1
	saltLength      = 8
)

type Sha3Hasher struct{}

func (s Sha3Hasher) CreateHash(secretKey string) (string, error) {
	secretKeyBytes := []byte(secretKey)
	hashInput := make([]byte, saltLength, saltLength+len(secretKeyBytes))
	_, err := rand.Read(hashInput)
	if err != nil {
		return "", fmt.Errorf("unable to read random values for salt: %w", err)
	}
	hashInput = append(hashInput, secretKeyBytes...)
	hash := sha3.Sum512(hashInput)
	encSalt := base64.RawStdEncoding.EncodeToString(hashInput[:saltLength])
	encKey := base64.RawStdEncoding.EncodeToString(hash[:])
	return fmt.Sprintf(sha3HashFormat, SHA3Version, sha512variation, encSalt, encKey), nil
}

func (s Sha3Hasher) VerifyHash(hash, secretKey string) error {
	if !strings.HasPrefix(hash, "$") {
		return fmt.Errorf("hash format invalid")
	}
	splitHash := strings.Split(strings.TrimPrefix(hash, "$"), ":")
	if len(splitHash) != 4 {
		return fmt.Errorf("hash format invalid")
	}

	version, err := strconv.Atoi(splitHash[0])
	if err != nil {
		return err
	}
	if HashVersion(version) != SHA3Version {
		return fmt.Errorf("hash version %d does not match package version %d", version, SHA3Version)
	}

	variationVersion, err := strconv.Atoi(splitHash[1])
	if err != nil {
		return fmt.Errorf("unable to convert hash variation to an int")
	}
	if variationVersion != sha512variation {
		return fmt.Errorf("sha3 variation %d is not a known variation: [%d]", variationVersion, sha512variation)
	}

	salt, enc := splitHash[2], splitHash[3]
	decodedKey, err := base64.RawStdEncoding.DecodeString(enc)
	if err != nil {
		return err
	}
	if len(decodedKey) < 1 {
		return fmt.Errorf("secretKey hash does not match")
	}
	decodedSalt, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return err
	}
	hashedSecretKey := sha3.Sum512([]byte(fmt.Sprintf("%s%s", string(decodedSalt), secretKey)))
	if subtle.ConstantTimeCompare(decodedKey, hashedSecretKey[:]) == 0 {
		return fmt.Errorf("secretKey hash does not match")
	}
	return nil
}
