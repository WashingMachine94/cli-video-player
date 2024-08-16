#!/bin/bash

# Linux ARM64 build
export GOOS=linux
export GOARCH=arm64
go build -o "build/linux_arm64/play"

# Linux AMD64 build
export GOOS=linux
export GOARCH=amd64
go build -o "build/linux_amd64/play"

# Windows build (cross-compiling from Linux)
export GOOS=windows
export GOARCH=amd64
go build -o "build/windows/play.exe"

# Reset environment variables
unset GOOS
unset GOARCH

echo "Build completed successfully!"

