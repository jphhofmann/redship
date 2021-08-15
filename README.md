# RedShip - Redis to Elasticsearch Shipper

RedShip ships json encoded data from redis to elasticsearch and allows transforming fields.

## Motivation

At combahton, flowShield DDoS-Protection outputs flow and log data as json (for performance reasons) to redis using hiredis C library in seperate threads. As we heavily use elasticsearch as data storage, later represented as json from REST APIs, we were in need of a redis->elasticsearch bridge. First concept was running for almost six months and were built in python3.

Due to the design of python, this costs a lot of (wasted) cpu ressources, python also doesnt provide real threading, instead it spawns childs (forks) running in their seperated memory region. Switching from redship built in python to go, we noticed eight times less cpu usage while handling identical workload (around ~10k documents per second).

## Features

- Handle each redis key (e.g. wildcard prefix) as individual go routine and output data to a individual elasticsearch index
- Transformation of json fields -> E.g. add GeoIP data, convert ip-addresses to decimal or from decimal, convert unix timestamp to RFC3339 timestamp
- Exposes comprehensive prometheus metrics about uptime, errors and shipping rate using it's inbuilt exporter

## Configuration

A sample configuration is available in "sample.yaml".