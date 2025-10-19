#!/bin/zsh

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./Server/main  main.go || echo "编译linux版本失败"

scp ./Server/main root@wzs.app:/root/main