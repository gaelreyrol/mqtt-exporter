package mqttexporter

import (
	"context"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	Config   *Config
	Registry *prometheus.Registry
	Server   *http.Server
	Client   *mqtt.Client
	topics   map[string]chan *map[string]float64
}

type TopicContext struct {
	topic    *Topic
	registry *prometheus.Registry
	counter  *prometheus.Counter
	gauge    *prometheus.GaugeVec
}

func NewApp(configPath string, listenAddress string, metricsPath string) (*App, error) {
	app := App{
		Config:   NewConfig(),
		Registry: prometheus.NewRegistry(),
		Server:   &http.Server{Addr: listenAddress},
	}

	if err := app.Config.FromFile(configPath); err != nil {
		return nil, err
	}

	app.Client = NewClient(*app.Config)
	app.topics = make(map[string]chan *map[string]float64, len(app.Config.Topics))

	http.Handle(metricsPath, promhttp.HandlerFor(
		app.Registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
			Registry:          app.Registry,
		},
	))

	return &app, nil
}

func (app *App) Start() error {
	for _, topic := range app.Config.Topics {
		app.topics[topic.Name] = make(chan *map[string]float64)

		topicReg := prometheus.NewRegistry()
		app.Registry.Register(topicReg)

		go SubscribeTopic(app, &topic)

		context := InitializeTopicContext(topic, topicReg)
		go ExportTopicMessages(app, context)
	}

	return app.Server.ListenAndServe()
}

func (app *App) Shutdown(ctx context.Context) error {
	(*app.Client).Disconnect(0)
	return app.Server.Shutdown(ctx)
}
