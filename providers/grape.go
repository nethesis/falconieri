/*
 * Copyright (C) 2026 Nethesis S.r.l.
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

/*
 * Grape provisioning provider
 */
package providers

import (
	"errors"
	"net"
	"sync"

	"github.com/nethesis/falconieri/configuration"
	"github.com/nethesis/falconieri/libs/grape"
	"github.com/nethesis/falconieri/models"
)

var (
	grapeClient     *grape.Client
	grapeClientOnce sync.Once
)

// getGrapeClient returns a singleton Grape client instance.
// This allows API navigation caching to work effectively across multiple requests.
// The client is created using configuration loaded at startup.
func getGrapeClient() *grape.Client {
	grapeClientOnce.Do(func() {
		grapeClient = grape.NewClient(
			configuration.Config.Providers.Grape.BaseURL,
			configuration.Config.Providers.Grape.ClientID,
			configuration.Config.Providers.Grape.ClientSecret,
		)
	})
	return grapeClient
}

type GrapeDevice struct {
	Mac string
	Url string
}

func (d GrapeDevice) Register() error {
	client := getGrapeClient()

	err := client.RegisterDevice(d.Mac, d.Url)
	if err != nil {
		var apiErr grape.APIError
		if errors.As(err, &apiErr) {
			return models.ProviderError{
				Message:      "provider_remote_call_failed",
				WrappedError: err,
			}
		}

		var netErr net.Error
		if errors.As(err, &netErr) {
			return models.ProviderError{
				Message:      "connection_to_remote_provider_failed",
				WrappedError: err,
			}
		}

		return models.ProviderError{
			Message:      "provider_remote_call_failed",
			WrappedError: err,
		}
	}

	return nil
}
