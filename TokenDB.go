package main

// TokenDB provides parallel execution safe methods to store, update and retrieve tokens from data base

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// TokenDB is a structure to handle the DB token operations
type TokenDB struct {
	DB *sql.DB // database connection
}

// readDSN reads dsn from config file
func readDSN() (string, error) {
	cfgFile := ".cnf.json"
	f, err := os.Open(cfgFile)
	if err != nil {
		return "", fmt.Errorf("configuration file '%s' can't be read: %w", cfgFile, err)
	}
	defer f.Close()
	cfg := make(map[string]interface{})
	err = json.NewDecoder(f).Decode(&cfg)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("configuration file '%s' can be parsed: %w", cfgFile, err)
	}

	dsn, ok := cfg["URLSHORTENER_DSN"].(string)
	if !ok || dsn == "" {
		return "", fmt.Errorf("configuration file '%s' sould contain URLSHORTENER_DSN variable with value like '<user>:<password>@<protocol>(<host>:<port>)/<database>'", cfgFile)
	}
	return dsn, nil
}

// TokenDBNew - creates new TokenDB struct and connect to mysql server
func TokenDBNew() (*TokenDB, error) {
	var err error
	dsn := os.Getenv("URLSHORTENER_DSN")
	if dsn == "" {
		dsn, err = readDSN()
		if err != nil {
			return nil, fmt.Errorf("env variable URLSHORTENER_DSN is not set and error occurred while config file reading: %w", err)
		}
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &TokenDB{db}, nil
}

// New returns new token for given long URL and store the token expiration period (in days)
func (t *TokenDB) New(longURL string, expiration int) (string, error) {
	count := 0
	sToken := ""

	tran, err := t.DB.Begin()
	if err != nil {
		return "", fmt.Errorf("can't create transaction: %w", err)
	}

	// Try several times to create new token and insert it into DB table.
	// The token field is unique in DB so it's not possible to insert the same token twice.
	// But if the token already expired then try to update it (url and expiration).
	for {
		tc, err := ShortTokenNew()
		if err != nil {
			return "", err
		}
		sToken = tc.BASE64

		_, err = tran.Exec(
			"INSERT INTO urls (`token`, `url`, `exp`) VALUES (?, ?, ?)",
			tc.Bytes,
			longURL,
			expiration,
		)
		if err == nil {
			break
		} else { // the token ia alredy in use
			if strings.Contains(err.Error(), "Duplicate entry") {
				// try to update the token if it is expired
				result, err := tran.Exec("UPDATE `urls` SET `url`=?, `exp`=? WHERE `token` = ? and DATE_ADD(`ts`, INTERVAL `exp` DAY) < NOW()",
					longURL,
					expiration,
					tc.Bytes,
				)
				if err != nil {
					tran.Rollback()
					return "", fmt.Errorf("can't update token: %w", err)
				}
				affected, _ := result.RowsAffected()
				if affected == 1 {
					break
				}
				// token is not expired or some one else already updated this token, let's try to select new one token
				tran.Rollback()
			}
		}
		// try 3 time to insert 3 random tokens
		if count++; count > 2 {
			// if we can't insert random token for 3 tries, then it seems that all tokens are busy
			tran.Rollback()
			return "", fmt.Errorf("BD insert error; can't create new random token")
		}
	}
	tran.Commit()
	return sToken, nil
}

// Get returns long url for given token
func (t *TokenDB) Get(sToken string) (string, error) {
	tc, err := ShortTokenSet(sToken)
	if err != nil {
		return "", err
	}

	// get the url by token (ignore expiratinon)
	row := t.DB.QueryRow("SELECT url FROM urls WHERE token = ?", tc.Bytes)

	url := ""
	err = row.Scan(&url)
	if err != nil {
		return "", err
	}

	return url, nil
}

// common function for update token expiration
func (t *TokenDB) updateExpiration(sToken string, exp int) error {

	tc, err := ShortTokenSet(sToken)
	if err != nil {
		return err
	}

	tran, err := t.DB.Begin()
	if err != nil {
		return fmt.Errorf("can't create transaction: %w", err)
	}

	_, err = tran.Exec("UPDATE `urls` SET `exp`=? WHERE `token` = ? ", exp, tc.Bytes)
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
