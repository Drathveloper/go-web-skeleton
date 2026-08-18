package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

var (
	// ErrMalformedPasswordHash is returned when a stored password cannot be parsed as
	// a PHC string. It is a data integrity failure, not a wrong password: callers must
	// keep the two apart, since only a mismatch is a failed authentication attempt.
	ErrMalformedPasswordHash = errors.New("malformed password hash")

	// ErrUnsupportedPasswordHash is returned when a stored password is a well formed
	// PHC string produced by an algorithm or an argon2 version this build cannot verify.
	ErrUnsupportedPasswordHash = errors.New("unsupported password hash")
)

const decodePasswordHashErrMsg = "decode password hash failed"

// decoySecretLength is the entropy behind the decoy hash; it only has to be
// unguessable enough that no real password could ever match it.
const decoySecretLength = 32

// Argon2id cost parameters applied when hashing a new password. They are the current
// values only: every hash carries the parameters it was produced with inside its PHC
// string, so raising the cost here leaves already stored passwords verifiable.
const (
	argon2HashSize     = 64
	argon2HashSaltSize = 16
	argon2HashMemory   = 64 * 1024
	argon2HashTime     = 1
	argonCpus          = 1
)

const (
	// argon2idVariant is the only PHC identifier this package produces and accepts.
	argon2idVariant = "argon2id"

	// phcFieldCount is the number of `$` separated fields of
	// `$argon2id$v=19$m=<memory>,t=<time>,p=<parallelism>$<salt>$<hash>`, counting the
	// empty one before the leading separator.
	phcFieldCount = 6

	phcVariantField = 1
	phcVersionField = 2
	phcParamsField  = 3
	phcSaltField    = 4
	phcHashField    = 5

	// maxArgon2HashSize bounds the digest a stored PHC string may declare, so a crafted
	// value cannot make verification allocate an unreasonable amount of memory.
	maxArgon2HashSize = 1024
)

// argon2idParams are the cost parameters of one single hash: the ones read back from a
// stored PHC string when verifying, the current ones when hashing.
type argon2idParams struct {
	memory      uint32
	time        uint32
	keyLength   uint32
	parallelism uint8
}

func currentArgon2idParams() argon2idParams {
	return argon2idParams{
		memory:      argon2HashMemory,
		time:        argon2HashTime,
		keyLength:   argon2HashSize,
		parallelism: argonCpus,
	}
}

// hashPassword derives the password with the current cost parameters and returns the
// PHC encoding of salt, digest and the parameters used to produce it.
func hashPassword(plaintextPassword string, salt []byte) string {
	params := currentArgon2idParams()
	hash := argon2.IDKey(
		[]byte(plaintextPassword),
		salt,
		params.time,
		params.memory,
		params.parallelism,
		params.keyLength)
	return encodeArgon2idHash(params, salt, hash)
}

// verifyPassword re-derives the password with the parameters stored in encodedHash —
// not with the current constants — so hashes produced before a cost bump still verify.
// A malformed hash is reported as an error and never as a mismatch.
func verifyPassword(plaintextPassword, encodedHash string) (bool, error) {
	params, salt, hash, err := decodeArgon2idHash(encodedHash)
	if err != nil {
		return false, err
	}
	inputHash := argon2.IDKey(
		[]byte(plaintextPassword),
		salt,
		params.time,
		params.memory,
		params.parallelism,
		params.keyLength)
	return subtle.ConstantTimeCompare(inputHash, hash) == 1, nil
}

func encodeArgon2idHash(params argon2idParams, salt, hash []byte) string {
	return fmt.Sprintf(
		"$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2idVariant,
		argon2.Version,
		params.memory,
		params.time,
		params.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))
}

func decodeArgon2idHash(encodedHash string) (argon2idParams, []byte, []byte, error) {
	fields := strings.Split(encodedHash, "$")
	if len(fields) != phcFieldCount || fields[0] != "" {
		return argon2idParams{}, nil, nil, wrapDecodePasswordHashErr(ErrMalformedPasswordHash)
	}
	params, err := decodeArgon2idParams(fields[phcVariantField], fields[phcVersionField], fields[phcParamsField])
	if err != nil {
		return argon2idParams{}, nil, nil, err
	}
	salt, err := decodePHCField(fields[phcSaltField])
	if err != nil {
		return argon2idParams{}, nil, nil, err
	}
	hash, err := decodePHCField(fields[phcHashField])
	if err != nil {
		return argon2idParams{}, nil, nil, err
	}
	if len(hash) > maxArgon2HashSize {
		return argon2idParams{}, nil, nil, wrapDecodePasswordHashErr(ErrMalformedPasswordHash)
	}
	params.keyLength = uint32(len(hash)) //nolint:gosec // len(hash) is bounded by maxArgon2HashSize right above
	return params, salt, hash, nil
}

// decodeArgon2idParams reads back the cost the stored hash was produced with. Zero costs
// are rejected here because argon2.IDKey panics on them, and a stored hash is untrusted
// input as soon as the database is.
func decodeArgon2idParams(variant, versionField, costField string) (argon2idParams, error) {
	if variant != argon2idVariant {
		return argon2idParams{}, wrapDecodePasswordHashErr(ErrUnsupportedPasswordHash)
	}
	var version int
	if _, err := fmt.Sscanf(versionField, "v=%d", &version); err != nil {
		return argon2idParams{}, wrapDecodePasswordHashErr(ErrMalformedPasswordHash)
	}
	if version != argon2.Version {
		return argon2idParams{}, wrapDecodePasswordHashErr(ErrUnsupportedPasswordHash)
	}
	var params argon2idParams
	if _, err := fmt.Sscanf(
		costField, "m=%d,t=%d,p=%d", &params.memory, &params.time, &params.parallelism); err != nil {
		return argon2idParams{}, wrapDecodePasswordHashErr(ErrMalformedPasswordHash)
	}
	if params.memory == 0 || params.time == 0 || params.parallelism == 0 {
		return argon2idParams{}, wrapDecodePasswordHashErr(ErrMalformedPasswordHash)
	}
	return params, nil
}

func decodePHCField(field string) ([]byte, error) {
	value, err := base64.RawStdEncoding.Strict().DecodeString(field)
	if err != nil || len(value) == 0 {
		return nil, wrapDecodePasswordHashErr(ErrMalformedPasswordHash)
	}
	return value, nil
}

func wrapDecodePasswordHashErr(err error) error {
	return fmt.Errorf(constants.DefaultWrappedErrorTemplate, decodePasswordHashErrMsg, err)
}

// enumerationDecoyHash is a valid PHC hash of a random password, computed once
// at startup. Login verifies against it when the username does not exist, so
// an unknown account costs the same KDF work as a known one and cannot be
// distinguished by response time.
var enumerationDecoyHash = mustGenerateDecoyHash() //nolint:gochecknoglobals

func mustGenerateDecoyHash() string {
	secret := make([]byte, decoySecretLength)
	if _, err := rand.Read(secret); err != nil {
		panic("generate password enumeration decoy failed: " + err.Error())
	}
	salt := make([]byte, argon2HashSaltSize)
	if _, err := rand.Read(salt); err != nil {
		panic("generate password enumeration decoy failed: " + err.Error())
	}
	return hashPassword(base64.RawStdEncoding.EncodeToString(secret), salt)
}
