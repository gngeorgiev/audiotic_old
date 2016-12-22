#!/bin/bash

rm -rf dist
mkdir -p dist

cd www
npm run build
cd ..
cp -R www/build dist/www
rm -rf www/build

cd server
rm -f audiotic
go build -o audiotic main.go
cd ..
cp server/audiotic dist/audiotic
rm -f server/audiotic