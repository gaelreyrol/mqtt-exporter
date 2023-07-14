# mqtt-exporter

[![CI](https://github.com/gaelreyrol/mqtt-exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/gaelreyrol/mqtt-exporter/actions/workflows/ci.yml)

Export MQTT messages to Promotheus

> **Warning**
> This exporter allows you to simplify your IoT stack if you don't want to store messages on the long run. While InfluxDB or other timeseries database are well suited for IoT messages, I built this exporter because I don't need them, just a temporary storage (at least a few months). Also Prometheus is a piece of software that is already deployed in my stack along with Grafana, so I wanted to keep it simple and stupid.


This project is still a work in progress.
