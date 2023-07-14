package mqttexporter

import "github.com/prometheus/client_golang/prometheus"

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
			Name: "mqtt_topic_field",
			ConstLabels: prometheus.Labels{
				"topic": topic.Name,
			},
		},
		[]string{"name"},
	)
}
