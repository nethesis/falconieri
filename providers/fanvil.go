/*
 * Copyright (C) 2019 Nethesis S.r.l.
 * http://www.nethesis.it - info@nethesis.it
 *
 * This file is part of Falconieri project.
 *
 * Falconieri is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License,
 * or any later version.
 *
 * Falconieri is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Falconieri.  If not, see LICENSE.
 *
 * author: Matteo Valentini <matteo.valentini@nethesis.it>
 */

package providers

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/divan/gorilla-xmlrpc/xml"

	"github.com/nethesis/falconieri/configuration"
	"github.com/nethesis/falconieri/models"
	"github.com/nethesis/falconieri/utils"
)

type FanvilDevice struct {
	Mac string
	Url string
}

type fanvilDeviceLock struct {
	sync.Mutex
	users int
}

var (
	fanvilPassword      string
	fanvilPasswordMutex sync.Mutex
	fanvilHTTPClient    = &http.Client{Timeout: 30 * time.Second}
	fanvilDeviceLocks   = struct {
		sync.Mutex
		devices map[string]*fanvilDeviceLock
	}{devices: make(map[string]*fanvilDeviceLock)}
)

// fanvilLockDevice prevents requests for the same phone from interleaving.
func fanvilLockDevice(mac string) func() {
	fanvilDeviceLocks.Lock()
	lock := fanvilDeviceLocks.devices[mac]
	if lock == nil {
		lock = &fanvilDeviceLock{}
		fanvilDeviceLocks.devices[mac] = lock
	}
	lock.users++
	fanvilDeviceLocks.Unlock()

	lock.Lock()
	return func() {
		lock.Unlock()
		fanvilDeviceLocks.Lock()
		defer fanvilDeviceLocks.Unlock()
		lock.users--
		if lock.users == 0 {
			delete(fanvilDeviceLocks.devices, mac)
		}
	}
}

func (d FanvilDevice) Register() error {
	Url, err := url.Parse(d.Url)
	if err != nil || Url.Hostname() == "" {
		return errors.New("malformed_url")
	}

	var UrlScheme string
	switch Url.Scheme {
	case "ftp":
		UrlScheme = "1"
	case "tftp":
		UrlScheme = "2"
	case "http":
		UrlScheme = "4"
	case "https":
		UrlScheme = "5"
	default:
		return errors.New("malformed_url")
	}

	unlock := fanvilLockDevice(d.Mac)
	defer unlock()

	fanvilPasswordMutex.Lock()
	password := fanvilGetPassword()
	fanvilPasswordMutex.Unlock()

	// Fanvil refuses to delete a server while a device is still registered to it.
	buf, _ := xml.EncodeClientRequest("redirect.deRegisterDevice",
		&struct {
			Mac string
		}{Mac: d.Mac})

	req, _ := http.NewRequest("POST", configuration.Config.Providers.Fanvil.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Fanvil.User, password)

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", " Falconieri/1")

	resp, err := fanvilHTTPClient.Do(req)

	if err != nil {
		return models.ProviderError{
			Message:      "connection_to_remote_provider_failed",
			WrappedError: err,
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("provider_remote_call_failed")
	}

	err = fanvilParseResponse(resp.Body)
	if err != nil {
		cause := errors.Unwrap(err)
		// A device that has never been registered needs no cleanup.
		if cause == nil || cause.Error() != "Error:device_not_exist" {
			return err
		}
	}

	//Delete old server
	buf, _ = xml.EncodeClientRequest("redirect.deleteServer",
		&struct {
			GroupName string
		}{GroupName: d.Mac})

	req, _ = http.NewRequest("POST", configuration.Config.Providers.Fanvil.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Fanvil.User, password)

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", " Falconieri/1")

	resp, err = fanvilHTTPClient.Do(req)

	if err != nil {
		return models.ProviderError{
			Message:      "connection_to_remote_provider_failed",
			WrappedError: err,
		}

	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("provider_remote_call_failed")
	}

	err = fanvilParseResponse(resp.Body)

	if err != nil {
		cause := errors.Unwrap(err)
		// An absent server is expected when registering a device for the first time.
		if cause == nil || (cause.Error() != "Error:server_not_exist" && cause.Error() != "Error:server_not_existed") {
			return err
		}
	}

	/* Server/Group configuration:
	 *   - Disable DHCP provision: cfgDhcpOpt=false
	 *   - Disable PnP provision: cfgPnpEnable=false
	 *   - Apply the configuration after the reboot: cfgPfMode=1
	 */

	buf, _ = xml.EncodeClientRequest("redirect.addMaterialServer",
		&struct {
			ServerConfigs []string
		}{ServerConfigs: []string{"cfgName=" + d.Mac, "cfgPfMode=1",
			"cfgDhcpOpt=false", "cfgPnpEnable=false",
			"cfgPfSrv=" + Url.Host, "cfgPfName=" + strings.TrimPrefix(Url.Path, "/"),
			"cfgPfProt=" + UrlScheme}})

	req, _ = http.NewRequest("POST", configuration.Config.Providers.Fanvil.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Fanvil.User, password)

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", " Falconieri/1")

	resp, err = fanvilHTTPClient.Do(req)

	if err != nil {
		return models.ProviderError{
			Message:      "connection_to_remote_provider_failed",
			WrappedError: err,
		}

	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("provider_remote_call_failed")
	}

	err = fanvilParseResponse(resp.Body)

	if err != nil {
		return err
	}

	//Register the device
	buf, _ = xml.EncodeClientRequest("redirect.registerDevice",
		&struct {
			Mac        string
			ServerName string
		}{Mac: d.Mac, ServerName: d.Mac})

	req, _ = http.NewRequest("POST", configuration.Config.Providers.Fanvil.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Fanvil.User, password)

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", " Falconieri/1")

	resp, err = fanvilHTTPClient.Do(req)

	if err != nil {
		return models.ProviderError{
			Message:      "connection_to_remote_provider_failed",
			WrappedError: err,
		}

	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("provider_remote_call_failed")
	}

	err = fanvilParseResponse(resp.Body)

	return err
}

func fanvilGetPassword() string {

	if fanvilPassword == "" {

		h := md5.New()
		io.WriteString(h, configuration.Config.Providers.Fanvil.Password)
		d1 := fmt.Sprintf("%x", h.Sum(nil))

		h2 := md5.New()
		io.WriteString(h2, d1)
		fanvilPassword = fmt.Sprintf("%x", h2.Sum(nil))

		return fanvilPassword

	} else {
		return fanvilPassword
	}
}

func fanvilParseResponse(body io.ReadCloser) error {

	var response_regexp = `(?s).*<boolean>(.*)</boolean>.*<string>(.*)</string>.*`

	respBytes, err := io.ReadAll(body)

	if err != nil {
		return errors.New("read_remote_response_failed")
	}

	re := regexp.MustCompile(response_regexp)

	if re.MatchString(string(respBytes)) {

		response := re.FindStringSubmatch(string(respBytes))

		var success bool

		success, _ = strconv.ParseBool(response[1])
		message := response[2]

		if !success {
			return utils.ParseProviderError(message)
		}

	} else {
		return errors.New("unknown_response_from_provider")
	}

	return nil
}
