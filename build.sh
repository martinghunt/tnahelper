#!/usr/bin/env bash
set -exu

version=$1

for os in windows linux darwin
do
    for arch in amd64 arm64
    do
        GOOS=$os GOARCH=$arch go build -ldflags="-X main.Version=${version}" -o tnahelper_${os}_${arch} main.go
    done
done
