#!/bin/bash

cd console-web
yarn install
CI=false yarn build
cd ../
