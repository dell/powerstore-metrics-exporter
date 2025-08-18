package generalCollector

import (
	"powerstore-metrics-exporter/collector/client"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

var metricAppliancePerfCollectorMetric = []string{
	"avg_read_latency",
	"avg_latency",
	"avg_write_latency",
	"avg_read_iops",
	"avg_read_bandwidth",
	"avg_total_iops",
	"avg_total_bandwidth",
	"avg_write_iops",
	"avg_write_bandwidth",
	"avg_io_workload_cpu_utilization",
	"avg_io_size",
	"avg_read_size",
	"avg_write_size",
}

// performance description
var metricAppliancePerfDescMap = map[string]string{
	"avg_read_latency":                "Average read latency in microseconds,unit is ms",
	"avg_latency":                     "Average read and write latency in microseconds,unit is ms",
	"avg_write_latency":               "Average write latency in microseconds,unit is ms",
	"avg_read_iops":                   "Total read operations per second,unit is iops",
	"avg_read_bandwidth":              "Read rate in bytes per second,unit is bps",
	"avg_total_iops":                  "Total read and write operations per second,unit is iops",
	"avg_total_bandwidth":             "Total data transfer rate in bytes per second,unit is bps",
	"avg_write_iops":                  "Total write operations per second,unit is iops",
	"avg_write_bandwidth":             "Write rate in bytes per second,unit is bps",
	"avg_io_workload_cpu_utilization": "The percentage of CPU Utilization on the cores dedicated to servicing storage I/O requests.unit is %",
	"avg_io_size":                     "Average size of read and write operations in bytes.unit is bytes",
	"avg_write_size":                  "Average write size in bytes.unit is bytes",
	"avg_read_size":                   "Average read size in bytes.unit is bytes",
}

type metricApplianceCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewMetricApplianceCollector(api *client.Client, logger log.Logger) *metricApplianceCollector {
	metrics := getMetricApplianceMetrics(api.IP)
	return &metricApplianceCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *metricApplianceCollector) Collect(ch chan<- prometheus.Metric) {
	level.Info(c.logger).Log("msg", "Start collecting appliance performance data")
	startTime := time.Now()
	var wg sync.WaitGroup
	applianceArray := client.PowerstoreModuleID[c.client.IP]
	for applianceID, applianceName := range applianceArray["appliance"] {
		wg.Add(1)
		go func(applianceID, applianceName string) {
			defer wg.Done()
			perfData, err := c.client.GetPerf(applianceID)
			if err != nil {
				level.Warn(c.logger).Log("msg", "get appliance performance data error", "err", err)
			}
			appliancePerformanceArray := gjson.Parse(perfData).Array()
			appliancePerformance := appliancePerformanceArray[len(appliancePerformanceArray)-1]
			for _, metricName := range metricAppliancePerfCollectorMetric {
				metricValue := appliancePerformance.Get(metricName)
				metricDesc := c.metrics["appliance"+"_"+metricName]
				if metricValue.Exists() && metricValue.Type != gjson.Null {
					ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), applianceID, applianceName)
				}
			}
		}(applianceID, applianceName.String())
	}
	wg.Wait()
	level.Info(c.logger).Log("msg", "Obtaining the performance appliance is successful", "time", time.Since(startTime))
}

func (c *metricApplianceCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getMetricApplianceMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, metricName := range metricAppliancePerfCollectorMetric {
		res["appliance"+"_"+metricName] = prometheus.NewDesc(
			"powerstore_perf_"+metricName,
			getMetricApplianceDescByType(metricName),
			[]string{"appliance_id", "appliance_name"},
			prometheus.Labels{"IP": ip})
	}
	return res
}

func getMetricApplianceDescByType(key string) string {
	if v, ok := metricAppliancePerfDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
