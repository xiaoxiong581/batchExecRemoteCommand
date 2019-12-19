#!/bin/bash

GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ../batchExecRemoteCommand.exe ../server/main.go
#GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../batchExecRemoteCommand ../server/main.go