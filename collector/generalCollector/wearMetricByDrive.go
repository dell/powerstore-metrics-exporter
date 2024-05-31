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
	driveArray := client.PowerstoreModuleID[c.client.IP]
	for _, driveID := range gjson.Parse(driveArray["drive"]).Array() {
		id := driveID.Get("id").String()
		name := driveID.Get("name").String()
		wearMetricData, err := c.client.GetWearMetricByDrive(id)
		if err != nil {
			level.Warn(c.logger).Log("msg", "get disk performance data error", "err", err)
			continue
		}
		metricWearArray := gjson.Parse(wearMetricData).Array()
		if len(metricWearArray) == 0 {
			continue
		}
		wearData := metricWearArray[len(metricWearArray)-1]
		applianceID := wearData.Get("appliance_id").String()
		metricsValue := wearData.Get("percent_endurance_remaining")
		metricDesc := c.metrics["wear"]
		if metricsValue.Exists() && metricsValue.Type != gjson.Null {
			ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricsValue.Float(), name, applianceID)
		}
	}

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
		"this is the percent of endurance remaining about drives",
		[]string{"name", "appliance_id"},
		prometheus.Labels{"IP": ip})
	return res
}
