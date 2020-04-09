package main

// ShortToken is string of the number of BASE64 symbols from the url safe alphabet.
// It represent 6*len(token) bits of data. ShortToken is not correct BASE64 data
// representation as number of bits is not always a multiple of 8 (1 byte).

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// ShortToken - interface for short token creation
type ShortToken interface {
	Get() (string, error)
	Check(string) error
}

// NewShortToken returns new ShortToken instance
func NewShortToken(len int) ShortToken {
	return &shortToken{
		len: len,
	}
}

type shortToken struct {
	len int // token length
}

// Get creates the token from random or debugging source
func (s *shortToken) Get() (string, error) {

	// prepare bytes bufer
	bLen := s.len*6/8 + 1
	buf := make([]byte, bLen)

	// get secure random bytes
	n, err := rand.Read(buf)
	if err != nil || n != bLen {
		return "", fmt.Errorf("error while retriving random data: %v", err.Error())
	}
	// return shortened to tokenLenS BASE64 representation
	return base64.URLEncoding.EncodeToString(buf)[:s.len], nil
}

func (s *shortToken) Check(sToken string) error {
	if len(sToken) != s.len {
		return errors.New("wrong token length")
	}

	// check base64 alphabet
	if _, err := base64.URLEncoding.DecodeString(sToken + "AAAA"[:4-len(sToken)%4]); err != nil {
		return errors.New("wrong token")
	}
	return nil
}
