package main

// TokenDB provides parallel execution safe methods to store, update and retrieve tokens from data base

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// TokenDB is a structure to handle the DB token operations
type TokenDB struct {
	DB *sql.DB // database connection
}

// TokenDBNew - creates new TokenDB struct and connect to mysql server
func TokenDBNew() (*TokenDB, error) {

	db, err := sql.Open("mysql", CONFIG.DSN)
	if err != nil {
		return nil, err
	}

	// set the connection pool size
	db.SetMaxOpenConns(CONFIG.MaxOpenConns)

	// Check the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &TokenDB{db}, nil
}

// New returns new token for given long URL and store the token expiration period (in days)
func (t *TokenDB) New(longURL string, expiration int) (string, error) {

	// Begin transaction
	tran, err := t.DB.Begin()
	if err != nil {
		return "", fmt.Errorf("can't create transaction: %w", err)
	}

	sToken := "" // token of saved long URL

	// Try 3 times to create new token and insert it into DB table.
	// The token field is unique in DB so it's not possible to insert the same token twice.
	// But if the token is already expired then try to update it (url and expiration).
	for tryCnt := 0; tryCnt < 3; tryCnt++ {
		sToken, err = ShortTokenNew()
		if err != nil {
			return "", err
		}
		// try to insert new token
		_, err = tran.Exec(
			"INSERT INTO urls (`token`, `url`, `exp`) VALUES (?, ?, ?)",
			sToken,
			longURL,
			expiration,
		)
		if err == nil {
			break // the token is successfully inserted
		}
		if !strings.Contains(err.Error(), "Duplicate entry") {
			tran.Rollback()
			return "", fmt.Errorf("can't insert token: %w", err)
		}
		// the token is already in use: try to update the token if it is expired
		result, err := tran.Exec("UPDATE `urls` SET `url`=?, `exp`=? WHERE `token` = ? and DATE_ADD(`ts`, INTERVAL `exp` DAY) < NOW()",
			longURL,
			expiration,
			sToken,
		)
		if err != nil {
			tran.Rollback()
			return "", fmt.Errorf("can't update token: %w", err)
		}

		if affected, _ := result.RowsAffected(); affected == 1 {
			break // the token is successfully updated
		}

		// token is not expired, let's try to select a new one token
		sToken = "" // reset bad token

	}

	if sToken == "" {
		// if we can't insert random token for 3 tries, then it seems that all tokens are busy
		tran.Rollback()
		return "", fmt.Errorf("BD insert error; can't create new random token")
	}
	tran.Commit() // commit the successful insert or update
	return sToken, nil
}

// Get returns long url for given token
func (t *TokenDB) Get(sToken string) (string, error) {
	if len(sToken) != 6 {
		return "", errors.New("wrong token length")
	}

	// get the url by token (ignore expiratinon)
	row := t.DB.QueryRow("SELECT url FROM urls WHERE token = ?", sToken)

	url := ""
	err := row.Scan(&url)
	if err != nil {
		return "", err
	}

	return url, nil
}

// common function for update token expiration
func (t *TokenDB) updateExpiration(sToken string, exp int) error {

	tran, err := t.DB.Begin()
	if err != nil {
		return fmt.Errorf("can't create transaction: %w", err)
	}

	_, err = tran.Exec("UPDATE `urls` SET `exp`=? WHERE `token` = ? ", exp, sToken)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("can't update token: %w", err)
	}
	tran.Commit()
	return nil
}

// Expire - make the token as expired
func (t *TokenDB) Expire(sToken string) error {
	return t.updateExpiration(sToken, -1)
}

// Prolong prolongs token on specified number of days from current datetime
func (t *TokenDB) Prolong(sToken string, exp int) error {
	return t.updateExpiration(sToken, exp)
}
