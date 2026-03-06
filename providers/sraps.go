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
 * SRAPS (Secure Redirection and Provisioning Service) provider
 * Reuses the grape library with a different base URL
 */
package providers

import (
	"errors"
	"net"
	"net/url"
	"sync"

	"github.com/nethesis/falconieri/configuration"
	"github.com/nethesis/falconieri/libs/grape"
	"github.com/nethesis/falconieri/models"
)

var (
	srapsClient     *grape.Client
	srapsClientOnce sync.Once
)

// getSrapsClient returns a singleton SRAPS client instance.
// SRAPS uses the same protocol as Grape but with a different base URL
// and uses "setting_server" instead of "ProvisioningServer" for the setting name.
func getSrapsClient() *grape.Client {
	srapsClientOnce.Do(func() {
		srapsClient = grape.NewClient(
			configuration.Config.Providers.Sraps.BaseURL,
			configuration.Config.Providers.Sraps.ClientID,
			configuration.Config.Providers.Sraps.ClientSecret,
		)
		srapsClient.ProvisioningSettingName = "setting_server"
	})
	return srapsClient
}

type SrapsDevice struct {
	Mac string
	Url string
}

func (d SrapsDevice) Register() error {
	client := getSrapsClient()

	err := client.RegisterDevice(d.Mac, d.Url)
	if err != nil {
		var apiErr grape.APIError
		if errors.As(err, &apiErr) {
			return models.ProviderError{
				Message:      "provider_remote_call_failed",
				WrappedError: apiErr,
			}
		}

		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return models.ProviderError{
				Message:      "connection_to_remote_provider_failed",
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

		return err
	}

	return nil
}
