# URLshortener
[![CircleCI](https://circleci.com/gh/slytomcat/URLshortener.svg?style=svg)](https://circleci.com/gh/slytomcat/URLshortener)
[![DeepSource](https://img.shields.io/badge/Deepsource-Passed-brightgreen)](https://deepsource.io/gh/slytomcat/URLshortener)
[![Platform: linux-64](https://img.shields.io/badge/Platform-linux--64-blue)]()
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Docker image](https://img.shields.io/badge/Docker-image-blue)](https://hub.docker.com/r/slytomcat/urlshortener)

URLshortener is a micro-service to shorten long URLs and to handle the redirection by generated short URLs.

The service requires Redis database connection. See example how to run Redis in Docker in [redisDockerRun.sh](https://github.com/slytomcat/URLshortener/blob/master/redisDockerRun.sh)


### Request for short URL:

URL: `<host>[:<port>]/api/v1/token`

Method: `POST`

Request body: JSON with following parameters:

- `url`: URL to shorten, mandatory
- `exp`: short URL expiration in days, optional, default: value of "DefaultExp" from configuration file

Success response: HTTP 200 OK with body containing JSON with following parameters:

- `token`: token for short URL
- `url`: short URL

Note: Token is created as random and the saving it to DB may cause duplicate error. In order to avoid such error service makes several attempts to store random token. The number of attempts is limited by the `Timeout` configuration value by time, not by amount. When time-out expired and no one attempt was not successful then service returns response code 504 Gateway Timeout. This response mean that the request can be repeated.  

While performing health-check the maximum number of possible attempts to store token during time-out is measured and displayed on the home page. The measurement is also written in log.

If measured number of attempts is too small (1-5) then consider increasing of `Timeout` configuration value. Number of attempts above 200 is more then enough, you may consider to decrease `Timeout` configuration value. 30-40 attempts allows to fulfill the space of tokens up to 70-80% before timeout errors (during request for short token) occasionally appears.

When the service time-out errors appears often and log contains many errors like `can't store a new token for 75 attempts` then it most probably means that active (not expired) token amount is near to maximum possible tokens amount (for configured token length). Consider increasing of token length (`TokenLength` configuration value) or decrease token expiration (`DefaultExp` configuration value and/or `exp` parameter in the request for new short URL).



### Request for set new expiration of token:

URL: `<host>[:<port>]/api/v1/expire`

Method: `POST`

Request body: JSON with following parameter:

- `token`: token for short URL, mandatory.
- `exp`: new expiration in days from now, optional. Default 0, value that marks token as expired.

Success response: HTTP 200 OK with empty body

### Redirect to long URL:
URL: `<host>[:<port>]/<token>` - URL from response on request for short URL

Method: `GET`

Response contain the redirection to long URL (response code: HTTP 301 'Moved permanently' with 'Location' = long URL in response header)

### Health-check:
URL: `<host>[:<port>]/`

Method: `GET`

Response: simple home page and HTTP 200 OK in case of good service health or HTTP 500 Server error in case of bad service health


### Sevice configuration

Configuration file must have a name `cnfr.json` and it should be placed in the same folder where URLshortener was run. The file content must be the following correct JSON value:

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
    - `Addrs` - Redis single node address or addresses of cluster/sentinel nodes (mandatory 1 address for single node or several addresses for cluster/sentinel nodes)
    - `Password` - Password for Redis authorization (mandatory for remote redis connections)
    - `DB` - database to be selected after connecting to Redis DB (optional, only for single mode and fail-over connection, default 0)
    - ... all possible connection options can be fount [here](https://godoc.org/github.com/go-redis/redis#UniversalOptions)
- `TokenLength` - number of BASE64 symbols in token
- `Timeout` - new token creation time-out in milliseconds (optional, default 500)
- `ListenHostPort` - host and port to listen on (optional, default localhost:8080)
- `DefaultExp` - default token expiration period in days (optional, default 1)
- `ShortDomain` - short domain name for short URL creation (optional, default localhost:8080)
- `Mode` - service mode (optional, default 0). Possible values:
    - 0 - service handles all requests
    - 1 - request for redirect is disabled
    - 2 - request for short URL is disabled
    - 4 - request for set new expiration of token is disabled

Value of `Mode` can be a sum of several modes, for example `"Mode":"6"` disables two requests: request for short URL and request to set new expiration of token.

Configuration data can be also provided via environment variables URLSHORTENER_ConnectOptions (JSON string with Redis connection options), URLSHORTENER_Timeout, URLSHORTENER_ListenHostPort, URLSHORTENER_DefaultExp, URLSHORTENER_ShortDomain and URLSHORTENER_Mode.

Configuration file values have more priority then environment variables.
