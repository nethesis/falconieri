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

package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type ProviderConf struct {
	Password string `json:"password"`
	User     string `json:"user"`
	RpcUrl   string `json:"rpc_url"`
}

type GigasetConf struct {
	ProviderConf
	DisableCrc bool `json:"disable_crc"`
}

type Configuration struct {
	Providers struct {
		Snom    ProviderConf `json:"snom"`
		Gigaset GigasetConf  `json:"gigaset"`
	} `json: "providers"`
}

var Config = Configuration{}

func Init(ConfigFilePtr *string) {

	// read configuration
	if _, err := os.Stat(*ConfigFilePtr); err == nil {
		file, _ := os.Open(*ConfigFilePtr)
		decoder := json.NewDecoder(file)
		// check errors or parse JSON
		err := decoder.Decode(&Config)
		if err != nil {
			fmt.Println("Configuration parsing error:", err)
		}
	}

	if os.Getenv("SNOM_USER") != "" {
		Config.Providers.Snom.User = os.Getenv("SNOM_USER")
	}

	if os.Getenv("SNOM_PASSWORD") != "" {
		Config.Providers.Snom.Password = os.Getenv("SNOM_PASSWORD")
	}

	if os.Getenv("SNOM_RPC_URL") != "" {
		Config.Providers.Snom.RpcUrl = os.Getenv("SNOM_RPC_URL")
	}

	if os.Getenv("GIGASET_USER") != "" {
		Config.Providers.Gigaset.User = os.Getenv("GIGASET_USER")
	}

	if os.Getenv("GIGASET_PASSWORD") != "" {
		Config.Providers.Gigaset.Password = os.Getenv("GIGASET_PASSWORD")
	}

	if os.Getenv("GIGASET_RPC_URL") != "" {
		Config.Providers.Gigaset.RpcUrl = os.Getenv("GIGASET_RPC_URL")
	}

	if os.Getenv("GIGASET_DISABLE_CRC") != "" {

		disableCrc, err := strconv.ParseBool(os.Getenv("GIGASET_DISABLE_CRC"))
		if err != nil {
			Config.Providers.Gigaset.DisableCrc = disableCrc
		} else {
			Config.Providers.Gigaset.DisableCrc = false
		}
	}
}
