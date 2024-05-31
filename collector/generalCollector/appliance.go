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

var metricApplianceDescMap = map[string]string{
	"service_tag": "service tag information",
}

type applianceCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewApplianceCollector(api *client.Client, logger log.Logger) *applianceCollector {
	metrics := getApplianceMetrics(api.IP)
	return &applianceCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *applianceCollector) Collect(ch chan<- prometheus.Metric) {
	applianceData, err := c.client.GetAppliance()
	if err != nil {
		level.Warn(c.logger).Log("msg", "get appliance data error", "err", err)
		return
	}
	for _, appliance := range gjson.Parse(applianceData).Array() {
		tag := appliance.Get("service_tag")
		applianceID := appliance.Get("id").String()
		metricDesc := c.metrics["tag"]
		if tag.Exists() && tag.Type != gjson.Null {
			ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, 0, tag.String(), applianceID)
		}
	}
}

func (c *applianceCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getApplianceMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	res["tag"] = prometheus.NewDesc(
		"powerstore_appliance",
		getApplianceDescByType("service_tag"),
		[]string{"service_tag", "appliance_id"},
		prometheus.Labels{"IP": ip})

	return res
}

func getApplianceDescByType(key string) string {
	if v, ok := metricApplianceDescMap[key]; ok {
		return v
	} else {
		return "this is " + key
	}
}
