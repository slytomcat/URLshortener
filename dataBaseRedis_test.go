package main

import (
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

type testDBErr struct {
}

func (t testDBErr) Set(sToken, longURL string, expiration int) (bool, error) {
	if longURL == "http://localhost:8080/favicon.ico" {
		return true, nil
	}
	return false, errors.New("test Set() error")
}
func (t testDBErr) Get(sToken string) (string, error) {
	if sToken == "Debug.Token" {
		return "http://localhost:8080/favicon.ico", nil
	}
	return "", errors.New("test Get() error")
}
func (t testDBErr) Expire(sToken string, expiration int) error {
	return errors.New("test Expire() error")
}
func (t testDBErr) Delete(sToken string) error {
	return errors.New("test Delete() error")
}
func (t testDBErr) Close() error {
	return errors.New("test Close() error")
}
func testDBNewTokenDB(_ redis.UniversalOptions) (TokenDB, error) {
	return &testDBErr{}, nil
}

// test new TokenDB creation errors
func Test05DBR01NewTokenDBError(t *testing.T) {

	connect, _ := parseConOpt(`{"Addrs":["Wrong.Host:6379"]}`)

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
