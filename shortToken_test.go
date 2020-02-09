package main

import (
	"strings"
	"testing"
)

// SetDebug sets debuging mode and return
// func() that resets the debog mode to 0.
func SetDebug(mode int) func() {
	SetShortTokenDebug(mode)
	return func() { SetShortTokenDebug(0) }
}

// try to create new token from debugging source
func Test00ST05NewShortTokenFake(t *testing.T) {

	defer SetDebug(1)()

	DEBUGToken := strings.Repeat("_", 6)
	tc, err := NewShortToken(6)
	if err != nil {
		t.Error("error of ShortToken creation from debug source:", err)
	}
	if tc != DEBUGToken {
		t.Errorf("wrong token BASE64 representation: expected: '%s', received '%s'", DEBUGToken, tc)
	}
}

// try to make two tokens from random source and compare them
func Test00ST07NewShortTokenReal(t *testing.T) {
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
func Test00ST07NewShortTokenReal2(t *testing.T) {
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

// test debug error
func Test00ST08DebugError(t *testing.T) {

	defer SetDebug(-1)()

	_, err := NewShortToken(2)
	if err == nil {
		t.Error("no error when expected:")
	}
}
