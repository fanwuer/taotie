#!/bin/bash
mkdir /opt
chmod 777 /opt
mkdir /opt/mydocker
chmod 777 /opt/mydocker
mkdir -p /opt/mydocker/redis/data
mkdir -p /opt/mydocker/redis/conf
mkdir -p /opt/mydocker/mysql/data
mkdir -p /opt/mydocker/mysql/conf
cp my.cnf /opt/mydocker/mysql/conf/my.cnf
chmod 644 /opt/mydocker/mysql/conf/my.cnf
cp redis.conf /opt/mydocker/redis/conf/redis.conf
docker-compose stop
docker-compose rm -f
docker-compose up -d