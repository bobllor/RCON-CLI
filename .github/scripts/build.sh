#!/usr/bin/env bash

# Build binaries and compress into files.

set -e

version_tag=$1

if [[ -z $version_tag ]]; then
    echo "Missing version argument"
    exit 1
fi

files=("build/linux/gorcon" "build/darwin/gorcon" "build/windows/gorcon.exe")

for file in "$files[@]"; do
    rm -f "$file"
done

GOOS=linux GOARCH=amd64 go build -o build/linux/gorcon \
    -ldflags="-X 'github.com/bobllor/rcon-cli/app/root.ProgramVersion=$($version_tag)'"
echo "Created Linux binary (amd64)"

GOOS=darwin GOARCH=arm64 go build -o build/darwin/gorcon \
    -ldflags="-X 'github.com/bobllor/rcon-cli/app/root.ProgramVersion=$($version_tag)'"
echo "Created macOS binary (arm64)"

GOOS=windows GOARCH=amd64 go build -o build/windows/gorcon.exe \
    -ldflags="-X 'github.com/bobllor/rcon-cli/app/root.ProgramVersion=$($version_tag)'"

echo "Created Windows executable (amd64)"

cd build/linux
tar -czf ../gorcon.linux.amd64.tar.gz ./ && echo "Linux tar created"

cd ../darwin
tar -czf ../gorcon.macos.arm64.tar.gz ./ && echo "MacOS tar created"

cd ../windows
zip ../gorcon.windows.amd64.zip * && echo "Windows ZIP created"