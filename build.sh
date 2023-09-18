#!/bin/bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -trimpath -ldflags="-s -w -X main.version=$(git branch --show-current)-$(git rev-parse --short HEAD)" ./...
upx -qq --best URLshortener
