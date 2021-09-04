package routines

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jphhofmann/redship/pkg/config"
	"github.com/jphhofmann/redship/pkg/redis"

	"github.com/jphhofmann/redship/pkg/elastic"

	"github.com/jphhofmann/redship/pkg/geoip"

	"github.com/jphhofmann/redship/pkg/prom"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/prometheus/client_golang/prometheus"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func ConvTime(field int64) string {
	return time.Unix(field, 0).Format(time.RFC3339)
}

/* Check type of ip-address */
func CheckIPFamily(ip string) int {
	if net.ParseIP(ip) == nil {
		return 0
	}
	for i := 0; i < len(ip); i++ {
		switch ip[i] {
		case '.':
			return 4
		case ':':
			return 6
		}
	}
	return 0
}

/* Convert IPv4 to decimal */
func IP4ToInt(IPv4Addr string) int64 {
	bits := strings.Split(IPv4Addr, ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])
	var sum int64
	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)
	return sum
}

/* Convert decimal to IPv4 */
func IntToIP4(ipInt int64) string {
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}

/* Convert IPv6 to decimal */
func IP6ToInt(IPv6Addr string) *big.Int {
	ip := net.ParseIP(IPv6Addr)
	IPv6Int := big.NewInt(0)
	IPv6Int.SetBytes(ip.To16())
	return IPv6Int
}

/* Convert decimal to IPv6 */
func IntToIP6(intipv6 *big.Int) string {
	ip := intipv6.Bytes()
	return string(ip)
}

func Transform(output map[string]interface{}, entry string, value config.Key) map[string]interface{} {
	/* Apply transformations */
	if value.Transform.Time {
		switch data := output[entry].(type) {
		case float64:
			output[entry] = ConvTime(int64(data))
		case string:
			output[entry], _ = strconv.ParseInt(output[entry].(string), 10, 64)
			output[entry] = ConvTime(output[entry].(int64))
		}
	}

	/* Apply GeoIP data */
	if value.Transform.GeoIP && config.Cfg.GeoIP.Enable {
		switch data := output[entry].(type) {
		case string:
			ip := net.ParseIP(data)
			asn, err := geoip.DB_ASN.ASN(ip)
			if err == nil {
				output[entry+"_geoip_isp"] = asn.AutonomousSystemOrganization
				output[entry+"_geoip_asn"] = strconv.Itoa(int(asn.AutonomousSystemNumber))
			}
			country, err := geoip.DB_Country.Country(ip)
			if err == nil {
				output[entry+"_geoip_country"] = country.Country.IsoCode
			}
		}
	}

	/* Apply IP to integer */
	if value.Transform.IP2Int {
		switch data := output[entry].(type) {
		case string:
			family := CheckIPFamily(data)
			if family == 4 {
				output[entry+"_int"] = fmt.Sprintf("%v", IP4ToInt(data))
			} else if family == 6 {
				output[entry+"_int"] = fmt.Sprintf("%v", IP6ToInt(data))
			}
		}
	}

	/* Apply integer to IP */
	if value.Transform.Int2IP {
		switch data := output[entry].(type) {
		case float64:
			if data <= 4294967295 {
				output[entry+"_ip"] = IntToIP4(int64(data))
			} else {
				output[entry+"_ip"] = IntToIP6(output[entry].(*big.Int))
			}
		case string:
			output[entry+"_ip"] = IntToIP6(output[entry].(*big.Int))
		}
	}

	return output
}

func Routine(routine string) {
	for {
		output := make(map[string]interface{})
		var fields map[string]interface{}
		var cursor uint64
		var keys []string
		var err error
		keys, cursor, err = redis.Client.Scan(redis.Ctx, cursor, config.Cfg.Routines[routine].Prefix, 10000).Result()
		if err != nil {
			prom.Metrics.Routines.Errors.With(prometheus.Labels{"routine": routine}).Inc()
			log.Errorf("Failed to retrieve keys, %v", err)
		} else {
			for _, key := range keys {
				result, err := redis.Client.Get(redis.Ctx, key).Result()
				if err != nil {
					prom.Metrics.Routines.Errors.With(prometheus.Labels{"routine": routine}).Inc()
					log.Errorf("Failed to get key %v, %v", key, err)
				} else {
					err := json.Unmarshal([]byte(result), &fields)
					if err != nil {
						prom.Metrics.Routines.Errors.With(prometheus.Labels{"routine": routine}).Inc()
						log.Errorf("Caught invalid json, %v", err)
					} else {
						/* Get all fields */
						output = fields

						/* Apply actions on fields */
						for key, value := range config.Cfg.Routines[routine].Keys {
							/* Rename field? */
							var entry string = key
							if len(value.Rename) != 0 {
								entry = value.Rename
							}

							/* Final field */
							output[entry] = fields[key]
							output = Transform(output, entry, value)
						}
					}

					/* Create new json from output interface */
					result, err := json.Marshal(output)
					if err != nil {
						prom.Metrics.Routines.Errors.With(prometheus.Labels{"routine": routine}).Inc()
						log.Errorf("Failed to marshal json, %v", err)
					} else {
						/* Ship to elasticsearch */
						req := esapi.IndexRequest{
							Index:      config.Cfg.Routines[routine].Index + "-" + time.Now().Format("2006-01-02"),
							DocumentID: uuid.NewV4().String(),
							Body:       strings.NewReader(string(result)),
							Refresh:    "false",
						}
						/* Carry out ES Doc creation */
						res, err := req.Do(context.Background(), elastic.Client)
						if err != nil {
							prom.Metrics.Routines.Errors.With(prometheus.Labels{"routine": routine}).Inc()
							log.Errorf("Error getting response, %v", err)
						} else {
							prom.Metrics.Routines.Exported.With(prometheus.Labels{"routine": routine}).Inc()
							res.Body.Close()
						}
					}
					redis.Client.Del(redis.Ctx, key)
				}
			}
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func UDPRoutine() {

	/* Open UDP Connection */
	s, err := net.ResolveUDPAddr("udp4", "0.0.0.0:1286")
	if err != nil {
		log.Error(err)
		return
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		log.Error(err)
		return
	}

	defer connection.Close()
	buffer := make([]byte, 1024)
	output := make(map[string]interface{})
	var fields map[string]interface{}
	for {
		n, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			log.Error(err)
		} else {
			err := json.Unmarshal(buffer[0:n], &fields)
			if err != nil {
				prom.Metrics.Routines.Errors.With(prometheus.Labels{"routine": "undefined"}).Inc()
				log.Errorf("Caught invalid json, %v", err)
			} else {
				switch r := fields["routine"].(type) {
				case string:
					/* Get all fields */
					output = fields

					/* Apply actions on fields */
					for key, value := range config.Cfg.Routines[r].Keys {
						/* Rename field? */
						var entry string = key
						if len(value.Rename) != 0 {
							entry = value.Rename
						}

						/* Final field */
						output[entry] = fields[key]
						output = Transform(output, entry, value)
					}

					/* Create new json from output interface */
					result, err := json.Marshal(output)
					if err != nil {
						prom.Metrics.Routines.Errors.With(prometheus.Labels{"routine": r}).Inc()
						log.Errorf("Failed to marshal json, %v", err)
					} else {
						/* Ship to elasticsearch */
						req := esapi.IndexRequest{
							Index:      config.Cfg.Routines[r].Index + "-" + time.Now().Format("2006-01-02"),
							DocumentID: uuid.NewV4().String(),
							Body:       strings.NewReader(string(result)),
							Refresh:    "false",
						}
						/* Carry out ES Doc creation */
						res, err := req.Do(context.Background(), elastic.Client)
						if err != nil {
							prom.Metrics.Routines.Errors.With(prometheus.Labels{"routine": r}).Inc()
							log.Errorf("Error getting response, %v", err)
						} else {
							prom.Metrics.Routines.Exported.With(prometheus.Labels{"routine": r}).Inc()
							res.Body.Close()
						}
					}
				}
			}
		}
	}
}
