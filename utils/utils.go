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

package utils

import (
	"errors"
	"regexp"
	"strings"

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
		return macAddress, errors.New("malformed_mac_address")
	}
}

func ParseProviderError(message string) error {

	switch {
	case message == "Error:malformed_url", //snom
		message == "Error: url_format_error",                                                            //fanvil
		strings.HasPrefix(message, "url_invalid:"),                                                      //gigaset
		message == "Error:The url can only begin with 'http://' or 'https://' or 'ftp://' or 'tftp://'": //yealink

		return errors.New("malformed_url")

	case message == "Error:malformed_mac", //snom
		message == "Error: mac_format_error",         //fanvil
		strings.HasPrefix(message, "mac_not_exist:"): //gigaset

		return errors.New("not_valid_mac_address")

	case message == "Error:owned_by_other_user", //snom
		message == "Error: device_had_existed",                         //fanvil
		strings.HasPrefix(message, "mac_already_in_use:"),              //gigaset
		regexp.MustCompile(`^Error:[a-z0-9]{10}`).MatchString(message): //yealink

		return errors.New("device_owned_by_other_user")

	default:
		return models.ProviderError{
			Message:      "unknown_provider_error: " + message,
			WrappedError: errors.New(message),
		}
	}
}
