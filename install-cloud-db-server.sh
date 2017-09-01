#!/bin/bash

dataDir=""
mountDir=""
containerName="ddn-server"
publicPort="7010"

if [[ $dataDir == "" ]] || [[ $mountDir == "" ]]; then
	echo "please set 'dataDir' and 'mountDir' in the script to point to valid locations"
	exit 1
fi

echo "pulling djavorszky/ddn:latest"
docker pull djavorszky/ddn:latest

echo "creating folders"
mkdir -p $dataDir $mountDir

if [[ ! -f $dataDir/srv.conf ]]; then
	echo "downloading configuration file"
	wget --quiet https://raw.githubusercontent.com/djavorszky/ddn/master/server/srv.conf
	mv srv.conf $dataDir/srv.conf

	echo "Please update $dataDir/srv.conf to complete the configuration"
	echo "and rerun this script once done."

	exit 0
fi

echo "trying to stop and remove previous container, if exists"
docker stop $containerName
docker rm $containerName

echo "starting container"
docker run -dit -p $publicPort:7010 --name $containerName -v $dataDir:/ddn/data -v $mountDir:/ddn/ftp djavorszky/ddn:latest

echo "done"
