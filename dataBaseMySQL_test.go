package main

import (
	"testing"

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
	token, err := tDB.New(url, 1, CONFIG.Timeout)
	if err != nil || token == "" {
		t.Errorf("unexpected error: %s; token: %s", err, token)
	} else {
		t.Logf("expected result: token for %s: %v\n", url, token)
	}
	token1, err := tDB.New(url, 1, CONFIG.Timeout)
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
