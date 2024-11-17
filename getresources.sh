#! /bin/bash

set -eux

for dir in mcgalaxyrunner httpserver/static; do
    if [ -d $dir ]; then
        mkdir -p $dir
    fi
done

# Server binaries
curl -L https://cdn.classicube.net/client/mcg/release/MCGalaxy.zip -o mcgalaxyrunner/MCGalaxy.zip

# Web client resources
curl -L https://classicube.net/static/default.zip -o httpserver/static/default.zip
curl -L https://cs.classicube.net/client/latest/ClassiCube.js -o httpserver/static/ClassiCube.js
