package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// saveEnv stores configuration environment variables and returns
// function that restores orininal values
func saveEnv() func() {
	vars := []string{
		"URLSHORTENER_REDISOPTIONS",
		"URLSHORTENER_REDISADDRS",
		"URLSHORTENER_TOKENLENGTH",
		"URLSHORTENER_TIMEOUT",
		"URLSHORTENER_LISTENHOSTPORT",
		"URLSHORTENER_DEFAULTEXP",
		"URLSHORTENER_SHORTDOMAIN",
		"URLSHORTENER_MODE",
	}
	save := make(map[string]string, len(vars))
	for _, name := range vars {
		if val, ok := os.LookupEnv(name); ok {
			save[name] = val
		}
	}
	// returned func restores enviroment to original
	return func() {
		for _, name := range vars {
			if value, ok := save[name]; ok {
				os.Setenv(name, value)
			} else {
				os.Unsetenv(name)
			}
		}
	}
}

// test config reading with empty URLSHORTENER_CONNECTOPTIONS
func Test01Tools00NoRequiredField(t *testing.T) {

	defer saveEnv()()

	os.Unsetenv("URLSHORTENER_REDISADDRS")

	_, err := readConfig()

	assert.Error(t, err)
	assert.Equal(t, "config error: required key URLSHORTENER_REDISADDRS missing value", err.Error())
}

func Test01Tools01WrongMode(t *testing.T) {

	defer saveEnv()()

	os.Setenv("URLSHORTENER_REDISADDRS", `"localhost:1234"`)
	os.Setenv("URLSHORTENER_MODE", fmt.Sprint(disableUI<<1))

	_, err := readConfig()

	assert.Error(t, err)
	assert.Equal(t, "config error: wrong mode value: 10", err.Error())
}

func Test01Tools02Success(t *testing.T) {

	defer saveEnv()()
	godotenv.Load(".env_sample")
	_, err := readConfig()
	assert.NoError(t, err)
}
