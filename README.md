# URLshortener
[![CircleCI](https://circleci.com/gh/slytomcat/URLshortener.svg?style=svg)](https://circleci.com/gh/slytomcat/URLshortener)
[![DeepSource](https://img.shields.io/badge/Deepsource-Passed-brightgreen)](https://deepsource.io/gh/slytomcat/URLshortener)
[![Platform: linux-64](https://img.shields.io/badge/Platform-linux--64-blue)]()
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Docker image](https://img.shields.io/badge/Docker-image-blue)](https://hub.docker.com/r/slytomcat/urlshortener)

URLshortener is a micro-service to shorten long URLs and to handle the redirection by generated short URLs.

The service requires and MySQL server connection and database structure described in [schema.sql](https://github.com/slytomcat/URLshortener/blob/master/schema.sql)


### Request for short URL:

URL: `<host>[:<port>]/api/v1/token`

Method: `POST`

Request body: JSON with following parameters:

- url: URL to shorten, mandatory
- exp: short URL expiration in days, optional, default: value of "DefaultExp" from configuration file

Success response: HTTP 200 OK with body containing JSON with following parameters:

- token: token for short URL
- url: short URL

### Request for set new expiration of token:

URL: `<host>[:<port>]/api/v1/expire`

Method: `POST`

Request body: JSON with following parameter:

- token: token for short URL, mandatory.
- exp: new expiration in days from now, optional. Default 0, value that marks token as expired. 

Success response: HTTP 200 OK with empty body

### Redirect to long URL:
URL: `<host>[:<port>]/<token>` - URL from response on request for short URL

Method: `GET`

Response contain the redirection to long URL (response code: HTTP 301 'Moved permanently' with 'Location' = long URL in response header)

### Health-check:
URL: `<host>[:<port>]/`

Method: `GET`

Response: simple home page and HTTP 200 OK in case of good service health or HTTP 500 Server error in case of bad service health


### Configuration file

Configuration file must have a name `cnf.json` and it should be placed in the same folder where URLshortener was run. The file content must be the following correct JSON value: 

    {

    "DSN":"shortener:<password>@<protocol>(<host>:<port>)/shortener_DB",

    "MaxOpenConns":"33",

    "ListenHostPort":"0.0.0.0:80",

    "DefaultExp":"30",

    "ShortDomain":"<shortDomain>",

    "Mode":"0"

    }

Where:

- `DSN` - MySQL connection string (mandatory, also can set via URLSHORTENER_DSN environment variable)
- `MaxOpenConns` - DataBase connections pool size (optional, default 10)
- `ListenHostPort` - host and port to listen on (optional, default localhost:8080)
- `DefaultExp` - default token expiration period in days (optional, default 1)
- `ShortDomain` - short domain name for short URL creation (optional, default localhost:8080)
- `Mode` - service mode (optional, default 0). Possible values:

   0 - service handles all requests

   1 - request for redirect is disabled

   2 - request for short URL is disabled

   4 - request for set new expiration of token is disabled

Value of `Mode` can be a sum of several modes, for example `"Mode":"6"` disables two requests: request for short URL and request to set new expiration of token.
