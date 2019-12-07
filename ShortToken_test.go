package main

import (
	"strings"
	"testing"
)

// try to create new token from debugging source
func Test05NewShortTokenFake(t *testing.T) {
	DEBUG = true
	defer func() { DEBUG = true }()
	tc, err := NewShortToken()
	if err != nil {
		t.Error("error of token creation from debug source:", err)
	}

	if tc != strings.Repeat("_", tokenLenS) {
		t.Error("wrong BASE64 representation of debug token")
	}
}

// try to make two tokens from random source and compare them
func Test07NewShortTokenReal(t *testing.T) {
	DEBUG = false
	tc, err := NewShortToken()
	if err != nil {
		t.Error("error of ShortToken creation from random:", err)
	}

	tc1, _ := NewShortToken()

	if len(tc) != len(tc1) || len(tc) != tokenLenS {
		t.Error("wrong token length")
	}

	if tc == tc1 {
		t.Error("2 sequential token are equal")
	}

	for i := 0; i < 300; i++ {
		if tci, err := NewShortToken(); err != nil {
			t.Errorf("error of %d token creation: %s\n", i, err)
			if tci == tc {
				t.Error("2 tokens are equal")
			}
		}
	}
}
