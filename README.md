# URLshortener
[![tests&bild&container](https://github.com/slytomcat/URLshortener/actions/workflows/go.yml/badge.svg)](https://github.com/slytomcat/URLshortener/actions/workflows/go.yml)
[![Platform: linux-64](https://img.shields.io/badge/Platform-linux--64-blue)]()
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Docker image](https://img.shields.io/badge/Docker-image-blue)](https://hub.docker.com/r/slytomcat/urlshortener)

`URLshortener` is a micro-service to shorten long URLs and to handle the redirection by generated short URLs.

The service requires Redis database connection. See example how to run Redis in Docker in [redisDockerRun.sh](https://github.com/slytomcat/URLshortener/blob/master/redisDockerRun.sh)

When `URLshortener` starts it also performs a self-healthcheck. If `URLshortener` misconfigured or initial healthcheck failed then it returns non zero exit code

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

Note: Token is created as random and the saving it to DB may cause duplicate error. In order to avoid such error the service makes several attempts to store random token. The number of attempts is limited by the `Timeout` configuration value by time, not by amount. When time-out expired and no one attempt was not successful then service returns response code `408 Request Timeout`. This response mean that the request can be repeated.

The maximum number of possible attempts to store token during time-out is calculated every time a new token stored. The last measured value is displayed on the homepage.

If measured number of attempts is too small (1-5) then log can contain warnings like: `Warning: Too low number of attempts: 5 per timeout (500 ms)`. In such a case consider increasing of `Timeout` configuration value. Number of attempts above 200 is more then enough, you may consider to decrease `Timeout` configuration value. 30-40 attempts allows to fulfill the space of tokens up to 70-80% before timeout errors (during request for short token) occasionally appears.

When the service time-out errors appears often and log contains many errors like `can't store a new token for 75 attempts` then it most probably means that active (not expired) token amount is near to maximum possible tokens amount (for configured token length). Consider increasing of token length (`TokenLength` configuration value) or decrease token expiration (`DefaultExp` configuration value and/or `exp` parameter in the request for new short URL).

Note also the log warnings such as `Warning: Measured 45 attempts for 423621 ns. Calculated 62 max attempts per 500 ms`. Such warnings also can be a signal that token space is filled near to maximum capacity.


### Request for set new expiration of token:

URL: `<host>[:<port>]/api/v1/expire`

Method: `POST`

Request body: JSON with following parameter:

- `token`: string, token for short URL, mandatory.
- `exp`: int, new expiration in days from now, optional. Default: 0 - value that marks token as expired.

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

Path to configuration file can be provided via optional `-config` command line option (default path is `./cnfr.json`). The file content must be the following correct JSON value:

    {
    "ConnectOptions": {
        "Addrs": [ "<RedisHost>:6379" ],
        "Password": "Long long password that is configured for Redis authorization",
        "DB": 7
    },
    "TokenLength":5,
    "Timeout":777,
    "ListenHostPort":"0.0.0.0:80",
    "DefaultExp":30,
    "ShortDomain":"<shortDomain>",
    "Mode":0
    }

Where:

- `ConnectOptions` - Redis connection options (mandatory):
    - `Addrs` - array of strings: Redis single node address or list of addresses of cluster/sentinel nodes (mandatory)
    - `Password` - string, password for Redis authorization (mandatory for connections to remote Redis node/cluster)
    - `DB` - int, database to be selected after connecting to Redis DB (optional, applicable only for connection to single node and fail-over nodes, default: 0)
    - ... all possible connection options can be found [here](https://godoc.org/github.com/go-redis/redis#UniversalOptions)
- `TokenLength` - int, number of BASE64 symbols in token
- `Timeout` - int, new token creation time-out in milliseconds (optional, default: 500)
- `ListenHostPort` - string: host and port to listen on (optional, default: "localhost:8080")
- `DefaultExp` - int, default token expiration period in days (optional, default: 1)
- `ShortDomain` - string, short domain name for short URL creation (optional, default: "localhost:8080")
- `Mode` - int, service mode (optional, default: 0). Possible values:
    - 0 - service handles all requests
    - 1 - request for redirect is disabled
    - 2 - request for short URL is disabled
    - 4 - request for set new expiration of token is disabled
    - 8 - WEB UI is disabled

Value of `Mode` can be a sum of several modes, for example `"Mode":6` disables two requests: request for short URL and request to set new expiration of token.

Configuration data can be also provided via environment variables URLSHORTENER_ConnectOptions (JSON string with Redis connection options), URLSHORTENER_Timeout, URLSHORTENER_ListenHostPort, URLSHORTENER_DefaultExp, URLSHORTENER_ShortDomain and URLSHORTENER_Mode.

When some configuration value is set in both configuration file and environment variable then value from environment is used.


### Logs

Log is written to output. It contains access log, request results and some warnings about the the measurements of attempts per time-out. 
