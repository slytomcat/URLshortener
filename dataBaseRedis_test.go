package main

import (
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	testDBConfig *Config
	testDB       Token
)

// test new TokenDB creation errors
func Test05DBR01NewTokenDBError(t *testing.T) {

	connect, _ := parseConOpt(`{"Addrs":["Wrong.Host:6379"]}`)

	_, err := NewTokenDB(connect, 500, 5)
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

	testDB, err = NewTokenDB(testDBConfig.ConnectOptions, testDBConfig.Timeout, testDBConfig.TokenLength)
	if err != nil {
		t.Error(err)
	}

	testDB.Delete(strings.Repeat("_", testDBConfig.TokenLength))
}

// try to add 2 tokens
func Test05DBR15OneTokenTwice(t *testing.T) {

	defer SetDebug(1)()

	testDB.Delete(strings.Repeat("_", testDBConfig.TokenLength))

	url := "https://golang.org/pkg/time/"
	token, err := testDB.New(url, 1)
	if err != nil || token == "" {
		t.Errorf("unexpected error: %s; token: %s", err, token)
	} else {
		t.Logf("expected result: token for %s: %v\n", url, token)
	}
	token1, err := testDB.New(url, 1)
	if err != nil {
		t.Logf("expected error: %s\n", err)
	} else {
		t.Errorf("wrong result: token for %s: %v\n", url, token1)
	}
	// clear
	testDB.Delete(strings.Repeat("_", testDBConfig.TokenLength))
}

// concurrent goroutines tries to make new short URL in the same time with the same token (debugging)
func raceNewToken(db Token, url string, t *testing.T) {

	defer SetDebug(1)()

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

		token, err := db.New(url, 1)

		if err != nil {
			t.Logf("%v Racer %d: %v \n", time.Now(), i, err)
			atomic.AddInt64(&fail, 1)
			return
		}

		t.Logf("%v Racer %d: Token for %s: %v\n", time.Now(), i, url, token)
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
	err := testDB.Expire(strings.Repeat("_", testDBConfig.TokenLength), -1)
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

	lURL, err := testDB.Get(strings.Repeat("_", testDBConfig.TokenLength))
	if err != nil {
		t.Error(err)
	}
	t.Logf("URL for token %s: %s\n", strings.Repeat("_", testDBConfig.TokenLength), lURL)
}

// try to delete token
func Test05DBR40DelToken(t *testing.T) {

	err := testDB.Delete(strings.Repeat("_", testDBConfig.TokenLength))
	if err != nil {
		t.Error(err)
	}

	_, err = testDB.Get(strings.Repeat("_", testDBConfig.TokenLength))
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to expire non existing token
func Test05DBR45ExpNonExisting(t *testing.T) {
	err := testDB.Expire(strings.Repeat("$", testDBConfig.TokenLength), -1)
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to expire token with wrong length
func Test05DBR45ExpNonExisting1(t *testing.T) {
	err := testDB.Expire(strings.Repeat("$", testDBConfig.TokenLength+1), -1)
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to delete non existing token
func Test05DBR50DelNonExisting(t *testing.T) {
	err := testDB.Delete(strings.Repeat("$", testDBConfig.TokenLength))
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to delete token with wrong length
func Test05DBR51DelNonExisting1(t *testing.T) {
	err := testDB.Delete(strings.Repeat("$", testDBConfig.TokenLength+1))
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to get non existing token
func Test05DBR55GetNonExisting(t *testing.T) {
	_, err := testDB.Get(strings.Repeat("$", testDBConfig.TokenLength))
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to get token with wrong length
func Test05DBR55GetNonExisting1(t *testing.T) {
	_, err := testDB.Get(strings.Repeat("$", testDBConfig.TokenLength+1))
	if err == nil {
		t.Error("no error when expected")
	}
}

// test debug error
func Test05DBR60DebugError(t *testing.T) {

	defer SetDebug(-1)()

	_, err := testDB.New("someUrl.com", 1)
	if err == nil {
		t.Error("no error when expected")
	}
}

// get the Attempts

func Test05DBRAttempts(t *testing.T) {
	if testDB.Attempts() == 0 {
		t.Error("no attempts calculated")
	}
}
