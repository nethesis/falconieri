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

package main

import (
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nethesis/falconieri/configuration"
	"github.com/nethesis/falconieri/methods"
	"github.com/nethesis/falconieri/middleware"
)

func DefineAPI(router *gin.Engine) {

	router.GET("/health/check", methods.HealthCheck)

	providers := router.Group("/providers")

	providers.Use(middleware.FalconieriAuth)
	{
		providers.PUT("/:provider/:mac", methods.ProviderDispatch)
	}

	// handle missing endpoint
	router.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

}

func main() {

	// read and init configuration
	ConfigFilePtr := flag.String("c", "/opt/falconieri/conf.json", "Path to configuration file")
	flag.Parse()
	configuration.Init(ConfigFilePtr)
	// init routers
	router := gin.Default()

	DefineAPI(router)

	router.Run()
}
