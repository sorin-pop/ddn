FROM debian:stretch-slim

RUN apt-get update && apt-get install sqlite3 -y
RUN mkdir -p /ddn/data

ADD ddn-server.tar.gz /ddn

VOLUME /ddn/data

EXPOSE 7010

ENTRYPOINT ["/ddn/server", "-p", "/ddn/data/prod.conf", "-l", "/ddn/data/srv.log"]
