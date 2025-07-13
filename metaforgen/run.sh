#!/bin/bash
rm -rf output
go run main.go
cp -r workerpool ./output