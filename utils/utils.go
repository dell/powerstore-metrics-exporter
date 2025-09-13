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

package utils

import (
	stdlog "log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

var (
	ReqCounter chan int
)

func InitReqCounter(MaxReq int) {
	ReqCounter = make(chan int, MaxReq)
}

type Storage struct {
	Ip       string `yaml:"ip"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Version  string `yaml:"apiVersion"`
	Limit    int    `yaml:"apiLimit"`
}

type Exporter struct {
	Port     int    `yaml:"port"`
	ReqLimit int    `yaml:"reqLimit"`
	Cert     string `yaml:"cert"`
	Key      string `yaml:"key"`
}

type Logs struct {
	Type  string `yaml:"type"`
	Path  string `yaml:"path"`
	Level string `yaml:"level"`
}

type Config struct {
	Exporter    Exporter  `yaml:"exporter"`
	StorageList []Storage `yaml:"storageList"`
	Log         Logs      `yaml:"log"`
}

func GetConfig(configPath string) *Config {
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		stdlog.Fatalf("Error reading configuration file: %s\n", err)
	}
	expandedyaml := os.ExpandEnv(string(yamlFile))
	config := Config{}
	err = yaml.Unmarshal([]byte(expandedyaml), &config)
	if err != nil {
		stdlog.Fatalf("Error Unmarshal yamL file: %s\n", err)
	}
	return &config
}

func PrometheusHandler(registry *prometheus.Registry, logger log.Logger) gin.HandlerFunc {
	//Check for cert and key file

	handlerOpts := promhttp.HandlerOpts{
		ErrorLog:      stdlog.New(log.NewStdlibAdapter(level.Error(logger)), "", 0),
		ErrorHandling: promhttp.ContinueOnError,
	}
	h := promhttp.HandlerFor(registry, handlerOpts)
	return func(context *gin.Context) {
		h.ServeHTTP(context.Writer, context.Request)
	}
}
