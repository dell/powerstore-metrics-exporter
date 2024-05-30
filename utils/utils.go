package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	stdlog "log"
)

const (
	MaxReq = 50
)

var (
	ReqCounter chan int
)

func init() {
	ReqCounter = make(chan int, MaxReq)
}

type Storage struct {
	Ip       string `yaml:"ip"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Version  string `yaml:"apiVersion"`
}

type Exporter struct {
	Port int `yaml:"port"`
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
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		stdlog.Fatalf("Error reading configuration file: %s\n", err)
	}
	config := Config{}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		stdlog.Fatalf("Error Unmarshal yamL file: %s\n", err)
	}
	return &config
}

func PrometheusHandler(registry *prometheus.Registry, logger log.Logger) gin.HandlerFunc {
	handlerOpts := promhttp.HandlerOpts{
		ErrorLog:      stdlog.New(log.NewStdlibAdapter(level.Error(logger)), "", 0),
		ErrorHandling: promhttp.ContinueOnError,
	}
	h := promhttp.HandlerFor(registry, handlerOpts)
	return func(context *gin.Context) {
		h.ServeHTTP(context.Writer, context.Request)
	}
}
