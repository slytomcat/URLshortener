package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
)

// Token is the interface to token database
type Token interface {
	New(longURL string, expiration int) (string, error)
	Get(sToken string) (string, error)
	Expire(sToken string, expiration int) error
	Delete(sToken string) error
	Test() (int, error)
}

var (
	// TokenDB - Database interface
	TokenDB Token
)

// tokenDBR is a structure to handle the DB token operations via Redis databasa
type tokenDBR struct {
	db redis.UniversalClient
}

// NewTokenDB creates new database interface to Redis database
func NewTokenDB() error {

	// create new UniversalClient from CONFIG.ConnectOptions
	db := redis.NewUniversalClient(&CONFIG.ConnectOptions)

	// try to ping data base
	if _, err := db.Ping().Result(); err != nil {
		return err
	}

	// initialize the global variable
	TokenDB = &tokenDBR{db}

	return nil
}

// New creates new token for given long URL
func (t *tokenDBR) New(longURL string, expiration int) (string, error) {

	// Using many attempts to store the new random token dramatically increases maximum amount of
	// used tokens since:
	// probability of the failure of n attempts = (probability of failure of single attempt)^n.

	// Limit number of attempts by time not by count

	stop := time.After(time.Millisecond * time.Duration(CONFIG.Timeout)) // time-out chanel

	// start trying to store new token
	attempt := 0
	for {
		select {
		case <-stop:
			// stop loop if timeout exceeded
			return "", fmt.Errorf("can't store a new token for %d attempts", attempt)
		default:
			attempt++
			sToken, err := NewShortToken(CONFIG.TokenLength)
			if err != nil {
				return "", err
			}

			// try to store token
			ok, err := t.db.SetNX(sToken, longURL, time.Hour*24*time.Duration(expiration)).Result()
			if err == nil && ok {
				// token stored successfully
				return sToken, nil
			}
			if err != nil {
				return "", err
			}
			// !ok mean that duplicate detected
			// try to make an another attempt
		}
	}
}

// Get returns the long URL for given token
func (t *tokenDBR) Get(sToken string) (string, error) {
	// just return result of standard call
	return t.db.Get(sToken).Result()
}

// Expire sets new expire datetime for given token
func (t *tokenDBR) Expire(sToken string, expiration int) error {
	ok, err := t.db.Expire(sToken, time.Hour*24*time.Duration(expiration)).Result()
	if !ok {
		return errors.New("Token is not exists")
	}
	return err
}

// Delete removes token from database
func (t *tokenDBR) Delete(sToken string) error {
	deleted, err := t.db.Del(sToken).Result()
	if deleted == 0 {
		return errors.New("Token is not exists")
	}
	return err
}

// Test measures number of attempts to store token during the new token request time-out
func (t *tokenDBR) Test() (int, error) {

	attempt := 0

	// get token for test
	testToken, err := t.New("test.url", 1)
	if err != nil {
		return 0, err
	}
	defer t.Delete(testToken)

	stop := time.After(time.Millisecond * time.Duration(CONFIG.Timeout)) // time-out chanel

	for {
		select {
		case <-stop:
			return attempt, nil
		default:
			attempt++
			// get new random token to emulate the timings of standard routine
			_, err := NewShortToken(CONFIG.TokenLength)
			if err != nil {
				return attempt, err
			}
			// try to store testToken twice
			ok, err := t.db.SetNX(testToken, "test.url", time.Hour*24).Result()
			if err == nil && ok {
				// token stored twice successfully: it is not good
				return attempt, errors.New("token successfilly stored twice")
			}
			if err != nil {
				// duplicate case should not return error! Something is going wrong
				return attempt, err
			}
			// !ok mean that duplicate detected
			// it's as expected, continue
		}

	}
}
