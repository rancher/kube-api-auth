package hashers

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	sha256HashFormat = "$%d:%s:%s"
)

type Sha256Hasher struct{}

func (s Sha256Hasher) CreateHash(secretKey string) (string, error) {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s", salt, secretKey)))
	encSalt := base64.RawStdEncoding.EncodeToString(salt)
	encKey := base64.RawStdEncoding.EncodeToString(hash[:])
	return fmt.Sprintf(sha256HashFormat, SHA256Version, encSalt, encKey), nil
}

func (s Sha256Hasher) VerifyHash(hash, secretKey string) error {
	if !strings.HasPrefix(hash, "$") {
		return errors.New("hash format invalid")
	}
	splitHash := strings.SplitN(strings.TrimPrefix(hash, "$"), ":", 3)
	if len(splitHash) != 3 {
		return errors.New("hash format invalid")
	}

	version, err := strconv.Atoi(splitHash[0])
	if err != nil {
		return err
	}
	if HashVersion(version) != SHA256Version {
		return fmt.Errorf("hash version %d does not match package version %d", version, SHA256Version)
	}

	salt, enc := splitHash[1], splitHash[2]
	decodedKey, err := base64.RawStdEncoding.DecodeString(enc)
	if err != nil {
		return err
	}
	if len(decodedKey) < 1 {
		return errors.New("secretKey hash does not match")
	}
	decodedSalt, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return err
	}
	hashedSecretKey := sha256.Sum256([]byte(string(decodedSalt) + secretKey))
	if subtle.ConstantTimeCompare(decodedKey, hashedSecretKey[:]) == 0 {
		return errors.New("secretKey hash does not match")
	}
	return nil
}
