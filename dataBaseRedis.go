package main

import (
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v7"
)

// TokenDB is the interface to token database
type TokenDB interface {
	Set(sToken, longURL string, expiration int) (bool, error)
	Get(sToken string) (string, error)
	Expire(sToken string, expiration int) error
	Delete(sToken string) error
	Close() error
}

// tokenDBR is a structure to handle the DB token operations via Redis database
type tokenDBR struct {
	db redis.UniversalClient
}

// NewTokenDB creates new database interface to Redis database
func NewTokenDB(connect redis.UniversalOptions) (TokenDB, error) {

	// create new UniversalClient from CONFIG.ConnectOptions
	db := redis.NewUniversalClient(&connect)

	// try to ping data base
	if _, err := db.Ping().Result(); err != nil {
		return nil, err
	}

	return &tokenDBR{db}, nil
}

// New creates new token for given long URL
func (t *tokenDBR) Set(sToken, longURL string, expiration int) (bool, error) {

	// try to store token
	return t.db.SetNX(sToken, longURL, time.Hour*24*time.Duration(expiration)).Result()
}

// Get returns the long URL for given token
func (t *tokenDBR) Get(sToken string) (string, error) {

	// if length is ok than just return result of standard call
	return t.db.Get(sToken).Result()
}

// Expire sets new expire datetime for given token
func (t *tokenDBR) Expire(sToken string, expiration int) error {

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

	deleted, err := t.db.Del(sToken).Result()
	// check the number deleted tokens
	if err == nil && deleted == 0 {
		return errors.New("token is not exists")
	}
	return err
}

// Close - flush data and close connection to database
func (t *tokenDBR) Close() error {
	_, err := t.db.BgSave().Result()
	if err != nil {
		log.Printf("BGSave error: %v", err)
	}
	return t.db.Close()
}
