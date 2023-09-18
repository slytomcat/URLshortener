package main

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDBToken string = "AAAA"

type mockDB struct {
	setFunc   func(string, string, int) (bool, error)
	getFunc   func(string) (string, error)
	expFunc   func(string, int) error
	delFunc   func(string) error
	closeFunc func() error
}

func (m *mockDB) Set(sToken, longURL string, expiration int) (bool, error) {
	return m.setFunc(sToken, longURL, expiration)
}

func (m *mockDB) Get(sToken string) (string, error) {
	return m.getFunc(sToken)
}

func (m *mockDB) Expire(sToken string, expiration int) error {
	return m.expFunc(sToken, expiration)
}

func (m *mockDB) Delete(sToken string) error {
	return m.delFunc(sToken)
}

func (m *mockDB) Close() error {
	return m.closeFunc()
}

func newMockDB() *mockDB {
	return &mockDB{
		setFunc:   func(_, _ string, _ int) (bool, error) { return true, nil },
		getFunc:   func(_ string) (string, error) { return "http://localhost:8080/favicon.ico", nil },
		expFunc:   func(_ string, _ int) error { return nil },
		delFunc:   func(_ string) error { return nil },
		closeFunc: func() error { return nil },
	}
}

// test new TokenDB creation errors
func Test05DBR01NewTokenDBError(t *testing.T) {

	connect := redis.UniversalOptions{}
	json.Unmarshal([]byte(`{"Addrs":["Wrong.Host:6379"]}`), &connect)

	_, err := NewTokenDB([]string{"Wrong.Host:6379"}, "")
	require.Error(t, err)
}

// concurrent goroutines tries to make new short URL in the same time with the same token (debugging)
func raceNewToken(db TokenDB, url string, t *testing.T) {

	var wg sync.WaitGroup
	var success, fail, cnt, i int64
	cnt = 5
	var start sync.RWMutex

	start.Lock()

	racer := func(i int64) {
		defer wg.Done()

		time.Sleep(time.Duration(rand.Intn(42)) * time.Microsecond * 100)
		start.RLock()

		ok, err := db.Set(testDBToken, url, 1)

		if err != nil {
			atomic.AddInt64(&fail, 1)
			return
		}

		if !ok {
			atomic.AddInt64(&fail, 1)
			return
		}

		atomic.AddInt64(&success, 1)
	}

	for i = 0; i < cnt; i++ {
		wg.Add(1)
		go racer(i)
	}
	time.Sleep(time.Millisecond * 20)
	start.Unlock()
	wg.Wait()

	if success != 1 || fail != cnt-1 {
		t.Errorf("Concurrent update error: success=%d, fail=%d, total=%d", success, fail, cnt)
	}
}

// test new TokenDBR creation
func Test05DBR10All(t *testing.T) {
	envSet(t)

	testDBConfig, err := readConfig()
	require.NoError(t, err)

	testDB, err := NewTokenDB(testDBConfig.RedisAddrs, testDBConfig.RedisPassword)
	require.NoError(t, err)

	testDB.Delete(testDBToken)
	defer testDB.Delete(testDBToken)

	t.Run("store 2 equal tokens: fail", func(t *testing.T) {

		url := "https://golang.org/pkg/time/"
		ok, err := testDB.Set(testDBToken, url, 1)
		require.NoError(t, err)
		require.True(t, ok)

		ok, err = testDB.Set(testDBToken, url, 1)
		require.NoError(t, err)
		require.False(t, ok)
		// clear
		testDB.Delete(testDBToken)
	})

	t.Run("detect race on token store: success", func(t *testing.T) {
		raceNewToken(testDB, "https://golang.org", t)
		require.NoError(t, testDB.Expire(testDBToken, -1))
	})

	t.Run("one more time: success", func(t *testing.T) {
		raceNewToken(testDB, "https://golang.org/pkg/time/error", t)
	})

	t.Run("get: success", func(t *testing.T) {

		lURL, err := testDB.Get(testDBToken)
		require.NoError(t, err)
		require.NotEmpty(t, lURL)
	})
	t.Run("del: success", func(t *testing.T) {

		require.NoError(t, testDB.Delete(testDBToken))

		_, err := testDB.Get(testDBToken)
		require.Error(t, err)
	})

	t.Run("expire non existing token", func(t *testing.T) {
		require.Error(t, testDB.Expire(testDBToken+"$", -1))
	})
	t.Run("delete non existing token", func(t *testing.T) {
		require.Error(t, testDB.Delete(testDBToken+"$"))
	})
	t.Run("get non existing token", func(t *testing.T) {
		_, err := testDB.Get(testDBToken + "$")
		require.Error(t, err)
	})

	t.Run("close db: success", func(t *testing.T) {
		require.NoError(t, testDB.Close())
	})
}

func Benchmark05DBR10set(b *testing.B) {
	envSet(b)

	testDBConfig, err := readConfig()
	require.NoError(b, err)

	testDB, err := NewTokenDB(testDBConfig.RedisAddrs, testDBConfig.RedisPassword)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		_, err := testDB.Set(strconv.Itoa(i), "test", 0)
		assert.NoError(b, err)
	}
}

func Benchmark05DBR00del(b *testing.B) {
	envSet(b)

	testDBConfig, err := readConfig()
	require.NoError(b, err)

	testDB, err := NewTokenDB(testDBConfig.RedisAddrs, testDBConfig.RedisPassword)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		err := testDB.Delete(strconv.Itoa(i))
		if err != nil {
			b.Logf("i=%v err=%v", i, err)
		}
	}
}
