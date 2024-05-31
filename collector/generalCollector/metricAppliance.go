/*
 Copyright (c) 2023-2024 Dell Inc. or its subsidiaries. All Rights Reserved.

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
	"powerstore/collector/client"
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
}

// performance description
var metricAppliancePerfDescMap = map[string]string{
	"avg_read_latency":                "avg latency of read , unit is ms",
	"avg_latency":                     "avg latency , unit is ms",
	"avg_write_latency":               "avg latency of write , unit is ms",
	"avg_read_iops":                   "iops of read , unit is iops",
	"avg_read_bandwidth":              "throughput of read , unit is bps",
	"avg_total_iops":                  "iops total , unit is iops",
	"avg_total_bandwidth":             "total throughput , unit is bps",
	"avg_write_iops":                  "iops of write , unit is iops",
	"avg_write_bandwidth":             "throughput of write , unit is bps",
	"avg_io_workload_cpu_utilization": "usage of CPU for IO workload ",
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
	applianceArray := client.PowerstoreModuleID[c.client.IP]
	for _, applianceID := range gjson.Parse(applianceArray["appliance"]).Array() {
		id := applianceID.Get("id").String()
		perfData, err := c.client.GetPerf(id)
		if err != nil {
			level.Warn(c.logger).Log("msg", "get appliance performance data error", "err", err)
			continue
		}
		appliancePerformanceArray := gjson.Parse(perfData).Array()
		appliancePerformance := appliancePerformanceArray[len(appliancePerformanceArray)-1]
		for _, metricName := range metricAppliancePerfCollectorMetric {
			metricValue := appliancePerformance.Get(metricName)
			metricDesc := c.metrics["appliance"+"_"+metricName]
			if metricValue.Exists() && metricValue.Type != gjson.Null {
				ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), id)
			}
		}
	}

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
			[]string{"appliance_id"},
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
