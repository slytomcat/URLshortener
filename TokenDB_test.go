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

func Test10NewTokenDB(t *testing.T) {
	var err error
	err = readConfig(".cnf.json")
	if err != nil {
		t.Error(err)
	}

	tDB, err = TokenDBNew()
	if err != nil {
		t.Error(err)
	}
}

func Test15ClearTable(t *testing.T) {
	tx, _ := tDB.DB.Begin()
	_, err := tx.Exec("DELETE FROM urls")
	if err != nil {
		t.Errorf("Can't clear table: %v", err)
	}
	tx.Commit()
}

func raceNewToken(url string, t *testing.T) {
	DEBUG = true

	var wg sync.WaitGroup
	var succes, fail, cnt, i int64
	cnt = 5
	var start sync.RWMutex

	start.Lock()
	rand.Seed(time.Now().UnixNano())

	racer := func(i int64) {
		defer wg.Done()

		time.Sleep(time.Duration(rand.Intn(10)) * time.Microsecond * 100)
		startTime := time.Now()
		start.RLock()
		fmt.Printf("%v Racer %d: Ready!\n", startTime, i)

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
	//fmt.Printf("%v Ready?\n", time.Now())
	time.Sleep(time.Second * 2)
	start.Unlock()
	//fmt.Printf("%v Go!!!\n", time.Now())
	wg.Wait()

	if succes != 1 || fail != cnt-1 {
		t.Errorf("Concurent update error: success=%d, fail=%d, total=%d", succes, fail, cnt)
	}
}

func Test20NewToken(t *testing.T) {
	raceNewToken("https://golang.org", t)
}

func Test30OneMoreToken(t *testing.T) {
	DEBUG = true
	url := "https://golang.org/pkg/time/"
	token1, err := tDB.New(url, 1)
	if err != nil {
		fmt.Println("ERROR (expected, don't panic): ", err)
	} else {
		fmt.Printf("Token for %s: %v\n", url, token1)
	}
}

func Test35ExpireToken(t *testing.T) {
	err := tDB.Expire("______")
	if err != nil {
		t.Error(err)
	}
}

func Test40OneMoreTokenRace(t *testing.T) {
	raceNewToken("https://golang.org/pkg/time/error", t)
}

func Test45GetToken(t *testing.T) {

	lURL, err := tDB.Get("______")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("URL for token ______: %s", lURL)
}

func Test50Prolong(t *testing.T) {
	err := tDB.Expire("______")
	if err != nil {
		t.Error(err)
	}
	err = tDB.Prolong("______", 1)
	if err != nil {
		t.Error(err)
	}
	DEBUG = true
	url := "https://golang.org/pkg/sometime/"
	token1, err := tDB.New(url, 1)
	if err != nil {
		fmt.Println("ERROR (don't panic): ", err)
	} else {
		fmt.Printf("Token for %s: %v\n", url, token1)
	}

}
