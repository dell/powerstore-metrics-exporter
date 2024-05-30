package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore/collector/client"
)

var volumeCollectorMetrics = []string{
	"state",
	"size",
	"logical_used",
}

var metricVolumeDescMap = map[string]string{
	"state":        "1 is ready ,0 is other",
	"size":         "the unit is B",
	"logical_used": "the unit is B",
}

var statusVolumeMetricsMap = map[string]map[string]int{
	"state": {"Ready": 1, "other": 0},
}

type volumeCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewVolumeCollector(api *client.Client, logger log.Logger) *volumeCollector {
	metrics := getVolumeMetrics(api.IP)
	return &volumeCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *volumeCollector) Collect(ch chan<- prometheus.Metric) {
	volumeData, err := c.client.GetVolume()
	if err != nil {
		level.Warn(c.logger).Log("msg", "get volume data error", "err", err)
		return
	}
	for _, volume := range gjson.Parse(volumeData).Array() {
		name := volume.Get("name").String()
		id := volume.Get("appliance_id").String()
		for _, metricName := range volumeCollectorMetrics {
			metricValue := getVolumeFloatDate(metricName, volume.Get(metricName))
			metricDesc := c.metrics[metricName]
			ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue, name, id)
		}
	}
}

func (c *volumeCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getVolumeFloatDate(key string, value gjson.Result) float64 {
	if v, ok := statusVolumeMetricsMap[key]; ok {
		if res, ok2 := v[value.String()]; ok2 {
			return float64(res)
		} else {
			return float64(v["other"])
		}
	} else {
		return value.Float()
	}
}

func getVolumeMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, metricName := range volumeCollectorMetrics {
		res[metricName] = prometheus.NewDesc(
			"powerstore_volume_"+metricName,
			getVolumeDescByType(metricName),
			[]string{"name", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}

	return res
}

func getVolumeDescByType(key string) string {
	if v, ok := metricVolumeDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
