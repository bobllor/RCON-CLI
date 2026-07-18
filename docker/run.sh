#!/usr/bin/env bash

set -e

if [[ ! -e "minecraft_server.26.2.jar" ]]; then
    curl -o "minecraft_server.26.2.jar" https://piston-data.mojang.com/v1/objects/823e2250d24b3ddac457a60c92a6a941943fcd6a/server.jar

    echo "eula=true" > eula.txt
fi

java -Xmx2G -Xms2G -jar minecraft_server.26.2.jar nogui