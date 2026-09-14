#!/bin/bash
set -euo pipefail

# install dependencies
yarn install

# build
CI=false yarn build
