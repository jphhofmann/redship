package prom

import (
	"net/http"

	"github.com/jphhofmann/redship/pkg/config"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metrics struct {
	Uptime   prometheus.Gauge
	Routines struct {
		Exported *prometheus.GaugeVec
		Errors   *prometheus.GaugeVec
	}
}

var Metrics metrics

func Exporter() {

	/* Initialize metrics */
	Metrics.Uptime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "uptime", Help: "redship uptime"})
	Metrics.Routines.Exported = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "routines",
			Name:      "exported",
			Help:      "Exported keys",
		},
		[]string{"routine"})
	Metrics.Routines.Errors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "routines",
			Name:      "errors",
			Help:      "Export errors",
		},
		[]string{"routine"})

	/* Initialize metrics */
	prometheus.MustRegister(Metrics.Uptime)
	prometheus.MustRegister(Metrics.Routines.Errors)
	prometheus.MustRegister(Metrics.Routines.Exported)

	/* Start Webserver */
	log.Infof("Beginning to serve metrics on %v", config.Cfg.Prometheus.Listen)
	log.Fatal(http.ListenAndServe(config.Cfg.Prometheus.Listen, promhttp.Handler()))
}
