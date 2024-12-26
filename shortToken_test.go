package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
func TestMockShortToken(t *testing.T) {
	st := mockShortToken(6)
	require.Equal(t, strings.Repeat("_", 6), st.Get())
}

// try to make two tokens from random source and compare them
func TestNewShortTokenReal6(t *testing.T) {
	st := NewShortToken(6)

	tc := st.Get()

	tc1 := st.Get()

	require.Equal(t, 6, len(tc))
	require.Equal(t, 6, len(tc1))
	require.NotEqual(t, tc, tc1)
	t.Log(tc, tc1)
}

// try to make two very short tokens from random source and compare them
func TestNewShortTokenReal2(t *testing.T) {
	st := NewShortToken(2)

	tc := st.Get()

	tc1 := st.Get()

	require.Equal(t, 2, len(tc))
	require.Equal(t, 2, len(tc1))
	require.NotEqual(t, tc, tc1)
	t.Log(tc, tc1)
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

	require.NoError(t, st.Check(sToken))
}

// test Check with wrong token length
func Test00ST15CheckNoOk(t *testing.T) {
	st := NewShortToken(2)

	sToken := st.Get()

	require.Error(t, st.Check(sToken+"wrong"))
}

// test Check with wrong token alphabet
func Test00ST15CheckNoOk2(t *testing.T) {
	st := NewShortToken(2)

	require.Error(t, st.Check("#$")) // check nonBase64 symbols
}

// Benchmark for the 2 symbols token
func Benchmark00ST00Create2(b *testing.B) {
	st := NewShortToken(2)
	for range b.N {
		_ = st.Get()
	}
}

// Benchmark for the 6 symbols token
func Benchmark00ST00Create6(b *testing.B) {
	st := NewShortToken(6)
	for range b.N {
		_ = st.Get()
	}
}

// Benchmark for the 8 symbols token
func Benchmark00ST00Create8(b *testing.B) {
	st := NewShortToken(8)
	for range b.N {
		_ = st.Get()
	}
}

// experimental token generator that uses sync.Pool for random buffer
// unfortunately it is slower and requires more memory than original simple version.
func NewBShortToken(length int) ShortToken {
	pool := &sync.Pool{}
	pool.New = func() interface{} {
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

// Get creates the token from random or debugging source
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

func (s *shortBToken) Check(sToken string) error {
	return nil
}

// Benchmark for the 2 symbols token
func Benchmark00ST00Create2B(b *testing.B) {
	st := NewBShortToken(2)
	for range b.N {
		_ = st.Get()
	}
}

// Benchmark for the 6 symbols token
func Benchmark00ST00Create6B(b *testing.B) {
	st := NewBShortToken(6)
	for range b.N {
		_ = st.Get()
	}
}

// Benchmark for the 8 symbols token
func Benchmark00ST00Create8B(b *testing.B) {
	st := NewBShortToken(8)
	for range b.N {
		_ = st.Get()
	}
}

var (
	tokens = []string{"ABCDEFGH", "ABCDEFG&", "ABCDEF&G", "ABCDE&FG", "ABCD&EFG", "&ABCDEFG", "ABCDEFGH"}
)

func TestCheck(t *testing.T) {
	st := NewShortToken(8)
	require.NoError(t, st.Check(tokens[0]))
	require.NoError(t, check1(tokens[0], 8))
	require.EqualError(t, st.Check(tokens[1]), "wrong token alphabet")
	require.EqualError(t, check1(tokens[1], 8), "incorrect symbol: '&'")

}

func BenchmarkTokenCheck(b *testing.B) {
	st := NewShortToken(8)
	for range b.N {
		for _, t := range tokens {
			st.Check(t)
		}
	}
}

func BenchmarkTokenCheck1(b *testing.B) {
	_ = NewShortToken(8)
	for range b.N {
		for _, t := range tokens {
			check1(t, 8)
		}
	}
}

func check1(token string, l int) error {
	// check length
	if len(token) != l {
		return errors.New("wrong token length")
	}

	for _, s := range token {
		if !((s >= 'A' && s <= 'Z') || (s >= 'a' && s <= 'z') || (s <= '0' && s >= '9') || s == '_' || s == '-') {
			return fmt.Errorf("incorrect symbol: '%c'", s)
		}
	}
	return nil
}
