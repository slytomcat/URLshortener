package main

import (
	"math/rand"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var tDBr Token

// test new TokenDBR creation
func Test10DBRNewTokenDB(t *testing.T) {
	var err error

	saveCONFIG := CONFIG
	saveDriver := os.Getenv("URLSHORTENER_DBdriver")
	saveDSN := os.Getenv("URLSHORTENER_DSN")
	saveTokenDB := TokenDB
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_DBdriver", saveDriver)
		os.Setenv("URLSHORTENER_DSN", saveDSN)
		TokenDB = saveTokenDB
	}()

	os.Setenv("URLSHORTENER_DSN", os.Getenv("URLSHORTENER_DSNR"))
	os.Setenv("URLSHORTENER_DBdriver", "Redis")

	err = readConfig("cnfr.json")
	if err != nil {
		t.Error(err)
	}

	err = NewTokenDB()
	if err != nil {
		t.Error(err)
	}

	tDBr = TokenDB

	CONFIG.DBdriver = "wrongValue"

	err = NewTokenDB()
	if err == nil {
		t.Error("No error when expected")
	}

	CONFIG.DBdriver = "Redis"
	CONFIG.DSN = "wrongValue"

	err = NewTokenDB()
	if err == nil {
		t.Error("No error when expected")
	}

	CONFIG.DBdriver = "Redis"
	CONFIG.DSN = ":wrongPass@tcp(wrongHost:6379)/0"

	err = NewTokenDB()
	if err == nil {
		t.Error("No error when expected")
	}

	tDBr.Delete(DEBUGToken)
}

func Test13DBROneTokenTwice(t *testing.T) {
	DEBUG = true
	defer func() { DEBUG = false }()

	url := "https://golang.org/pkg/time/"
	token, err := tDBr.New(url, 1, newTokenTimeOut)
	if err != nil || token == "" {
		t.Errorf("unexpected error: %s; token: %s", err, token)
	} else {
		t.Logf("expected result: token for %s: %v\n", url, token)
	}
	token1, err := tDBr.New(url, 1, newTokenTimeOut)
	if err != nil {
		t.Logf("expected error: %s\n", err)
	} else {
		t.Errorf("wrong result: token for %s: %v\n", url, token1)
	}
	// clear
	err = tDBr.Delete(DEBUGToken)
	if err != nil {
		t.Errorf("Can't delete token: %v", err)
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

		token, err := db.New(url, 1, newTokenTimeOut)

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
func Test15DBRNewTokenRace(t *testing.T) {
	raceNewToken(tDBr, "https://golang.org", t)
}

// try to make token expired
func Test20DBRExpireToken(t *testing.T) {
	err := tDBr.Expire(DEBUGToken, -1)
	if err != nil {
		t.Error(err)
	}
}

// try to update expired token from several concurrent goroutines
func Test23DBROneMoreTokenRace(t *testing.T) {
	raceNewToken(tDBr, "https://golang.org/pkg/time/error", t)
}

// try to receive long URL by token
func Test25DBRGetToken(t *testing.T) {

	lURL, err := tDBr.Get(DEBUGToken)
	if err != nil {
		t.Error(err)
	}
	t.Logf("URL for token %s: %s\n", DEBUGToken, lURL)
}

// try to delete token
func Test30DBRDelToken(t *testing.T) {

	err := tDBr.Delete(DEBUGToken)
	if err != nil {
		t.Error(err)
	}

	_, err = tDBr.Get(DEBUGToken)
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to expire non existing token
func Test35DBRExpNonExisting(t *testing.T) {
	err := tDBr.Expire("#$%^&*(", -1)
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to delete non existing token
func Test40DBRDelNonExisting(t *testing.T) {
	err := tDBr.Delete("#$%^&*(")
	if err == nil {
		t.Error("no error when expected")
	}
}

// try to get non existing token
func Test40DBRGetNonExisting(t *testing.T) {
	_, err := tDBr.Get("#$%^&*(")
	if err == nil {
		t.Error("no error when expected")
	}
}

// Test debug error
func Test50DBRDebugError(t *testing.T) {
	DEBUG = true
	DEBUGToken = "error"
	defer func() {
		DEBUG = false
		DEBUGToken = strings.Repeat("_", tokenLenS)
	}()
	_, err := tDBr.New("someUrl.com", 1, 10)
	if err == nil {
		t.Error("no error when expected")
	}
}
