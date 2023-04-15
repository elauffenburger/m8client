#!/usr/bin/env bash

set -o errexit -o nounset

go build -o build/m8client
./build/m8client