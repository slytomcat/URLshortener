#!/bin/bash

#### IMPORTANT NOTE ####
## Suggested method must be used only for test purposes!
## Never pass the password through the command line parameter in production.
## The full command line (with password) can be obtained via `ps` comand by ANY OS user!
########################

docker run --name redis -d -p 6379:6379  redis:5.0.7-alpine redis-server --requirepass "some very long password that is provided through ConnectOptions.Password configuration value"