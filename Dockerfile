FROM debian:bookworm

RUN apt-get update && apt-get install -y debhelper
# Dev dependencies
RUN apt-get install -y python3 systemd vim

WORKDIR /src
COPY . .

RUN dpkg-buildpackage -us -uc
RUN dpkg -c ../nmealogger_1.0-1_all.deb
