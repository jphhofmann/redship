package elastic

import (
	"net"
	"net/http"
	"time"

	"github.com/jphhofmann/redship/pkg/config"

	"github.com/elastic/go-elasticsearch"
)

var Client *elasticsearch.Client

/* Open elasticsearch handle */
func Open() error {
	cfg := elasticsearch.Config{
		Addresses: config.Cfg.Elasticsearch.Hosts,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Duration(config.Cfg.Elasticsearch.Timeout.ResponseHeader) * time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Duration(config.Cfg.Elasticsearch.Timeout.Dial) * time.Second}).DialContext,
		},
	}
	var err error = nil
	Client, err = elasticsearch.NewClient(cfg)
	if err != nil {
		return err
	}
	return nil
}
