package main

import (
	"errors"
)

// Token is the interface to token database
type Token interface {
	New(longURL string, expiration int) (string, error)
	Get(sToken string) (string, error)
	Expire(sToken string, expiration int) error
	Delete(sToken string) error
}

var (
	// TokenDB - Database interface
	TokenDB Token
)

// NewTokenDB creates new data base interface
func NewTokenDB() (err error) {
	switch CONFIG.DBdriver {
	case "MySQL":
		if TokenDB, err = TokenDBNewM(); err != nil {
			return err
		}
	case "Redis":
		if TokenDB, err = TokenDBNewR(); err != nil {
			return err
		}
	default:
		return errors.New("wrong value of DBdriver configuration parameter")
	}
	return nil
}
