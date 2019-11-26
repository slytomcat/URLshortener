package main

import (
	"bytes"
	"testing"
)

func Test00ShortTokenNewFake(t *testing.T) {
	DEBUG = true
	tc, err := ShortTokenNew()
	if err != nil {
		t.Error("error of ShortToken creation from random")
	}

	if !bytes.Equal(tc.Bytes, []byte{0xff, 0xff, 0xff, 0xff, 0xf0}) {
		t.Error("wrong token []bytes representation")
	}

	if tc.BASE64 != "______" {
		t.Error("wrong token BASE64 representation")
	}
}

func Test03ShortTokenNewReal(t *testing.T) {
	DEBUG = false
	tc, err := ShortTokenNew()
	if err != nil {
		t.Error("error of ShortToken creation from random")
	}

	tc1, _ := ShortTokenNew()

	if bytes.Equal(tc.Bytes, tc1.Bytes) {
		t.Error("2 sequential token are equal by bytes")
	}

	if tc.BASE64 == tc1.BASE64 {
		t.Error("2 sequential token are equal by BASE64")
	}
}

func Test05ShortTokenSet(t *testing.T) {
	tc, err := ShortTokenSet("Abys-_")
	if err != nil {
		t.Error("error of ShortToken creation from BASE64")
	}

	if !bytes.Equal(tc.Bytes, []byte{1, 188, 172, 251, 240}) {
		t.Error("wrong token []bytes representation")
	}

	if tc.BASE64 != "Abys-_" {
		t.Error("wrong token BASE64 representation")
	}

	_, err = ShortTokenSet("a#$%s-")
	if err.Error() != "illegal base64 data at input byte 1" {
		t.Error("wrong error for wrong alphabet: " + err.Error())
	}

	_, err = ShortTokenSet("-_")
	if err.Error() != "wrong token length" {
		t.Error("wrong error for wrong token lenght: " + err.Error())
	}

}
