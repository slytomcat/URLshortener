package main

import (
	"testing"
)

var tDBr Token

// test new TokenDBR creation
func Test10DBRNewTokenDB(t *testing.T) {
	var err error
	err = readConfig("cnfr.json")
	if err != nil {
		t.Error(err)
	}

	tDBr, err = TokenDBNewR()
	if err != nil {
		t.Error(err)
	}
	tDBr.Delete(DEBUGToken)
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
