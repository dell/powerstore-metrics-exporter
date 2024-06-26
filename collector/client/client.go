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
	"bytes"
	"crypto/tls"
	"errors"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"io"
	"net"
	"net/http"
	"powerstore/utils"
	"time"
)

type Client struct {
	IP       string
	username string
	password string
	version  string
	baseUrl  string
	http     *http.Client
	token    string
	cookie   string
	logger   log.Logger
}

func NewClient(config utils.Storage, logger log.Logger) (*Client, error) {
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
	client := &Client{
		IP:       config.Ip,
		username: config.User,
		password: config.Password,
		version:  config.Version,
		baseUrl:  baseUrl,
		http:     httpClient,
		logger:   logger,
	}
	return client, nil
}

func (c *Client) InitLogin() error {
	reqUrl := c.baseUrl + "login_session"
	request, err := http.NewRequest("GET", reqUrl, bytes.NewBuffer([]byte("")))
	if err != nil {
		return err
	}
	request.SetBasicAuth(c.username, c.password)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		level.Warn(c.logger).Log("msg", "Request URL error!")
		return err
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated:
		c.token = response.Header.Get("Dell-Emc-Token")
		cookies := response.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "auth_cookie" {
				c.cookie = cookie.Value
			}
		}
		return nil
	default:
		body, err := io.ReadAll(response.Body)
		level.Warn(c.logger).Log("msg", "get token error", "err", err)
		return errors.New("get token error: " + string(body))
	}
}

func (c *Client) getResource(method, uri, body string) (string, error) {
	reqUrl := c.baseUrl + uri
	request, err := http.NewRequest(method, reqUrl, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("DELL-EMC-TOKEN", c.token)
	request.Header.Set("Cookie", "auth_cookie="+c.cookie)

	response, err := c.http.Do(request)
	if err != nil {
		level.Warn(c.logger).Log("msg", "Request URL error!")
		return "", err
	}

	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusPartialContent:
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return "", errors.New("get resource error: " + string(body))
		}
		return string(body), nil
	case http.StatusUnauthorized, http.StatusFound:
		level.Warn(c.logger).Log("msg", "authentication token is invalid, relogin...", "err", err)
		err = c.InitLogin()
		if err != nil {
			level.Warn(c.logger).Log("msg", "init auth error", "err", err)
			return "", err
		} else {
			return c.getResource(method, uri, body)
		}
	default:
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return "", errors.New("get resource error ReadAll err is not nil: " + string(body))
		}
		return "", errors.New("get resource error ReadAll err is nil: " + string(body))
	}

}
