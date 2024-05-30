package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore/collector/client"
)

var metricEthPortCollectorMetric = []string{
	"avg_bytes_rx_ps",
	"avg_bytes_tx_ps",
	"avg_pkt_rx_crc_error_ps",
	"avg_pkt_rx_no_buffer_error_ps",
	"avg_pkt_rx_ps",
	"avg_pkt_tx_error_ps",
	"avg_pkt_tx_ps",
}

var metricMetricEthPortDescMap = map[string]string{
	"bytes_rx_ps":               "receive bytes in a second",
	"bytes_tx_ps":               "send bytes in a second",
	"pkt_rx_crc_error_ps":       "packet receive crc error in a second",
	"pkt_rx_no_buffer_error_ps": "packet receive no buffer error in a second",
	"pkt_rx_ps":                 "packet receive in a second",
	"pkt_tx_error_ps":           "packet send error in a second",
	"pkt_tx_ps":                 "packet get in a second",
}

type metricEthPortCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewMetricEthPortCollector(api *client.Client, logger log.Logger) *metricEthPortCollector {
	metrics := getMetricEthPortfMetrics(api.IP)
	return &metricEthPortCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *metricEthPortCollector) Collect(ch chan<- prometheus.Metric) {
	ethPortArray := client.PowerstoreModuleID[c.client.IP]
	for _, portId := range gjson.Parse(ethPortArray["ethport"]).Array() {
		id := portId.Get("id").String()
		name := portId.Get("name").String()
		ethPortsData, err := c.client.GetMetricEthPort(id)
		if err != nil {
			level.Warn(c.logger).Log("msg", "get ethPort performance data error", "err", err)
			continue
		}
		ethPortDataArray := gjson.Parse(ethPortsData).Array()
		if len(ethPortDataArray) == 0 {
			continue
		}
		ethPortData := ethPortDataArray[len(ethPortDataArray)-1]
		applianceID := ethPortData.Get("appliance_id").String()
		for _, metricName := range metricEthPortCollectorMetric {
			metricValue := ethPortData.Get(metricName)
			metricDesc := c.metrics["ethport"+"_"+metricName]
			if metricValue.Exists() && metricValue.Type != gjson.Null {
				ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), name, applianceID)
			}
		}
	}
}

func (c *metricEthPortCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getMetricEthPortfMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, metricName := range metricEthPortCollectorMetric {
		res["ethport"+"_"+metricName] = prometheus.NewDesc(
			"powerstore_metricEthPort_"+metricName,
			getMetricEthPortDescByType(metricName),
			[]string{"eth_port_id", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}
	return res
}

func getMetricEthPortDescByType(key string) string {
	if v, ok := metricMetricEthPortDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
