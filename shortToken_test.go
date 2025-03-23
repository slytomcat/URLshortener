package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shortTokenD - debugging ShortToken interface realization
type shortTokenD struct{ length int }

// mockShortToken returns the shortToken interface that always returns the same token
func mockShortToken(length int) ShortToken { return &shortTokenD{length} }

func (s shortTokenD) Get() string { return strings.Repeat("_", s.length) }

func (s shortTokenD) CheckAlphabet(_ string) error { return nil }

func (s shortTokenD) CheckLength(_ string) error { return nil }

// try to create new token from debugging source
func TestMockShortToken(t *testing.T) {
	st := mockShortToken(6)
	require.Equal(t, strings.Repeat("_", 6), st.Get())
}

const urlEncodingString = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

func TestCheckAlphabet(t *testing.T) {
	st := NewShortToken(8)
	require.NoError(t, st.CheckAlphabet(urlEncodingString))
	for _, ch := range "~`!№@#$%^&*/+=(){}[]<>?,.\"':;|\\" {
		assert.Error(t, st.CheckAlphabet(string(ch)))
	}
}

// try to make two tokens from random source and compare them
func TestNewShortTokenReal6(t *testing.T) {
	st := NewShortToken(6)
	tc := st.Get()
	tc1 := st.Get()
	t.Log(tc, tc1)
	require.NoError(t, st.CheckLength(tc))
	require.NoError(t, st.CheckLength(tc1))
	require.NoError(t, st.CheckAlphabet(tc), tc)
	require.NoError(t, st.CheckAlphabet(tc1), tc1)
	require.NotEqual(t, tc, tc1)
}

// try to make two very short tokens from random source and compare them
func TestNewShortTokenReal2(t *testing.T) {
	st := NewShortToken(2)
	tc := st.Get()
	tc1 := st.Get()
	t.Log(tc, tc1)
	require.NoError(t, st.CheckLength(tc))
	require.NoError(t, st.CheckLength(tc1))
	require.NoError(t, st.CheckAlphabet(tc), tc)
	require.NoError(t, st.CheckAlphabet(tc1), tc1)
	require.NotEqual(t, tc, tc1)
}

func TestAsyncGet(t *testing.T) {
	st := NewShortToken(6)
	cnt := 500
	resBuf := make([]string, cnt)
	wg := sync.WaitGroup{}
	wg.Add(cnt)
	for i := range cnt {
		go func(i int) {
			defer wg.Done()
			resBuf[i] = st.Get()
		}(i)
	}
	wg.Wait()
	for i := range cnt {
		for j := i + 1; j < cnt; j++ {
			assert.NotEqual(t, resBuf[i], resBuf[j])
		}
	}
}

// test Check with correct token
func Test00ST10CheckOk(t *testing.T) {
	st := NewShortToken(2)
	sToken := st.Get()
	require.NoError(t, st.CheckLength(sToken))
	require.NoError(t, st.CheckAlphabet(sToken))
}

// test Check with wrong token length
func Test00ST15CheckNoOk(t *testing.T) {
	st := NewShortToken(2)
	sToken := st.Get()
	require.Error(t, st.CheckLength(sToken+"wrong"))
	require.Error(t, st.CheckAlphabet(sToken+"!@#"))
}

// Benchmark for the 2 symbols token
func Benchmark00ST00Create2(b *testing.B) {
	st := NewShortToken(2)
	for b.Loop() {
		_ = st.Get()
	}
}

// Benchmark for the 6 symbols token
func Benchmark00ST00Create6(b *testing.B) {
	st := NewShortToken(6)
	for b.Loop() {
		_ = st.Get()
	}
}

// Benchmark for the 8 symbols token
func Benchmark00ST00Create8(b *testing.B) {
	st := NewShortToken(8)
	for b.Loop() {
		_ = st.Get()
	}
}

// experimental token generator that uses sync.Pool for random buffer
// unfortunately it is slower and requires more memory than original simple version.
func NewBShortToken(length int) ShortToken {
	pool := &sync.Pool{}
	pool.New = func() any {
		return make([]byte, length*6/8+1)
	}
	return &shortBToken{
		length:  length,
		bufPool: pool,
	}
}

type shortBToken struct {
	length  int        // token length
	bufPool *sync.Pool // buffer pool for random bytes
}

// Get creates the token from random source
func (s *shortBToken) Get() string {
	// get secure random bytes
	buf := s.bufPool.Get().([]byte)
	defer s.bufPool.Put(buf)

	n, err := rand.Read(buf)
	if err != nil || n != len(buf) {
		panic(fmt.Errorf("error while retrieving random data: %d %v", n, err.Error()))
	}
	// return shortened to tokenLenS BASE64 representation
	return base64.URLEncoding.EncodeToString(buf)[:s.length]
}

func (s *shortBToken) CheckLength(sToken string) error { return nil }

func (s *shortBToken) CheckAlphabet(sToken string) error { return nil }

// Benchmark for the 2 symbols token
func Benchmark00ST00Create2B(b *testing.B) {
	st := NewBShortToken(2)
	for b.Loop() {
		_ = st.Get()
	}
}

// Benchmark for the 6 symbols token
func Benchmark00ST00Create6B(b *testing.B) {
	st := NewBShortToken(6)
	for b.Loop() {
		_ = st.Get()
	}
}

// Benchmark for the 8 symbols token
func Benchmark00ST00Create8B(b *testing.B) {
	st := NewBShortToken(8)
	for b.Loop() {
		_ = st.Get()
	}
}

var tokens = []string{"ABCDEFGH", "ABCDEFG&", "ABCDEF&G", "ABCDE&FG", "ABCD&EFG", "&ABCDEFG", "ABCDEFGH", "ABCD&EFG", "&ABCDEFG", "ABCDEFGH"}

func TestCheck(t *testing.T) {
	st := NewShortToken(8)
	stO := newBShortTokenOrig(8)
	stA := newBShortTokenAlt(8)
	require.NoError(t, st.CheckAlphabet(tokens[0]))
	require.NoError(t, stO.CheckAlphabet(tokens[0]))
	require.NoError(t, stA.CheckAlphabet(tokens[0]))
	require.EqualError(t, st.CheckAlphabet(tokens[1]), "illegal base64 data at input byte 7")
	require.EqualError(t, stO.CheckAlphabet(tokens[1]), "illegal base64 data at input byte 7")
	require.EqualError(t, stA.CheckAlphabet(tokens[1]), "illegal base64 data at input byte 7")
}

func BenchmarkTokenCheck(b *testing.B) {
	st := NewShortToken(8)
	for b.Loop() {
		for _, t := range tokens {
			st.CheckAlphabet(t)
		}
	}
}

func BenchmarkTokenCheckOrig(b *testing.B) {
	st := newBShortTokenOrig(8)
	for b.Loop() {
		for _, t := range tokens {
			st.CheckAlphabet(t)
		}
	}
}

func BenchmarkTokenCheckAlt(b *testing.B) {
	st := newBShortTokenAlt(8)
	for b.Loop() {
		for _, t := range tokens {
			st.CheckAlphabet(t)
		}
	}
}

type testSToken struct{ length int }

func newBShortTokenOrig(length int) ShortToken { return &testSToken{length: length} }

func (s *testSToken) Get() string { return "AAAAAAAA" }

func (s *testSToken) CheckLength(_ string) error { return nil }

func (s *testSToken) CheckAlphabet(sToken string) error {
	if _, err := base64.URLEncoding.DecodeString(sToken + "AAAA"[:4-s.length%4]); err != nil {
		return err
	}
	return nil
}

type testSTokenA struct{}

func newBShortTokenAlt(_ int) ShortToken { return &testSTokenA{} }

func (s *testSTokenA) Get() string { return "AAAAAAAA" }

func (s *testSTokenA) CheckLength(sToken string) error { return nil }

var urlEncodingBytes = []byte(urlEncodingString)

func (s *testSTokenA) CheckAlphabet(sToken string) error {
	for i, s := range sToken {
		if !bytes.ContainsRune(urlEncodingBytes, s) {
			return base64.CorruptInputError(i)
		}
	}
	return nil
}
