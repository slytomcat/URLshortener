package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	tDB *TokenDB
)

// test new TokenDB creation
func Test10NewTokenDB(t *testing.T) {
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
func Test11ClearTable(t *testing.T) {
	err := tDB.Delete(DEBUGToken)
	if err != nil {
		t.Errorf("Can't clear table: %v", err)
	}
}

// Try to insert the same token twice (error expected)
func Test13OneTokenTwice(t *testing.T) {
	DEBUG = true
	defer func() { DEBUG = false }()

	url := "https://golang.org/pkg/time/"
	token, err := tDB.New(url, 1)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else {
		fmt.Printf("expected result: token for %s: %v\n", url, token)
	}
	token1, err := tDB.New(url, 1)
	if err != nil {
		fmt.Printf("expected error: %s\n", err)
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
func raceNewToken(url string, t *testing.T) {
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
		fmt.Printf("%v Racer %d: Ready!\n", time.Now(), i)
		start.RLock()

		Token, err := tDB.New(url, 1)

		if err != nil {
			fmt.Printf("%v Racer %d: can't get token\n", time.Now(), i)
			atomic.AddInt64(&fail, 1)
			return
		}

		fmt.Printf("%v Racer %d: Token for %s: %v\n", time.Now(), i, url, Token)
		atomic.AddInt64(&succes, 1)
	}

	for i = 0; i < cnt; i++ {
		wg.Add(1)
		go racer(i)
	}
	fmt.Printf("%v Ready?\n", time.Now())
	time.Sleep(time.Second * 2)
	start.Unlock()
	fmt.Printf("%v Go!!!\n", time.Now())
	wg.Wait()

	if succes != 1 || fail != cnt-1 {
		t.Errorf("Concurent update error: success=%d, fail=%d, total=%d", succes, fail, cnt)
	}
}

// try to insert new token from concurrent goroutines
func Test15NewTokenRace(t *testing.T) {
	raceNewToken("https://golang.org", t)
}

// try to make token expired
func Test20ExpireToken(t *testing.T) {
	err := tDB.Expire(DEBUGToken, -1)
	if err != nil {
		t.Error(err)
	}
}

// try to update expired token from several concurrent goroutines
func Test23OneMoreTokenRace(t *testing.T) {
	raceNewToken("https://golang.org/pkg/time/error", t)
}

// try to receive long URL by token
func Test25GetToken(t *testing.T) {

	lURL, err := tDB.Get(DEBUGToken)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("URL for token %s: %s\n", DEBUGToken, lURL)
}

// try to prolong the token
func Test27Prolong(t *testing.T) {
	err := tDB.Expire(DEBUGToken, -1)
	if err != nil {
		t.Error(err)
	}
	err = tDB.Expire(DEBUGToken, 1)
	if err != nil {
		t.Error(err)
	}

	DEBUG = true
	defer func() { DEBUG = false }()

	url := "https://golang.org/pkg/sometime/"
	token1, err := tDB.New(url, 1)
	if err != nil {
		fmt.Printf("expected error (don't panic): %v\n", err)
	} else {
		t.Errorf("unexpected response: token for %s: %v\n", url, token1)
	}

}
