package main

// ShortToken supports two representation
// - bytes array: 5 bytes with 4 less significant bits always equal to zero (36 bits)
// - BASE64 encoded string: 6 BASE64 symbols from the url safe alphabet (36 bits)
//
// The BASE64 encoded string should be extended by 'A=' (that provides fixed 4 zero bits) to decode to 5 bytes

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

var (
	// DEBUG = true sets token as constant
	DEBUG = false
)

// ShortToken is a structure with 2 representation of token
type ShortToken struct {
	Bytes  []byte // bytes array token representation
	BASE64 string // BASE64 string token representation
}

// ShortTokenNew - creates, fills and returns new ShortToken from random source
func ShortTokenNew() (*ShortToken, error) {

	b := make([]byte, 5)
	if DEBUG {
		b[0], b[1], b[2], b[3], b[4] = 0xff, 0xff, 0xff, 0xff, 0xf0
	} else {
		_, err := rand.Read(b) // get 5 secure random bytes
		if err != nil {
			return nil, err
		}
		b[4] &= 0xf0 // zero last 4 bits as we need only 36 bits that equal to 6 BASE64 symbols
	}

	tc := ShortToken{
		Bytes: b,
		// as last 4 bit of token is always zero, the base64 encoded token always ends with "A=" (6 zero bits and 2-bits padding)
		BASE64: base64.URLEncoding.EncodeToString(b)[:6], // shorten to 6 symbols without padding
	}

	return &tc, nil
}

// ShortTokenSet creates, fills and returns new ShortToken from BASE64 encoded string
func ShortTokenSet(sToken string) (*ShortToken, error) {

	if len(sToken) != 6 {
		return nil, errors.New("wrong token length")
	}

	// add "A=" (6 zeros bits and 2-bits padding) to receive 5 bytes from decoding
	b, err := base64.URLEncoding.DecodeString(sToken + "A=")
	if err != nil {
		return nil, err
	}

	tc := ShortToken{
		BASE64: sToken,
		Bytes:  b,
	}

	return &tc, nil
}
