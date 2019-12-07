package main

// TokenDB provides parallel execution safe methods to store, update and
// retrieve tokens from data base.

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
	var err error
	// token of saved long URL
	sToken := ""

	// Try 3 times to create new token and insert it into DB table.
	// The token field is unique in DB so it's not possible to insert the same token twice.
	// But if the token is already expired then try to update it (url and expiration).
	for tryCnt := 0; tryCnt < 3; tryCnt++ {

		// get new token
		sToken, err = NewShortToken()
		if err != nil {
			return "", err
		}

		// begin transaction
		tran, err := t.DB.Begin()
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
			// the token is successfully stored
			tran.Commit()
			break
		}

		// handle error if it is not Duplicate entry error
		if !strings.Contains(err.Error(), "Duplicate entry") {
			tran.Rollback()
			return "", fmt.Errorf("can't insert token: %w", err)
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
			break
		}
		tran.Rollback()
		if err != nil {
			return "", fmt.Errorf("can't get affected rows: %w", err)
		}
		// token is not updated
		// reset bad token
		sToken = ""
	}

	if sToken == "" {
		// if we can't insert random token for 3 tries, then it seems that all tokens are busy
		return "", fmt.Errorf("can't create new token")
	}
	// commit the successful insert or update

	return sToken, nil
}

// Get returns long url for given token
func (t *TokenDB) Get(sToken string) (string, error) {
	if len(sToken) != tokenLenS {
		return "", errors.New("wrong token length")
	}

	// get the url by token checking expiratinon
	row := t.DB.QueryRow("SELECT url FROM urls WHERE token = ? and DATE_ADD(`ts`, INTERVAL `exp` DAY) > NOW()", sToken)

	url := ""
	err := row.Scan(&url)
	if err != nil {
		return "", err
	}

	return url, nil
}

// common function for update token expiration
func (t *TokenDB) updateExpiration(sToken string, exp int) error {

	// begin transaction
	tran, err := t.DB.Begin()
	if err != nil {
		return fmt.Errorf("can't create transaction: %w", err)
	}

	// update token
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

// Expire - make the token as expired
func (t *TokenDB) Expire(sToken string) error {
	return t.updateExpiration(sToken, -1)
}

// Prolong prolongs token on specified number of days from current datetime
func (t *TokenDB) Prolong(sToken string, exp int) error {
	return t.updateExpiration(sToken, exp)
}
