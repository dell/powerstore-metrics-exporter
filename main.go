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

package main

import (
	"flag"
	"powerstore-metrics-exporter/route"
	"powerstore-metrics-exporter/utils"

	"github.com/go-kit/log"
)

var (
	loggers    log.Logger
	config     *utils.Config
	configPath string
)

func init() {
	flag.StringVar(&configPath, "c", "config.yml", "powerstore exporter configuration file path")
	flag.Parse()
	config = utils.GetConfig(configPath)
	loggers = utils.GetLogger(config.Log.Level, config.Log.Path, config.Log.Type)
	utils.InitReqCounter(config.Exporter.ReqLimit)
}

func main() {
	route.Run(config, loggers)
}
