# URLshortener
[![CircleCI](https://circleci.com/gh/slytomcat/URLshortener.svg?style=svg)](https://circleci.com/gh/slytomcat/URLshortener)
[![DeepSource](https://static.deepsource.io/deepsource-badge-light.svg)](https://deepsource.io/gh/slytomcat/URLshortener/?ref=repository-badge)

URLshortener is a microservice to shorten long URLs and to handle the redirection by generated short URLs.

### Request for short URL:
`URL: <host>[:<port>]/token`

`Method: POST`

`Body: JSON with following parameters:`

- url: URL to shorten, mandatory
- exp: short URL expiration in days, optional

`Response: JSON with following parameters:`

- token: token for short URL
- url: short URL

### Redirect to long URL:
`URL: <host>[:<port>]/<token> - URL from response on request for short URL`

`Method: GET`

`No parameters`

`Response contain the redirection to long URL`

### Helth-check:
`URL: <host>[:<port>]/`

`Method: GET`

`No parameters`

`Response: simple home page and HTTP 200 OK in case of good service health or HTTP 500 Server error in case of bad service health`


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

- DSN - MySQL connection string
- MaxOpenConns - DataBase connections pool size
- ListenHostPort - host and port to listen on
- DefaultExp - default token expiration period in days
- ShortDomain - short domain name for short URL creation
- Mode - service mode:

   1 - service handles only request for short URLs and health check request
   
   2 - service handles only request for redirect and health check request
   
   0 - (default) service handles all requests

