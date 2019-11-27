#!/bin/bash
mkdir /opt
chmod 777 /opt
mkdir /opt/mydocker
chmod 777 /opt/mydocker
mkdir -p /opt/mydocker/taotie
mkdir -p /opt/mydocker/taotie/data
cp config.yaml /opt/mydocker/taotie/config.yaml
docker-compose stop
docker-compose rm -f
docker-compose up -d