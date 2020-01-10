package main

// ShortToken is string of the number of BASE64 symbols from the url safe alphabet.
// It represent 6*len(token) bits of data. ShortToken is not correct BASE64 data
// representation as number of bits is not always a multiple of 8 (1 byte).

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
)

var (
	// DEBUG = true sets token as constant DEBUGToken
	DEBUG = false
	// DEBUGToken is token that returned in debug mode
	DEBUGToken = ""
)

// NewShortToken creates the token (`length` BASE64 symbols) from random or debugging source
func NewShortToken(length int) (string, error) {

	if DEBUG {
		if DEBUGToken == "error" {
			DEBUGToken = strings.Repeat("_", length)
			// handle debugging error
			return "", errors.New("debug error")
		}
		// setub debuggig token
		if DEBUGToken == "" {
			DEBUGToken = strings.Repeat("_", length)
		}
		return DEBUGToken, nil
	}

	buf := make([]byte, length*6/8+1)

	// get secure random bytes
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	// shorten BASE64 representation to tokenLenS symbols
	return base64.URLEncoding.EncodeToString(buf)[:length], nil
}
