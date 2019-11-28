package main

import (
	"testing"
)

// try to create new token from debugging source
func Test05ShortTokenNewFake(t *testing.T) {
	DEBUG = true
	tc, err := ShortTokenNew()
	if err != nil {
		t.Error("error of ShortToken creation from debug source:", err)
	}

	if tc != "______" {
		t.Error("wrong token BASE64 representation")
	}
}

// try to make two tokens from random source and compare them
func Test07ShortTokenNewReal(t *testing.T) {
	DEBUG = false
	tc, err := ShortTokenNew()
	if err != nil {
		t.Error("error of ShortToken creation from random:", err)
	}

	tc1, _ := ShortTokenNew()

	if tc == tc1 {
		t.Error("2 sequential token are equal by BASE64")
	}
}
