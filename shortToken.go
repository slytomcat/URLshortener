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
	Get() string                // returns new random short token
	CheckLength(string) error   // true when the token length is correct
	CheckAlphabet(string) error // check the token alphabet
}
type shortToken struct {
	length  int // token length
	bufSize int // bytes buffer size
}

var (
	lengthError = errors.New("wrong token length")
)

// NewShortToken returns new ShortToken instance
func NewShortToken(length int) ShortToken {
	return &shortToken{
		length:  length,
		bufSize: length*6/8 + 1,
	}
}

// Get creates the token from random source
func (s *shortToken) Get() string {
	// prepare bytes buffer
	buf := make([]byte, s.bufSize)
	// get secure random bytes
	n, err := rand.Read(buf)
	if err != nil || n != s.bufSize {
		panic(fmt.Errorf("error while retrieving random data: %d %v", n, err.Error()))
	}
	// return shortened to tokenLenS BASE64 representation
	return base64.URLEncoding.EncodeToString(buf)[:s.length]
}

// Check checks the length of token
func (s *shortToken) CheckLength(sToken string) error {
	if len(sToken) == s.length {
		return nil
	}
	return lengthError
}

// Check checks the token alphabet
func (s *shortToken) CheckAlphabet(sToken string) error {
	// check base64 URL safe alphabet
	for i, s := range sToken {
		if !((s >= 'A' && s <= 'Z') || (s >= 'a' && s <= 'z') || (s >= '0' && s <= '9') || s == '_' || s == '-') {
			return base64.CorruptInputError(i)
		}
	}
	return nil
}
