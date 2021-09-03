package config

import (
	"os"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Key struct {
	Rename    string `yaml:"rename"` //Rename into
	Transform struct {
		Time   bool `yaml:"time"`    //Transform into time
		GeoIP  bool `yaml:"geoip"`   //Add GeoIP data from field
		IP2Int bool `yaml:"iptoint"` //Add IP2Int from field
		Int2IP bool `yaml:"inttoip"` //Add Int2IP from field
	} `yaml:"transform"` //Transform options
}

type Routine struct {
	Index      string         `yaml:"index"`  //Elasticsearch Index
	Prefix     string         `yaml:"prefix"` //Redis Prefix
	Keys       map[string]Key `yaml:"keys"`   //JSON fields
	UDPRoutine bool           `yaml:"udp"`    //UDP Routine?
}

type config struct {
	Prometheus struct {
		Listen string `yaml:"listen"`
	} `yaml:"prometheus"`
	Elasticsearch struct {
		Hosts   []string `yaml:"hosts"`
		Timeout struct {
			ResponseHeader int `yaml:"response_header"`
			Dial           int `yaml:"dial"`
		} `yaml:"timeout"`
	} `yaml:"elasticsearch"`
	Redis struct {
		Server   string `yaml:"server"`   //Redis host:port
		Password string `yaml:"password"` //Redis password
		Database int    `yaml:"database"` //Redis database
	} `yaml:"redis"`
	GeoIP struct {
		Enable   bool `yaml:"enable"`
		Database struct {
			ASN     string `yaml:"asn"`
			Country string `yaml:"country"`
		} `yaml:"database"`
	} `yaml:"geoip"`
	Routines   map[string]Routine `yaml:"routines"`    //Routine configs
	UDPRoutine bool               `yaml:"udp_routine"` //Enable UDP Routine
}

var Cfg config

func Load() {
	f, err := os.Open("/etc/redship.yaml")
	if err != nil {
		log.Fatalf("Failed to load config, %v", err)
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&Cfg)
}
