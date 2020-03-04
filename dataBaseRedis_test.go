package main

import (
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// test new TokenDB creation errors
func Test05DBR01NewTokenDBError(t *testing.T) {

	defer saveEnv()()

	CONFIG.ConnectOptions, _ = parseConOpt(`{"Addrs":["Wrong.Host:6379"]}`)

	err := NewTokenDB()
	if err == nil {
		t.Error("No error when expected")
	}
}

// test new TokenDBR creation
func Test05DBR10NewTokenDB(t *testing.T) {

	err := readConfig("cnfr.json")
	if err != nil {
		t.Error(err)
	}

	err = NewTokenDB()
	if err != nil {
		t.Error(err)
	}

	TokenDB.Delete(strings.Repeat("_", CONFIG.TokenLength))
}

// try to add 2 tokens
func Test05DBR15OneTokenTwice(t *testing.T) {

	defer SetDebug(1)()

	TokenDB.Delete(strings.Repeat("_", CONFIG.TokenLength))

	url := "https://golang.org/pkg/time/"
	token, err := TokenDB.New(url, 1)
	if err != nil || token == "" {
		t.Errorf("unexpected error: %s; token: %s", err, token)
	} else {
		t.Logf("expected result: token for %s: %v\n", url, token)
	}
	token1, err := TokenDB.New(url, 1)
	if err != nil {
		t.Logf("expected error: %s\n", err)
	} else {
		t.Errorf("wrong result: token for %s: %v\n", url, token1)
	}
	// clear
	TokenDB.Delete(strings.Repeat("_", CONFIG.TokenLength))
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
	raceNewToken(TokenDB, "https://golang.org", t)
}

// try to make token expired
func Test05DBR25ExpireToken(t *testing.T) {
	err := TokenDB.Expire(strings.Repeat("_", CONFIG.TokenLength), -1)
	if err != nil {
		t.Error(err)
	}
}

// try to update expired token from several concurrent goroutines
func Test05DBR30OneMoreTokenRace(t *testing.T) {
	raceNewToken(TokenDB, "https://golang.org/pkg/time/error", t)
}

// try to receive long URL by token
func Test05DBR35GetToken(t *testing.T) {

	lURL, err := TokenDB.Get(strings.Repeat("_", CONFIG.TokenLength))
	if err != nil {
		t.Error(err)
	}
	t.Logf("URL for token %s: %s\n", strings.Repeat("_", CONFIG.TokenLength), lURL)
}

// try to delete token
func Test05DBR40DelToken(t *testing.T) {

	err := TokenDB.Delete(strings.Repeat("_", CONFIG.TokenLength))
	if err != nil {
		t.Error(err)
	}

	_, err = TokenDB.Get(strings.Repeat("_", CONFIG.TokenLength))
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to expire non existing token
func Test05DBR45ExpNonExisting(t *testing.T) {
	err := TokenDB.Expire("#$%^&*(", -1)
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to delete non existing token
func Test05DBR50DelNonExisting(t *testing.T) {
	err := TokenDB.Delete("#$%^&*(")
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to get non existing token
func Test05DBR55GetNonExisting(t *testing.T) {
	_, err := TokenDB.Get("#$%^&*(")
	if err == nil {
		t.Error("no error when expected")
	}
}

// test debug error
func Test05DBR60ebugError(t *testing.T) {

	defer SetDebug(-1)()

	_, err := TokenDB.New("someUrl.com", 1)
	if err == nil {
		t.Error("no error when expected")
	}
}

// get the test results
