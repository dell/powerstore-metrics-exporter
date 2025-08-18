package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore-metrics-exporter/collector/client"
	"sync"
	"time"
)

type metricWearMetricCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewWearMetricCollector(api *client.Client, logger log.Logger) *metricWearMetricCollector {
	metrics := getWearMetrics(api.IP)
	return &metricWearMetricCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *metricWearMetricCollector) Collect(ch chan<- prometheus.Metric) {
	level.Info(c.logger).Log("msg", "Start collecting driver percent endurance remaining data")
	startTime := time.Now()
	var wg sync.WaitGroup
	driveArray := client.PowerstoreModuleID[c.client.IP]
	for driveID, driveName := range driveArray["drive"] {
		wg.Add(1)
		go func(driveID, driveName string) {
			defer wg.Done()
			result, err := c.client.GetWearMetricByDrive(driveID)
			if err != nil {
				level.Warn(c.logger).Log("msg", "get driver percent endurance remaining data error", "driver_id", driveID, "err", err)
				return
			}
			metricWearArray := gjson.Parse(result).Array()
			if len(metricWearArray) == 0 {
				level.Warn(c.logger).Log("msg", "get driver percent endurance remaining data empty", "driver_id")
				return
			}
			wearData := metricWearArray[len(metricWearArray)-1]
			applianceID := wearData.Get("appliance_id").String()
			metricsValue := wearData.Get("percent_endurance_remaining")
			metricDesc := c.metrics["wear"]
			if metricsValue.Exists() && metricsValue.Type != gjson.Null {
				ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricsValue.Float(), driveName, applianceID)
			}
		}(driveID, driveName.String())
	}
	wg.Wait()
	level.Info(c.logger).Log("msg", "Obtaining the driver percent endurance remaining is successful", "time", time.Since(startTime))
}

func (c *metricWearMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getWearMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	res["wear"] = prometheus.NewDesc(
		"powerstore_wear_metrics_by_drive",
		"The percentage of drive wear remaining.",
		[]string{"name", "appliance_id"},
		prometheus.Labels{"IP": ip})
	return res
}
