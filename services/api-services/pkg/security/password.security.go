package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"

	"golang.org/x/crypto/argon2"
)

const (
	Argon2idVersion = argon2.Version
	SaltLength      = 16
)

// Argon2Params defines the parameters for Argon2id password hashing.
type Argon2Params struct {
	Memory  uint32
	Time    uint32
	Threads uint8
	KeyLen  uint32
}

// parsedHash represents the components of an Argon2id hashed password string.
type parsedHash struct {
	Version int
	Memory  uint32
	Time    uint32
	Threads uint8
	Salt    []byte
	Hash    []byte
}

// DefaultArgon2Params returns a recommended set of parameters for Argon2id.
func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:  64 * 1024,
		Time:    3,
		Threads: 4,
		KeyLen:  32,
	}
}

// HashPassword generates a secure hash of the provided password using Argon2id.
func HashPassword(password string, params *Argon2Params) (string, error) {
	if params == nil {
		params = DefaultArgon2Params()
	}

	salt := make([]byte, SaltLength)
	if _, err := rand.Read(salt); err != nil {
		logger.Error("Failed to generate salt for password hashing", logger.ErrorField(err))
		return "", fmt.Errorf("security: failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Time,
		params.Memory,
		params.Threads,
		params.KeyLen,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	hashedPassword := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		Argon2idVersion,
		params.Memory,
		params.Time,
		params.Threads,
		b64Salt,
		b64Hash,
	)

	return hashedPassword, nil
}

// parseArgon2HashString parses the Argon2id hash string into its constituent parts.
func parseArgon2HashString(hashStr string) (*parsedHash, error) {
	parts := strings.Split(hashStr, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		err := fmt.Errorf("security: invalid Argon2id hash format for string: %s", hashStr)
		logger.Warn("Invalid Argon2id hash format", logger.String("hash_prefix", hashStr[:min(len(hashStr), 20)]), logger.ErrorField(err))
		return nil, err
	}

	var parsed parsedHash

	if !strings.HasPrefix(parts[2], "v=") {
		err := fmt.Errorf("security: invalid Argon2id version format in hash string")
		logger.Warn("Invalid Argon2id version format", logger.String("version_part", parts[2]), logger.ErrorField(err))
		return nil, err
	}
	versionStr := strings.TrimPrefix(parts[2], "v=")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		err = fmt.Errorf("security: invalid Argon2id version value: %w", err)
		logger.Warn("Invalid Argon2id version value", logger.String("version_str", versionStr), logger.ErrorField(err))
		return nil, err
	}
	parsed.Version = version

	paramParts := strings.Split(parts[3], ",")
	if len(paramParts) != 3 {
		err := fmt.Errorf("security: invalid Argon2id parameter format in hash string")
		logger.Warn("Invalid Argon2id parameter format", logger.String("param_part", parts[3]), logger.ErrorField(err))
		return nil, err
	}

	if !strings.HasPrefix(paramParts[0], "m=") {
		err := fmt.Errorf("security: invalid Argon2id memory parameter format")
		logger.Warn("Invalid Argon2id memory parameter format", logger.String("memory_part", paramParts[0]), logger.ErrorField(err))
		return nil, err
	}
	memoryStr := strings.TrimPrefix(paramParts[0], "m=")
	memory, err := strconv.ParseUint(memoryStr, 10, 32)
	if err != nil {
		err = fmt.Errorf("security: invalid Argon2id memory value: %w", err)
		logger.Warn("Invalid Argon2id memory value", logger.String("memory_str", memoryStr), logger.ErrorField(err))
		return nil, err
	}
	parsed.Memory = uint32(memory)

	if !strings.HasPrefix(paramParts[1], "t=") {
		err := fmt.Errorf("security: invalid Argon2id time parameter format")
		logger.Warn("Invalid Argon2id time parameter format", logger.String("time_part", paramParts[1]), logger.ErrorField(err))
		return nil, err
	}
	timeStr := strings.TrimPrefix(paramParts[1], "t=")
	time, err := strconv.ParseUint(timeStr, 10, 32)
	if err != nil {
		err = fmt.Errorf("security: invalid Argon2id time value: %w", err)
		logger.Warn("Invalid Argon2id time value", logger.String("time_str", timeStr), logger.ErrorField(err))
		return nil, err
	}
	parsed.Time = uint32(time)

	if !strings.HasPrefix(paramParts[2], "p=") {
		err := fmt.Errorf("security: invalid Argon2id threads parameter format")
		logger.Warn("Invalid Argon2id threads parameter format", logger.String("threads_part", paramParts[2]), logger.ErrorField(err))
		return nil, err
	}
	threadsStr := strings.TrimPrefix(paramParts[2], "p=")
	threads, err := strconv.ParseUint(threadsStr, 10, 8)
	if err != nil {
		err = fmt.Errorf("security: invalid Argon2id threads value: %w", err)
		logger.Warn("Invalid Argon2id threads value", logger.String("threads_str", threadsStr), logger.ErrorField(err))
		return nil, err
	}
	parsed.Threads = uint8(threads)

	parsed.Salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		err = fmt.Errorf("security: failed to decode salt: %w", err)
		logger.Warn("Failed to decode salt from hash string", logger.String("salt_part", parts[4]), logger.ErrorField(err))
		return nil, err
	}

	parsed.Hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		err = fmt.Errorf("security: failed to decode hash: %w", err)
		logger.Warn("Failed to decode hash from hash string", logger.String("hash_part", parts[5]), logger.ErrorField(err))
		return nil, err
	}

	return &parsed, nil
}

// VerifyPassword checks if the provided password matches the stored Argon2id hash.
func VerifyPassword(storedHash string, password string) bool {
	parsed, err := parseArgon2HashString(storedHash)
	if err != nil {
		logger.Error("Failed to parse stored hash during password verification", logger.ErrorField(err))
		return false
	}

	if parsed.Version != Argon2idVersion {
		logger.Warn("Argon2id version mismatch during password verification",
			logger.Int("expected_version", Argon2idVersion),
			logger.Int("parsed_version", parsed.Version),
		)
		return false
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		parsed.Salt,
		parsed.Time,
		parsed.Memory,
		parsed.Threads,
		uint32(len(parsed.Hash)),
	)

	result := subtle.ConstantTimeCompare(parsed.Hash, computedHash) == 1

	if !result {
		logger.Debug("Password verification failed: hash mismatch",
			logger.String("stored_hash_prefix", storedHash[:min(len(storedHash), 20)]),
		)
	} else {
		logger.Debug("Password verification successful")
	}

	return result
}
