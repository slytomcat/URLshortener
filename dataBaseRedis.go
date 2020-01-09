package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
)

// TokenDBR is a structure to handle the DB token operations via Redis databasa
type TokenDBR struct {
	db redis.Client
}

// TokenDBNewR creates new database interface to Redis DB
func TokenDBNewR() (*TokenDBR, error) {

	var err error
	d := 0

	res := regexp.MustCompile(`.*:(.*)@(.*)\((.*)\)/(.*)`).FindAllStringSubmatch(CONFIG.DSN, -1)
	if len(res) > 0 {
		d, err = strconv.Atoi(res[0][4])
	}
	if err != nil || len(res) == 0 {
		return nil, errors.New("wrong format of DSN config parameter")
	}

	db := redis.NewClient(&redis.Options{
		Network:  res[0][2],
		Addr:     res[0][3],
		Password: res[0][1],
		DB:       d,
	})

	if _, err := db.Ping().Result(); err != nil {
		return nil, err
	}

	return &TokenDBR{*db}, nil
}

// New creates new token for given long URL
func (t *TokenDBR) New(longURL string, expiration int, timeout int) (string, error) {

	// Using many attempts to setNX token dramatically increases maximum amount of
	// used tokens since:
	// probability of the failure of n attempts = (probability of failure of single attempt)^n.

	// Limit attempts by time not by count
	type replay struct {
		sToken string
		err    error
	}

	rep := make(chan replay)                                      // replay chanel
	stop := time.After(time.Millisecond * time.Duration(timeout)) // time-out chanel

	// start trying to store new token in separate goroutine
	go func() {
		attempt := 0
		for {
			attempt++
			sToken, err := NewShortToken()
			if err != nil {
				rep <- replay{sToken, err}
				return
			}

			ok, err := t.db.SetNX(sToken, longURL, time.Hour*24*time.Duration(expiration)).Result()
			if err == nil && ok {
				// Token stored successfully
				rep <- replay{sToken, nil}
				return
			}
			if err != nil {
				rep <- replay{"", err}
				return
			}

			// stop loop if timeout exceeded
			select {
			case <-stop:
				rep <- replay{"", fmt.Errorf("can't store a new token for %d attempts", attempt)}
				return
			default:
			}
		}
	}()

	r := <-rep

	if r.err != nil {
		return "", r.err
	}

	return r.sToken, nil
}

// Get returns the long URL for given token
func (t *TokenDBR) Get(sToken string) (string, error) {
	return t.db.Get(sToken).Result()
}

// Expire sets new expire datetime for given token
func (t *TokenDBR) Expire(sToken string, expiration int) error {
	return t.db.Expire(sToken, time.Hour*24*time.Duration(expiration)).Err()
}

// Delete removes token from database
func (t *TokenDBR) Delete(sToken string) error {
	return t.db.Del(sToken).Err()
}
