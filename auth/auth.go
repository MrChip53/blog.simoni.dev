package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

type Argon2Params struct {
	argon2id    bool
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func HashPassword(password string) (string, error) {
	params := &Argon2Params{
		argon2id:    true,
		memory:      128 * 1024,
		iterations:  15,
		parallelism: 6,
		saltLength:  64,
		keyLength:   128,
	}

	salt := make([]byte, params.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash, err := hashPassword(password, salt, params)
	if err != nil {
		return "", err
	}

	return formatHash(params, salt, hash), nil
}

func VerifyPassword(password string, hash string) (bool, error) {
	passwordBytes, saltBytes, params, err := parseKey(hash)
	if err != nil {
		return false, err
	}

	cmpHash, err := hashPassword(password, saltBytes, params)
	if err != nil {
		return false, err
	}

	return bytes.Equal(cmpHash, passwordBytes), nil
}

func formatHash(params *Argon2Params, salt []byte, hash []byte) string {
	argonStr := "argon2id"
	if !params.argon2id {
		argonStr = "argon2"
	}

	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s$%s",
		argonStr,
		argon2.Version,
		params.memory,
		params.iterations,
		params.parallelism,
		base64.RawStdEncoding.Strict().EncodeToString(salt),
		base64.RawStdEncoding.Strict().EncodeToString(hash),
		"simoni.dev",
	)
}

func hashPassword(password string, salt []byte, params *Argon2Params) ([]byte, error) {
	if params.argon2id {
		return argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength), nil
	}

	return argon2.Key([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength), nil
}

func parseKey(hash string) (password []byte, salt []byte, params *Argon2Params, err error) {
	parts := strings.Split(hash, "$")

	var version uint
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, nil
	}

	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible version of argon2")
	}

	splitCosts := strings.Split(parts[3], ",")

	var memory uint32
	_, err = fmt.Sscanf(splitCosts[0], "m=%d", &memory)
	if err != nil {
		return nil, nil, nil, err
	}

	var timeCost uint32
	_, err = fmt.Sscanf(splitCosts[1], "t=%d", &timeCost)
	if err != nil {
		return nil, nil, nil, err
	}

	var threads uint8
	_, err = fmt.Sscanf(splitCosts[2], "p=%d", &threads)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}

	password, err = base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}

	params = &Argon2Params{
		argon2id:    parts[1] == "argon2id",
		memory:      memory,
		iterations:  timeCost,
		parallelism: threads,
		saltLength:  uint32(len(salt)),
		keyLength:   uint32(len(password)),
	}

	return password, salt, params, nil
}
