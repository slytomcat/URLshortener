package main

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
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

func testSet(sToken, longURL string, expiration int) (bool, error) {
	if longURL == "http://localhost:8080/favicon.ico" {
		return true, nil
	}
	return false, errors.New("test Set() error")
}

func testGet(sToken string) (string, error) {
	if sToken == "Debug.Token" {
		return "http://localhost:8080/favicon.ico", nil
	}
	return "", errors.New("test Get() error")
}

func testExpire(sToken string, expiration int) error {
	return errors.New("test Expire() error")
}

func testDelete(sToken string) error {
	return errors.New("test Delete() error")
}

func testClose() error {
	return errors.New("test Close() error")
}

func newMockDB() TokenDB {
	return &mockDB{
		setFunc:   testSet,
		getFunc:   testGet,
		expFunc:   testExpire,
		delfunc:   testDelete,
		closeFunc: testClose,
	}
}

// test new TokenDB creation errors
func Test05DBR01NewTokenDBError(t *testing.T) {

	connect := redis.UniversalOptions{}
	json.Unmarshal([]byte(`{"Addrs":["Wrong.Host:6379"]}`), &connect)

	_, err := NewTokenDB(connect)
	if err == nil {
		t.Error("No error when expected")
	}
}

// test new TokenDBR creation
func Test05DBR10NewTokenDB(t *testing.T) {
	var err error
	testDBConfig, err = readConfig("cnfr.json")
	if err != nil {
		t.Error(err)
	}

	testDB, err = NewTokenDB(testDBConfig.ConnectOptions)
	if err != nil {
		t.Error(err)
	}

	testDB.Delete(testDBToken)
}

// try to add 2 same tokens
func Test05DBR15OneTokenTwice(t *testing.T) {

	testDB.Delete(testDBToken)

	url := "https://golang.org/pkg/time/"
	ok, err := testDB.Set(testDBToken, url, 1)
	if err != nil || !ok {
		t.Errorf("unexpected error: %s; Ok: %v", err, ok)
	} else {
		t.Logf("expected result: token for %s: %v\n", url, testDBToken)
	}
	ok, err = testDB.Set(testDBToken, url, 1)
	if err != nil {
		t.Errorf("unexpected error: %s; token: %s", err, testDBToken)
	}
	if ok {
		t.Error("same token stored twice succesfily")
	}
	// clear
	testDB.Delete(testDBToken)
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
		t.Logf("%v Racer %d: Ready!\n", time.Now(), i)
		start.RLock()

		ok, err := db.Set(testDBToken, url, 1)

		if err != nil {
			t.Logf("%v Racer %d: %v \n", time.Now(), i, err)
			atomic.AddInt64(&fail, 1)
			return
		}

		if !ok {
			t.Logf("%v Racer %d: %v \n", time.Now(), i, "dubicate detected")
			atomic.AddInt64(&fail, 1)
			return
		}

		t.Logf("%v Racer %d: Token for %s: %v\n", time.Now(), i, url, testDBToken)
		atomic.AddInt64(&succes, 1)
	}

	for i = 0; i < cnt; i++ {
		wg.Add(1)
		go racer(i)
	}
	t.Logf("%v Ready?\n", time.Now())
	time.Sleep(time.Second * 2)
	start.Unlock()
	t.Logf("%v Go!!!\n", time.Now())
	wg.Wait()

	if succes != 1 || fail != cnt-1 {
		t.Errorf("Concurent update error: success=%d, fail=%d, total=%d", succes, fail, cnt)
	}
}

// try to insert new token from concurrent goroutines
func Test05DBR20NewTokenRace(t *testing.T) {
	raceNewToken(testDB, "https://golang.org", t)
}

// try to make token expired
func Test05DBR25ExpireToken(t *testing.T) {
	err := testDB.Expire(testDBToken, -1)
	if err != nil {
		t.Error(err)
	}
}

// try to update expired token from several concurrent goroutines
func Test05DBR30OneMoreTokenRace(t *testing.T) {
	raceNewToken(testDB, "https://golang.org/pkg/time/error", t)
}

// try to receive long URL by token
func Test05DBR35GetToken(t *testing.T) {

	lURL, err := testDB.Get(testDBToken)
	if err != nil {
		t.Error(err)
	}
	t.Logf("URL for token %s: %s\n", testDBToken, lURL)
}

// try to delete token
func Test05DBR40DelToken(t *testing.T) {

	err := testDB.Delete(testDBToken)
	if err != nil {
		t.Error(err)
	}

	_, err = testDB.Get(testDBToken)
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to expire non existing token
func Test05DBR45ExpNonExisting(t *testing.T) {
	err := testDB.Expire(testDBToken+"$", -1)
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to delete non existing token
func Test05DBR50DelNonExisting(t *testing.T) {
	err := testDB.Delete(testDBToken + "$")
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to get non existing token
func Test05DBR55GetNonExisting(t *testing.T) {
	_, err := testDB.Get(testDBToken + "$")
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to close connection
func Test05DBR65Close(t *testing.T) {
	if err := testDB.Close(); err != nil {
		t.Errorf("error DB connection closing: %v", err)
	}
}

func Benchmark05DBR10set(b *testing.B) {
	var err error
	testDBConfig, err = readConfig("cnfr.json")
	if err != nil {
		b.Error(err)
	}

	testDB, err = NewTokenDB(testDBConfig.ConnectOptions)
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
	testDBConfig, err = readConfig("cnfr.json")
	if err != nil {
		b.Error(err)
	}

	testDB, err = NewTokenDB(testDBConfig.ConnectOptions)
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
