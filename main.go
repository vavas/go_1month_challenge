package main

import (

	"github.com/vavas/go_services/gnats"
	"github.com/vavas/go_services/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/vavas/go_mc_gateway/config"
	"github.com/vavas/go_mc_gateway/gateway"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configPath = kingpin.Flag("config", "Path to config file.").Short('c').ExistingFile()
)

// go run ~/Work/telemetrytv/src/github.com/vavas/go_mc_gateway/main.go -c ~/Work/telemetrytv/src/github.com/vavas/go_mc_gateway/config.toml

func main() {
	kingpin.Parse()

	// Load config
	conf := config.ReadConfig(*configPath)

	println("")
	println("")
	println("__   ____ ___   ____ _ ___")
	println("\\ \\ / / _` \\ \\ / / _` / __|")
	println(" \\ V / (_| |\\ V / (_| \\__ \\")
	println("  \\_/ \\__,_| \\_/ \\__,_|___/")
	println("  API Gateway")
	println("")
	println("")

	// Init logger
	logger.InitLogging("go_mc_gateway")
	log := logger.Logger

	// Connect to Nats
	// TODO: Nats client should be an object managed by gateway object internally
	log.Debug("Connecting to Gnats")
	if err := gnats.Connect(conf.Gnats); err != nil {
		log.Fatal("gnats connects fail")
	}

	gtw, err := gateway.New(conf, log)
	if err != nil {
		log.Fatal("gnats connects fail")
	}
	log.Debug("API gateway created")

	log.Debug("Starting API gateway")
	gtw.Start()

	log.Debug("API gateway started")
	log.Debug("All systems online")

	// Gracefully shutdownn on kill signals
	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(
		osSignalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	osSignal := <-osSignalChan

	log.Sugar().Debugf("The OS has requested program termination with signal: %v\n", osSignal.String())
	log.Debug("Beginning system shutdown")
	if err := gtw.Stop(); err != nil {
		log.Fatal("Gateway stop fail")
	}
	log.Debug("All systems are offline")
	log.Debug("Goodbye!")

}
