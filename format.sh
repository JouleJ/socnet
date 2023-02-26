#! /usr/bin/env bash

for f in `find . -name '*.go'`
do
    go fmt $f
done
