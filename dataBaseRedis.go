package main

import (
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v7"
)

// Token is the interface to token database
type Token interface {
	New(longURL string, expiration int) (string, error)
	Get(sToken string) (string, error)
	Expire(sToken string, expiration int) error
	Delete(sToken string) error
	Attempts() int
}

// tokenDBR is a structure to handle the DB token operations via Redis database
type tokenDBR struct {
	db          redis.UniversalClient
	timeout     int   // new token record creation timeout
	tokenLength int   // length of token
	attempts    int32 // number of attemppts during timeout (calculated)
}

// NewTokenDB creates new database interface to Redis database
func NewTokenDB(connect redis.UniversalOptions, timeout, tokenLength int) (Token, error) {

	// create new UniversalClient from CONFIG.ConnectOptions
	db := redis.NewUniversalClient(&connect)

	// try to ping data base
	if _, err := db.Ping().Result(); err != nil {
		return nil, err
	}

	return &tokenDBR{
			db:          db,
			timeout:     timeout,
			tokenLength: tokenLength,
		},
		nil
}

// New creates new token for given long URL
func (t *tokenDBR) New(longURL string, expiration int) (string, error) {

	// Using many attempts to store the new random token dramatically increases maximum amount of
	// used tokens since:
	// probability of the failure of n attempts = (probability of failure of single attempt)^n.

	// Limit number of attempts by time not by count

	// Count attempts and time for reports
	var attempt, startTime int64

	// Calculate statistics and report if some dangerous situation appears
	defer func() {
		elapsedTime := time.Now().UnixNano() - startTime
		// perform statistical calculation and reporting in another go-routine
		go func() {
			if attempt > 0 {
				MaxAtt := attempt * int64(t.timeout) * 1000000 / elapsedTime
				// use atomic to avoid race conditions
				atomic.StoreInt32(&t.attempts, int32(MaxAtt))
				// report warnings of some not good measurements
				if MaxAtt*3/4 < attempt {
					log.Printf("Warning: Measured %d attempts for %d ns. Calculated %d max attempts per %d ms\n", attempt, elapsedTime, MaxAtt, t.timeout)
				}
				if MaxAtt > 0 && MaxAtt < 10 {
					log.Printf("Warning: Too low number of attempts: %d per timeout (%d ms)\n", MaxAtt, t.timeout)
				}
			}
		}()
	}()

	// make time-out chanel
	stop := time.After(time.Millisecond * time.Duration(t.timeout))

	// Remember starting time
	startTime = time.Now().UnixNano()

	// start trying to store new token
	for {
		select {
		case <-stop:
			// stop loop if timeout exceeded
			return "", fmt.Errorf("can't store a new token for %d attempts", attempt)
		default:
			// get random token
			sToken, err := NewShortToken(t.tokenLength)
			if err != nil {
				return "", fmt.Errorf("NewShortToken error: %w", err)
			}

			// count attempts
			attempt++

			// try to store token
			ok, err := t.db.SetNX(sToken, longURL, time.Hour*24*time.Duration(expiration)).Result()
			if err == nil && ok {
				// token stored successfully
				return sToken, nil
			}
			if err != nil {
				return "", fmt.Errorf("SetNX operation error: %w", err)
			}
			// !ok mean that duplicate detected
			// try to make an another attempt
		}
	}
}

// checkTokenLenth do the work as described in name
func (t *tokenDBR) checkTokenLenth(sToken string) error {
	if len(sToken) != t.tokenLength {
		return errors.New("wrong token length")
	}
	return nil
}

// Get returns the long URL for given token
func (t *tokenDBR) Get(sToken string) (string, error) {

	// check token length
	if err := t.checkTokenLenth(sToken); err != nil {
		return "", err
	}

	// if length is ok than just return result of standard call
	return t.db.Get(sToken).Result()
}

// Expire sets new expire datetime for given token
func (t *tokenDBR) Expire(sToken string, expiration int) error {

	// check token length
	if err := t.checkTokenLenth(sToken); err != nil {
		return err
	}

	// try to change the token expiration
	ok, err := t.db.Expire(sToken, time.Hour*24*time.Duration(expiration)).Result()
	// check the result status
	if err == nil && !ok {
		return errors.New("token is not exists")
	}
	return err
}

// Delete removes token from database
func (t *tokenDBR) Delete(sToken string) error {

	// check token length
	if err := t.checkTokenLenth(sToken); err != nil {
		return err
	}

	deleted, err := t.db.Del(sToken).Result()
	// check the number deleted tokens
	if err == nil && deleted == 0 {
		return errors.New("token is not exists")
	}
	return err
}

// Attempts returns the number of attempts (calculated) during the time-out
func (t *tokenDBR) Attempts() int {
	return int(atomic.LoadInt32(&t.attempts))
}
