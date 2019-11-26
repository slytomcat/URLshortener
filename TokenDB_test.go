package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var (
	tDB *TokenDB
)

func Test10NewTokenDB(t *testing.T) {
	var err error
	tDB, err = TokenDBNew()
	if err != nil {
		t.Error(err)
	}
}

func Test15ClearTable(t *testing.T) {
	tx, err := tDB.DB.Begin()
	_, err = tx.Exec("DELETE FROM urls")
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

	racer := func(i int64) {
		defer wg.Done()
		Token, err := tDB.New(url, 1)
		if err != nil {
			fmt.Printf("Racer %d: can't get token\n", i)
			atomic.AddInt64(&fail, 1)
			return
		}
		fmt.Printf("Racer %d: Token for %s: %v\n", i, url, Token)
		atomic.AddInt64(&succes, 1)
	}

	for i = 0; i < cnt; i++ {
		wg.Add(1)
		go racer(i)
	}
	wg.Wait()

	if succes != 1 && fail != cnt-1 {
		t.Errorf("Concurent update error: success=%d, fail=%d", succes, fail)
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
		panic(err)
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
