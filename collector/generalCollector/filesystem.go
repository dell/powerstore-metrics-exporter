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
