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
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore-metrics-exporter/collector/client"
	"sync"
	"time"
)

var metricFilesystemCollectorMetric = []string{
	"avg_read_latency",
	"avg_latency",
	"avg_write_latency",
	"avg_read_iops",
	"avg_read_bandwidth",
	"avg_total_iops",
	"avg_total_bandwidth",
	"avg_write_iops",
	"avg_write_bandwidth",
	"avg_size",
	"avg_read_size",
	"avg_write_size",
	"avg_block_write_iops",
	"avg_mirror_write_iops",
	"avg_block_write_bandwidth",
	"avg_mirror_write_bandwidth",
	"avg_block_write_latency",
	"avg_mirror_overhead_latency",
}

var metricMetricFilesystemDescMap = map[string]string{
	"avg_read_latency":            "Average read latency in microseconds,unit is ms",
	"avg_latency":                 "Average read and write latency in microseconds,unit is ms",
	"avg_write_latency":           "Average write latency in microseconds,unit is ms",
	"avg_read_iops":               "Total read operations per second,unit is iops",
	"avg_read_bandwidth":          "Read rate in bytes per second,unit is bps",
	"avg_total_iops":              "Total read and write operations per second,unit is iops",
	"avg_total_bandwidth":         "Total data transfer rate in bytes per second,unit is bps",
	"avg_write_iops":              "Total write operations per second,unit is iops",
	"avg_write_bandwidth":         "Write rate in bytes per second,unit is bps",
	"avg_block_write_iops":        "Total block write operations per second,unit is iops",
	"avg_mirror_write_iops":       "Total mirror write operations per second,unit is iops",
	"avg_block_write_bandwidth":   "Block write rate in byte/sec,unit is bps",
	"avg_mirror_write_bandwidth":  "Mirror write rate in byte/sec,unit is bps",
	"avg_block_write_latency":     "Average block write latency in microsecond,unit is ms",
	"avg_mirror_overhead_latency": "Average additional latency incurred on the source in order to do the remote mirror writes in microseconds,unit is ms",
	"avg_size":                    "Average size of read and write operations in bytes.unit is bytes",
	"avg_write_size":              "Average write size in bytes.unit is bytes",
	"avg_read_size":               "Average read size in bytes.unit is bytes",
}

type metricFilesystemCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewMetricFilesystemCollector(api *client.Client, logger log.Logger) *metricFilesystemCollector {
	metrics := getMetricFilesystemMetrics(api.IP)
	return &metricFilesystemCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *metricFilesystemCollector) Collect(ch chan<- prometheus.Metric) {
	level.Info(c.logger).Log("msg", "Start collecting filesystem performance data")
	startTime := time.Now()
	var wg sync.WaitGroup
	fileSystemArray := client.PowerstoreModuleID[c.client.IP]
	for filesystemId, filesystemName := range fileSystemArray["filesystem"] {
		wg.Add(1)
		go func(filesystemId, filesystemName string) {
			defer wg.Done()
			filesystemData, err := c.client.GetMetricsFilesystem(filesystemId)
			if err != nil {
				level.Warn(c.logger).Log("msg", "get filesystem performance data error", "err", err)
				return
			}
			filesystemArray := gjson.Parse(filesystemData).Array()
			if len(filesystemArray) == 0 {
				level.Warn(c.logger).Log("msg", "get filesystem performance data is null")
				return
			}
			applianceID := filesystemArray[len(filesystemArray)-1].Get("appliance_id").String()
			for _, metricName := range metricFilesystemCollectorMetric {
				metricValue := filesystemArray[len(filesystemArray)-1].Get(metricName)
				metricDesc := c.metrics["filesystem"+"_"+metricName]
				if metricValue.Exists() && metricValue.Type != gjson.Null {
					ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), filesystemName, applianceID)
				}
			}
		}(filesystemId, filesystemName.String())
	}
	wg.Wait()
	level.Info(c.logger).Log("msg", "Obtaining the performance filesystem is successful", "time", time.Since(startTime))
}

func (c *metricFilesystemCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getMetricFilesystemMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}

	for _, metricName := range metricFilesystemCollectorMetric {
		res["filesystem"+"_"+metricName] = prometheus.NewDesc(
			"powerstore_metricFilesystem_"+metricName,
			getMetricFilesystemDescByType(metricName),
			[]string{"name", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}
	return res
}

func getMetricFilesystemDescByType(key string) string {
	if v, ok := metricMetricFcPortDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
