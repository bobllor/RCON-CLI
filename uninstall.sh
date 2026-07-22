#!/usr/bin/env bash

set -e

file="gorcon"
path="$HOME/.local/bin"

echo "Starting gorcon uninstallation..."

if [[ -e "$path/$file" ]]; then
    echo "Removing $file from $path..."
    rm "$path/$file"

    echo "Successfully uninstalled gorcon"
else
    echo "gorcon installation not found"
fi

# path is not removed from zshrc/bashrc.
# dont think we need to since $HOME/.local/bin is an XDG standard directory
# it may also conflict with other PATH exports