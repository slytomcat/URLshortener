package main

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, strings.Repeat("_", 6), st.Get())
}

// try to make two tokens from random source and compare them
func TestNewShortTokenReal6(t *testing.T) {
	st := NewShortToken(6)

	tc := st.Get()

	tc1 := st.Get()

	assert.Equal(t, 6, len(tc))
	assert.Equal(t, 6, len(tc1))
	assert.NotEqual(t, tc, tc1)
	t.Log(tc, tc1)
}

// try to make two very short tokens from random source and compare them
func TestNewShortTokenReal2(t *testing.T) {
	st := NewShortToken(2)

	tc := st.Get()

	tc1 := st.Get()

	assert.Equal(t, 2, len(tc))
	assert.Equal(t, 2, len(tc1))
	assert.NotEqual(t, tc, tc1)
	t.Log(tc, tc1)
}

func TestAsyncGet(t *testing.T) {
	st := NewShortToken(6)
	cnt := 500
	resBuf := make([]string, cnt)
	wg := sync.WaitGroup{}
	wg.Add(cnt)
	for i := 0; i < cnt; i++ {
		go func(i int) {
			defer wg.Done()
			resBuf[i] = st.Get()
		}(i)
	}

	wg.Wait()
	// Check that all token are different
	for i := 0; i < cnt; i++ {
		//wg.Add(1)
		//go func(i int) {
		//	defer wg.Done()
		for j := i + 1; j < cnt; j++ {
			assert.NotEqual(t, resBuf[i], resBuf[j])
			// t.Log(resBuf[i], resBuf[j])
		}
		//}(i)
	}
	//wg.Wait()
}

// test Check with correct token
func Test00ST10CheckOk(t *testing.T) {
	st := NewShortToken(2)

	sToken := st.Get()

	assert.NoError(t, st.Check(sToken))
}

// test Check with wrong token length
func Test00ST15CheckNoOk(t *testing.T) {
	st := NewShortToken(2)

	sToken := st.Get()

	assert.Error(t, st.Check(sToken+"wrong"))
}

// test Check with wrong token alphabet
func Test00ST15CheckNoOk2(t *testing.T) {
	st := NewShortToken(2)

	assert.Error(t, st.Check("#$")) // check nonBase64 symbols
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
