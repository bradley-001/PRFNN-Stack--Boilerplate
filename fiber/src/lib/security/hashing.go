package security

import (
	"api/src/lib/general"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Hash512Result struct {
	HashHex string  `json:"cipher_text"`
	Salt    *string `json:"salt,omitempty"`
}

var pepper = general.GetEnv("HASH_PEPPER", "")
var cost = general.GetEnv("BCRYPT_COST", 16)

func Hash512(text string, salt *string) (Hash512Result, error) {
	// Takes in text, and an optional salt
	// If salt does not exist/ is not passed in, generate one and include in output
	// Otherwise, salt will return as nil

	// Hash standard is 512-bit
	// Salt is a 256-bit random value

	if pepper == "" {
		return Hash512Result{}, errors.New("no HASH_PEPPER specified in .env")
	}

	var saltBytes []byte
	var returnSalt *string

	if salt == nil || *salt == "" {
		// If there is no salt provided

		saltBytes = make([]byte, 32)

		_, err := rand.Read(saltBytes)
		if err != nil {
			// This should never error. It's an OS call
			return Hash512Result{}, errors.New("byte rand.Read failure of saltBytes")
		}

		saltHex := hex.EncodeToString(saltBytes)
		returnSalt = &saltHex
	} else {
		var err error
		saltBytes, err = hex.DecodeString(*salt)
		if err != nil {
			// Salt is likely in an invalid format.
			return Hash512Result{}, errors.New("salt is incorrectly formatted")
		}

		if len(saltBytes) != 32 {
			return Hash512Result{}, errors.New("salt is not 32 bytes")
		}

		returnSalt = nil

	}

	combined := append([]byte(text), saltBytes...)

	// Hash using SHA-512
	hash := sha512.Sum512(combined)
	hashHex := hex.EncodeToString(hash[:])

	return Hash512Result{HashHex: hashHex, Salt: returnSalt}, nil
}

func CheckHash512(text string, hashHex string, salt string) (bool, error) {
	a, err := Hash512(text, &salt)
	if err != nil {
		return false, err
	}

	if a.HashHex == hashHex {
		return true, nil
	}

	return false, nil

}

func HashBcrypt(text string) (string, error) {
	if cost == 0 {
		return "", errors.New("no BCRYPT_COST specified in .env")
	}

	if pepper == "" {
		return "", errors.New("no HASH_PEPPER specified in .env")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(text+pepper), cost)
	if err != nil {
		return "", errors.New("failed to hash password")
	}

	return string(hash), nil
}

func CheckHashBcrypt(text string, hash string) (bool, error) {
	if pepper == "" {
		return false, errors.New("no HASH_PEPPER specified in .env")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(text+pepper))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
