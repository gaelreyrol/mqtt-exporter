# mqtt-exporter

[![CI](https://github.com/gaelreyrol/mqtt-exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/gaelreyrol/mqtt-exporter/actions/workflows/ci.yml)

Export MQTT messages to Promotheus

> **Warning**
> This exporter allows you to simplify your IoT monitoring stack if you don't want to store messages on the long run. While InfluxDB or other timeseries database are well suited for IoT messages, I built this exporter because I don't need those databases, I just need a temporary storage (at least a few months). Also Prometheus is a piece of software that is already deployed in my stack along with Grafana, so I wanted to keep it simple and stupid.

> **Warning**
> Only JSON is supported with no level of depth, every  value must be at the root of the JSON object.


This project is still a work in progress.

# Usage

```bash
go install github.com/gaelreyrol/mqtt-exporter
```

You change:

- the listening address server with `-listen-addr`, defaults to `:8181`.
- the telemetry path with `-telemetry-path`, defaults to `/metrics`.
- the config file path with `-config-path`, defaults to `/etc/mqtt-exporter.toml`.

# Configuration

The configuration file follows the [TOML](https://toml.io/en/) format.

Here is an configuration example:

```toml
broker = "localhost:1883"

[[topics]]
name = "zigbee2mqtt/my_thermostat"
fields = [
    "outside_temperature",
    "inside_temperature",
]
```

## `broker`

The MQTT broker TCP address.

It does no yet support encryption nor authentication.

## `topics`

Topics represent each topic messages you want to export to Prometheus.

The `name` field represents the topic name available in your broker.
The `fields` field represents each key that should be exported to Prometheus from a JSON payload received in the topic.

For example if the following JSON payload is received with two fields defined `"outside_temperature", "inside_temperature"`:

```json
{
    "outside_temperature": 12.6,
    "inside_temperature": 19,
    "garage_temperature": 14
}
```

The `garage_temperature` field will not be exported to Prometheus.

Each value's field extracted from the JSON payload must be `float` compatible.

Strings or child object path values are not supported.
