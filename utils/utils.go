/*
 * Copyright (C) 2019 Nethesis S.r.l.
 * http://www.nethesis.it - info@nethesis.it
 *
 * This file is part of Falconieri project.
 *
 * Icaro is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License,
 * or any later version.
 *
 * Icaro is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Icaro.  If not, see LICENSE.
 *
 * author: Matteo Valentini <matteo.valentini@nethesis.it>
 */

package utils

import (
	"errors"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/nethesis/falconieri/models"
)

func GetMacAddress(mac string) (models.MacAddress, error) {

	var macRegexp = regexp.MustCompile(`^([A-Z0-9]{2})-([A-Z0-9]{2})-([A-Z0-9]{2})-([A-Z0-9]{2})-([A-Z0-9]{2})-([A-Z0-9]{2})$`)

	var macAddress models.MacAddress

	if macRegexp.MatchString(mac) {

		macAddr := macRegexp.FindStringSubmatch(mac)

		macAddress.A0 = macAddr[1]
		macAddress.A1 = macAddr[2]
		macAddress.A2 = macAddr[3]
		macAddress.A3 = macAddr[4]
		macAddress.A4 = macAddr[5]
		macAddress.A5 = macAddr[6]

		return macAddress, nil

	} else {
		return macAddress, errors.New("Malformed mac address.")
	}
}

func ParseRegisterResponse(provider string, resp io.Reader) error {

	var regexp1 = `<params><param><value><array><data><value><boolean>(.*)</boolean></value><value><string>(.*)</string></value></data></array></value></param></params>`

	respBytes, err := ioutil.ReadAll(resp)

	if err != nil {
		return errors.New("Error on reading " + provider + " response")
	}

	respString := string(respBytes)

	switch provider {
	case "snom":
		re := regexp.MustCompile(regexp1)
		if re.MatchString(respString) {

			response := re.FindStringSubmatch(respString)

			var success bool

			success, _ = strconv.ParseBool(response[1])
			message := response[2]

			if !success {
				return errors.New("Error to register Device on " + provider + " provider: " + message)
			}

		} else {
			return errors.New("Unknow response from" + provider + " provider")
		}
	}

	return nil
}
