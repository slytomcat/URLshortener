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
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var (
	testDBConfig *Config
	testDB       TokenDB
	testDBerr    TokenDB
	testDBToken  string = "AAAA"
)

type mockDB struct {
	setFunc   func(string, string, int) (bool, error)
	getFunc   func(string) (string, error)
	expFunc   func(string, int) error
	delfunc   func(string) error
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
	return m.delfunc(sToken)
}

func (m *mockDB) Close() error {
	return m.closeFunc()
}

func newMockDB() *mockDB {
	return &mockDB{
		setFunc:   func(_, _ string, _ int) (bool, error) { return true, nil },
		getFunc:   func(_ string) (string, error) { return "http://localhost:8080/favicon.ico", nil },
		expFunc:   func(_ string, _ int) error { return nil },
		delfunc:   func(_ string) error { return nil },
		closeFunc: func() error { return nil },
	}
}

// test new TokenDB creation errors
func Test05DBR01NewTokenDBError(t *testing.T) {

	connect := redis.UniversalOptions{}
	json.Unmarshal([]byte(`{"Addrs":["Wrong.Host:6379"]}`), &connect)

	_, err := NewTokenDB([]string{"Wrong.Host:6379"}, "")
	assert.Error(t, err)
}

// concurrent goroutines tries to make new short URL in the same time with the same token (debugging)
func raceNewToken(db TokenDB, url string, t *testing.T) {

	var wg sync.WaitGroup
	var succes, fail, cnt, i int64
	cnt = 5
	var start sync.RWMutex

	start.Lock()
	rand.Seed(time.Now().UnixNano())

	racer := func(i int64) {
		defer wg.Done()

		time.Sleep(time.Duration(rand.Intn(42)) * time.Microsecond * 100)
		//t.Logf("%v Racer %d: Ready!\n", time.Now(), i)
		start.RLock()

		ok, err := db.Set(testDBToken, url, 1)

		if err != nil {
			//t.Logf("%v Racer %d: %v \n", time.Now(), i, err)
			atomic.AddInt64(&fail, 1)
			return
		}

		if !ok {
			//t.Logf("%v Racer %d: %v \n", time.Now(), i, "dubicate detected")
			atomic.AddInt64(&fail, 1)
			return
		}

		//t.Logf("%v Racer %d: Token for %s: %v\n", time.Now(), i, url, testDBToken)
		atomic.AddInt64(&succes, 1)
	}

	for i = 0; i < cnt; i++ {
		wg.Add(1)
		go racer(i)
	}
	//t.Logf("%v Ready?\n", time.Now())
	time.Sleep(time.Second * 2)
	start.Unlock()
	//t.Logf("%v Go!!!\n", time.Now())
	wg.Wait()

	if succes != 1 || fail != cnt-1 {
		t.Errorf("Concurent update error: success=%d, fail=%d, total=%d", succes, fail, cnt)
	}
}

// test new TokenDBR creation
func Test05DBR10All(t *testing.T) {

	godotenv.Load()

	testDBConfig, err := readConfig()
	assert.NoError(t, err)

	testDB, err = NewTokenDB(testDBConfig.RedisAddrs, testDBConfig.RedisPassword)
	assert.NoError(t, err)

	testDB.Delete(testDBToken)
	defer testDB.Delete(testDBToken)

	t.Run("store 2 equal tokens: fail", func(t *testing.T) {

		url := "https://golang.org/pkg/time/"
		ok, err := testDB.Set(testDBToken, url, 1)
		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = testDB.Set(testDBToken, url, 1)
		assert.NoError(t, err)
		assert.False(t, ok)
		// clear
		testDB.Delete(testDBToken)
	})

	t.Run("detect race on token store: success", func(t *testing.T) {
		raceNewToken(testDB, "https://golang.org", t)
		assert.NoError(t, testDB.Expire(testDBToken, -1))
	})

	t.Run("one more time: success", func(t *testing.T) {
		raceNewToken(testDB, "https://golang.org/pkg/time/error", t)
	})

	t.Run("get: success", func(t *testing.T) {

		lURL, err := testDB.Get(testDBToken)
		assert.NoError(t, err)
		assert.NotEmpty(t, lURL)
	})
	t.Run("del: success", func(t *testing.T) {

		assert.NoError(t, testDB.Delete(testDBToken))

		_, err := testDB.Get(testDBToken)
		assert.Error(t, err)
	})

	t.Run("expire non existing token", func(t *testing.T) {
		assert.Error(t, testDB.Expire(testDBToken+"$", -1))
	})
	t.Run("delete non existing token", func(t *testing.T) {
		assert.Error(t, testDB.Delete(testDBToken+"$"))
	})
	t.Run("get non existing token", func(t *testing.T) {
		_, err := testDB.Get(testDBToken + "$")
		assert.Error(t, err)
	})

	t.Run("close db: success", func(t *testing.T) {
		assert.NoError(t, testDB.Close())
	})
}

func Benchmark05DBR10set(b *testing.B) {
	var err error
	testDBConfig, err = readConfig()
	if err != nil {
		b.Error(err)
	}

	testDB, err = NewTokenDB(testDBConfig.RedisAddrs, testDBConfig.RedisPassword)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		_, err := testDB.Set(strconv.Itoa(i), "test", 0)
		if err != nil {
			b.Error(err)
		}
	}
}

func Benchmark05DBR00del(b *testing.B) {
	var err error
	testDBConfig, err = readConfig()
	if err != nil {
		b.Error(err)
	}

	testDB, err = NewTokenDB(testDBConfig.RedisAddrs, testDBConfig.RedisPassword)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		err := testDB.Delete(strconv.Itoa(i))
		if err != nil {
			b.Logf("i=%v err=%v", i, err)
		}
	}
}
