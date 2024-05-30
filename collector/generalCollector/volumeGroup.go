package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore/collector/client"
)

var volumeGroupCollectorMetrics = []string{
	"logical_provisioned",
	"logical_used",
}

var metricVolumeGroupDescMap = map[string]string{
	"logical_provisioned": "logical provisioned,unit is B",
	"logical_used":        "logical used,unit is B",
}

type volumeGroupCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewVolumeGroupCollector(api *client.Client, logger log.Logger) *volumeGroupCollector {
	metrics := getVolumeGroupMetrics(api.IP)
	return &volumeGroupCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *volumeGroupCollector) Collect(ch chan<- prometheus.Metric) {
	volumeGroupData, err := c.client.GetVolumeGroup()
	if err != nil {
		level.Warn(c.logger).Log("msg", "get volume group data error", "err", err)
		return
	}
	for _, volumeGroup := range gjson.Parse(volumeGroupData).Array() {
		name := volumeGroup.Get("name").String()
		for _, applianceID := range volumeGroup.Get("appliance_ids").Array() {
			for _, metricName := range volumeGroupCollectorMetrics {
				metricValue := volumeGroup.Get(metricName)
				metricDesc := c.metrics[metricName]
				if metricValue.Exists() && metricValue.Type != gjson.Null {
					ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), name, applianceID.String())
				}
			}
		}
	}
}

func (c *volumeGroupCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getVolumeGroupMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, metricName := range volumeGroupCollectorMetrics {
		res[metricName] = prometheus.NewDesc(
			"powerstore_volumegroup_"+metricName,
			getVolumeGroupDescByType(metricName),
			[]string{"name", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}
	return res
}

func getVolumeGroupDescByType(key string) string {
	if v, ok := metricVolumeGroupDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
