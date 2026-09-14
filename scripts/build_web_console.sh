#!/bin/bash

# match server version/git metadata (scripts/version.sh)
. ./scripts/version.sh

cd web-console
./scripts/build.sh
cd ../
