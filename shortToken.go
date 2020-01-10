package main

// ShortToken is string of tokenLenS BASE64 symbols from the url safe alphabet.
// It represent 6*tokenLenS bits of data. ShortToken is not correct BASE64 data
// representation as number of bits is not always a multiple of 8 (1 byte).

// If you need longer/shorten token length adjust tokenLenS constant and `token`
// field length in schema.sql file

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
)

const (
	// tokenLenS - number of BASE64 symbols in token
	tokenLenS = 6
)

var (
	// DEBUG = true sets token as constant DEBUGToken
	DEBUG = false
	// DEBUGToken is token that returned in debug mode
	DEBUGToken = strings.Repeat("_", tokenLenS)
)

// NewShortToken creates the token (tokenLength BASE64 symbols) from random or debugging source
func NewShortToken() (string, error) {

	if DEBUG {
		if DEBUGToken == "error" {
			return "", errors.New("debug error")
		}
		return DEBUGToken, nil
	}

	buf := make([]byte, tokenLenS*6/8+1)

	_, err := rand.Read(buf) // get secure random bytes
	if err != nil {
		return "", err
	}

	// shorten BASE64 representation to tokenLenS symbols
	return base64.URLEncoding.EncodeToString(buf)[:tokenLenS], nil
}
