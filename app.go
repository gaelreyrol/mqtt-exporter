package mqttexporter

import (
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type App struct {
	Config   *Config
	Registry *prometheus.Registry
	Server   *http.Server
	Client   *mqtt.Client
}

func NewApp(configPath string, listenAddress string) App {
	app := App{
		Config:   NewConfig(),
		Registry: prometheus.NewRegistry(),
		Server:   &http.Server{Addr: listenAddress},
	}

	if err := app.Config.FromFile(configPath); err != nil {
		log.Fatal().Str("context", "app").Msg(err.Error())
	}

	app.Client = NewClient(*app.Config)

	return app
}
