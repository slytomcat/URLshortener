package main

import (
	"errors"
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
	if len(res) == 0 {
		return nil, errors.New("wrong format of DSN config parameter")
	}

	d, err = strconv.Atoi(res[0][4])
	if err != nil {
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
func (t *TokenDBR) New(longURL string, expiration int) (string, error) {

	sToken, err := NewShortToken()
	if err != nil {
		return "", err
	}

	ok, err := t.db.SetNX(sToken, longURL, time.Hour*24*time.Duration(expiration)).Result()
	if !ok {
		return "", errors.New("can't store a new token")
	}
	if err != nil {
		return "", err
	}

	return sToken, nil
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
