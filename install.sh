#!/usr/bin/env bash

set -e

os=$(uname -s)

base_url="https://github.com/bobllor/rcon/releases/latest"
file_name=""

case "$os" in
    Linux)
        file_name="gorcon.linux.amd64.tar"
        ;;
    Darwin):
        file_name="gorcon.darwin.amd64.tar"
        ;;
    *)
        echo "OS is unsupported ($os)"
        exit 1
        ;;
esac

url="$base_url/download/$file_name"

# folder setup
temp_folder="/tmp/gorcontemp"
mkdir -p "$temp_folder"
mkdir -p "$HOME/.local/bin"

cd "$temp_folder"

curl -L "$url" -o "$temp_folder/$file_name"
tar -xf "$file_name" -C "$HOME/.local/bin"

rm -rf "$temp_folder"

cd -

bash_file="$HOME/.bashrc"

if [[ -z $(cat "$bash_file" | grep 'export PATH="$PATH:$HOME/.local/bin"') ]]; then
    echo "" >> "$bash_file"
    echo 'export PATH="$PATH:$HOME/.local/bin"' >> "$bash_file"

    echo "Added PATH $HOME/.local/bin"
fi

source "$bash_file"