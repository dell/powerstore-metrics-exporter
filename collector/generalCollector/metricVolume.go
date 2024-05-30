package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore/collector/client"
)

var metricVolumeCollectorMetric = []string{
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

var metricMetricVolumeDescMap = map[string]string{
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

type metricVolumeCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewMetricVolumeCollector(api *client.Client, logger log.Logger) *metricVolumeCollector {
	metrics := getMetricVolumeMetrics(api.IP)
	return &metricVolumeCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *metricVolumeCollector) Collect(ch chan<- prometheus.Metric) {
	volumeArray := client.PowerstoreModuleID[c.client.IP]
	for _, volId := range gjson.Parse(volumeArray["volume"]).Array() {
		id := volId.Get("id").String()
		name := volId.Get("name").String()
		metricVolData, err := c.client.GetMetricVolume(id)
		if err != nil {
			level.Warn(c.logger).Log("msg", "get volume performance data error", "err", err)
			continue
		}
		volumeDataArray := gjson.Parse(metricVolData).Array()
		if len(volumeDataArray) == 0 {
			continue
		}
		volumeData := volumeDataArray[len(volumeDataArray)-1]
		applianceID := volumeData.Get("appliance_id").String()
		for _, metricName := range metricVolumeCollectorMetric {
			metricValue := volumeData.Get(metricName)
			metricDesc := c.metrics["volume"+"_"+metricName]
			if metricValue.Exists() && metricValue.Type != gjson.Null {
				ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), name, applianceID)
			}
		}
	}
}

func (c *metricVolumeCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getMetricVolumeMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, metricName := range metricVolumeCollectorMetric {
		res["volume"+"_"+metricName] = prometheus.NewDesc(
			"powerstore_metricVolume_"+metricName,
			getMetricVolumeDescByType(metricName),
			[]string{"volume_id", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}
	return res
}

func getMetricVolumeDescByType(key string) string {
	if v, ok := metricMetricVolumeDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
