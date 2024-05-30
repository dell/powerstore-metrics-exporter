package main

import (
	"flag"
	"github.com/go-kit/log"
	"powerstore/route"
	"powerstore/utils"
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
}

func main() {
	route.Run(config, loggers)
}
