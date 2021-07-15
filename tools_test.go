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
		"URLSHORTENER_REDISPASSWORD",
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
			os.Unsetenv(name)
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
	c, err := readConfig()
	assert.NoError(t, err)
	assert.Equal(t, []string{"<RedisHost>:6379", "<BackupRedisHost>:6379"}, c.RedisAddrs)
	assert.Equal(t, "Some long password that is configured for Redis authorization", c.RedisPassword)
	assert.Equal(t, 5, c.TokenLength)
	assert.Equal(t, "0.0.0.0:80", c.ListenHostPort)
	assert.Equal(t, 777, c.Timeout)
	assert.Equal(t, 2, c.DefaultExp)
	assert.Equal(t, "<short.Domain>", c.ShortDomain)
	assert.Equal(t, uint(4), c.Mode)
}
