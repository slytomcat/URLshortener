package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func envSet(t testing.TB, files ...string) {
	envs, _ := godotenv.Read(files...)
	for k, v := range envs {
		t.Setenv(k, v)
	}
}

// test config reading with empty URLSHORTENER_CONNECTOPTIONS
func Test01Tools00NoRequiredField(t *testing.T) {

	t.Setenv("URLSHORTENER_REDISADDRS", "")

	_, err := readConfig()

	require.Error(t, err)
	require.Equal(t, "config error: required key URLSHORTENER_REDISADDRS missing value", err.Error())

	os.Unsetenv("URLSHORTENER_REDISADDRS")
	_, err = readConfig()

	require.Error(t, err)
	require.Equal(t, "config error: required key URLSHORTENER_REDISADDRS missing value", err.Error())

}

func Test01Tools01WrongMode(t *testing.T) {
	t.Setenv("URLSHORTENER_REDISADDRS", `"localhost:1234"`)
	t.Setenv("URLSHORTENER_MODE", fmt.Sprint(disableUI<<1))

	_, err := readConfig()

	require.Error(t, err)
	require.Equal(t, "config error: wrong mode value: 10", err.Error())
}

func Test01Tools02Success(t *testing.T) {
	envSet(t, ".env_sample")

	c, err := readConfig()
	require.NoError(t, err)
	require.Equal(t, []string{"<RedisHost>:6379", "<BackupRedisHost>:6379"}, c.RedisAddrs)
	require.Equal(t, "Some long password that is configured for Redis authorization", c.RedisPassword)
	require.Equal(t, 5, c.TokenLength)
	require.Equal(t, "0.0.0.0:80", c.ListenHostPort)
	require.Equal(t, 777, c.Timeout)
	require.Equal(t, 2, c.DefaultExp)
	require.Equal(t, "<short.Domain>", c.ShortDomain)
	require.Equal(t, uint(4), c.Mode)
}
