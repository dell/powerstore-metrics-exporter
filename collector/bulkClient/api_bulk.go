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

package bulkClient

import (
	"archive/tar"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gocarina/gocsv"
	"github.com/tidwall/gjson"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"powerstore-metrics-exporter/utils"
	"time"
)

var moduleToCsvFileMap = map[string]string{
	"PerformanceMetricsByAppliance":  "performance_metrics_by_appliance.csv",
	"PerformanceMetricsByFeEthPort":  "performance_metrics_by_fe_eth_port.csv",
	"PerformanceMetricsByFeFcPort":   "performance_metrics_by_fe_fc_port.csv",
	"PerformanceMetricsByFileSystem": "performance_metrics_by_file_system.csv",
	"PerformanceMetricsByNasServer":  "performance_metrics_by_nas_server.csv",
	"PerformanceMetricsByVolume":     "performance_metrics_by_volume.csv",
	"PerformanceMetricsByVg":         "performance_metrics_by_vg.csv",
	"SpaceMetricsByAppliance":        "space_metrics_by_appliance.csv",
	"SpaceMetricsByFilesystem":       "space_metrics_by_file_system.csv",
	"WearMetricsByDrive":             "wear_metrics_by_drive.csv",
}

type BulkClient struct {
	IsEnable  bool
	IP        string
	username  string
	password  string
	baseUrl   string
	http      *http.Client
	outputDir string
	logger    log.Logger
}

func NewBulkClient(config utils.Storage, bulkDir string, logger log.Logger) (*BulkClient, error) {
	if config.Ip == "" || config.User == "" || config.Password == "" || config.Version == "" {
		return nil, errors.New("please check config file ,Some parameters are null")
	}
	baseUrl := "https://" + config.Ip + "/api/rest/"
	var httpClient *http.Client
	httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 60 * time.Second,
	}
	client := &BulkClient{
		IsEnable:  config.Bulk,
		IP:        config.Ip,
		username:  config.User,
		password:  config.Password,
		baseUrl:   baseUrl,
		http:      httpClient,
		outputDir: bulkDir,
		logger:    logger,
	}
	return client, nil
}

// BulkEnable Turn on the PowerStore function to obtain performance data in batches
func (bc *BulkClient) BulkEnable() error {
	enableUrl := bc.baseUrl + "latest_five_min_metrics/enable"
	enableReq, err := http.NewRequest("POST", enableUrl, nil)
	if err != nil {
		return err
	}
	enableReq.SetBasicAuth(bc.username, bc.password)
	enableReq.Header.Set("DELL-VISIBILITY", "Partner")

	enableResp, err := bc.http.Do(enableReq)
	if err != nil {
		return errors.New("Request URL error:" + err.Error())
	}
	defer enableResp.Body.Close()

	if enableResp.StatusCode != 204 {
		enableRespBody, err := io.ReadAll(enableResp.Body)
		if err != nil {
			return errors.New("get resource error: " + string(enableRespBody))
		}
		enableRespBodyJson := gjson.Parse(string(enableRespBody))
		msg := enableRespBodyJson.Get("messages").Array()
		if len(msg) > 0 {
			return errors.New("Failed to enable batch acquisition performance data function: " + enableRespBodyJson.Get("messages").Array()[0].Get("message_l10n").String())
		} else {
			return errors.New("Failed to enable batch acquisition performance data function. Status: " + enableResp.Status)
		}
	}
	level.Info(bc.logger).Log("msg", "Successfully enabled the batch acquisition performance data function")
	return nil
}

// DownloadBulkData call batch to get data api and save files
func (bc *BulkClient) DownloadBulkData() error {
	level.Info(bc.logger).Log("msg", "download data to start", "time", time.Now().Format("2006-01-02 15:04:05"))
	enableUrl := bc.baseUrl + "latest_five_min_metrics/download"
	enableReq, err := http.NewRequest("POST", enableUrl, nil)
	if err != nil {
		return err
	}
	enableReq.SetBasicAuth(bc.username, bc.password)
	enableReq.Header.Set("DELL-VISIBILITY", "Partner")
	enableReq.Header.Set("If-None-Match", "start")

	enableResp, err := bc.http.Do(enableReq)
	if err != nil {
		return errors.New("Request URL error:" + err.Error())
	}
	defer enableResp.Body.Close()
	var outputFilePath string
	outputFilePath = filepath.Join(bc.outputDir, fmt.Sprintf("pst_bulk_%s.tar.gz", bc.IP))
	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return errors.New("Creating output file failed: " + err.Error())
	}
	defer outFile.Close()

	bytesWritten, err := io.Copy(outFile, enableResp.Body)
	if err != nil {
		return errors.New("Failed to write response content to file: " + err.Error())
	}
	level.Info(bc.logger).Log("msg", "Successfully written bulk api data to file", "size", bytesWritten, "path", outputFilePath)
	return nil
}

func (bc *BulkClient) ReadCsvData(moduleType string) (string, error) {
	var modelFilename string
	if filename, ok := moduleToCsvFileMap[moduleType]; !ok {
		return "", errors.New("model type error")
	} else {
		modelFilename = filename
	}
	var bulkFilepath = filepath.Join(bc.outputDir, fmt.Sprintf("pst_bulk_%s.tar.gz", bc.IP))
	level.Info(bc.logger).Log("msg", "Start reading bulk api result file", "filepath", bulkFilepath, "type", moduleType)
	file, err := os.Open(bulkFilepath)
	if err != nil {
		return "", errors.New("open tar file error:" + err.Error())
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", errors.New("gzip reader error: " + err.Error())
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			level.Info(bc.logger).Log("msg", "Failed to read tar header", "err", err)
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(header.Name) != modelFilename {
			continue
		}
		switch modelFilename {
		case "performance_metrics_by_appliance.csv":
			var records []*PerformanceMetricsByAppliance
			if err := gocsv.Unmarshal(tr, &records); err != nil {
				return "", errors.New("error parsing csv file: " + err.Error())
			}
			recordsJson, err := json.Marshal(records)
			if err != nil {
				return "", errors.New("error json marshal: " + err.Error())
			}
			return string(recordsJson), nil
		case "performance_metrics_by_fe_eth_port.csv":
			var records []*PerformanceMetricsByFeEthPort
			if err := gocsv.Unmarshal(tr, &records); err != nil {
				return "", errors.New("error parsing csv file: " + err.Error())
			}
			recordsJson, err := json.Marshal(records)
			if err != nil {
				return "", errors.New("error json marshal: " + err.Error())
			}
			return string(recordsJson), nil
		case "performance_metrics_by_fe_fc_port.csv":
			var records []*PerformanceMetricsByFeFcPort
			if err := gocsv.Unmarshal(tr, &records); err != nil {
				return "", errors.New("error parsing csv file: " + err.Error())
			}
			recordsJson, err := json.Marshal(records)
			if err != nil {
				return "", errors.New("error json marshal: " + err.Error())
			}
			return string(recordsJson), nil
		case "performance_metrics_by_file_system.csv":
			var records []*PerformanceMetricsByFileSystem
			if err := gocsv.Unmarshal(tr, &records); err != nil {
				return "", errors.New("error parsing csv file: " + err.Error())
			}
			recordsJson, err := json.Marshal(records)
			if err != nil {
				return "", errors.New("error json marshal: " + err.Error())
			}
			return string(recordsJson), nil
		case "performance_metrics_by_nas_server.csv":
			var records []*PerformanceMetricsByNasServer
			if err := gocsv.Unmarshal(tr, &records); err != nil {
				return "", errors.New("error parsing csv file: " + err.Error())
			}
			recordsJson, err := json.Marshal(records)
			if err != nil {
				return "", errors.New("error json marshal: " + err.Error())
			}
			return string(recordsJson), nil
		case "performance_metrics_by_volume.csv":
			var records []*PerformanceMetricsByVolume
			if err := gocsv.Unmarshal(tr, &records); err != nil {
				return "", errors.New("error parsing csv file: " + err.Error())
			}
			recordsJson, err := json.Marshal(records)
			if err != nil {
				return "", errors.New("error json marshal: " + err.Error())
			}
			return string(recordsJson), nil
		case "performance_metrics_by_vg.csv":
			var records []*PerformanceMetricsByVg
			if err := gocsv.Unmarshal(tr, &records); err != nil {
				return "", errors.New("error parsing csv file: " + err.Error())
			}
			recordsJson, err := json.Marshal(records)
			if err != nil {
				return "", errors.New("error json marshal: " + err.Error())
			}
			return string(recordsJson), nil
		case "space_metrics_by_appliance.csv":
			var records []*SpaceMetricsByAppliance
			if err := gocsv.Unmarshal(tr, &records); err != nil {
				return "", errors.New("error parsing csv file: " + err.Error())
			}
			recordsJson, err := json.Marshal(records)
			if err != nil {
				return "", errors.New("error json marshal: " + err.Error())
			}
			return string(recordsJson), nil
		case "space_metrics_by_file_system.csv":
			var records []*SpaceMetricsByFilesystem
			if err := gocsv.Unmarshal(tr, &records); err != nil {
				return "", errors.New("error parsing csv file: " + err.Error())
			}
			recordsJson, err := json.Marshal(records)
			if err != nil {
				return "", errors.New("error json marshal: " + err.Error())
			}
			return string(recordsJson), nil
		case "wear_metrics_by_drive.csv":
			var records []*WearMetricsByDrive
			if err := gocsv.Unmarshal(tr, &records); err != nil {
				return "", errors.New("error parsing csv file: " + err.Error())
			}
			recordsJson, err := json.Marshal(records)
			if err != nil {
				return "", errors.New("error json marshal: " + err.Error())
			}
			return string(recordsJson), nil
		default:
			return "", errors.New("model type error")
		}
	}
	return "", errors.New("get data empty")
}
