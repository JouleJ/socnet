#! /usr/bin/env bash

VOLUME_PATH="$PWD/volume"
SALT=$1

echo "Volume path: $VOLUME_PATH"

docker build -t socnet . && \
    docker run -v "$VOLUME_PATH:/app/volume" --env "SALT=$SALT" -p 80:80 socnet
