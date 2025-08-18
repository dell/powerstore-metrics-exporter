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
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

var statuNasMetricsMap = map[string]map[string]int{
	"operational_status": {"Started": 1, "other": 0},
}

var metricNasDescMap = map[string]string{
	"operational_status": "NAS server operational status,Started is 1 other is 0",
}

type nasCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewNasCollector(api *client.Client, logger log.Logger) *nasCollector {
	metrics := getNasMetrics(api.IP)
	return &nasCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *nasCollector) Collect(ch chan<- prometheus.Metric) {
	level.Info(c.logger).Log("msg", "Start collecting nas server data")
	startTime := time.Now()
	nasData, err := c.client.GetNas()
	if err != nil {
		level.Warn(c.logger).Log("msg", "get Nas data error", "err", err)
		return
	}
	for _, nas := range gjson.Parse(nasData).Array() {
		name := nas.Get("name").String()
		state := nas.Get("operational_status")
		value := getNasFloatData("operational_status", state)
		metricDesc := c.metrics["operational_status"]
		if state.Exists() && state.Type != gjson.Null {
			ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, value, name)
		}
	}
	level.Info(c.logger).Log("msg", "Obtaining the nas server is successful", "time", time.Since(startTime))
}

func getNasFloatData(key string, value gjson.Result) float64 {
	if v, ok := statuNasMetricsMap[key]; ok {
		if res, ok2 := v[value.String()]; ok2 {
			return float64(res)
		} else {
			return float64(v["other"])
		}
	} else {
		return value.Float()
	}
}

func (c *nasCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getNasMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	res["operational_status"] = prometheus.NewDesc(
		"powerstore_nas_server_operational_status",
		getNasDescByType("operational_status"),
		[]string{"name"},
		prometheus.Labels{"IP": ip})
	return res
}

func getNasDescByType(key string) string {
	if v, ok := metricNasDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
