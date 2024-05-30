package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore/collector/client"
)

var metricApplianceDescMap = map[string]string{
	"service_tag": "service tag information",
}

type applianceCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewApplianceCollector(api *client.Client, logger log.Logger) *applianceCollector {
	metrics := getApplianceMetrics(api.IP)
	return &applianceCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *applianceCollector) Collect(ch chan<- prometheus.Metric) {
	applianceData, err := c.client.GetAppliance()
	if err != nil {
		level.Warn(c.logger).Log("msg", "get appliance data error", "err", err)
		return
	}
	for _, appliance := range gjson.Parse(applianceData).Array() {
		tag := appliance.Get("service_tag")
		applianceID := appliance.Get("id").String()
		metricDesc := c.metrics["tag"]
		if tag.Exists() && tag.Type != gjson.Null {
			ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, 0, tag.String(), applianceID)
		}
	}
}

func (c *applianceCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getApplianceMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	res["tag"] = prometheus.NewDesc(
		"powerstore_appliance",
		getApplianceDescByType("service_tag"),
		[]string{"service_tag", "appliance_id"},
		prometheus.Labels{"IP": ip})

	return res
}

func getApplianceDescByType(key string) string {
	if v, ok := metricApplianceDescMap[key]; ok {
		return v
	} else {
		return "this is " + key
	}
}
