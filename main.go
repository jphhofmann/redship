package main

import (
	"redship/package/config"
	"redship/package/elastic"
	"redship/package/prom"
	"redship/package/routines"
	"time"

	"redship/package/redis"

	"redship/package/geoip"

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
