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
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/divan/gorilla-xmlrpc/xml"

	"github.com/nethesis/falconieri/configuration"
)

type FanvilDevice struct {
	Mac string
	Url string
}

var fanvilPassword string

func (d FanvilDevice) Register() error {

	password := fanvilGetPassword()

	//Create Server
	buf, _ := xml.EncodeClientRequest("redirect.addServer",
		&struct {
			GroupName string
			GroupUrl  string
		}{GroupName: d.Url, GroupUrl: d.Url})

	req, _ := http.NewRequest("POST", configuration.Config.Providers.Fanvil.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Fanvil.User, password)

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", " Falconieri/1")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Remote call failed")
	}

	err = fanvilParseResponse(resp.Body)

	if (err != nil) && (err.Error() != "Error:server_had_existed") {
		return err
	}

	//Deregister the device
	buf, _ = xml.EncodeClientRequest("redirect.deRegisterDevice",
		&struct {
			Mac string
		}{Mac: d.Mac})

	req, _ = http.NewRequest("POST", configuration.Config.Providers.Fanvil.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Fanvil.User, password)

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", " Falconieri/1")

	resp, err = http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Remote call failed")
	}

	//Register the device
	buf, _ = xml.EncodeClientRequest("redirect.registerDevice",
		&struct {
			Mac        string
			ServerName string
		}{Mac: d.Mac, ServerName: d.Url})

	req, _ = http.NewRequest("POST", configuration.Config.Providers.Fanvil.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Fanvil.User, password)

	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("User-Agent", " Falconieri/1")

	resp, err = http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Remote call failed")
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

	respBytes, err := ioutil.ReadAll(body)

	if err != nil {
		return errors.New("Error on reading Fanvil response")
	}

	re := regexp.MustCompile(response_regexp)

	if re.MatchString(string(respBytes)) {

		response := re.FindStringSubmatch(string(respBytes))

		var success bool

		success, _ = strconv.ParseBool(response[1])
		message := response[2]

		if !success {
			return errors.New(message)
		}

	} else {
		return errors.New("Unknow response from Fanvil provider")
	}

	return nil
}
