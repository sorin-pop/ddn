#!/bin/bash

rootloc=`pwd`

echo "building binary of server.."
docker build -f Dockerfile.build -t djavorszky/ddn:build .

docker container create --name extract djavorszky/ddn:build  
docker container cp extract:/go/src/github.com/djavorszky/ddn/server/server ./dist/server
docker container rm -f extract

echo "updating libraries"
cd $rootloc/server/web
npm install -u

echo "copying server.."
cp -r $rootloc/server/web $rootloc/dist/web

cd $rootloc/dist

echo "building server image"
docker build -t djavorszky/ddn .

echo "stopping previous version"
docker stop ddn-server
docker rm ddn-server

echo "starting container.."
docker run -dit -p 7010:7010 --name ddn-server -v $rootloc/dist/data:/ddn/data -v $rootloc/dist/ftp:/ddn/ftp djavorszky/ddn:latest

echo "removing artefacts.."
rm -rf $rootloc/dist/server $rootloc/dist/web 
