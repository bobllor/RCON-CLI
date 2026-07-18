#!/usr/bin/env bash

set -e

os=$(uname -s)

file_name=""

echo -e "Starting gorcon installation...\n"

case "$os" in
    Linux)
        file_name="gorcon.linux.amd64.tar.gz"
        ;;
    Darwin):
        file_name="gorcon.darwin.amd64.tar.gz"
        ;;
    *)
        echo "OS is unsupported ($os)"
        exit 1
        ;;
esac

base_url="https://github.com/bobllor/rcon-cli/releases/latest"
url="$base_url/download/$file_name"

if ! curl -fsI "$url" > /dev/null; then
    echo "Failed to fetch download link: $url"
    exit 1
fi

# folder setup
temp_folder="/tmp/gorcontemp"
mkdir -p "$temp_folder"
mkdir -p "$HOME/.local/bin"

cd "$temp_folder"

echo "Downloading $file_name..."
# yes this is formatting
echo ""

curl -L "$url" -o "$temp_folder/$file_name"

echo ""

echo "Extracting files..."
tar -xzf "$file_name" -C "$HOME/.local/bin"

bash_file="$HOME/.bashrc"

if [[ -z $(cat "$bash_file" | grep 'export PATH="$PATH:$HOME/.local/bin"') ]]; then
    echo "PATH not found"
    echo "Configuring PATH..."

    echo "" >> "$bash_file"
    echo 'export PATH="$PATH:$HOME/.local/bin"' >> "$bash_file"

    echo "Configured PATH ($HOME/.local/bin)"
fi

echo "Cleaning up..."
rm -rf "$temp_folder"

source "$bash_file"

echo -e "\nSuccessfully installed!"
echo "Run \"gorcon\" to get started"