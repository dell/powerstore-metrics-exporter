package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore/collector/client"
)

var metricFileSystemCollector = []string{
	"size_total",
	"size_used",
}

// file description
var metricFileSystemDescMap = map[string]string{
	"size_total": "filesystem total size",
	"size_used":  "filesystem used size",
}

type fileSystemCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewFileCollector(api *client.Client, logger log.Logger) *fileSystemCollector {
	metrics := getFileSystemMetrics(api.IP)
	return &fileSystemCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *fileSystemCollector) Collect(ch chan<- prometheus.Metric) {
	fileData, err := c.client.GetFile()
	if err != nil {
		level.Warn(c.logger).Log("msg", "get file system data error", "err", err)
		return
	}
	for _, file := range gjson.Parse(fileData).Array() {
		name := file.Get("name").String()
		id := file.Get("appliance_id").String()
		for _, metricName := range metricFileSystemCollector {
			metricValue := file.Get(metricName)
			metricDesc := c.metrics[metricName]
			if metricValue.Exists() && metricValue.Type != gjson.Null {
				ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), name, id)
			}
		}
	}
}

func (c *fileSystemCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getFileSystemMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, metricName := range metricFileSystemCollector {
		res[metricName] = prometheus.NewDesc(
			"powerstore_filesystem_"+metricName,
			getFileSystemDescByType(metricName),
			[]string{"name", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}
	return res
}

func getFileSystemDescByType(key string) string {
	if v, ok := metricFileSystemDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
