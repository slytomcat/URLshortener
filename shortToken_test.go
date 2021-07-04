package main

import (
	"strings"
	"testing"
)

// shortTokenD - debugging ShortToken interface realization
type shortTokenD struct {
	length int
}

// mockShortToken returns the shortToken interface that always returns the same token
func mockShortToken(length int) ShortToken {
	return &shortTokenD{length}
}

func (s shortTokenD) Get() string {
	return strings.Repeat("_", s.length)
}

func (s shortTokenD) Check(_ string) error {
	return nil
}

// try to create new token from debugging source
func Test00ST05NewShortTokenFake(t *testing.T) {
	st := mockShortToken(6)

	DEBUGToken := strings.Repeat("_", 6)
	tc := st.Get()
	if tc != DEBUGToken {
		t.Errorf("wrong token BASE64 representation: expected: '%s', received '%s'", DEBUGToken, tc)
	}
}

// try to make two tokens from random source and compare them
func Test00ST07NewShortTokenReal(t *testing.T) {
	st := NewShortToken(6)

	tc := st.Get()

	tc1 := st.Get()

	if len(tc) != len(tc1) || len(tc) != 6 {
		t.Error("wrong token length")
	}

	if tc == tc1 {
		t.Errorf("2 sequential token are equal: '%s' == '%s'", tc, tc1)
	}
}

// try to make two very short tokens from random source and compare them
func Test00ST07NewShortTokenReal2(t *testing.T) {
	st := NewShortToken(2)

	tc := st.Get()

	tc1 := st.Get()

	if len(tc) != len(tc1) || len(tc) != 2 {
		t.Error("wrong token length")
	}

	if tc == tc1 {
		t.Errorf("2 sequential token are equal: '%s' == '%s'", tc, tc1)
	}
}

// test Check with correct token
func Test00ST10CheckOk(t *testing.T) {
	st := NewShortToken(2)

	sToken := st.Get()

	err := st.Check(sToken)
	if err != nil {
		t.Errorf("token check error: %v", err)
	}
}

// test Check with wrong token length
func Test00ST15CheckNoOk(t *testing.T) {
	st := NewShortToken(2)

	sToken := st.Get()

	err := st.Check(sToken + "wrong")
	if err == nil {
		t.Error("no error when expected")
	}
}

// test Check with wrong token alphabet
func Test00ST15CheckNoOk2(t *testing.T) {
	st := NewShortToken(2)

	err := st.Check("#$") // check nonBase64 symbols
	if err == nil {
		t.Error("no error when expected")
	}
}

// Benchmark for the 2 symbols token
func Benchmark00ST00Create2(b *testing.B) {
	st := NewShortToken(2)
	for i := 0; i < b.N; i++ {
		_ = st.Get()
	}
}

// Benchmark for the 6 symbols token
func Benchmark00ST00Create6(b *testing.B) {
	st := NewShortToken(6)
	for i := 0; i < b.N; i++ {
		_ = st.Get()
	}
}

// Benchmark for the 8 symbols token
func Benchmark00ST00Create8(b *testing.B) {
	st := NewShortToken(8)
	for i := 0; i < b.N; i++ {
		_ = st.Get()
	}
}
