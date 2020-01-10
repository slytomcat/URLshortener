package main

// Token is the interface to token database
type Token interface {
	New(longURL string, expiration int, timeout int) (string, error)
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
	TokenDB, err = TokenDBNewR()
	return err
}
