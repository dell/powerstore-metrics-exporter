package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore/collector/client"
)

var metricVgCollectorMetric = []string{
	"avg_read_latency",
	"avg_latency",
	"avg_write_latency",
	"avg_read_iops",
	"avg_read_bandwidth",
	"avg_total_iops",
	"avg_total_bandwidth",
	"avg_write_iops",
	"avg_write_bandwidth",
}

var metricMetricVgDescMap = map[string]string{
	"avg_read_latency":    "avg latency time of read,unit is ms",
	"avg_latency":         "avg latency time,unit is ms",
	"avg_write_latency":   "avg latency time of write,unit is ms",
	"avg_read_iops":       "iops of read,unit is iops",
	"avg_read_bandwidth":  "bandwidth of read,unit is bps",
	"avg_total_iops":      "total iops,unit is iops",
	"avg_total_bandwidth": "total bandwidth,unit is bps",
	"avg_write_iops":      "iops of write,unit is iops",
	"avg_write_bandwidth": "bandwidth of write,unit is bps",
}

type metricVgCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewMetricVgCollector(api *client.Client, logger log.Logger) *metricVgCollector {
	metrics := getMetricVgfMetrics(api.IP)
	return &metricVgCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *metricVgCollector) Collect(ch chan<- prometheus.Metric) {
	vgArray := client.PowerstoreModuleID[c.client.IP]
	for _, vgId := range gjson.Parse(vgArray["volumegroup"]).Array() {
		id := vgId.Get("id").String()
		name := vgId.Get("name").String()
		applianceIDs := vgId.Get("appliance_ids").Array()
		metricVgData, err := c.client.GetMetricVg(id)
		if err != nil {
			level.Warn(c.logger).Log("msg", "get volume group performance data error", "err", err)
			continue
		}
		vgDataArray := gjson.Parse(metricVgData).Array()
		if len(vgDataArray) == 0 {
			continue
		}
		vgData := vgDataArray[len(vgDataArray)-1]
		for _, applianceID := range applianceIDs {
			for _, metricName := range metricVgCollectorMetric {
				metricValue := vgData.Get(metricName)
				metricDesc := c.metrics["vg"+"_"+metricName]
				if metricValue.Exists() && metricValue.Type != gjson.Null {
					ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), name, applianceID.String())
				}
			}
		}
	}
}

func (c *metricVgCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getMetricVgfMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, metricName := range metricVgCollectorMetric {
		res["vg"+"_"+metricName] = prometheus.NewDesc(
			"powerstore_metricVg_"+metricName,
			getMetricVgDescByType(metricName),
			[]string{"volume_group_id", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}
	return res
}

func getMetricVgDescByType(key string) string {
	if v, ok := metricMetricVgDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
