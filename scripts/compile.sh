#!/bin/bash

mkdir -p /go/src/core
rm -r /go/src/core/*
yes | cp /src/* /go/src/core
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /workspace/main /workspace/main.go
