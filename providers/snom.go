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

package providers

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/divan/gorilla-xmlrpc/xml"

	"github.com/nethesis/falconieri/configuration"
	"github.com/nethesis/falconieri/utils"
)

type SnomDevice struct {
	Mac string
	Url string
}

func (s SnomDevice) Register() error {

	buf, _ := xml.EncodeClientRequest("redirect.registerPhone", &s)

	req, _ := http.NewRequest("POST", configuration.Config.Providers.Snom.RpcUrl,
		bytes.NewReader(buf))

	req.SetBasicAuth(configuration.Config.Providers.Snom.User,
		configuration.Config.Providers.Snom.Password)

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

	if err = utils.ParseRegisterResponse("snom", resp.Body); err != nil {
		return err
	}

	return nil

}
