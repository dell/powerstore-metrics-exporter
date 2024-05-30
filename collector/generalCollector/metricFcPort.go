package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore/collector/client"
)

var metricFcPortCollectorMetric = []string{
	"avg_read_latency",
	"avg_latency",
	"avg_write_latency",
	"avg_total_iops",
	"avg_total_bandwidth",
	"avg_dumped_frames_ps",
	"avg_loss_of_signal_count_ps",
	"avg_invalid_crc_count_ps",
	"avg_loss_of_sync_count_ps",
	"avg_invalid_tx_word_count_ps",
	"avg_prim_seq_prot_err_count_ps",
	"avg_link_failure_count_ps",
}

var metricMetricFcPortDescMap = map[string]string{
	"avg_read_latency":               "avg latency time of read,unit is ms",
	"avg_latency":                    "avg latency time,unit is ms",
	"avg_write_latency":              "avg latency time of write,unit is ms",
	"avg_total_iops":                 "Total IOPS,unit is bps",
	"avg_total_bandwidth":            "Total Bandwidth,unit is bps",
	"avg_dumped_frames_ps":           "count of dumped frames in a second",
	"avg_loss_of_signal_count_ps":    "count of loss of signal in a second",
	"avg_invalid_crc_count_ps":       "count of invalid useless in a second",
	"avg_loss_of_sync_count_ps":      "count of loss of sync in a second",
	"avg_invalid_tx_word_count_ps":   "count of invalid send word in a second",
	"avg_prim_seq_prot_err_count_ps": "count of prim seq prot err in a second",
	"avg_link_failure_count_ps":      "count of link failure in a second",
}

type metricFcPortCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewMetricFcPortCollector(api *client.Client, logger log.Logger) *metricFcPortCollector {
	metrics := getMetricFcPortMetrics(api.IP)
	return &metricFcPortCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *metricFcPortCollector) Collect(ch chan<- prometheus.Metric) {
	fcPortArray := client.PowerstoreModuleID[c.client.IP]
	for _, portId := range gjson.Parse(fcPortArray["fcport"]).Array() {
		id := portId.Get("id").String()
		name := portId.Get("name").String()
		fcPortsData, err := c.client.GetMetricFcPort(id)
		if err != nil {
			level.Warn(c.logger).Log("msg", "get fcPort performance data error", "err", err)
			continue
		}
		fcPortDataArray := gjson.Parse(fcPortsData).Array()
		if len(fcPortDataArray) == 0 {
			continue
		}
		fcPortData := fcPortDataArray[len(fcPortDataArray)-1]
		applianceID := fcPortData.Get("appliance_id").String()
		for _, metricName := range metricFcPortCollectorMetric {
			metricValue := fcPortData.Get(metricName)
			metricDesc := c.metrics["fcport"+"_"+metricName]
			if metricValue.Exists() && metricValue.Type != gjson.Null {
				ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), name, applianceID)
			}
		}
	}
}

func (c *metricFcPortCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getMetricFcPortMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}

	for _, metricName := range metricFcPortCollectorMetric {
		res["fcport"+"_"+metricName] = prometheus.NewDesc(
			"powerstore_metricFcPort_"+metricName,
			getMetricFcPortDescByType(metricName),
			[]string{"fc_port_id", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}
	return res
}

func getMetricFcPortDescByType(key string) string {
	if v, ok := metricMetricFcPortDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
