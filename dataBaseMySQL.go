package main

// TokenDB provides parallel execution safe methods to store, update and
// retrieve tokens from data base.

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// TokenDBM is a structure to handle the DB token operations
type TokenDBM struct {
	db *sql.DB // database connection
}

// TokenDBNewM - creates new TokenDB struct and connect to mysql server
func TokenDBNewM() (*TokenDBM, error) {

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

	return &TokenDBM{db}, nil
}

// New returns new token for given long URL and store the token expiration period (in days)
func (t *TokenDBM) New(longURL string, expiration int, timeout int) (string, error) {

	// Try several times to create new token and insert it into DB table.
	// The token field is unique in DB so it's not possible to insert the same token twice.
	// But if the token is already expired then try to update it (set new url and expiration).

	// Using several attempts to store new token dramatically increases maximum amount of
	// used tokens since :
	// probability of the failure of n attempts = (probability of failure of single attempt)^n.

	// Limit attempts by time not by count
	stop := time.After(time.Millisecond * time.Duration(timeout))

	attempt := 0
	for {
		attempt++
		sToken, err := NewShortToken()
		if err != nil {
			return sToken, err
		}

		// begin transaction
		tran, err := t.db.Begin()
		if err != nil {
			return "", fmt.Errorf("can't create transaction: %w", err)
		}

		// try to store new token
		_, err = tran.Exec(
			"INSERT INTO urls (`token`, `url`, `exp`) VALUES (?, ?, ?)",
			sToken,
			longURL,
			expiration,
		)
		if err == nil {
			// the token is successfully inserted
			tran.Commit()
			return sToken, nil
		}

		// handle error if it is not Duplicate entry error
		if !strings.Contains(err.Error(), "Duplicate entry") {
			tran.Rollback()
			return "", fmt.Errorf("can't insert token: %w", err)
		}

		// close unsuccessful transaction and create new one to avoid deadlocks
		tran.Rollback()
		tran, err = t.db.Begin()
		if err != nil {
			return "", fmt.Errorf("can't create transaction: %w", err)
		}

		// the token is already exists: try to update the token if it is expired
		result, err := tran.Exec("UPDATE `urls` SET `url`=?, `exp`=? WHERE `token` = ? and DATE_ADD(`ts`, INTERVAL `exp` DAY) < NOW()",
			longURL,
			expiration,
			sToken,
		)
		if err != nil {
			tran.Rollback()
			return "", fmt.Errorf("can't update token: %w", err)
		}

		// check affected rows
		affected, err := result.RowsAffected()
		if err == nil && affected == 1 {
			// token successfully updated
			tran.Commit()
			return sToken, nil
		}
		tran.Rollback()
		if err != nil {
			return "", fmt.Errorf("can't get affected rows: %w", err)
		}

		// we didn't manage to insert or update the token

		select {
		case <-stop:
			// stop loop if time-out exceeded
			return "", fmt.Errorf("can't store a new token for %d attempts", attempt)
		default:
			// make next attempt as time-out is not exceeded yet
		}

	}
}

// Get returns long url for given token
func (t *TokenDBM) Get(sToken string) (string, error) {
	if len(sToken) != tokenLenS {
		return "", errors.New("wrong token length")
	}

	// get the url by token checking expiratinon
	row := t.db.QueryRow("SELECT url FROM urls WHERE token = ? and DATE_ADD(`ts`, INTERVAL `exp` DAY) > NOW()", sToken)

	url := ""
	err := row.Scan(&url)
	if err != nil {
		return "", err
	}

	return url, nil
}

// Expire - set new expiration on the token
// Use zero or negative exp value to expire token
func (t *TokenDBM) Expire(sToken string, exp int) error {

	// begin transaction
	tran, err := t.db.Begin()
	if err != nil {
		return fmt.Errorf("can't create transaction: %w", err)
	}

	// try update token
	result, err := tran.Exec("UPDATE `urls` SET `exp`=? WHERE `token` = ? ", exp, sToken)
	if err != nil {
		tran.Rollback()
		return fmt.Errorf("can't update token: %w", err)
	}

	// check affected rows
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		tran.Rollback()
		if err != nil {
			return fmt.Errorf("can't get affected rows: %w", err)
		}
		return errors.New("token is not found")
	}

	// commit transaction
	tran.Commit()
	return nil
}

// Delete deletes specified token. It requered mostly for tests
func (t *TokenDBM) Delete(sToken string) error {
	tx, _ := t.db.Begin()
	_, err := tx.Exec("DELETE FROM urls WHERE token=?", DEBUGToken)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
