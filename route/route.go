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
	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"powerstore-metrics-exporter/collector/bulkClient"
	"powerstore-metrics-exporter/collector/client"
	"powerstore-metrics-exporter/collector/generalCollector"
	"powerstore-metrics-exporter/utils"
	"strconv"
	"time"
)

func Run(config *utils.Config, logger log.Logger) {
	r := gin.New()
	r.Use(gin.Recovery())
	gin.SetMode(gin.ReleaseMode)
	for _, storage := range config.StorageList {
		client, err := client.NewClient(storage, logger)
		if err != nil {
			level.Error(logger).Log("msg", "init PowerStore client error", "err", err, "ip", storage.Ip)
		}

		var bc = &bulkClient.BulkClient{
			IsEnable: false,
		}
		if storage.Bulk {
			bc.IsEnable = true
			bc, err = bulkClient.NewBulkClient(storage, config.Exporter.BulkDir, logger)
			if err != nil {
				level.Error(logger).Log("msg", "init PowerStore bulk client error", "err", err, "ip", storage.Ip)
			}
			err = bc.BulkEnable()
			if err != nil {
				level.Error(logger).Log("msg", "failed to enable batch request api", "err", err, "ip", storage.Ip)
			}
			// It takes at least one minute to start the API by opening the BULK API
			level.Info(logger).Log("msg", "The bulk api is starting...", "ip", storage.Ip)
			time.Sleep(60 * time.Second)
			level.Info(logger).Log("msg", "The bulk api startup is completed", "ip", storage.Ip)
			// Initialize download bulk api data
			if err := bc.DownloadBulkData(); err != nil {
				level.Error(logger).Log("msg", "error downloading batch data", "err", err, "ip", storage.Ip)
			}
			// Add timed tasks and download bulk API data regularly
			c := cron.New()
			var bulkCron string
			if config.Exporter.BulkCron == "" {
				bulkCron = "*/5 * * * *"
			} else {
				bulkCron = config.Exporter.BulkCron
			}
			_, err = c.AddFunc(bulkCron, func() {
				level.Info(logger).Log("msg", "Start a scheduled task", "ip", storage.Ip)
				if err := bc.DownloadBulkData(); err != nil {
					level.Error(logger).Log("msg", "error downloading batch data", "err", err, "ip", storage.Ip)
				}
			})
			if err != nil {
				level.Error(logger).Log("msg", "task addition failed", "err", err, "ip", storage.Ip)
			} else {
				// Start collecting bulk API data regularly
				c.Start()
			}
		}
		// Initialize the corresponding relationship between each component id and component name
		client.InitModuleID(logger)

		// Generate the registry for each component collector
		ClusterRegistry := prometheus.NewPedanticRegistry()
		PortRegistry := prometheus.NewPedanticRegistry()
		FileSystemRegistry := prometheus.NewPedanticRegistry()
		HardwareRegistry := prometheus.NewPedanticRegistry()
		VolumeRegistry := prometheus.NewPedanticRegistry()
		ApplianceRegistry := prometheus.NewPedanticRegistry()
		NasRegistry := prometheus.NewPedanticRegistry()
		VolumeGroupRegistry := prometheus.NewPedanticRegistry()
		CapacityRegistry := prometheus.NewPedanticRegistry()

		// The collector that registers each component in the registry
		ClusterRegistry.MustRegister(generalCollector.NewClusterCollector(client, logger))
		ClusterRegistry.MustRegister(generalCollector.NewMetroCollector(client, logger))
		PortRegistry.MustRegister(generalCollector.NewPortCollector(client, logger))
		HardwareRegistry.MustRegister(generalCollector.NewHardwareCollector(client, logger))
		VolumeRegistry.MustRegister(generalCollector.NewVolumeCollector(client, logger))
		ApplianceRegistry.MustRegister(generalCollector.NewApplianceCollector(client, logger))
		NasRegistry.MustRegister(generalCollector.NewNasCollector(client, logger))
		VolumeGroupRegistry.MustRegister(generalCollector.NewVolumeGroupCollector(client, logger))
		CapacityRegistry.MustRegister(generalCollector.NewCapacityCollector(client, logger))
		FileSystemRegistry.MustRegister(generalCollector.NewFileCollector(client, bc, logger))
		// Performance data
		ApplianceRegistry.MustRegister(generalCollector.NewMetricApplianceCollector(client, bc, logger))
		PortRegistry.MustRegister(generalCollector.NewMetricFcPortCollector(client, bc, logger))
		PortRegistry.MustRegister(generalCollector.NewMetricEthPortCollector(client, bc, logger))
		NasRegistry.MustRegister(generalCollector.NewMetricNasCollector(client, bc, logger))
		FileSystemRegistry.MustRegister(generalCollector.NewMetricFilesystemCollector(client, bc, logger))
		VolumeRegistry.MustRegister(generalCollector.NewMetricVolumeCollector(client, bc, logger))
		VolumeGroupRegistry.MustRegister(generalCollector.NewMetricVgCollector(client, bc, logger))
		HardwareRegistry.MustRegister(generalCollector.NewWearMetricCollector(client, bc, logger))

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
			metricsGroup.GET("all", utils.PrometheusHandler(prometheus.Gatherers{
				ClusterRegistry,
				PortRegistry,
				FileSystemRegistry,
				HardwareRegistry,
				VolumeRegistry,
				ApplianceRegistry,
				NasRegistry,
				VolumeGroupRegistry,
				CapacityRegistry,
			}, logger))
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
	level.Info(logger).Log("http-port", httpPort, "https", config.Exporter.Https.Enable)
	// Determine whether https is enabled
	if config.Exporter.Https.Enable {
		if config.Exporter.Https.CrtPath == "" || config.Exporter.Https.KeyPath == "" {
			level.Error(logger).Log("msg", "certificate missing", "crt", config.Exporter.Https.CrtPath, "key", config.Exporter.Https.KeyPath)
		}
		err := r.RunTLS(httpPort, config.Exporter.Https.CrtPath, config.Exporter.Https.KeyPath)
		if err != nil {
			level.Error(logger).Log("msg", "Service startup failed", "err", err)
		}
	} else {
		err := r.Run(httpPort)
		if err != nil {
			level.Error(logger).Log("msg", "Service startup failed", "err", err)
		}
	}
}
