#!/bin/bash

# match server version/git metadata (scripts/version.sh)
. ./scripts/version.sh

corepack enable

cd web-console
./scripts/build.sh
cd ../
