package main

import (
	"time"

	"github.com/jphhofmann/redship/pkg/elastic"
	"github.com/jphhofmann/redship/pkg/prom"
	"github.com/jphhofmann/redship/pkg/redis"
	"github.com/jphhofmann/redship/pkg/routines"

	"github.com/jphhofmann/redship/pkg/config"

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

	/* Enable Redis import */
	if config.Cfg.RedisRoutine {
		/* Open redis client */
		redis.Client = redis.Connect()
		/* Spawn go routines */
		for routine, cfg := range config.Cfg.Routines {
			log.Infof("Starting redship routine %v", routine)
			if !cfg.UDPRoutine {
				go routines.Routine(routine)
			}
		}
	}

	/* Enable UDP import */
	if config.Cfg.UDPRoutine {
		go routines.UDPRoutine()
	}

	/* Update uptime */
	for {
		if prom.Metrics.Uptime != nil {
			prom.Metrics.Uptime.Add(1)
		}
		time.Sleep(1 * time.Second)
	}
}
