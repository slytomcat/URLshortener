-- This script contains all necessary commands to create DB structure for URLshortener
-- It should be run from root user in MySQL console

CREATE DATABASE IF NOT EXISTS shortener_DB
CHARACTER SET `utf8`;

USE shortener_DB

-- actually, there is no matter which TimeZone is used by server. 
-- Expiration of token is detected by comparison the timestamp of 
-- stored token + expiration period and the current time (Now()). 
-- Timestamp and now() are in the same TimeZone and such comparison
-- works independently on used timezone. But let it be UTC.  
SET time_zone = '+00:00';

-- If you need longer/shorten token then adjust the tokenLenS constant 
-- in ShortToken.go in accordance to length of `token` field 

DROP TABLE IF EXISTS `urls`;
CREATE TABLE `urls` (
  `token` CHAR(6) NOT NULL ,
  `url` VARCHAR(255) NOT NULL,
  `ts` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `exp` INT DEFAULT NULL,
  PRIMARY KEY (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- keep the generated password to use in in DSN (DB connection string)

DROP USER IF EXISTS `shortener`; 
CREATE USER `shortener`@`%` IDENTIFIED BY RANDOM PASSWORD;

-- NOTE:
-- DELETE right required only for test purpose. Do not grant DELETE right in production environment.

-- GRANT command for test environment:
-- GRANT DELETE, SELECT, INSERT(`token`, `url`, `exp`), UPDATE(`token`, `url`, `exp`) ON `shortener_DB`.`urls` TO `shortener`@`%`;
-- GRANT command for production environment:
GRANT SELECT, INSERT(`token`, `url`, `exp`), UPDATE(`token`, `url`, `exp`) ON `shortener_DB`.`urls` TO `shortener`@`%`;
