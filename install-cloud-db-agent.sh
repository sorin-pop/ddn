#!/bin/bash

echo "grabbing latest release"
downloadURL=`curl -s https://api.github.com/repos/djavorszky/ddn/releases/latest | grep "browser_download_url" | cut -d"\"" -f4`

wget --quiet $downloadURL

echo "extracting"
tar xzf agent*.tar.gz

echo "cleaning up"
rm agent*.tar.gz

if [[ ! -f ddnc.conf ]]; then
	echo "downloading default configuration file"
	wget --quiet https://raw.githubusercontent.com/djavorszky/ddn/master/agent/con.conf

	mv default.conf ddnc.conf

	echo "configuration file downloaded, please configure agent and then start it"
	exit 1
fi

echo "all done. simply run the agent, or add -h for additional information."
