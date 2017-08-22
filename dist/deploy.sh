#!/bin/bash

echo "building binary.."
cd ../server
go build

echo "archiving.."
tar czf ddn-server.tar.gz server web
cd ../dist

mv ../server/ddn-server.tar.gz .

echo "building image"
docker build -t ddn/server .

echo "starting container.."
docker run -dit -p 7010:7010 -v `pwd`/data:/ddn/data ddn/server:latest

echo "removing artefacts.."
rm -rf ddn-server.tar.gz
