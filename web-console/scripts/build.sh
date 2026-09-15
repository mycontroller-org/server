#!/bin/bash
set -euo pipefail

# install dependencies
yarn install

# build
yarn build
