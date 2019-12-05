package main

// ShortToken is string of 6 BASE64 symbols from the url safe alphabet.
// It represent 36 bits of data. ShortToken is not correct BASE64 data
// representation.
// If you need longer/shorten token length adjust tokenLenS

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

const (
	// tokenLenS - number of BASE64 symbols in token
	tokenLenS = 5
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
