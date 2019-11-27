package main

// ShortToken is string of 6 BASE64 symbols from the url safe alphabet.
// It represent 36 bits of data. ShortToken is not correct BASE64 data
// representation. It should be extended by 'A=' (that provides additional
// 4 zero bits) to decode into 5 bytes with 4 less significant bits equal to zero.

import (
	"crypto/rand"
	"encoding/base64"
)

var (
	// DEBUG = true sets token as constant
	DEBUG = false
)

// ShortTokenNew -
func ShortTokenNew() (string, error) {

	b := make([]byte, 5)
	if DEBUG {
		b[0], b[1], b[2], b[3], b[4] = 0xff, 0xff, 0xff, 0xff, 0xf0
	} else {
		_, err := rand.Read(b) // get 5 secure random bytes
		if err != nil {
			return "", err
		}
		b[4] &= 0xf0 // zero last 4 bits as we need only 36 bits that equal to 6 BASE64 symbols
	}

	// shorten to 6 symbols without last 4 zero bits and padding ('A=')
	return base64.URLEncoding.EncodeToString(b)[:6], nil
}
