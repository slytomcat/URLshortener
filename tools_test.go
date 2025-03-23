package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func envSet(t testing.TB, files ...string) {
	if os.Getenv("CI") != "true" {
		envs, _ := godotenv.Read(files...)
		for k, v := range envs {
			t.Setenv(k, v)
		}
	}
}

// test config reading with empty URLSHORTENER_CONNECTOPTIONS
func Test01Tools00NoRequiredField(t *testing.T) {

	t.Setenv("URLSHORTENER_REDISADDRS", "")

	c, err := readConfig()

	require.Nil(t, c)

	require.Error(t, err)
	require.Equal(t, "config error: wrong or missed value of URLSHORTENER_REDISADDRS", err.Error())

	os.Unsetenv("URLSHORTENER_REDISADDRS")
	_, err = readConfig()

	require.Error(t, err)
	require.Equal(t, "config error: wrong or missed value of URLSHORTENER_REDISADDRS", err.Error())

}
func Test01Tools01WrongTimeout(t *testing.T) {
	t.Setenv("URLSHORTENER_REDISADDRS", `"localhost:1234"`)
	t.Setenv("URLSHORTENER_TIMEOUT", "z")
	_, err := readConfig()
	require.Error(t, err)
	require.EqualError(t, err, "config error: wrong value of URLSHORTENER_TIMEOUT: strconv.ParseUint: parsing \"z\": invalid syntax")
	t.Setenv("URLSHORTENER_TIMEOUT", "-2")
	_, err = readConfig()
	require.Error(t, err)
	require.EqualError(t, err, "config error: wrong value of URLSHORTENER_TIMEOUT: strconv.ParseUint: parsing \"-2\": invalid syntax")
}

func Test01Tools02WrongTokenLength(t *testing.T) {
	t.Setenv("URLSHORTENER_REDISADDRS", `"localhost:1234"`)
	t.Setenv("URLSHORTENER_TOKENLENGTH", "z")
	_, err := readConfig()
	require.Error(t, err)
	require.EqualError(t, err, "config error: wrong value of URLSHORTENER_TOKENLENGTH: strconv.ParseUint: parsing \"z\": invalid syntax")
	t.Setenv("URLSHORTENER_TOKENLENGTH", "-2")
	_, err = readConfig()
	require.Error(t, err)
	require.EqualError(t, err, "config error: wrong value of URLSHORTENER_TOKENLENGTH: strconv.ParseUint: parsing \"-2\": invalid syntax")
}

func Test01Tools02WrongExp(t *testing.T) {
	t.Setenv("URLSHORTENER_REDISADDRS", `"localhost:1234"`)
	t.Setenv("URLSHORTENER_DEFAULTEXP", "z")
	_, err := readConfig()
	require.Error(t, err)
	require.EqualError(t, err, "config error: wrong value of URLSHORTENER_DEFAULTEXP: strconv.ParseUint: parsing \"z\": invalid syntax")
	t.Setenv("URLSHORTENER_DEFAULTEXP", "-2")
	_, err = readConfig()
	require.Error(t, err)
	require.EqualError(t, err, "config error: wrong value of URLSHORTENER_DEFAULTEXP: strconv.ParseUint: parsing \"-2\": invalid syntax")
}

func Test01Tools04WrongMode(t *testing.T) {
	t.Setenv("URLSHORTENER_REDISADDRS", `"localhost:1234"`)
	t.Setenv("URLSHORTENER_MODE", "z")

	_, err := readConfig()

	require.Error(t, err)
	require.Equal(t, "config error: wrong value of URLSHORTENER_MODE: strconv.ParseUint: parsing \"z\": invalid syntax", err.Error())
	t.Setenv("URLSHORTENER_MODE", fmt.Sprint(incorrectOption))

	_, err = readConfig()

	require.Error(t, err)
	require.Equal(t, "config error: wrong value of URLSHORTENER_MODE: 20H (32)", err.Error())
}

func Test01Tools05Success(t *testing.T) {
	t.Setenv(envRedisAddrs, "<RedisHost>:6379,<BackupRedisHost>:6379")
	t.Setenv(envRedisPassword, "Some long password that is configured for Redis authorization")
	t.Setenv(envTokenLength, "5")
	t.Setenv(envListenHostPort, "0.0.0.0:80")
	t.Setenv(envTimeout, "777")
	t.Setenv(envDefaultExp, "2")
	t.Setenv(envShortDomain, "<short.Domain>")
	t.Setenv(envMode, "4")
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
