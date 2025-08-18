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

var volumeGroupCollectorMetrics = []string{
	"logical_provisioned",
	"logical_used",
}

var metricVolumeGroupDescMap = map[string]string{
	"logical_provisioned": "The size of the capacity that has been provisioned by the volume group,unit is B",
	"logical_used":        "Current amount of data (in bytes) host has written to a volume without dedupe, compression or sharing,unit is B",
}

type volumeGroupCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewVolumeGroupCollector(api *client.Client, logger log.Logger) *volumeGroupCollector {
	metrics := getVolumeGroupMetrics(api.IP)
	return &volumeGroupCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *volumeGroupCollector) Collect(ch chan<- prometheus.Metric) {
	level.Info(c.logger).Log("msg", "Start collecting volume group data")
	startTime := time.Now()
	volumeGroupData, err := c.client.GetVolumeGroup()
	if err != nil {
		level.Warn(c.logger).Log("msg", "get volume group data error", "err", err)
		return
	}
	for _, volumeGroup := range gjson.Parse(volumeGroupData).Array() {
		name := volumeGroup.Get("name").String()
		for _, applianceID := range volumeGroup.Get("appliance_ids").Array() {
			for _, metricName := range volumeGroupCollectorMetrics {
				metricValue := volumeGroup.Get(metricName)
				metricDesc := c.metrics[metricName]
				if metricValue.Exists() && metricValue.Type != gjson.Null {
					ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), name, applianceID.String())
				}
			}
		}
	}
	level.Info(c.logger).Log("msg", "Obtaining the volume group is successful", "time", time.Since(startTime))
}

func (c *volumeGroupCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getVolumeGroupMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, metricName := range volumeGroupCollectorMetrics {
		res[metricName] = prometheus.NewDesc(
			"powerstore_volumegroup_"+metricName,
			getVolumeGroupDescByType(metricName),
			[]string{"name", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}
	return res
}

func getVolumeGroupDescByType(key string) string {
	if v, ok := metricVolumeGroupDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
