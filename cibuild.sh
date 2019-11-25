#!/bin/bash -e

go test

if [ -d "outputs/" ]; then rm -rf outputs/; fi

mkdir -p outputs
GOOS=linux GOARCH=amd64 go build -o outputs/smartbackup.linux64
GOOS=linux GOARCH=386 go build -o outputs/smartbackup.linux32
GOOS=darwin go build -o outputs/smartbackup.macos
GOOS=windows GOARCH=amd64 go build -o outputs/smartbackup64.exe
GOOS=windows GOARCH=386 go build -o outputs/smartbackup32.exe
