#!/bin/bash

echo "grabbing latest release"
downloadURL=`curl -s https://api.github.com/repos/djavorszky/ddn/releases/latest | grep "browser_download_url" | cut -d"\"" -f4`

wget --quiet $downloadURL

echo "extracting"
tar xzf connector*.tar.gz

echo "cleaning up"
rm connector*.tar.gz

if [[ ! -f ddnc.conf ]]; then
	echo "downloading default configuration file"
	wget --quiet https://raw.githubusercontent.com/djavorszky/ddn/master/connector/con.conf

	mv con.conf ddnc.conf

	echo "configuration file downloaded, please configure connector and then start it"
	exit 1
fi



echo "all done. simply run the connector, or add -h for additional information."
