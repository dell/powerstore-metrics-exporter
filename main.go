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
