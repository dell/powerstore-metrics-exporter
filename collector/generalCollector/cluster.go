package generalCollector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"powerstore/collector/client"
)

var statuMetricsMap = map[string]map[string]int{
	"cluster_state": {"Configured": 1, "other": 0},
}

var metricClusterDescMap = map[string]string{
	"cluster": "cluster state ,1 is Configured,0 other",
}

type clusterCollector struct {
	client  *client.Client
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

func NewClusterCollector(api *client.Client, logger log.Logger) *clusterCollector {
	metrics := getClusterMetrics(api.IP)
	return &clusterCollector{
		client:  api,
		metrics: metrics,
		logger:  logger,
	}
}

func (c *clusterCollector) Collect(ch chan<- prometheus.Metric) {
	clusterData, err := c.client.GetCluster()
	if err != nil {
		level.Warn(c.logger).Log("msg", "get cluster data error", "err", err)
		return
	}
	for _, cluster := range gjson.Parse(clusterData).Array() {
		stateValue := getFloatData("cluster_state", cluster.Get("state"))
		id := cluster.Get("master_appliance_id").String()
		clusterName := cluster.Get("name").String()
		clusterId := cluster.Get("global_id").String()
		clusterIp := cluster.Get("management_address").String()
		metricDesc := c.metrics["cluster"]
		if cluster.Exists() && cluster.Type != gjson.Null {
			ch <- prometheus.MustNewConstMetric(metricDesc, prometheus.GaugeValue, stateValue, id, clusterId, clusterIp, clusterName)
		}
	}
}

func getFloatData(key string, value gjson.Result) float64 {
	if v, ok := statuMetricsMap[key]; ok {
		if res, ok2 := v[value.String()]; ok2 {
			return float64(res)
		} else {
			return float64(v["other"])
		}
	} else {
		return value.Float()
	}
}

func (c *clusterCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, descMap := range c.metrics {
		ch <- descMap
	}
}

func getClusterMetrics(ip string) map[string]*prometheus.Desc {
	res := map[string]*prometheus.Desc{}
	res["cluster"] = prometheus.NewDesc(
		"powerstore_cluster",
		getClusterDescByType("cluster"),
		[]string{"master_appliance_id", "global_id", "management_address", "name"},
		prometheus.Labels{"IP": ip})
	return res
}

func getClusterDescByType(key string) string {
	if v, ok := metricClusterDescMap[key]; ok {
		return v
	} else {
		return key
	}
}
