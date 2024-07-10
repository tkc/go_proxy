#!/bin/sh

# Create the output directory
mkdir -p bin

# Build for M1 Mac (arm64)
GOOS=darwin GOARCH=arm64 go build -o go-proxy-server-darwin-arm64 main.go

# Build for Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o go-proxy-server-windows-amd64.exe main.go

echo "Build completed. Binaries are in the bin/ directory."
