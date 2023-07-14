package mqttexporter

import (
	"encoding/json"
	"fmt"
	"math"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

var unknownMessagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Info().Str("context", "mqtt").Msgf("received message: %s from unknown topic: %s", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Info().Str("context", "mqtt").Msg("connected to MQTT broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("connection lost: %v", err)
}

func NewClient(config Config) *mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", config.Broker))
	opts.SetClientID("mqtt_exporter")

	opts.SetDefaultPublishHandler(unknownMessagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal().Str("context", "mqtt").Msg(token.Error().Error())
	}

	return &client
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

func handleMessage(topic Topic, messageCounter *prometheus.Counter, gauge *prometheus.GaugeVec, msg *mqtt.Message) {
	(*messageCounter).Inc()
	payload := make(map[string]interface{})
	if err := json.Unmarshal((*msg).Payload(), &payload); err != nil {
		log.Error().Str("context", "mqtt").Msgf("failed to parse message from topic: %s", topic.Name)
		return
	}

	for key, value := range payload {
		if topic.IsFieldFiltered(key) {
			log.Info().Str("context", "mqtt").Msgf("skipping field key %s", key)
			continue
		}
		floatVal, err := fieldValue(value, key)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		gauge.WithLabelValues(key).Set(floatVal)
	}
}

func Subscribe(app *App, topic Topic, registry *prometheus.Registry) {
	messageCounter := NewTopicMessagesTotalCounter(topic)
	registry.MustRegister(messageCounter)

	gauge := NewTopicFieldsGaugeVec(topic)
	app.Registry.MustRegister(gauge)

	token := (*app.Client).Subscribe(topic.Name, 0, func(client mqtt.Client, msg mqtt.Message) {
		messageCounter.Inc()
		handleMessage(topic, &messageCounter, gauge, &msg)
	})
	token.Wait()
	log.Info().Str("context", "mqtt").Msgf("subscribed to topic '%s'", topic.Name)
}
