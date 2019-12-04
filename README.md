# URLshortener
[![CircleCI](https://circleci.com/gh/slytomcat/URLshortener.svg?style=svg)](https://circleci.com/gh/slytomcat/URLshortener)
[![DeepSource](https://static.deepsource.io/deepsource-badge-light.svg)](https://deepsource.io/gh/slytomcat/URLshortener/?ref=repository-badge)

URLshortener is a microservice to shorten long URLs and to handle the redirection by generated short URLs.

The service requires and MsSQL server conection and database structure described in [schema.sql](https://github.com/slytomcat/URLshortener/blob/master/schema.sql)

Docker image: https://hub.docker.com/r/slytomcat/urlshortener

### Request for short URL:

URL: `<host>[:<port>]/token`

Method: `POST`

Body: JSON with following parameters:

- url: URL to shorten, mandatory
- exp: short URL expiration in days, optional

Response: JSON with following parameters:

- token: token for short URL
- url: short URL

### Redirect to long URL:
URL: `<host>[:<port>]/<token>` - URL from response on request for short URL

Method: `GET`

Response contain the redirection to long URL (response code: HTTP 301 'Moved permanently' with 'Location' = long URL in header)

### Helth-check:
URL: `<host>[:<port>]/`

Method: `GET`

Response: simple home page and HTTP 200 OK in case of good service health or HTTP 500 Server error in case of bad service health


### Configuration file

    {
    
    "DSN":"shortener:<password>@<protocol>(<host>:<port>)/shortener_DB",
    
    "MaxOpenConns":"33",
    
    "ListenHostPort":"0.0.0.0:80",
    
    "DefaultExp":"30",
    
    "ShortDomain":"<shortDomain>",
    
    "Mode":"0"
    
    }

Where:

- DSN - MySQL connection string (mandatory, also can set via URLSHORTENER_DSN environment variable)
- MaxOpenConns - DataBase connections pool size (optional, default 10)
- ListenHostPort - host and port to listen on (optional, default localhost:8080)
- DefaultExp - default token expiration period in days (optional, default 1)
- ShortDomain - short domain name for short URL creation (optional, default localhost:8080)
- Mode - service mode (optional, default 0):

   0 - service handles all requests

   1 - service handles only request for short URLs and health check request

   2 - service handles only request for redirect and health check request
   

