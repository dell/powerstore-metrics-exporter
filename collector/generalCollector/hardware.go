package generalCollector

import (
	"powerstore-metrics-exporter/collector/client"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

var hardwareCollectorType = []string{
	"Drive",
	"Fan",
	"Power_Supply",
	"Battery",
}

var metricHardwareDescMap = map[string]string{
	"size":            "Size of the drive in bytes,unit is B",
	"lifecycle_state": "The lifecycle state of the Hardware,Healthy is 1",
}

var metricHardwareValueMap = map[string]map[string]int{
	"lifecycle_state": {"Healthy": 1, "others": 0},
}

type hardwareCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewHardwareCollector(api *client.Client, logger log.Logger) *hardwareCollector {
	metrics := getHardwareMetrics(api.IP)
	return &hardwareCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *hardwareCollector) Collect(ch chan<- prometheus.Metric) {
	level.Info(c.logger).Log("msg", "Start collecting hardware data")
	startTime := time.Now()
	nodeData, err := c.client.GetHardware("Node")
	if err != nil {
		level.Warn(c.logger).Log("msg", "get node data error", "err", err)
		return
	}
	for _, node := range gjson.Parse(nodeData).Array() {
		id := node.Get("appliance_id").String()
		nodeName := node.Get("name").String()
		sn := node.Get("serial_number").String()
		state := node.Get("lifecycle_state").String()
		if node.Exists() && node.Type != gjson.Null {
			metricDesc := c.metrics["node"+id]
			ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, 0, nodeName, sn, state, id)
		}
	}
	level.Info(c.logger).Log("msg", "Obtaining the node status is successful")
	for _, types := range hardwareCollectorType {
		hardwareData, err := c.client.GetHardware(types)
		if err != nil {
			level.Warn(c.logger).Log("msg", "get hardware data error", "err", err)
		}
		for _, hardware := range gjson.Parse(hardwareData).Array() {
			id := hardware.Get("appliance_id").String()
			name := hardware.Get("name").String()
			state := hardware.Get("lifecycle_state")
			stateValue := getHardwareFloatDate("lifecycle_state", state)
			if state.Exists() && state.Type != gjson.Null {
				metricDesc := c.metrics[types+"state"]
				ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, stateValue, name, id)
			}
			if hardware.Get("type").String() == "Drive" {
				size := hardware.Get("extra_details").Get("size")
				if size.Exists() && size.Type != gjson.Null {
					metricDesc := c.metrics["size"]
					ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, size.Float(), name, id, hardware.Get("extra_details").Get("drive_type").String())
				}
			}
		}
	}
	level.Info(c.logger).Log("msg", "Obtaining the hardware status is successful", "time", time.Since(startTime))
}

func (c *hardwareCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getHardwareFloatDate(key string, value gjson.Result) float64 {
	if v, ok := metricHardwareValueMap[key]; ok {
		if res, ok2 := v[value.String()]; ok2 {
			return float64(res)
		} else {
			return float64(v["other"])
		}
	} else {
		return value.Float()
	}
}

func getHardwareMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}

	res["size"] = prometheus.NewDesc(
		"powerstore_hardware_drive_size",
		getHardwareDescByType("size"),
		[]string{"name", "appliance_id", "drive_type"},
		prometheus.Labels{"IP": ip})

	for _, types := range hardwareCollectorType {
		res[types+"state"] = prometheus.NewDesc(
			"powerstore_hardware_"+types+"_state",
			getHardwareDescByType("lifecycle_state"),
			[]string{"name", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}

	for applianceID, _ := range client.PowerstoreModuleID[ip]["appliance"] {
		res["node"+applianceID] = prometheus.NewDesc(
			"powerstore_hardware_node_state",
			getHardwareDescByType("lifecycle_state"),
			[]string{"name", "serial_number", "state", "appliance_id"},
			prometheus.Labels{"IP": ip})
	}

	return res
}

func getHardwareDescByType(key string) string {
	if v, ok := metricHardwareDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
