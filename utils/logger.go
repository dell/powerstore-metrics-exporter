package utils

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"io"
	"os"
	"strings"
	"time"
)

func GetLogger(loglevel, logPath, logfmt string) log.Logger {
	var out *os.File
	if logPath == "" {
		logPath = "/var/log/Exporter/Exporter.out.log"
	}
	out, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("open log file error: %s\n", err)
	}
	//defer out.Close()
	var logCreator func(io.Writer) log.Logger
	switch strings.ToLower(logfmt) {
	case "json":
		logCreator = log.NewJSONLogger
	case "logfmt":
		logCreator = log.NewLogfmtLogger
	default:
		logCreator = log.NewLogfmtLogger
	}

	nw := io.MultiWriter(os.Stdout, out)
	// create a logger
	logger := logCreator(log.NewSyncWriter(nw))

	// set loglevel
	var loglevelFilterOpt level.Option
	switch strings.ToLower(loglevel) {
	case "debug":
		loglevelFilterOpt = level.AllowDebug()
	case "info":
		loglevelFilterOpt = level.AllowInfo()
	case "warn":
		loglevelFilterOpt = level.AllowWarn()
	case "error":
		loglevelFilterOpt = level.AllowError()
	default:
		loglevelFilterOpt = level.AllowInfo()
	}
	logger = level.NewFilter(logger, loglevelFilterOpt)
	logger = log.With(logger,
		"ts", log.TimestampFormat(time.Now, time.RFC3339),
		"caller", log.DefaultCaller,
	)
	return logger
}
