#!/bin/bash

docker run --name redis -d -p 6379:6379  redis:5.0.7-alpine redis-server --requirepass "some very long password that is provided through ConnectOptions.Password configuration value"