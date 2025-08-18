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

var metricNasCollectorMetric = []string{
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
	"avg_write_size",
	"avg_read_size",
}

var metricMetricNasDescMap = map[string]string{
	"avg_read_latency":    "Average read latency in microseconds,unit is ms",
	"avg_latency":         "Average read and write latency in microseconds,unit is ms",
	"avg_write_latency":   "Average write latency in microseconds,unit is ms",
	"avg_read_iops":       "Total read operations per second,unit is iops",
	"avg_read_bandwidth":  "Read rate in bytes per second,unit is bps",
	"avg_total_iops":      "Total read and write operations per second,unit is iops",
	"avg_total_bandwidth": "Total data transfer rate in bytes per second,unit is bps",
	"avg_write_iops":      "Total write operations per second,unit is iops",
	"avg_write_bandwidth": "Write rate in bytes per second,unit is bps",
	"avg_size":            "Average size of read and write operations in bytes.unit is bytes",
	"avg_write_size":      "Average write size in bytes.unit is bytes",
	"avg_read_size":       "Average read size in bytes.unit is bytes",
}

type metricNasCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewMetricNasCollector(api *client.Client, logger log.Logger) *metricNasCollector {
	metrics := getMetricNasMetrics(api.IP)
	return &metricNasCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *metricNasCollector) Collect(ch chan<- prometheus.Metric) {
	level.Info(c.logger).Log("msg", "Start collecting nas server performance data")
	startTime := time.Now()
	var wg sync.WaitGroup
	vgArray := client.PowerstoreModuleID[c.client.IP]
	for nasId, nasName := range vgArray["nas"] {
		wg.Add(1)
		go func(nasId, nasName string) {
			defer wg.Done()
			metricNasData, err := c.client.GetMetricByNas(nasId)
			if err != nil {
				level.Warn(c.logger).Log("msg", "get nas server performance data error", "err", err)
				return
			}
			nasDataArray := gjson.Parse(metricNasData).Array()
			if len(nasDataArray) == 0 {
				level.Warn(c.logger).Log("msg", "get nas server performance data is null")
				return
			}

			nasData := nasDataArray[len(nasDataArray)-1]
			for _, metricName := range metricVgCollectorMetric {
				metricValue := nasData.Get(metricName)
				metricDesc := c.metrics["nas"+"_"+metricName]
				if metricValue.Exists() && metricValue.Type != gjson.Null {
					ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), nasName)
				}
			}
		}(nasId, nasName.String())
	}
	wg.Wait()
	level.Info(c.logger).Log("msg", "Obtaining the performance nas server is successful", "time", time.Since(startTime))
}

func (c *metricNasCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getMetricNasMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, metricName := range metricVgCollectorMetric {
		res["nas"+"_"+metricName] = prometheus.NewDesc(
			"powerstore_metricNas_"+metricName,
			getMetricNasDescByType(metricName),
			[]string{"nas_id"},
			prometheus.Labels{"IP": ip})
	}
	return res
}

func getMetricNasDescByType(key string) string {
	if v, ok := metricMetricVgDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
