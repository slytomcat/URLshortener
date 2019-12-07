#! /bin/bash

# run mysql container
docker run --name=mysql1 -e MYSQL_ROOT_PASSWORD=root -p 127.0.0.1:3306:3306 -d mysql/mysql-server:latest

# wait initialisation  
while [[ $(docker ps | grep "(healthy)") == "" ]]
do 
  sleep 1
done

# create db structure
docker exec -i mysql1 mysql -uroot -proot < schema.sql

# change shortener password and add delete right required for tests
docker exec  mysql1 mysql -uroot -proot -e "ALTER USER 'shortener' IDENTIFIED BY 'test'"
docker exec  mysql1 mysql -uroot -proot -e 'GRANT DELETE ON `shortener_DB`.`urls` TO `shortener`@`%`'

# define DSN 
# export URLSHORTENER_DSN="shortener:test@tcp(localhost:3306)/shortener_DB"