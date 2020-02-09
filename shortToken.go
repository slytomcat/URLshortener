package main

// ShortToken is string of the number of BASE64 symbols from the url safe alphabet.
// It represent 6*len(token) bits of data. ShortToken is not correct BASE64 data
// representation as number of bits is not always a multiple of 8 (1 byte).

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"sync/atomic"
)

var (
	// debugMode - debugging mode. See SetShortTokenDebug for details
	debugMode int32
)

// SetShortTokenDebug sets debuging mode:
// 0 - normal operation: NewShortToken returns random token
// 1 - debugging mode: NewShortToken returns debug token
// -1 - error mode: NewShortToken returns error
func SetShortTokenDebug(mode int) {
	atomic.StoreInt32(&debugMode, int32(mode))
}

// NewShortToken creates the token (`length` BASE64 symbols) from random or debugging source
func NewShortToken(length int) (string, error) {
	switch atomic.LoadInt32(&debugMode) {
	case 1:
		return strings.Repeat("_", length), nil
	case -1:
		return "", errors.New("debug error")
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
