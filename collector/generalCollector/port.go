package generalCollector

import (
	"powerstore-metrics-exporter/collector/client"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

var portTypes = []string{
	"eth_port",
	"fc_port",
}

var portCollectorMetrics = []string{
	"is_link_up",
	"current_speed",
}

var portStatusMetricMap = map[string]map[string]int{
	"is_link_up": {"true": 1, "false": 0},
}

// port description
var metricPortDescMap = map[string]string{
	"is_link_up":    "Indicates whether the port's link is up:true is 1,false is 0",
	"current_speed": "Supported Ethernet front-end port transmission speeds,units is Gps",
}

type portCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewPortCollector(api *client.Client, logger log.Logger) *portCollector {
	metrics := getPortMetrics(api.IP)
	return &portCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *portCollector) Collect(ch chan<- prometheus.Metric) {
	level.Info(c.logger).Log("msg", "Start collecting port data")
	startTime := time.Now()
	for _, portType := range portTypes {
		portTypeData, err := c.client.GetPort(portType)
		if err != nil {
			level.Warn(c.logger).Log("msg", "get "+portType+" data error", "err", err)
			return
		}
		for _, data := range gjson.Parse(portTypeData).Array() {
			name := data.Get("name").String()
			id := data.Get("appliance_id").String()
			for _, metricName := range portCollectorMetrics {
				metricValue := getPortFloatDate(metricName, data.Get(metricName))
				metricDesc := c.metrics[portType+metricName]
				ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue, id, name)
			}
		}
	}
	level.Info(c.logger).Log("msg", "Obtaining the port is successful", "time", time.Since(startTime))
}

func (c *portCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getPortFloatDate(key string, value gjson.Result) float64 {
	if v, ok := portStatusMetricMap[key]; ok {
		if res, ok2 := v[value.String()]; ok2 {
			return float64(res)
		} else {
			return float64(v["other"])
		}
	} else if key == "current_speed" {
		if value.Type == gjson.Null {
			return 0
		}
		rs := []rune(value.String())
		speed := string(rs[0:strings.Index(value.String(), "_")])
		result, _ := strconv.Atoi(speed)
		return float64(result)
	} else {
		return value.Float()
	}
}

func getPortMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, portType := range portTypes {
		for _, metricName := range portCollectorMetrics {
			res[portType+metricName] = prometheus.NewDesc(
				"powerstore_"+portType+"_"+metricName,
				getPortDescByType(metricName),
				[]string{"appliance_id", portType + "_id"},
				prometheus.Labels{"IP": ip})
		}
	}
	return res
}

func getPortDescByType(key string) string {
	if v, ok := metricPortDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
