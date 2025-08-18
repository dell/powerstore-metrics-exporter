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

package client

import (
	"encoding/json"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/tidwall/gjson"
	"powerstore-metrics-exporter/utils"
	"strconv"
)

type RequestBody struct {
	Entity   string `json:"entity"`
	EntityID string `json:"entity_id"`
	Interval string `json:"interval"`
}

// PowerstoreModuleID This map stores the mapping relationships of the ip, module type, module id, and module name of the powerstore
var PowerstoreModuleID = make(map[string]map[string]map[string]gjson.Result)

func (c *Client) getData(path, method, body string) (string, error) {
	utils.ReqCounter <- 1
	result, err := c.getResource(method, path, body)
	<-utils.ReqCounter
	return result, err
}

func (c *Client) GetCluster() (string, error) {
	return c.getData("cluster?select=*&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetPort(portType string) (string, error) {
	return c.getData(portType+"?select=*&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetHardware(hardwareType string) (string, error) {
	return c.getData("hardware?select=*&type=eq."+hardwareType+"&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetVolume() (string, error) {
	if c.version == "v3" {
		return c.getData("volume_list_cma_view?select=*&limit="+strconv.Itoa(c.limit), "GET", "")
	}
	return c.getData("volume?select=*&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetAppliance() (string, error) {
	return c.getData("appliance?select=*&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetNas() (string, error) {
	return c.getData("nas_server?select=*&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetNasDetail() (string, error) {
	return c.getData("nas_server_list_cma_view?select=*&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetVolumeGroup() (string, error) {
	return c.getData("volume_group_list_cma_view?select=*&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetPerf(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "performance_metrics_by_appliance",
		EntityID: id,
		Interval: "Five_Mins",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetCap(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "space_metrics_by_appliance",
		EntityID: id,
		Interval: "One_Day",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetMetricVg(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "performance_metrics_by_vg",
		EntityID: id,
		Interval: "Five_Mins",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetMetricVolume(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "performance_metrics_by_volume",
		EntityID: id,
		Interval: "Five_Mins",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetMetricFcPort(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "performance_metrics_by_fe_fc_port",
		EntityID: id,
		Interval: "Five_Mins",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetMetricEthPort(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "performance_metrics_by_fe_eth_port",
		EntityID: id,
		Interval: "Five_Mins",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetMetricAppliance(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "performance_metrics_by_appliance",
		EntityID: id,
		Interval: "Five_Mins",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetWearMetricByDrive(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "wear_metrics_by_drive",
		EntityID: id,
		Interval: "Five_Mins",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetMetricByNas(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "performance_metrics_by_nas_server",
		EntityID: id,
		Interval: "Five_Mins",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetFilesystemCap(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "space_metrics_by_file_system",
		EntityID: id,
		Interval: "Five_Mins",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetMetricsFilesystem(id string) (string, error) {
	var body = &RequestBody{
		Entity:   "performance_metrics_by_file_system",
		EntityID: id,
		Interval: "Five_Mins",
	}
	entityBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return c.getData("metrics/generate", "POST", string(entityBody))
}

func (c *Client) GetApplianceId() (string, error) {
	return c.getData("appliance?select=id,name&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetVolumeGroupId() (string, error) {
	return c.getData("volume_group_list_cma_view?select=id,name,appliance_ids&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetVolumeId() (string, error) {
	if c.version == "v3" {
		return c.getData("volume_list_cma_view?select=id,name&limit="+strconv.Itoa(c.limit), "GET", "")
	}
	return c.getData("volume?select=id,name&type=eq.Drive&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetEthPortId() (string, error) {
	return c.getData("eth_port?select=id,name&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetFcPortId() (string, error) {
	return c.getData("fc_port?select=id,name&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetDrivesId() (string, error) {
	return c.getData("hardware?select=id,name&type=eq.Drive&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetNasId() (string, error) {
	return c.getData("nas_server_list_cma_view?select=id,name&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) GetFilesystemId() (string, error) {
	return c.getData("file_system?select=id,name&limit="+strconv.Itoa(c.limit), "GET", "")
}

func (c *Client) InitModuleID(logger log.Logger) {
	ModuleIdToNameMap := make(map[string]map[string]gjson.Result)
	applianceIdToName, err := c.GetApplianceId()
	if err != nil {
		level.Error(logger).Log("msg", "Init appliance id list error", "err", err, "ip", c.IP)
	}
	ModuleIdToNameMap["appliance"] = resultToMap(applianceIdToName)

	volumeIdToName, err := c.GetVolumeId()
	if err != nil {
		level.Error(logger).Log("msg", "Init volume id list error", "err", err, "ip", c.IP)
	}
	ModuleIdToNameMap["volume"] = resultToMap(volumeIdToName)

	volumeGroupIdToName, err := c.GetVolumeGroupId()
	if err != nil {
		level.Error(logger).Log("msg", "Init volume group id list error", "err", err, "ip", c.IP)
	}
	ModuleIdToNameMap["volumegroup"] = resultToMap(volumeGroupIdToName)

	ethPortIdToName, err := c.GetEthPortId()
	if err != nil {
		level.Error(logger).Log("msg", "Init eth port id list error", "err", err, "ip", c.IP)
	}
	ModuleIdToNameMap["ethport"] = resultToMap(ethPortIdToName)

	fcPortIdToName, err := c.GetFcPortId()
	if err != nil {
		level.Error(logger).Log("msg", "Init fc port id list error", "err", err, "ip", c.IP)
	}
	ModuleIdToNameMap["fcport"] = resultToMap(fcPortIdToName)

	drivesIdToName, err := c.GetDrivesId()
	if err != nil {
		level.Error(logger).Log("msg", "Init drives id list error", "err", err, "ip", c.IP)
	}
	ModuleIdToNameMap["drive"] = resultToMap(drivesIdToName)

	nasIdToName, err := c.GetNasId()
	if err != nil {
		level.Error(logger).Log("msg", "Init nas server id list error", "err", err, "ip", c.IP)
	}
	ModuleIdToNameMap["nas"] = resultToMap(nasIdToName)

	filesystemIdToName, err := c.GetFilesystemId()
	if err != nil {
		level.Error(logger).Log("msg", "Init filesystem server id list error", "err", err, "ip", c.IP)
	}
	ModuleIdToNameMap["filesystem"] = resultToMap(filesystemIdToName)
	PowerstoreModuleID[c.IP] = ModuleIdToNameMap
}

// resultToMap Convert http response body to map structure
func resultToMap(result string) map[string]gjson.Result {
	var resultMap = make(map[string]gjson.Result)
	for _, entity := range gjson.Parse(result).Array() {
		resultMap[entity.Get("id").String()] = entity.Get("name")
	}
	return resultMap
}
