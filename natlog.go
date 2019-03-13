package main

import (
	"flag"
	"fmt"
	"github.com/ircop/natlog/cfg"
	"github.com/ircop/natlog/chwriter"
	"github.com/ircop/natlog/parser"
	"go.uber.org/zap"
	"log"
	"os"
	"gopkg.in/mcuadros/go-syslog.v2"
)

func main() {
	configPath := flag.String("c", "./natlog.toml", "Config file location")
	flag.Parse()
	config, err := cfg.NewConfig(*configPath)
	if nil != err {
		log.Fatalf("Error reading config: %s", err.Error())
	}

	var logger *zap.Logger
	if os.Getenv("ENV") == "dev" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		 log.Fatalf("error initializing logger: %s", err.Error())
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	zap.L().Info("Starting syslog collector")

	// todo: init CH

	dataChan := make(chan *parser.NatRecord, 10000)

	parser.Init(dataChan)

	if err = chwriter.Init(config.Ch.ConnectionString, config.Ch.Interval, config.Ch.Count, dataChan); err != nil {
		zap.L().Error("Failed to init chwriter", zap.Error(err))
		return
	}

	logChan := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(logChan)
	for i := 0; i < config.Listener.Workers; i++ {
		go logHandler(logChan)
	}

	// syslog server
	srv := syslog.NewServer()

	srv.SetFormat(syslog.RFC3164)
	srv.SetHandler(handler)
	if err = srv.ListenUDP(fmt.Sprintf("%s:%d", config.Listener.IP, config.Listener.Port)); err != nil {
		zap.L().Error("Failed to init listener", zap.Error(err))
		return
	}
	if err = srv.Boot(); err != nil {
		zap.L().Error("Failed to start syslog server", zap.Error(err))
		return
	}

	// go logHandler(logChan)

	srv.Wait()
}


func logHandler(channel syslog.LogPartsChannel) {
	for parts := range channel {
		parser.ParseMessage(parts["content"].(string))
	}
}

