#!/bin/bash
rm -rf output
go run main.go
cp -r workerpool ./output
cd output/wiring
go mod download
go get
go run main.go -o build -w docker
cd build/docker
cp ../.local.env .env
sudo docker compose build
sudo docker compose up -d
cd ../../../../
