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

package methods

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/nethesis/falconieri/configuration"
	"github.com/nethesis/falconieri/models"
	"github.com/nethesis/falconieri/providers"
	"github.com/nethesis/falconieri/utils"
)

func ProviderDispatch(c *gin.Context) {

	var device interface {
		Register() error
	}

	switch provider := c.Param("provider"); {

	case (provider == "snom") && !configuration.Config.Providers.Snom.Disable:
		mac, url, err := parseParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		device = providers.SnomDevice{
			Mac: mac.A0 + mac.A1 + mac.A2 + mac.A3 + mac.A4 + mac.A5,
			Url: url,
		}

	case (provider == "gigaset") && !configuration.Config.Providers.Gigaset.Disable:

		mac, url, err := parseParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var mac_address string

		if configuration.Config.Providers.Gigaset.DisableCrc {

			mac_address = mac.A0 + mac.A1 + mac.A2 + mac.A3 + mac.A4 + mac.A5

		} else {

			if regexp.MustCompile(`^[A-Z0-9]{4}$`).MatchString(c.Query("crc")) {

				mac_address = mac.A0 + mac.A1 + mac.A2 + mac.A3 + mac.A4 + mac.A5 + "-" + c.Query("crc")

			} else if c.Query("crc") == "" {

				c.JSON(http.StatusBadRequest, gin.H{"error": "missing_mac-id_crc"})
				return

			} else {

				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_mac-id_crc_format"})
				return
			}

		}

		device = providers.GigasetDevice{
			Mac:      mac_address,
			Url:      url,
			Provider: "Falconieri",
		}

	case (provider == "fanvil") && !configuration.Config.Providers.Fanvil.Disable:
		mac, url, err := parseParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		device = providers.FanvilDevice{
			Mac: mac.A0 + mac.A1 + mac.A2 + mac.A3 + mac.A4 + mac.A5,
			Url: url,
		}

	case (provider == "yealink") && !configuration.Config.Providers.Yealink.Disable:
		mac, url, err := parseParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		device = providers.YealinkDevice{
			Mac:        mac.A0 + "-" + mac.A1 + "-" + mac.A2 + "-" + mac.A3 + "-" + mac.A4 + "-" + mac.A5,
			Url:        url,
			ServerName: "Falconieri",
			Override:   "1",
		}

	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "provider_not_supported"})
		return
	}

	err := device.Register()
	if err != nil {
		if errors.Unwrap(err) != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(),
				"message": errors.Unwrap(err).Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusOK)
}

func parseParams(c *gin.Context) (models.MacAddress, string, error) {

	var url models.Url

	mac, err := utils.GetMacAddress(c.Param("mac"))
	if err != nil {
		return mac, url.Url, errors.New("missing_mac_address")
	}

	if err := c.BindJSON(&url); err != nil {
		return mac, url.Url, errors.New("missing_url")
	}

	return mac, url.Url, nil
}
