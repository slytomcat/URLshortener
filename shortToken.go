package main

// ShortToken is string of the number of BASE64 symbols from the url safe alphabet.
// It represent 6*len(token) bits of data. ShortToken is not correct BASE64 data
// representation as number of bits is not always a multiple of 8 (1 byte).

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
)

// ShortToken - interface for short token creation
type ShortToken interface {
	Get() string        // returns new random short token
	Check(string) error // check the token length and alphabet
}

// NewShortToken returns new ShortToken instance
func NewShortToken(length int) ShortToken {
	pool := sync.Pool{}
	pool.New = func() interface{} {
		return make([]byte, length*6/8+1)
	}
	return &shortToken{
		length:  length,
		bufPool: pool,
	}
}

type shortToken struct {
	length  int       // token length
	bufPool sync.Pool // buffer pool for random bytes
}

// Get creates the token from random or debugging source
func (s *shortToken) Get() string {
	// get secure random bytes
	buf := s.bufPool.Get().([]byte)
	defer s.bufPool.Put(buf)

	n, err := rand.Read(buf)
	if err != nil || n != len(buf) {
		panic(fmt.Errorf("error while retriving random data: %d %v", n, err.Error()))
	}
	// return shortened to tokenLenS BASE64 representation
	return base64.URLEncoding.EncodeToString(buf)[:s.length]
}

// Check checks the lenght of token and its alphabet
func (s *shortToken) Check(sToken string) error {

	// check lenght
	if len(sToken) != s.length {
		return errors.New("wrong token length")
	}

	// check base64 alphabet
	if _, err := base64.URLEncoding.DecodeString(sToken + "AAAA"[:4-s.length%4]); err != nil {
		return errors.New("wrong token alphabet")
	}
	return nil
}
