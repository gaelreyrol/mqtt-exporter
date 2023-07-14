package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqttexporter "github.com/gaelreyrol/mqtt-exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var listenAddress = flag.String("listen-address", ":8181", "The address to listen on for HTTP requests.")
var metricsPath = flag.String("telemetry-path", "/metrics", "The path under which to expose metrics.")
var configPath = flag.String("config-path", "/etc/mqtt-exporter.toml", "The configuration file path to use.")

var app mqttexporter.App

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func main() {
	flag.Parse()

	app = mqttexporter.NewApp(*configPath, *listenAddress)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for _, topic := range app.Config.Topics {
		topicReg := prometheus.NewRegistry()
		app.Registry.Register(topicReg)

		mqttexporter.Subscribe(&app, topic, topicReg)
	}

	http.Handle(*metricsPath, promhttp.HandlerFor(
		app.Registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
			Registry:          app.Registry,
		},
	))

	go func() {
		if err := app.Server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Str("context", "main").Msgf("Server exited with error: %v", err)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Server.Shutdown(ctx); err != nil {
		panic(err)
	}
}
