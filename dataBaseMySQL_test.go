package main

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	tDB Token
)

// test new TokenDBM creation
func Test10DBMNewTokenDB(t *testing.T) {
	var err error
	err = readConfig("cnf.json")
	if err != nil {
		t.Error(err)
	}

	tDB, err = TokenDBNewM()
	if err != nil {
		t.Error(err)
	}
}

// Clear the table - it is not a test
func Test11DBMClearTable(t *testing.T) {
	err := tDB.Delete(DEBUGToken)
	if err != nil {
		t.Errorf("Can't clear table: %v", err)
	}
}

// Try to insert the same token twice (error expected)
func Test13DBMOneTokenTwice(t *testing.T) {
	DEBUG = true
	defer func() { DEBUG = false }()

	url := "https://golang.org/pkg/time/"
	token, err := tDB.New(url, 1)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else {
		t.Logf("expected result: token for %s: %v\n", url, token)
	}
	token1, err := tDB.New(url, 1)
	if err != nil {
		t.Logf("expected error: %s\n", err)
	} else {
		t.Errorf("wrong result: token for %s: %v\n", url, token1)
	}
	// clear
	err = tDB.Delete(DEBUGToken)
	if err != nil {
		t.Errorf("Can't clear table: %v", err)
	}
}

// Concurrent goroutines tries to make new short URL in the same time with the same token (debugging)
func raceNewToken(db Token, url string, t *testing.T) {
	DEBUG = true
	defer func() { DEBUG = false }()

	var wg sync.WaitGroup
	var succes, fail, cnt, i int64
	cnt = 4
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
			t.Logf("%v Racer %d: can't get token\n", time.Now(), i)
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
func Test15DBMNewTokenRace(t *testing.T) {
	raceNewToken(tDB, "https://golang.org", t)
}

// try to make token expired
func Test20DBMExpireToken(t *testing.T) {
	err := tDB.Expire(DEBUGToken, -1)
	if err != nil {
		t.Error(err)
	}
}

// try to update expired token from several concurrent goroutines
func Test23DBMOneMoreTokenRace(t *testing.T) {
	raceNewToken(tDB, "https://golang.org/pkg/time/error", t)
}

// try to receive long URL by token
func Test25DBMGetToken(t *testing.T) {

	lURL, err := tDB.Get(DEBUGToken)
	if err != nil {
		t.Error(err)
	}
	t.Logf("URL for token %s: %s\n", DEBUGToken, lURL)
}

// try to delete token
func Test30DBMDelToken(t *testing.T) {

	err := tDB.Delete(DEBUGToken)
	if err != nil {
		t.Error(err)
	}

	_, err = tDB.Get(DEBUGToken)
	if err == nil {
		t.Error("no error when expected")
	}
}
