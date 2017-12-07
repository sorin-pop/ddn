#!/bin/bash

rootloc="`pwd`/.."

echo "building binary of server.."
cd $rootloc/server
go build -ldflags "-X main.version=`date -u +%Y%m%d.%H%M%S`"

echo "updating libraries"
cd $rootloc/server/web
npm install -u
cd ..

echo "copying server.."
cp $rootloc/server/server $rootloc/dist/server
cp -r $rootloc/server/web $rootloc/dist/web

cd $rootloc/dist

echo "building server image"
docker build -t djavorszky/ddn .

echo "stopping previous version"
docker stop ddn-server
docker rm ddn-server

echo "starting container.."
docker run -dit -p 7010:7010 --name ddn-server -v /home/javdaniel/go/src/github.com/djavorszky/ddn/dist/data:/ddn/data -v /home/javdaniel/go/src/github.com/djavorszky/ddn/dist/ftp:/ddn/ftp djavorszky/ddn:latest

echo "removing artefacts.."
rm -rf server web 
