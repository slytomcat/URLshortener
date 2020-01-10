package main

import (
	"strings"
	"testing"
)

// try to create new token from debugging source
func Test05NewShortTokenFake(t *testing.T) {
	DEBUG = true
	DEBUGToken = strings.Repeat("_", 6)
	tc, err := NewShortToken(6)
	if err != nil {
		t.Error("error of ShortToken creation from debug source:", err)
	}
	if tc != DEBUGToken {
		t.Errorf("wrong token BASE64 representation: expected: '%s', received '%s'", DEBUGToken, tc)
	}
}

// try to make two tokens from random source and compare them
func Test07NewShortTokenReal(t *testing.T) {
	DEBUG = false
	tc, err := NewShortToken(6)
	if err != nil {
		t.Error("error of ShortToken creation from random:", err)
	}

	tc1, _ := NewShortToken(6)

	if len(tc) != len(tc1) || len(tc) != 6 {
		t.Error("wrong token length")
	}

	if tc == tc1 {
		t.Errorf("2 sequential token are equal by BASE64: '%s' == '%s'", tc, tc1)
	}
}

// try to make two tokens from random source and compare them
func Test07NewShortTokenReal2(t *testing.T) {
	DEBUG = false
	tc, err := NewShortToken(2)
	if err != nil {
		t.Error("error of ShortToken creation from random:", err)
	}

	tc1, _ := NewShortToken(2)

	if len(tc) != len(tc1) || len(tc) != 2 {
		t.Error("wrong token length")
	}

	if tc == tc1 {
		t.Error("2 sequential token are equal by BASE64")
	}
}
