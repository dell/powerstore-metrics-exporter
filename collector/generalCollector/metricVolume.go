/*
 Copyright (c) 2024-2025 Dell Inc. or its subsidiaries. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

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
	"avg_io_size",
	"avg_read_size",
	"avg_write_size",
}

var metricMetricVolumeDescMap = map[string]string{
	"avg_read_latency":    "Average read latency in microseconds,unit is ms",
	"avg_latency":         "Average read and write latency in microseconds,unit is ms",
	"avg_write_latency":   "Average write latency in microseconds,unit is ms",
	"avg_read_iops":       "Total read operations per second,unit is iops",
	"avg_read_bandwidth":  "Read rate in bytes per second,unit is bps",
	"avg_total_iops":      "Total read and write operations per second,unit is iops",
	"avg_total_bandwidth": "Total data transfer rate in bytes per second,unit is bps",
	"avg_write_iops":      "Total write operations per second,unit is iops",
	"avg_write_bandwidth": "Write rate in bytes per second,unit is bps",
	"avg_io_size":         "Average size of read and write operations in bytes.unit is bytes",
	"avg_write_size":      "Average write size in bytes.unit is bytes",
	"avg_read_size":       "Average read size in bytes.unit is bytes",
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
	level.Info(c.logger).Log("msg", "Start collecting volume performance data")
	startTime := time.Now()
	var wg sync.WaitGroup
	volumeArray := client.PowerstoreModuleID[c.client.IP]
	for volumeId, volumeName := range volumeArray["volume"] {
		wg.Add(1)
		go func(volumeId, volumeName string) {
			defer wg.Done()
			metricVolData, err := c.client.GetMetricVolume(volumeId)
			if err != nil {
				level.Warn(c.logger).Log("msg", "get volume performance data error", "err", err)
				return
			}
			volumeDataArray := gjson.Parse(metricVolData).Array()
			if len(volumeDataArray) == 0 {
				level.Warn(c.logger).Log("msg", "get volume performance data is null")
				return
			}
			volumeData := volumeDataArray[len(volumeDataArray)-1]
			applianceID := volumeData.Get("appliance_id").String()
			for _, metricName := range metricVolumeCollectorMetric {
				metricValue := volumeData.Get(metricName)
				metricDesc := c.metrics["volume"+"_"+metricName]
				if metricValue.Exists() && metricValue.Type != gjson.Null {
					ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), volumeName, applianceID)
				}
			}
		}(volumeId, volumeName.String())
	}
	wg.Wait()
	level.Info(c.logger).Log("msg", "Obtaining the performance volume is successful", "time", time.Since(startTime))
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
