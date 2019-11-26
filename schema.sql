
CREATE DATABASE IF NOT EXISTS shortener_DB
CHARACTER SET 'utf8';

use shortener_DB

SET time_zone = '+00:00';

DROP TABLE IF EXISTS `urls`;
CREATE TABLE `urls` (
  `token` BINARY(5) NOT NULL ,
  `url` VARCHAR(255) NOT NULL,
  `ts` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `exp` INT DEFAULT NULL,
  PRIMARY KEY (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

Create user `shortener`@`%` IDENTIFIED BY RANDOM PASSWORD;

grant all on shortener_DB.* to shortener@'%';
