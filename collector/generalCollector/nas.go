package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore/collector/client"
)

var statuNasMetricsMap = map[string]map[string]int{
	"operational_status": {"Started": 1, "other": 0},
}

var metricNasDescMap = map[string]string{
	"operational_status": "this is operational_status,Started is 1 other is 0",
}

type nasCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewNasCollector(api *client.Client, logger log.Logger) *nasCollector {
	metrics := getNasMetrics(api.IP)
	return &nasCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *nasCollector) Collect(ch chan<- prometheus.Metric) {
	nasData, err := c.client.GetNas()
	if err != nil {
		level.Warn(c.logger).Log("msg", "get Nas data error", "err", err)
		return
	}
	for _, nas := range gjson.Parse(nasData).Array() {
		id := nas.Get("appliance_id").String()
		name := nas.Get("name").String()
		state := nas.Get("operational_status")
		value := getNasFloatData("operational_status", state)
		metricDesc := c.metrics["operational_status"]
		if state.Exists() && state.Type != gjson.Null {
			ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, value, name, id)
		}
	}
}

func getNasFloatData(key string, value gjson.Result) float64 {
	if v, ok := statuNasMetricsMap[key]; ok {
		if res, ok2 := v[value.String()]; ok2 {
			return float64(res)
		} else {
			return float64(v["other"])
		}
	} else {
		return value.Float()
	}
}

func (c *nasCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getNasMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	res["operational_status"] = prometheus.NewDesc(
		"powerstore_nas_server_operational_status",
		getNasDescByType("operational_status"),
		[]string{"name", "appliance_id"},
		prometheus.Labels{"IP": ip})
	return res
}

func getNasDescByType(key string) string {
	if v, ok := metricNasDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
