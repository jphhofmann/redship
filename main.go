package main

import (
	"time"

	"github.com/jphhofmann/redship/pkg/elastic"
	"github.com/jphhofmann/redship/pkg/prom"
	"github.com/jphhofmann/redship/pkg/routines"

	"github.com/jphhofmann/redship/pkg/config"

	"github.com/jphhofmann/redship/pkg/redis"

	"github.com/jphhofmann/redship/pkg/geoip"

	log "github.com/sirupsen/logrus"
)

func main() {
	/* Load config file */
	config.Load()

	/* Run prometheus exporter */
	go prom.Exporter()

	/* Open GeoIP databases */
	if config.Cfg.GeoIP.Enable {
		geoip.Open()
	}

	/* Open elasticsearch client */
	err := elastic.Open()
	if err != nil {
		log.Fatalf("Failed to open elasticsearch client, %v", err)
	}

	/* Open redis client */
	redis.Client = redis.Connect()

	/* Run redship routines */
	for routine, _ := range config.Cfg.Routines {
		log.Infof("Starting redship routine %v", routine)
		go routines.Routine(routine)
	}

	/* Update uptime */
	for {
		if prom.Metrics.Uptime != nil {
			prom.Metrics.Uptime.Add(1)
		}
		time.Sleep(1 * time.Second)
	}
}
