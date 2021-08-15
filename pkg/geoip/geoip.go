package geoip

import (
	"log"

	"github.com/jphhofmann/redship/pkg/config"

	"github.com/oschwald/geoip2-golang"
)

var DB_ASN *geoip2.Reader
var DB_Country *geoip2.Reader

func Open() {
	var err error = nil
	DB_ASN, err = geoip2.Open(config.Cfg.GeoIP.Database.ASN)
	if err != nil {
		log.Fatal(err)
	}
	DB_Country, err = geoip2.Open(config.Cfg.GeoIP.Database.Country)
	if err != nil {
		log.Fatal(err)
	}
}
