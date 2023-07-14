package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttexporter "github.com/gaelreyrol/mqtt-exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var listenAddress = flag.String("listen-address", ":8181", "The address to listen on for HTTP requests.")
var metricsPath = flag.String("telemetry-path", "/metrics", "The path under which to expose metrics.")
var configPath = flag.String("config-path", "/etc/mqtt-exporter.toml", "The configuration file path to use.")

var unknownMessagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("received message: %s from unknown topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("connected to MQTT broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("connection lost: %v", err)
}

func registerFields(topic *mqttexporter.Topic, registry *prometheus.Registry) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mqtt_topic_field",
			ConstLabels: prometheus.Labels{
				"topic": topic.Name,
			},
		},
		[]string{"name"},
	)
	registry.MustRegister(gauge)
	return gauge
}

func fieldValue(value interface{}, key string) (float64, error) {
	switch i := value.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	default:
		return math.NaN(), fmt.Errorf("field value %s is of incompatible type", key)
	}
}

func sub(topic *mqttexporter.Topic, client mqtt.Client, registry *prometheus.Registry) {
	messageCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mqtt_topic_messages_total",
		ConstLabels: prometheus.Labels{
			"topic": topic.Name,
		},
	})
	registry.MustRegister(messageCounter)

	fields := registerFields(topic, registry)

	token := client.Subscribe(topic.Name, 0, func(client mqtt.Client, msg mqtt.Message) {
		messageCounter.Inc()
		payload := make(map[string]interface{})
		if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
			fmt.Printf("failed to parse message from topic: %s\n", msg.Topic())
			return
		}

		fmt.Printf("received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

		for key, value := range payload {
			if topic.IsFieldFiltered(key) {
				fmt.Printf("skipping field key %s\n", key)
				continue
			}
			floatVal, err := fieldValue(value, key)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			fields.WithLabelValues(key).Set(floatVal)
		}
	})
	token.Wait()
	fmt.Printf("subscribed to topic '%s'\n", topic.Name)
}

func main() {
	flag.Parse()

	config := mqttexporter.NewConfig()
	if err := config.FromFile(*configPath); err != nil {
		log.Fatal(err)
	}

	globalReg := prometheus.NewRegistry()

	server := &http.Server{Addr: *listenAddress}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", config.Broker))
	opts.SetClientID("mqtt-exporter")

	opts.SetDefaultPublishHandler(unknownMessagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for _, topic := range config.Topics {
		topicReg := prometheus.NewRegistry()
		globalReg.Register(topicReg)

		sub(topic, client, topicReg)
	}

	http.Handle(*metricsPath, promhttp.HandlerFor(
		globalReg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
			Registry:          globalReg,
		},
	))

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}
}
