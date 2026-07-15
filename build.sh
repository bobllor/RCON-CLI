#!/usr/bin/env bash

# Build binaries and compress into files.

set -e

files=("build/linux/gorcon" "build/darwin/gorcon" "build/windows/gorcon.exe")

for file in "$files[@]"; do
    rm -f "$file"
done

GOOS=linux GOARCH=amd64 go build -o build/linux/gorcon
echo "Created Linux binary (amd64)"

GOOS=darwin GOARCH=arm64 go build -o build/darwin/gorcon
echo "Created MacOS binary (arm64)"

GOOS=windows GOARCH=amd64 go build -o build/windows/gorcon.exe
echo "Created Windows executable (amd64)"

cd build/linux
tar -cf ../gorcon.linux.amd64.tar ./ && echo "Linux tar created"

cd ../darwin
tar -cf ../gorcon.macos.arm64.tar ./ && echo "MacOS tar created"

cd ../windows
zip ../gorcon.windows.amd64.zip * && echo "Windows ZIP created"