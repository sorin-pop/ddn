#!/bin/bash

echo "building binary.."
cd ../server
go build

echo "copying.."
cp server ../dist/server
cp -r web ../dist/web

cd ../dist

echo "building image"
docker build -t djavorszky/ddn .

echo "stopping previous version"
docker stop ddn-server
docker rm ddn-server

echo "starting container.."
docker run -dit -p 7010:7010 --name ddn-server -v /home/javdaniel/go/src/github.com/djavorszky/ddn/dist/data:/ddn/data -v /home/javdaniel/go/src/github.com/djavorszky/ddn/dist/ftp:/ddn/ftp djavorszky/ddn:latest

echo "removing artefacts.."
rm -rf server web