/*
 Copyright (c) 2023-2024 Dell Inc. or its subsidiaries. All Rights Reserved.

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
	"powerstore/utils"
)

type RequestBody struct {
	Entity   string `json:"entity"`
	EntityID string `json:"entity_id"`
	Interval string `json:"interval"`
}

var PowerstoreModuleID = map[string]map[string]string{}

func (c *Client) getData(path, method, body string) (string, error) {
	utils.ReqCounter <- 1
	result, err := c.getResource(method, path, body)
	<-utils.ReqCounter
	return result, err
}

func (c *Client) GetCluster() (string, error) {
	return c.getData("cluster?select=*", "GET", "")
}

func (c *Client) GetPort(portType string) (string, error) {
	return c.getData(portType+"?select=*", "GET", "")
}

func (c *Client) GetHardware(hardwareType string) (string, error) {
	return c.getData("hardware?select=*&type=eq."+hardwareType, "GET", "")
}

func (c *Client) GetVolume() (string, error) {
	if c.version == "v3" {
		return c.getData("volume_list_cma_view?select=*", "GET", "")
	}
	return c.getData("volume?select=*", "GET", "")
}

func (c *Client) GetAppliance() (string, error) {
	return c.getData("appliance?select=*", "GET", "")
}

func (c *Client) GetFile() (string, error) {
	return c.getData("file_system?select=*", "GET", "")
}

func (c *Client) GetNas() (string, error) {
	return c.getData("nas_server?select=*", "GET", "")
}

func (c *Client) GetVolumeGroup() (string, error) {
	return c.getData("volume_group_list_cma_view?select=*", "GET", "")
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
		Interval: "One_Hour",
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

func (c *Client) GetApplianceId() (string, error) {
	return c.getData("appliance?select=id,name", "GET", "")
}

func (c *Client) GetVolumeGroupId() (string, error) {
	return c.getData("volume_group_list_cma_view?select=id,name,appliance_ids", "GET", "")
}

func (c *Client) GetVolumeId() (string, error) {
	if c.version == "v3" {
		return c.getData("volume_list_cma_view?select=id,name", "GET", "")
	}
	return c.getData("volume?select=id,name", "GET", "")
}

func (c *Client) GetEthPortId() (string, error) {
	return c.getData("eth_port?select=id,name", "GET", "")
}

func (c *Client) GetFcPortId() (string, error) {
	return c.getData("fc_port?select=id,name", "GET", "")
}

func (c *Client) GetDrivesId() (string, error) {
	return c.getData("hardware?select=id,name", "GET", "")
}

func (c *Client) InitModuleID(logger log.Logger) {
	ModuleIDMap := make(map[string]string)
	ApplianceId, err := c.GetApplianceId()
	if err != nil {
		level.Error(logger).Log("msg", "Init appliance id list error", "err", err, "ip", c.IP)
	}
	ModuleIDMap["appliance"] = ApplianceId
	VolumeId, err := c.GetVolumeId()
	if err != nil {
		level.Error(logger).Log("msg", "Init volume id list error", "err", err, "ip", c.IP)
	}
	ModuleIDMap["volume"] = VolumeId
	VolumeGroupId, err := c.GetVolumeGroupId()
	if err != nil {
		level.Error(logger).Log("msg", "Init volume group id list error", "err", err, "ip", c.IP)
	}
	ModuleIDMap["volumegroup"] = VolumeGroupId
	EthPortId, err := c.GetEthPortId()
	if err != nil {
		level.Error(logger).Log("msg", "Init eth port id list error", "err", err, "ip", c.IP)
	}
	ModuleIDMap["ethport"] = EthPortId
	FcPortId, err := c.GetFcPortId()
	if err != nil {
		level.Error(logger).Log("msg", "Init fc port id list error", "err", err, "ip", c.IP)
	}
	ModuleIDMap["fcport"] = FcPortId
	DrivesId, err := c.GetDrivesId()
	if err != nil {
		level.Error(logger).Log("msg", "Init drives id list error", "err", err, "ip", c.IP)
	}
	ModuleIDMap["drive"] = DrivesId
	PowerstoreModuleID[c.IP] = ModuleIDMap
}
