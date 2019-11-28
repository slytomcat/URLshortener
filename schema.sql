-- This script contains all necessary commands to create DB structure for URLshortener
-- It should be run from root user in MySQL console

CREATE DATABASE IF NOT EXISTS shortener_DB
CHARACTER SET `utf8`;

USE shortener_DB

SET time_zone = '+00:00';

DROP TABLE IF EXISTS `urls`;
CREATE TABLE `urls` (
  `token` CHAR(6) NOT NULL ,
  `url` VARCHAR(255) NOT NULL,
  `ts` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `exp` INT DEFAULT NULL,
  PRIMARY KEY (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


-- keep the generated password to use in in DSN (DB connection string)
CREATE USER `shortener`@`%` IDENTIFIED BY RANDOM PASSWORD;

GRANT SELECT(`token`, `url`, `exp`), INSERT(`token`, `url`, `exp`), UPDATE(`token`, `url`, `exp`) ON `shortener_DB`.`urls` TO `shortener`@`%`;
