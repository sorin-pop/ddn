FROM debian:stretch-slim

RUN apt-get update && apt-get install sqlite3 -y
RUN mkdir -p /ddn/data /ddn/ftp /ddn/web /ddn/web/dumps

COPY server /ddn
COPY web /ddn/web

VOLUME ["/ddn/data", "/ddn/ftp"]

EXPOSE 7010

ENTRYPOINT ["/ddn/server", "-p", "/ddn/data/prod.conf", "-l", "/ddn/data/srv.log"]
