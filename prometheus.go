package mqttexporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

func getConstLabels(topic Topic) prometheus.Labels {
	return prometheus.Labels{
		"topic": topic.Name,
	}
}

func NewTopicMessagesTotalCounter(topic Topic) prometheus.Counter {
	return prometheus.NewCounter(prometheus.CounterOpts{
		Name:        "mqtt_topic_messages_total",
		ConstLabels: getConstLabels(topic),
	})
}

func NewTopicFieldsGaugeVec(topic Topic) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "mqtt_topic_field",
			ConstLabels: getConstLabels(topic),
		},
		[]string{"name"},
	)
}

func InitializeTopicContext(topic Topic, registry *prometheus.Registry) *TopicContext {
	counter := NewTopicMessagesTotalCounter(topic)
	registry.MustRegister(counter)
	log.Debug().Str("context", "prometheus").Msg("registered counter in global registry")

	gauge := NewTopicFieldsGaugeVec(topic)
	registry.MustRegister(gauge)
	log.Debug().Str("context", "prometheus").Msg("registered gauge in global registry")

	return &TopicContext{&topic, registry, &counter, gauge}
}

func ExportTopicMessages(app *App, context *TopicContext) {

	for {
		msg := <-app.topics[context.topic.Name]
		log.Debug().Str("context", "prometheus").Msgf("received payload from topic channels '%s'", context.topic.Name)
		(*context.counter).Inc()
		for key, value := range *msg {
			context.gauge.WithLabelValues(key).Set(value)
		}
	}
}
