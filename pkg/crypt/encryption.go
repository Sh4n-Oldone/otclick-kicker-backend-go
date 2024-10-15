package crypt

import (
	"bytes"
	"errors"
	"golang.org/x/crypto/scrypt"
)

var (
	ErrPasswordsNotMatch = errors.New("password comparison failed")
)

func EncryptPassword(password, salt []byte) ([]byte, error) {
	hash, err := scrypt.Key(password, salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func ComparePasswordAndHash(password, hash, salt []byte) error {
	newHash, err := EncryptPassword(password, salt)
	if err != nil {
		return err
	}

	isEqual := bytes.Equal(newHash, hash)
	if !isEqual {
		return ErrPasswordsNotMatch
	}
	return nil
}
