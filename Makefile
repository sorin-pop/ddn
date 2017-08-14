all:
	echo "use clean or build"

clean:
	rm -rf out/*

build:
	mkdir -p out

	go build -o ddn-server github.com/djavorszky/ddn/server
	cp -r server/web .
	tar czf ddn-server.tar.gz ddn-server web
	mv ddn-server.tar.gz out
	rm -r ddn-server web

	go build -o ddn-connector github.com/djavorszky/ddn/connector
	tar czf ddn-connector.tar.gz ddn-connector
	mv ddn-connector.tar.gz out/
	rm ddn-connector
