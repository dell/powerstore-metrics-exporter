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

package route

import (
	"fmt"
	"powerstore-metrics-exporter/collector/client"
	"powerstore-metrics-exporter/collector/generalCollector"
	"powerstore-metrics-exporter/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run(config *utils.Config, logger log.Logger) {
	r := gin.New()
	r.Use(gin.Recovery())
	gin.SetMode(gin.ReleaseMode)
	for _, storage := range config.StorageList {
		client, err := client.NewClient(storage, logger)
		if err != nil {
			level.Error(logger).Log("msg", "init Powerstore client error", "err", err, "ip", storage.Ip)
		}

		client.InitModuleID(logger)

		ClusterRegistry := prometheus.NewPedanticRegistry()
		PortRegistry := prometheus.NewPedanticRegistry()
		FileSystemRegistry := prometheus.NewPedanticRegistry()
		HardwareRegistry := prometheus.NewPedanticRegistry()
		VolumeRegistry := prometheus.NewPedanticRegistry()
		ApplianceRegistry := prometheus.NewPedanticRegistry()
		NasRegistry := prometheus.NewPedanticRegistry()
		VolumeGroupRegistry := prometheus.NewPedanticRegistry()
		CapacityRegistry := prometheus.NewPedanticRegistry()

		ClusterRegistry.MustRegister(generalCollector.NewClusterCollector(client, logger))
		PortRegistry.MustRegister(generalCollector.NewPortCollector(client, logger))
		PortRegistry.MustRegister(generalCollector.NewMetricFcPortCollector(client, logger))
		PortRegistry.MustRegister(generalCollector.NewMetricEthPortCollector(client, logger))
		FileSystemRegistry.MustRegister(generalCollector.NewFileCollector(client, logger))
		FileSystemRegistry.MustRegister(generalCollector.NewMetricFilesystemCollector(client, logger))
		HardwareRegistry.MustRegister(generalCollector.NewHardwareCollector(client, logger))
		HardwareRegistry.MustRegister(generalCollector.NewWearMetricCollector(client, logger))
		VolumeRegistry.MustRegister(generalCollector.NewVolumeCollector(client, logger))
		VolumeRegistry.MustRegister(generalCollector.NewMetricVolumeCollector(client, logger))
		ApplianceRegistry.MustRegister(generalCollector.NewApplianceCollector(client, logger))
		ApplianceRegistry.MustRegister(generalCollector.NewMetricApplianceCollector(client, logger))
		NasRegistry.MustRegister(generalCollector.NewNasCollector(client, logger))
		NasRegistry.MustRegister(generalCollector.NewMetricNasCollector(client, logger))
		VolumeGroupRegistry.MustRegister(generalCollector.NewVolumeGroupCollector(client, logger))
		VolumeGroupRegistry.MustRegister(generalCollector.NewMetricVgCollector(client, logger))
		CapacityRegistry.MustRegister(generalCollector.NewCapacityCollector(client, logger))

		metricsGroup := r.Group(fmt.Sprintf("/metrics/%s", storage.Ip))
		{
			metricsGroup.GET("cluster", utils.PrometheusHandler(ClusterRegistry, logger))
			metricsGroup.GET("port", utils.PrometheusHandler(PortRegistry, logger))
			metricsGroup.GET("file", utils.PrometheusHandler(FileSystemRegistry, logger))
			metricsGroup.GET("hardware", utils.PrometheusHandler(HardwareRegistry, logger))
			metricsGroup.GET("volume", utils.PrometheusHandler(VolumeRegistry, logger))
			metricsGroup.GET("appliance", utils.PrometheusHandler(ApplianceRegistry, logger))
			metricsGroup.GET("nas", utils.PrometheusHandler(NasRegistry, logger))
			metricsGroup.GET("volumeGroup", utils.PrometheusHandler(VolumeGroupRegistry, logger))
			metricsGroup.GET("capacity", utils.PrometheusHandler(CapacityRegistry, logger))
		}
		level.Info(logger).Log("msg", "The Powerstore is ready", "ip", storage.Ip)
	}

	// exporter Performance
	r.GET("/performance", func(context *gin.Context) {
		h := promhttp.Handler()
		h.ServeHTTP(context.Writer, context.Request)
	})

	httpPort := fmt.Sprintf(":%s", strconv.Itoa(config.Exporter.Port))
	level.Info(logger).Log("msg", "~~~~~~~~~~~~~Start PowerStore Exporter~~~~~~~~~~~~~~")
	level.Info(logger).Log("http-port", httpPort)

	// Start HTTP server with or without TLS based on the presence of cert and key
	if config.Exporter.Cert == "" || config.Exporter.Key == "" {
		level.Info(logger).Log("msg", "Starting HTTP server without TLS")
		err := r.Run(httpPort)
		if err != nil {
			level.Error(logger).Log("msg", "Service startup failed", "err", err)
		}
	} else {
		level.Info(logger).Log("msg", "Starting HTTP server with TLS")
		// Start HTTPS server
		err := r.RunTLS(httpPort, config.Exporter.Cert, config.Exporter.Key)
		if err != nil {
			level.Error(logger).Log("msg", "Service startup failed", "err", err)
		}
	}
}
