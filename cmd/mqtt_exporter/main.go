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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var listenAddress = flag.String("listen-address", ":8181", "The address to listen on for HTTP requests.")
var metricsPath = flag.String("telemetry-path", "/metrics", "The path under which to expose metrics.")
var configPath = flag.String("config-path", "/etc/mqtt-exporter.toml", "The configuration file path to use.")

var app *mqttexporter.App

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func main() {
	flag.Parse()

	var err error
	app, err = mqttexporter.NewApp(*configPath, *listenAddress, *metricsPath)
	if err != nil {
		log.Fatal().Str("context", "main").Msgf("fail to init app with error: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := app.Start(); err != http.ErrServerClosed {
			log.Fatal().Str("context", "main").Msgf("app exited with error: %v", err)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		panic(err)
	}
}
