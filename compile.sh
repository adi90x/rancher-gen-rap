#!/bin/bash
export GOPATH=$(pwd)/Godeps
go build -tags netgo -a -v  -ldflags "-X main.GitSHA=LocalBuild"
