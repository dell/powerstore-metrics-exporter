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

var metricFileSystemCollector = []string{
	"logical_provisioned",
	"logical_used",
	"thin_savings",
}

// file description
var metricFileSystemDescMap = map[string]string{
	"logical_provisioned": "Last logical provisioned space during the period.",
	"logical_used":        "Last logical used space during the period.",
	"thin_savings":        "Last thin savings ratio during the period.",
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
	level.Info(c.logger).Log("msg", "Start collecting filesystem data")
	startTime := time.Now()
	moduleIDArray := client.PowerstoreModuleID[c.client.IP]
	for filesystemID, filesystemName := range moduleIDArray["filesystem"] {
		filesystemData, err := c.client.GetFilesystemCap(filesystemID)
		if err != nil {
			level.Warn(c.logger).Log("msg", "get filesystem data error", "err", err)
			return
		}
		filesystemArray := gjson.Parse(filesystemData).Array()
		if len(filesystemArray) == 0 {
			continue
		}

		id := filesystemArray[len(filesystemArray)-1].Get("appliance_id").String()
		for _, metricName := range metricFileSystemCollector {
			metricValue := filesystemArray[len(filesystemArray)-1].Get(metricName)
			metricDesc := c.metrics["filesystem_"+metricName]
			if metricValue.Exists() && metricValue.Type != gjson.Null {
				ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, metricValue.Float(), filesystemName.String(), id)
			}
		}
	}
	level.Info(c.logger).Log("msg", "Obtaining the filesystem is successful", "time", time.Since(startTime))
}

func (c *fileSystemCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getFileSystemMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	for _, metricName := range metricFileSystemCollector {
		res["filesystem_"+metricName] = prometheus.NewDesc(
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
