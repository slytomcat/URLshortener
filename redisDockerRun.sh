#!/bin/bash

docker run --name redis -d -p 6379:6379  redis:5.0.7-alpine redis-server --requirepass some_wery_long_password_that_stored_in_ConnectOptions_configuration_value