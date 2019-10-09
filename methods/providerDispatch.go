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

package methods

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nethesis/falconieri/models"
	"github.com/nethesis/falconieri/providers"
	"github.com/nethesis/falconieri/utils"
)

func ProviderDispatch(c *gin.Context) {

	switch provider := c.Param("provider"); provider {

	case "snom":
		mac, url, err := parseParams(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		//registerDevice(c, "snom", mac.A0+mac.A1+mac.A2+mac.A3+mac.A4+mac.A5, url)

		device := providers.SnomDevice{
			Mac: mac.A0 + mac.A1 + mac.A2 + mac.A3 + mac.A4 + mac.A5,
			Url: url,
		}

		err = device.Register()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

	default:
		c.JSON(http.StatusNotFound, gin.H{"message": "provider not supported"})
		return
	}

	c.Status(http.StatusOK)
}

func parseParams(c *gin.Context) (models.MacAddress, string, error) {

	var url models.Url

	mac, err := utils.GetMacAddress(c.Param("mac"))
	if err != nil {
		return mac, url.Url, errors.New("Invalid mac address")
	}

	if err := c.BindJSON(&url); err != nil {
		return mac, url.Url, errors.New("Missing url")
	}

	return mac, url.Url, nil
}
