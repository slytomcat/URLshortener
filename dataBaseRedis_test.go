package main

import (
	"os"
	"testing"
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
