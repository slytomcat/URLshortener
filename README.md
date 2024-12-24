# URLshortener
[![tests&build&container](https://github.com/slytomcat/URLshortener/actions/workflows/go.yml/badge.svg)](https://github.com/slytomcat/URLshortener/actions/workflows/go.yml)
[![Platform: linux-64](https://img.shields.io/badge/Platform-linux--64-blue)]()
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Docker image](https://img.shields.io/badge/Docker-image-blue)](https://github.com/slytomcat/URLshortener/pkgs/container/urlshortener)

`URLshortener` is a micro-service to shorten long URLs and to handle the redirection by generated short URLs.

The service requires Redis database connection. See example how to run Redis in Docker in [redisDockerRun.sh](https://github.com/slytomcat/URLshortener/blob/master/redisDockerRun.sh)

`URLshortener` performs a self-health-check on start. If `URLshortener` misconfigured or initial health-check failed then it returns non zero exit code.

`URLshortener -v` outputs version info end exits with zero exit code.

### Web UI for short URL generation:

URL `<host>[:<port>]/ui/generate`

Method: `GET`

Shows the simple user interface for short URL generation. It also generated QR code containing the short URL.

Note that the lifetime of generated short URL is via `DefaultExp` value in the configuration file.

Request example using `s-t-c.tk` (micro-service demo):

Open in browser: `http://s-t-c.tk/ui/generate`

### Request for short URL:

URL: `<host>[:<port>]/api/v1/token`

Method: `POST`

Request body: JSON with following parameters:

- `url`: string, URL to shorten, mandatory
- `exp`: int, short URL expiration in days, optional, default: value of `"DefaultExp"` from configuration file

Success response: `HTTP 200 OK` with body containing JSON with following parameters:

- `token`: string, token for short URL
- `url`: string, short URL

Request example using `curl` and `s-t-c.tk` (micro-service demo):

`curl -v POST -H "Content-Type: application/json" -d '{"url":"<long url>","exp":10}' http://s-t-c.tk/api/v1/token`

Note: Token is created as random and the saving it to DB may cause duplicate error. In order to avoid such error the service makes several attempts to store random token. The number of attempts is limited by the `URLSHORTENER_TIMEOUT` configuration value by time, not by count of attempts. When time-out expired and no one attempt was successful then service returns response code `408 Request Timeout`. This response mean that the request can be repeated.

The maximum number of possible attempts to store token during time-out is calculated every time a new token stored. The last measured value is displayed on the homepage.

If measured number of attempts is too small (1-5) then log can contain warnings like: `Warning: Too low number of attempts: 5 per timeout (500 ms)`. In such a case consider increasing of `URLSHORTENER_TIMEOUT` configuration value. Number of attempts above 200 is more then enough, you may consider to decrease `URLSHORTENER_TIMEOUT` configuration value. 30-40 attempts allows to fulfill the space of tokens up to 70-80% before timeout errors (during request for short token) occasionally appears.

When the service time-out errors appears often and log contains many errors like `can't store a new token for 75 attempts` then it most probably means that active (not expired) token amount is near to maximum possible tokens amount (for configured token length). Consider increasing of token length (`TokenLength` configuration value) or decrease token expiration (`DefaultExp` configuration value and/or `exp` parameter in the request for new short URL). You can also delete some tokens from redis DB.

Note also the log warnings such as `Warning: Measured 45 attempts for 423621 ns. Calculated 62 max attempts per 500 ms`. Such warnings also can be a signal that token space is filled near to maximum capacity.


### Request for set new expiration of token:

URL: `<host>[:<port>]/api/v1/expire`

Method: `POST`

Request body: JSON with following parameter:

- `token`: string, token for short URL, mandatory.
- `exp`: int, new expiration in days from now, optional. Default: 0 - value that marks token as expired immediately.

Success response: `HTTP 200 OK` with empty body

Request example using `curl` and `s-t-c.tk` (micro-service demo):

`curl -v POST -H "Content-Type: application/json" -d '{"token":"<token>","exp":<exp>}' http://s-t-c.tk/api/v1/expire`


### Redirect to long URL:
URL: `<host>[:<port>]/<token>` - URL from response on request for short URL

Method: `GET`

Response contain the redirection to long URL (response code: HTTP 302 'Found' with 'Location' = long URL in response header)

Request example using `s-t-c.tk` (micro-service demo):

Via `curl`:

`curl -i -v http://s-t-c.tk/<token>`

Via browser:

`http://s-t-c.tk/<token>`


### Health-check:
URL: `<host>[:<port>]/`

Method: `GET`

Response: simple home page and `HTTP 200 OK` in case of successful self-health-check, or `HTTP 500 Server error` in case of any error during self-health-check.

Request example using `curl` and `s-t-c.tk` (micro-service demo):

`curl -i -v http://s-t-c.tk/`


### Service configuration

Service configuration is made via environment variables.

The following variables are read on start:
 - URLSHORTENER_REDISADDRS: comma separated list of redis cluster/sentinel nodes (address:port,address:port...) or a single address:port value for single node redis. The value is mandatory.
 - URLSHORTENER_REDISPASSWORD: password for Redis authorization. The value is optional (empty by default). But it is strongly recommended DO NOT USE THE REDIS WITHOUT AUTHORISATION!
 - URLSHORTENER_TOKENLENGTH: length of short token, default: 6
 - URLSHORTENER_LISTENHOSTPORT: service listening host:port, default: localhost:8080
 - URLSHORTENER_TIMEOUT: A new token creation timeout in milliseconds, default: 500
 - URLSHORTENER_DEFAULTEXP: Default token expiration time in days, default: 1
 - URLSHORTENER_SHORTDOMAIN: the short domain to use in short URL, default: localhost:8080
 - URLSHORTENER_MODE: The service mode: the value combined as summary of options (see below), default: 0

The service mode options are:
    1 - disable redirects
    2 - disable request for new short URL creation
    4 - disable expire request
    8 - disable UI page for short URL creation

### Logs

Log is written to output. It contains access log, request results and some warnings about the the measurements of attempts per time-out.
