package main

import (
	"befw/internal/agent"
	"befw/internal/logger"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
)

type Config struct {
	Address        string `long:"address" short:"a" env:"ADDRESS" default:"127.0.0.1:8080" description:"set remote server"`
	LogFile        string `long:"log_file" env:"LOG_FILE" default:"" description:"set log file"`
	LogLevel       string `long:"log_level" env:"LOG_LEVEL" default:"info" description:"set log level"`
	PollInterval   int    `long:"poll_interval" short:"p" env:"POLL_INTERVAL" default:"30" description:"set poll interval"`
	ReportInterval int    `long:"report_interval" short:"r" env:"REPORT_INTERVAL" default:"10" description:"set stats update interval"`
	Tenant         string `long:"tenant" short:"t" env:"TENANT" description:"set current tenant"`
}

func main() {
	var cfg Config

	// parse flags
	parser := flags.NewParser(&cfg, flags.HelpFlag)
	if _, err := parser.Parse(); err != nil {
		var e *flags.Error

		if errors.As(err, &e) {
			if e.Type == flags.ErrHelp {
				log.Printf("%s", e.Message)
				os.Exit(0)
			}
		}
		log.Fatalf("error parse arguments:%+v\n", err)
	}

	if cfg.Tenant == "" {
		log.Fatalf("invalid tenant\n")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Init logger
	logger.NewLogger(cfg.LogLevel, cfg.LogFile)
	defer logger.Close()

	logger.Info("Agent started")

	ipfwAgent := agent.New()
	ipfwAgent.Tenant = cfg.Tenant
	ipfwAgent.ServerURL = fmt.Sprintf("http://%s/api/v1/rule", cfg.Address)
	ipfwAgent.PollInterval = cfg.PollInterval
	ipfwAgent.ReportInterval = cfg.ReportInterval

	//start watcher
	go ipfwAgent.Watcher(ctx)
	go ipfwAgent.SendStat(ctx)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	<-sigint

	logger.Info("Agent stopped")
}
