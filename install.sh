#!/usr/bin/env bash

set -e

os=$(uname -s)

release_url=""
case "$os" in
    Linux)
        release_url="Linux"
        ;;
    Darwin):
        release_url="Darwin"
        ;;
    *)
        echo "OS is unsupported ($os)"
        exit 1
        ;;
esac

if [[ ! -d "$HOME/.local/bin" ]]; then
    mkdir -p "$HOME/.local/bin"
fi

bash_file="$HOME/.bashrc"

if [[ -z $(cat "$bash_file" | grep '$PATH:$HOME/.local/bin') ]]; then
    echo "" >> "$bash_file"
    echo 'export PATH="$PATH:$HOME/.local/bin"' >> "$bash_file"

    echo "Added PATH $HOME/.local/bin"
fi

source "$bash_file"